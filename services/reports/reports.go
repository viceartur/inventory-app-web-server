package reports

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/leekchan/accounting"
)

type Transaction struct {
	StockID           string    `field:"stock_id"`
	Description       string    `field:"description"`
	LocationName      string    `field:"location_name"`
	MaterialType      string    `field:"material_type"`
	Qty               int       `field:"quantity"`
	UnitCost          float64   `field:"unit_cost"`
	Cost              float64   `field:"cost"`
	UpdatedAt         time.Time `field:"updated_at"`
	TotalValue        float64   `field:"total_value"`
	SerialNumberRange string    `field:"serial_number_range"`
	JobTicket         string    `field:"job_ticket"`
	ReasonDescription string    `field:"reason_description"`
	CumulativeQty     int       `field:"cumulative_quantity"`
}

type SearchQuery struct {
	WarehouseID  int
	CustomerID   int
	StockId      string
	Owner        string
	MaterialType string
	DateFrom     string
	DateTo       string
	DateAsOf     string
}

type Report struct {
	DB *sql.DB
}

type TransactionReport struct {
	Report
	TrxFilter SearchQuery
}

type BalanceReport struct {
	Report
	BlcFilter SearchQuery
}

type WeeklyUsageReport struct {
	Report
	WeeklyUsgFilter SearchQuery
}

type TransactionLogReport struct {
	Report
	TrxLogFilter SearchQuery
}

type VaultReport struct {
	Report
	// no filters yet
}

type TransactionRep struct {
	StockID           string
	LocationName      string
	MaterialType      string
	Qty               string
	UnitCost          string
	Cost              string
	Date              string
	SerialNumberRange string
	JobTicket         string
	ReasonDescription string
	CumulativeQty     string
}

type BalanceRep struct {
	StockID      string
	Description  string
	MaterialType string
	Qty          string
	TotalValue   string
}

type WeeklyUsageRep struct {
	CustomerName   string  `field:"customer_name" json:"customerName"`
	StockID        string  `field:"stock_id" json:"stockId"`
	MaterialType   string  `field:"material_type" json:"materialType"`
	QtyOnRefDate   int32   `field:"quantity_on_ref_date" json:"qtyOnRefDate"`
	AvgWeeklyUsg   float32 `field:"avg_weekly_usage" json:"avgWeeklyUsg"`
	WeeksRemaining float32 `field:"weeks_of_stock_remaining" json:"weeksRemaining"`
}

type VaultRep struct {
	InnerLocation string `field:"inner_location" json:"innerLocation"`
	OuterLocation string `field:"outer_location" json:"outerLocation"`
	StockID       string `field:"stock_id" json:"stockId"`
	InnerVaultQty int    `field:"inner_vault_quantity" json:"innerVaultQty"`
	OuterVaultQty int    `field:"outer_vault_quantity" json:"outerVaultQty"`
	TotalQty      int    `field:"total_quantity" json:"totalQty"`
}

var accLib accounting.Accounting = accounting.Accounting{Symbol: "$", Precision: 4}

func (t TransactionReport) GetReportList() ([]TransactionRep, error) {
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
						LEFT JOIN customers c ON m.customer_id = c.customer_id
						WHERE
							($1 = 0 OR m.customer_id = $1) AND
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
						LEFT JOIN customers c ON m.customer_id = c.customer_id
						WHERE
						    ($1 = 0 OR m.customer_id = $1) AND
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
		t.TrxFilter.CustomerID,
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

func (b BalanceReport) GetReportList() ([]BalanceRep, error) {
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
			($1 = 0 OR m.customer_id = $1) AND
			($2 = '' OR m.material_type::TEXT = $2) AND
			($3 = '' OR tl.updated_at::TEXT <= $3) AND
			($4 = '' OR m.owner::TEXT = $4) AND
			m.location_id IS NOT NULL
		GROUP BY m.stock_id, m.description, m.material_type
		ORDER BY m.material_type ASC, m.description ASC;
`,
		b.BlcFilter.CustomerID, b.BlcFilter.MaterialType, b.BlcFilter.DateAsOf, b.BlcFilter.Owner,
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
func (w WeeklyUsageReport) GetReportList() ([]WeeklyUsageRep, error) {
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
					m.customer_id,
					SUM(m.quantity) - COALESCE(fu.future_quantity_change, 0) AS quantity_on_ref_date
				FROM
					materials m
					LEFT JOIN future_usage fu ON fu.stock_id = m.stock_id
				GROUP BY
					m.stock_id,
					m.material_type,
					m.customer_id,
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
			c.name AS customer_name,
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
			LEFT JOIN customers c ON c.customer_id = mq.customer_id
		WHERE
			mq.quantity_on_ref_date > 0
			AND au.avg_weekly_usage > 0
			AND ($2 = 0 OR c.customer_id = $2)
			AND ($3 = '' OR mq.stock_id = $3)
			AND ($4 = '' OR mq.material_type::TEXT = $4)
		ORDER BY
			c.name,
			mq.material_type,
			mq.stock_id;

		`,
		dateAsOf,
		w.WeeklyUsgFilter.CustomerID,
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
			&usageTransaction.CustomerName,
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
			CustomerName:   usageTransaction.CustomerName,
			StockID:        usageTransaction.StockID,
			MaterialType:   usageTransaction.MaterialType,
			QtyOnRefDate:   usageTransaction.QtyOnRefDate,
			AvgWeeklyUsg:   usageTransaction.AvgWeeklyUsg,
			WeeksRemaining: usageTransaction.WeeksRemaining,
		})
	}

	return weeklyUsgList, err
}

func (tl TransactionLogReport) GetReportList() ([]TransactionRep, error) {
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
		LEFT JOIN customers c ON m.customer_id = c.customer_id
		LEFT JOIN material_usage_reasons mus ON mus.reason_id = tl.reason_id
		WHERE
			($1 = 0 OR w.warehouse_id = $1) AND
			($2 = 0 OR m.customer_id = $2) AND
			($3 = '' OR m.material_type::TEXT = $3) AND
			($4 = '' OR tl.updated_at::TEXT >= $4) AND
			($5 = '' OR tl.updated_at::TEXT <= $5) AND
			($6 = '' OR m.owner::TEXT = $6)
		ORDER BY tl.transaction_id ASC;`,
		tl.TrxLogFilter.WarehouseID,
		tl.TrxLogFilter.CustomerID,
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

func (vr VaultReport) GetReportList() ([]VaultRep, error) {
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
