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
}

type SearchQuery struct {
	CustomerId   int
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

type TransactionRep struct {
	StockID           string
	MaterialType      string
	Qty               string
	UnitCost          string
	Cost              string
	Date              string
	SerialNumberRange string
}

type BalanceRep struct {
	StockID      string
	Description  string
	MaterialType string
	Qty          string
	TotalValue   string
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
							 ORDER BY tl.updated_at ASC;`,
		t.TrxFilter.CustomerId, t.TrxFilter.MaterialType, t.TrxFilter.DateFrom, t.TrxFilter.DateTo, t.TrxFilter.Owner)
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
		b.BlcFilter.CustomerId, b.BlcFilter.MaterialType, b.BlcFilter.DateAsOf, b.BlcFilter.Owner,
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
