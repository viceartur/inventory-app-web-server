package reports

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/leekchan/accounting"
)

type Report struct {
	DB                  *sql.DB
	TrxFilter           SearchQuery
	BlcFilter           SearchQuery
	WeeklyUsgFilter     SearchQuery
	TrxLogFilter        SearchQuery
	CustomerUsageFilter SearchQuery
}

var accLib accounting.Accounting = accounting.Accounting{Symbol: "$", Precision: 4}

func (t *Report) GetTransactions() ([]TransactionRep, error) {
	rows, err := t.DB.Query(`
			WITH
				-- Compute starting quantity before the 'from' date for each stock
				starting_quantities AS (
					SELECT
						m.stock_id,
						COALESCE(SUM(tl.quantity_change), 0) AS starting_qty
					FROM
						transactions_log tl
						LEFT JOIN prices p ON p.price_id = tl.price_id
						LEFT JOIN materials m ON m.material_id = p.material_id
						LEFT JOIN customer_programs c ON m.program_id = c.program_id
						WHERE
							($1 = 0 OR m.program_id = $1) AND
							($2 = '' OR m.material_type = $2::MATERIAL_TYPE) AND
							($5 = '' OR m.owner = $5::OWNER) AND
							($4 = '' OR tl.updated_at < $4::timestamp)
					GROUP BY
						m.stock_id
				),
				-- Select and enrich all relevant transactions
				ordered_transactions AS (
					SELECT
						m.stock_id,
						m.material_type,
						tl.quantity_change AS quantity,
						p.cost AS unit_cost,
						(tl.quantity_change * p.cost) AS cost,
						tl.updated_at,
						COALESCE(tl.serial_number_range, '') AS serial_number_range,
						tl.transaction_id,
						m.material_id
					FROM
						transactions_log tl
						LEFT JOIN prices p ON p.price_id = tl.price_id
						LEFT JOIN materials m ON m.material_id = p.material_id
						LEFT JOIN customer_programs c ON m.program_id = c.program_id
						WHERE
						    ($1 = 0 OR m.program_id = $1) AND
						    ($2 = '' OR m.material_type = $2::MATERIAL_TYPE) AND
						    ($5 = '' OR m.owner = $5::OWNER)
				)
				-- Final result: transactions with cumulative quantities
			SELECT
				ot.stock_id,
				ot.material_type,
				ot.quantity,
				ot.unit_cost,
				ot.cost,
				ot.updated_at,
				ot.serial_number_range,
				-- Calculate cumulative quantity (balance) per stock_id
				COALESCE(sq.starting_qty, 0) - SUM(ot.quantity) OVER (
					PARTITION BY
						ot.stock_id
					ORDER BY
						ot.updated_at,
						ot.transaction_id ROWS BETWEEN CURRENT ROW
						AND UNBOUNDED FOLLOWING
				) + ot.quantity AS cumulative_quantity
			FROM
				ordered_transactions ot
				LEFT JOIN starting_quantities sq ON sq.stock_id = ot.stock_id
				WHERE
				    ($3 = '' OR ot.updated_at >= $3::timestamp) AND
				    ($4 = '' OR ot.updated_at <= $4::timestamp)
			ORDER BY
				ot.updated_at,
				ot.transaction_id ASC;
	`,
		t.TrxFilter.ProgramID,
		t.TrxFilter.MaterialType,
		t.TrxFilter.DateFrom,
		t.TrxFilter.DateTo,
		t.TrxFilter.Owner,
	)
	if err != nil {
		return []TransactionRep{}, err
	}

	trxList := []TransactionRep{}

	for rows.Next() {
		trx := Transaction{}

		err := rows.Scan(
			&trx.StockID,
			&trx.MaterialType,
			&trx.Qty,
			&trx.UnitCost,
			&trx.Cost,
			&trx.UpdatedAt,
			&trx.SerialNumberRange,
			&trx.CumulativeQty,
		)
		if err != nil {
			return []TransactionRep{}, err
		}

		year, month, day := trx.UpdatedAt.Date()
		strDate := strconv.Itoa(int(month)) + "/" +
			strconv.Itoa(day) + "/" +
			strconv.Itoa(year)
		unitCost := accLib.FormatMoney(trx.UnitCost)
		cost := accLib.FormatMoney(trx.Cost)

		trxList = append(trxList, TransactionRep{
			StockID:           trx.StockID,
			MaterialType:      trx.MaterialType,
			Qty:               strconv.Itoa(trx.Qty),
			UnitCost:          unitCost,
			Cost:              cost,
			Date:              strDate,
			SerialNumberRange: trx.SerialNumberRange,
			CumulativeQty:     strconv.Itoa(trx.CumulativeQty),
		})
	}

	return trxList, nil
}

func (b *Report) GetBalance() ([]BalanceRep, error) {
	rows, err := b.DB.Query(`
		SELECT m.stock_id,
			m.description,
			m.material_type,
			SUM(tl.quantity_change) AS "quantity",
			SUM(tl.quantity_change * p.cost) AS "total_value"
		FROM transactions_log tl
		LEFT JOIN prices p ON p.price_id = tl.price_id
		LEFT JOIN materials m ON m.material_id = p.material_id
		WHERE
			($1 = 0 OR m.program_id = $1) AND
			($2 = '' OR m.material_type::TEXT = $2) AND
			($3 = '' OR tl.updated_at::TEXT <= $3) AND
			($4 = '' OR m.owner::TEXT = $4) AND
			m.location_id IS NOT NULL
		GROUP BY m.stock_id, m.description, m.material_type
		ORDER BY m.material_type ASC, m.description ASC;
`,
		b.BlcFilter.ProgramID, b.BlcFilter.MaterialType, b.BlcFilter.DateAsOf, b.BlcFilter.Owner,
	)
	if err != nil {
		return []BalanceRep{}, err
	}

	blcList := []BalanceRep{}

	for rows.Next() {
		balance := Transaction{}

		err := rows.Scan(
			&balance.StockID,
			&balance.Description,
			&balance.MaterialType,
			&balance.Qty,
			&balance.TotalValue,
		)
		if err != nil {
			return []BalanceRep{}, err
		}

		totalValue := accLib.FormatMoney(balance.TotalValue)
		blcList = append(blcList, BalanceRep{
			StockID:      balance.StockID,
			Description:  balance.Description,
			MaterialType: balance.MaterialType,
			Qty:          strconv.Itoa(balance.Qty),
			TotalValue:   totalValue,
		})
	}

	return blcList, err

}

// Generates a weekly usage report based on the provided filters in the WeeklyUsageReport struct.
// It calculates the quantity of materials on a reference date, their average weekly usage over the past 6 weeks,
// and the estimated number of weeks of stock remaining.
func (w *Report) GetWeeklyUsage() ([]WeeklyUsageRep, error) {
	// Use today's date if dateAsOf is not provided
	dateAsOf := w.WeeklyUsgFilter.DateAsOf
	if dateAsOf == "" {
		dateAsOf = time.Now().Format("2006-01-02")
	}
	rows, err := w.DB.Query(`
		WITH
			reference_date AS (
				SELECT
					$1::DATE AS ref_date
			),
			-- Total future usage grouped by stock_id
			future_usage AS (
				SELECT
					m.stock_id,
					SUM(tl.quantity_change) AS future_quantity_change
				FROM
					transactions_log tl
					JOIN prices p ON p.price_id = tl.price_id
					JOIN materials m ON m.material_id = p.material_id
					JOIN reference_date r ON true
				WHERE
					tl.updated_at > r.ref_date
				GROUP BY
					m.stock_id
			),
			-- Quantity on reference date grouped by stock_id
			material_quantities AS (
				SELECT
					m.stock_id,
					m.material_type,
					m.program_id,
					SUM(m.quantity) - COALESCE(fu.future_quantity_change, 0) AS quantity_on_ref_date
				FROM
					materials m
					LEFT JOIN future_usage fu ON fu.stock_id = m.stock_id
				GROUP BY
					m.stock_id,
					m.material_type,
					m.program_id,
					fu.future_quantity_change
			),
			-- Filtered transactions (past 6 weeks)
			filtered_transactions AS (
				SELECT
					tl.*,
					m.stock_id
				FROM
					transactions_log tl
					JOIN prices p ON p.price_id = tl.price_id
					JOIN materials m ON m.material_id = p.material_id
					JOIN reference_date r ON true
				WHERE
					tl.updated_at BETWEEN r.ref_date - INTERVAL '6 weeks' AND r.ref_date
			),
			-- Weekly usage per stock
			weekly_usage AS (
				SELECT
					stock_id,
					DATE_TRUNC ('week', updated_at) AS week_start,
					SUM(quantity_change) AS total_used
				FROM
					filtered_transactions
				WHERE
					quantity_change < 0
					AND notes NOT ILIKE 'moved from a location'
				GROUP BY
					stock_id,
					week_start
			),
			-- Average usage per stock
			average_usage AS (
				SELECT
					stock_id,
					ABS(AVG(total_used)) AS avg_weekly_usage
				FROM
					weekly_usage
				GROUP BY
					stock_id
			)
			-- Final output
		SELECT
			c.program_name,
			mq.stock_id,
			mq.material_type,
			mq.quantity_on_ref_date,
			ROUND(au.avg_weekly_usage) AS avg_weekly_usage,
			CASE
				WHEN au.avg_weekly_usage = 0 THEN NULL
				ELSE ROUND(mq.quantity_on_ref_date / au.avg_weekly_usage, 2)
			END AS weeks_of_stock_remaining
		FROM
			material_quantities mq
			LEFT JOIN average_usage au ON mq.stock_id = au.stock_id
			LEFT JOIN customer_programs c ON c.program_id = mq.program_id
		WHERE
			mq.quantity_on_ref_date > 0
			AND au.avg_weekly_usage > 0
			AND ($2 = 0 OR c.program_id = $2)
			AND ($3 = '' OR mq.stock_id = $3)
			AND ($4 = '' OR mq.material_type::TEXT = $4)
		ORDER BY
			c.program_name,
			mq.material_type,
			mq.stock_id;

		`,
		dateAsOf,
		w.WeeklyUsgFilter.ProgramID,
		w.WeeklyUsgFilter.StockId,
		w.WeeklyUsgFilter.MaterialType,
	)
	if err != nil {
		return []WeeklyUsageRep{}, err
	}

	weeklyUsgList := []WeeklyUsageRep{}

	for rows.Next() {
		usageTransaction := WeeklyUsageRep{}

		err := rows.Scan(
			&usageTransaction.ProgramName,
			&usageTransaction.StockID,
			&usageTransaction.MaterialType,
			&usageTransaction.QtyOnRefDate,
			&usageTransaction.AvgWeeklyUsg,
			&usageTransaction.WeeksRemaining,
		)
		if err != nil {
			return []WeeklyUsageRep{}, err
		}

		weeklyUsgList = append(weeklyUsgList, WeeklyUsageRep{
			ProgramName:    usageTransaction.ProgramName,
			StockID:        usageTransaction.StockID,
			MaterialType:   usageTransaction.MaterialType,
			QtyOnRefDate:   usageTransaction.QtyOnRefDate,
			AvgWeeklyUsg:   usageTransaction.AvgWeeklyUsg,
			WeeksRemaining: usageTransaction.WeeksRemaining,
		})
	}

	return weeklyUsgList, err
}

func (tl *Report) GetTransactionsLog() ([]TransactionRep, error) {
	rows, err := tl.DB.Query(`
		SELECT
			m.stock_id,
			m.material_type,
			COALESCE(l.name, 'None') AS location_name,
			COALESCE(tl.serial_number_range, ''),
			tl.quantity_change AS "quantity",
			tl.job_ticket,
			tl.updated_at,
			COALESCE(mus.description, '') AS reason_description
		FROM transactions_log tl
		LEFT JOIN prices p ON p.price_id = tl.price_id
		LEFT JOIN materials m ON m.material_id = p.material_id
		LEFT JOIN locations l ON l.location_id = m.location_id
		LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
		LEFT JOIN customer_programs c ON m.program_id = c.program_id
		LEFT JOIN material_usage_reasons mus ON mus.reason_id = tl.reason_id
		WHERE
			($1 = 0 OR w.warehouse_id = $1) AND
			($2 = 0 OR m.program_id = $2) AND
			($3 = '' OR m.material_type::TEXT = $3) AND
			($4 = '' OR tl.updated_at::TEXT >= $4) AND
			($5 = '' OR tl.updated_at::TEXT <= $5) AND
			($6 = '' OR m.owner::TEXT = $6)
		ORDER BY tl.transaction_id ASC;`,
		tl.TrxLogFilter.WarehouseID,
		tl.TrxLogFilter.ProgramID,
		tl.TrxLogFilter.MaterialType,
		tl.TrxLogFilter.DateFrom,
		tl.TrxLogFilter.DateTo,
		tl.TrxLogFilter.Owner,
	)
	if err != nil {
		return []TransactionRep{}, err
	}

	trxList := []TransactionRep{}

	for rows.Next() {
		trx := Transaction{}

		err := rows.Scan(
			&trx.StockID,
			&trx.MaterialType,
			&trx.LocationName,
			&trx.SerialNumberRange,
			&trx.Qty,
			&trx.JobTicket,
			&trx.UpdatedAt,
			&trx.ReasonDescription,
		)
		if err != nil {
			return []TransactionRep{}, err
		}

		year, month, day := trx.UpdatedAt.Date()
		strDate := strconv.Itoa(int(month)) + "/" +
			strconv.Itoa(day) + "/" +
			strconv.Itoa(year)

		trxList = append(trxList, TransactionRep{
			StockID:           trx.StockID,
			MaterialType:      trx.MaterialType,
			LocationName:      trx.LocationName,
			SerialNumberRange: trx.SerialNumberRange,
			Qty:               strconv.Itoa(trx.Qty),
			JobTicket:         trx.JobTicket,
			Date:              strDate,
			ReasonDescription: trx.ReasonDescription,
		})
	}

	return trxList, nil
}

func (vr *Report) GetVault() ([]VaultRep, error) {
	rows, err := vr.DB.Query(`
		WITH
			inner_vault AS (
				SELECT
					l.name AS location_name,
					m.stock_id,
					SUM(m.quantity) AS quantity
				FROM
					materials m
					LEFT JOIN locations l ON l.location_id = m.location_id
					LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
				WHERE
					w.name = 'Inner Vault'
				GROUP BY
					l.name,
					m.stock_id
			),
			outer_vault AS (
				SELECT
					l.name AS location_name,
					m.stock_id,
					SUM(m.quantity) AS quantity
				FROM
					materials m
					LEFT JOIN locations l ON l.location_id = m.location_id
					LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
				WHERE
					w.name = 'Outer Vault'
				GROUP BY
					l.name,
					m.stock_id
			)
		SELECT
			COALESCE(iv.location_name, '-') AS inner_location,
			COALESCE(ov.location_name, '-') AS outer_location,
			m.stock_id,
			COALESCE(iv.quantity, 0) AS inner_vault_quantity,
			COALESCE(ov.quantity, 0) AS outer_vault_quantity,
			SUM(m.quantity) AS total_quantity
		FROM
			materials m
			LEFT JOIN locations l ON l.location_id = m.location_id
			LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
			LEFT JOIN inner_vault iv ON iv.stock_id = m.stock_id
			LEFT JOIN outer_vault ov ON ov.stock_id = m.stock_id
		WHERE
			w.name IN ('Inner Vault', 'Outer Vault')
		GROUP BY
			iv.location_name,
			ov.location_name,
			m.stock_id,
			iv.quantity,
			ov.quantity
		ORDER BY
			iv.location_name,
			ov.location_name,
			m.stock_id;
	`)
	if err != nil {
		return []VaultRep{}, err
	}

	vaultReport := []VaultRep{}

	for rows.Next() {
		vault := VaultRep{}

		err := rows.Scan(
			&vault.InnerLocation,
			&vault.OuterLocation,
			&vault.StockID,
			&vault.InnerVaultQty,
			&vault.OuterVaultQty,
			&vault.TotalQty,
		)
		if err != nil {
			return []VaultRep{}, err
		}

		vaultReport = append(vaultReport, VaultRep{
			InnerLocation: vault.InnerLocation,
			OuterLocation: vault.OuterLocation,
			StockID:       vault.StockID,
			InnerVaultQty: vault.InnerVaultQty,
			OuterVaultQty: vault.OuterVaultQty,
			TotalQty:      vault.TotalQty,
		})
	}

	return vaultReport, nil
}

// GetReportList generates a customer usage report based on the provided filter criteria.
// It calculates starting quantity, received, used, spoiled, ending quantity, and average weekly usage
// for each material, grouped by program and material type, within the specified date range.
// If no date range is provided, it defaults to the last week (Sunday to Saturday).
// If a customer ID is specified, the report is filtered for that customer.
func (c *Report) GetCustomerUsage() ([]CustomerUsageRep, error) {
	customerId := c.CustomerUsageFilter.CustomerID
	dateFrom := c.CustomerUsageFilter.DateFrom
	dateTo := c.CustomerUsageFilter.DateTo

	if dateFrom == "" || dateTo == "" {
		return nil, fmt.Errorf("No period specified.")
	}

	baseQuery := `
		WITH
			constants AS (
				SELECT
					$1::DATE AS ref_start,
					$2::DATE AS ref_end
			),
			starting_qty AS (
				SELECT
					m.stock_id,
					COALESCE(SUM(tl.quantity_change), 0) AS qty_start
				FROM
					materials m
					LEFT JOIN prices p ON p.material_id = m.material_id
					LEFT JOIN transactions_log tl ON tl.price_id = p.price_id
					LEFT JOIN customer_programs cp ON cp.program_id = m.program_id
					%s
					JOIN constants const ON TRUE
				WHERE
					tl.updated_at < const.ref_start
					%s
				GROUP BY
					m.stock_id
			),
			received AS (
				SELECT
					m.stock_id,
					COALESCE(SUM(tl.quantity_change), 0) AS qty_received
				FROM
					materials m
					LEFT JOIN prices p ON p.material_id = m.material_id
					LEFT JOIN transactions_log tl ON tl.price_id = p.price_id
					LEFT JOIN customer_programs cp ON cp.program_id = m.program_id
					%s
					JOIN constants const ON TRUE
				WHERE
					tl.updated_at BETWEEN const.ref_start AND const.ref_end
					AND tl.quantity_change > 0
					AND tl.notes NOT ILIKE 'moved from%%'
					%s
				GROUP BY
					m.stock_id
			),
			used AS (
				SELECT
					m.stock_id,
					ABS(SUM(tl.quantity_change)) AS qty_used
				FROM
					materials m
					LEFT JOIN prices p ON p.material_id = m.material_id
					LEFT JOIN transactions_log tl ON tl.price_id = p.price_id
					LEFT JOIN customer_programs cp ON cp.program_id = m.program_id
					%s
					JOIN constants const ON TRUE
				WHERE
					tl.updated_at BETWEEN const.ref_start AND const.ref_end
					AND tl.quantity_change < 0
					AND tl.reason_id IS NULL
					AND tl.notes NOT ILIKE 'moved to%%'
					%s
				GROUP BY
					m.stock_id
			),
			spoiled AS (
				SELECT
					m.stock_id,
					ABS(SUM(tl.quantity_change)) AS qty_spoiled
				FROM
					materials m
					LEFT JOIN prices p ON p.material_id = m.material_id
					LEFT JOIN transactions_log tl ON tl.price_id = p.price_id
					LEFT JOIN customer_programs cp ON cp.program_id = m.program_id
					%s
					JOIN constants const ON TRUE
				WHERE
					tl.updated_at BETWEEN const.ref_start AND const.ref_end
					AND tl.quantity_change < 0
					AND tl.reason_id IS NOT NULL
					%s
				GROUP BY
					m.stock_id
			),
			ending_qty AS (
				SELECT
					m.stock_id,
					COALESCE(SUM(tl.quantity_change), 0) AS qty_end
				FROM
					materials m
					LEFT JOIN prices p ON p.material_id = m.material_id
					LEFT JOIN transactions_log tl ON tl.price_id = p.price_id
					LEFT JOIN customer_programs cp ON cp.program_id = m.program_id
					%s
					JOIN constants const ON TRUE
				WHERE
					tl.updated_at <= const.ref_end
					%s
				GROUP BY
					m.stock_id
			),
			filtered_transactions AS (
				SELECT
					tl.*,
					m.stock_id
				FROM
					materials m
					LEFT JOIN prices p ON p.material_id = m.material_id
					LEFT JOIN transactions_log tl ON tl.price_id = p.price_id
					LEFT JOIN customer_programs cp ON cp.program_id = m.program_id
					%s
					JOIN constants const ON TRUE
				WHERE
					tl.updated_at BETWEEN (const.ref_end - INTERVAL '6 weeks') AND const.ref_end
					%s
			),
			weekly_usage AS (
				SELECT
					stock_id,
					DATE_TRUNC ('week', updated_at) AS week_start,
					SUM(quantity_change) AS total_used
				FROM
					filtered_transactions
				WHERE
					quantity_change < 0
					AND notes NOT ILIKE 'moved from%%'
				GROUP BY
					stock_id,
					week_start
			),
			average_usage AS (
				SELECT
					stock_id,
					ABS(AVG(total_used)) AS avg_weekly_usage
				FROM
					weekly_usage
				GROUP BY
					stock_id
			)
		SELECT
			c.customer_id,
			cp.program_name,
			m.material_type,
			m.stock_id,
			COALESCE(sq.qty_start, 0) AS qty_start,
			COALESCE(r.qty_received, 0) AS qty_received,
			COALESCE(u.qty_used, 0) AS qty_used,
			COALESCE(s.qty_spoiled, 0) AS qty_spoiled,
			COALESCE(eq.qty_end, 0) AS qty_end,
			CASE
				WHEN m.material_type NOT IN ('CARDS (PVC)', 'CARDS (METAL)', 'CHIPS') THEN 0
				ELSE COALESCE(ROUND(au.avg_weekly_usage), 0)
			END AS six_week_avg_to_ref_end,
			CASE
				WHEN m.material_type NOT IN ('CARDS (PVC)', 'CARDS (METAL)', 'CHIPS') THEN 0
				WHEN au.avg_weekly_usage = 0 THEN NULL
				ELSE ROUND(COALESCE(eq.qty_end, 0) / au.avg_weekly_usage)
			END AS weeks_remaining
		FROM
			materials m
			LEFT JOIN customer_programs cp ON cp.program_id = m.program_id
			LEFT JOIN customers c ON c.customer_id = cp.customer_id
			JOIN constants const ON TRUE
			LEFT JOIN starting_qty sq ON sq.stock_id = m.stock_id
			LEFT JOIN received r ON r.stock_id = m.stock_id
			LEFT JOIN used u ON u.stock_id = m.stock_id
			LEFT JOIN spoiled s ON s.stock_id = m.stock_id
			LEFT JOIN ending_qty eq ON eq.stock_id = m.stock_id
			LEFT JOIN average_usage au ON au.stock_id = m.stock_id
		WHERE c.is_connected_to_reports = true
		%s
		GROUP BY
			c.customer_id,
			cp.program_name,
			m.stock_id,
			m.material_type,
			sq.qty_start,
			r.qty_received,
			u.qty_used,
			s.qty_spoiled,
			eq.qty_end,
			au.avg_weekly_usage
		ORDER BY
			c.customer_name,
			cp.program_name,
			m.material_type,
			m.stock_id;
	`

	var joinCustomer, whereCustomer string
	if customerId > 0 {
		joinCustomer = "LEFT JOIN customers c ON c.customer_id = cp.customer_id"
		whereCustomer = "AND c.customer_id = $3"
	}

	// Fill all join/where placeholders for each CTE and main query
	query := fmt.Sprintf(
		baseQuery,
		joinCustomer, whereCustomer, // starting_qty
		joinCustomer, whereCustomer, // received
		joinCustomer, whereCustomer, // used
		joinCustomer, whereCustomer, // spoiled
		joinCustomer, whereCustomer, // ending_qty
		joinCustomer, whereCustomer, // filtered_transactions
		whereCustomer, // main WHERE
	)

	var rows *sql.Rows
	var err error
	if customerId > 0 {
		rows, err = c.DB.Query(query, dateFrom, dateTo, customerId)
	} else {
		rows, err = c.DB.Query(query, dateFrom, dateTo)
	}

	if err != nil {
		return []CustomerUsageRep{}, err
	}

	reportList := []CustomerUsageRep{}

	for rows.Next() {
		var reportRow CustomerUsageRep
		var customerID sql.NullInt64
		var weeksRemaining sql.NullInt64

		err := rows.Scan(
			&customerID,
			&reportRow.ProgramName,
			&reportRow.MaterialType,
			&reportRow.StockID,
			&reportRow.QtyStart,
			&reportRow.QtyReceived,
			&reportRow.QtyUsed,
			&reportRow.QtySpoiled,
			&reportRow.QtyEnd,
			&reportRow.WeekAvg,
			&weeksRemaining,
		)
		if err != nil {
			return []CustomerUsageRep{}, err
		}

		// Check for NULL customer id
		if customerID.Valid {
			reportRow.CustomerID = int(customerID.Int64)
		} else {
			reportRow.CustomerID = 0
		}

		// Check for NULL weeks remaining
		if weeksRemaining.Valid {
			reportRow.WeeksRemaining = int(weeksRemaining.Int64)
		} else {
			reportRow.WeeksRemaining = 0
		}

		reportList = append(reportList, reportRow)
	}

	return reportList, nil
}
