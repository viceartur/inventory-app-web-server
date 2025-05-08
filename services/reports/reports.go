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

	// Weekly Usage Report:
	CustomerName   string  `field:"customer_name"`
	QtyOnRefDate   int32   `field:"quantity_on_ref_date"`
	AvgWeeklyUsg   float32 `field:"avg_weekly_usage"`
	WeeksRemaining float32 `field:"weeks_of_stock_remaining"`

	// Transactions Log Report:
	JobTicket string `field:"job_ticket"`
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
}

type BalanceRep struct {
	StockID      string
	Description  string
	MaterialType string
	Qty          string
	TotalValue   string
}

type WeeklyUsageRep struct {
	CustomerName   string  `json:"customerName"`
	StockID        string  `json:"stockId"`
	MaterialType   string  `json:"materialType"`
	QtyOnRefDate   int32   `json:"qtyOnRefDate"`
	AvgWeeklyUsg   float32 `json:"avgWeeklyUsg"`
	WeeksRemaining float32 `json:"weeksRemaining"`
}

var accLib accounting.Accounting = accounting.Accounting{Symbol: "$", Precision: 4}

func (t TransactionReport) GetReportList() ([]TransactionRep, error) {
	rows, err := t.DB.Query(`SELECT
								m.stock_id,
								m.material_type,
								tl.quantity_change as "quantity",
								p.cost as "unit_cost",
								(tl.quantity_change * p.cost) as "cost",
								tl.updated_at,
								COALESCE(tl.serial_number_range, '')
							 FROM transactions_log tl
							 LEFT JOIN prices p ON p.price_id = tl.price_id
							 LEFT JOIN materials m ON m.material_id = p.material_id
							 LEFT JOIN customers c ON m.customer_id = c.customer_id
							 WHERE 
								($1 = 0 OR m.customer_id = $1) AND
								($2 = '' OR m.material_type::TEXT = $2) AND
								($3 = '' OR tl.updated_at::TEXT >= $3) AND
								($4 = '' OR tl.updated_at::TEXT <= $4) AND
								($5 = '' OR m.owner::TEXT = $5)
							 ORDER BY tl.transaction_id ASC;`,
		t.TrxFilter.CustomerID, t.TrxFilter.MaterialType, t.TrxFilter.DateFrom, t.TrxFilter.DateTo, t.TrxFilter.Owner)
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
				SELECT $1::DATE AS ref_date
			),

			-- Get total quantity_change *after* the reference date
			future_usage AS (
				SELECT
					p.material_id,
					SUM(tl.quantity_change) AS future_quantity_change
				FROM
					transactions_log tl
					JOIN prices p ON p.price_id = tl.price_id
					JOIN reference_date r ON true
				WHERE
					tl.updated_at > r.ref_date
				GROUP BY
					p.material_id
			),

			-- Quantity on the reference date = current quantity - future usage
			material_quantities AS (
				SELECT
					m.material_id,
					m.stock_id,
					m.material_type,
					m.customer_id,
					m.quantity - COALESCE(fu.future_quantity_change, 0) AS quantity_on_ref_date
				FROM
					materials m
					LEFT JOIN future_usage fu ON fu.material_id = m.material_id
			),

			-- Filter transactions within the 6-week window before reference date
			filtered_transactions AS (
				SELECT
					tl.*,
					p.material_id
				FROM
					transactions_log tl
					JOIN prices p ON p.price_id = tl.price_id
					JOIN reference_date r ON true
				WHERE
					tl.updated_at BETWEEN r.ref_date - INTERVAL '6 weeks' AND r.ref_date
			),

			-- Weekly usage per material
			weekly_usage AS (
				SELECT
					material_id,
					DATE_TRUNC ('week', updated_at) AS week_start,
					SUM(quantity_change) AS total_used
				FROM
					filtered_transactions
				WHERE
					-- Only actual materials used are in the query
					quantity_change < 0
					AND notes NOT ILIKE 'moved from a location'
				GROUP BY
					material_id,
					week_start
			),

			-- Average usage per material over 6 weeks
			average_usage AS (
				SELECT
					material_id,
					ABS(AVG(total_used)) AS avg_weekly_usage
				FROM
					weekly_usage
				GROUP BY
					material_id
			)

		-- Final output:
		SELECT
			c.name as customer_name,
			mq.stock_id,
			mq.material_type,
			mq.quantity_on_ref_date,
			ROUND(au.avg_weekly_usage, 2) AS avg_weekly_usage,
			CASE
				WHEN au.avg_weekly_usage IS NULL OR au.avg_weekly_usage = 0 THEN NULL
				ELSE ROUND(mq.quantity_on_ref_date / au.avg_weekly_usage, 2)
			END AS weeks_of_stock_remaining
		FROM
			material_quantities mq
			LEFT JOIN average_usage au ON mq.material_id = au.material_id
			LEFT JOIN customers c ON c.customer_id = mq.customer_id
		WHERE
			mq.quantity_on_ref_date > 0 AND
			avg_weekly_usage > 0 AND
			($2 = 0 OR c.customer_id = $2) AND
			($3 = '' OR mq.stock_id = $3) AND
			($4 = '' OR mq.material_type::TEXT = $4);
		`,
		dateAsOf, w.WeeklyUsgFilter.CustomerID, w.WeeklyUsgFilter.StockId, w.WeeklyUsgFilter.MaterialType,
	)
	if err != nil {
		return []WeeklyUsageRep{}, err
	}

	weeklyUsgList := []WeeklyUsageRep{}

	for rows.Next() {
		usageTransaction := Transaction{}

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
			tl.updated_at
		FROM transactions_log tl
		LEFT JOIN prices p ON p.price_id = tl.price_id
		LEFT JOIN materials m ON m.material_id = p.material_id
		LEFT JOIN locations l ON l.location_id = m.location_id
		LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
		LEFT JOIN customers c ON m.customer_id = c.customer_id
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
			&trx.UpdatedAt)
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
		})
	}

	return trxList, nil
}
