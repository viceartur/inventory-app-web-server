package reports

import "time"

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
	CustomerID   int
	WarehouseID  int
	ProgramID    int
	StockId      string
	Owner        string
	MaterialType string
	DateFrom     string
	DateTo       string
	DateAsOf     string
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
	ProgramName    string  `field:"program_name" json:"programName"`
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

type CustomerUsageRep struct {
	CustomerID     int    `field:"customer_id" json:"customerId"`
	ProgramName    string `field:"program_name" json:"programName"`
	MaterialType   string `field:"material_type" json:"materialType"`
	StockID        string `field:"stock_id" json:"stockId"`
	QtyStart       int    `field:"qty_start" json:"qtyStart"`
	QtyReceived    int    `field:"qty_received" json:"qtyReceived"`
	QtyUsed        int    `field:"qty_used" json:"qtyUsed"`
	QtySpoiled     int    `field:"qty_spoiled" json:"qtySpoiled"`
	QtyEnd         int    `field:"qty_end" json:"qtyEnd"`
	WeekAvg        int    `field:"six_week_avg_to_ref_end" json:"weekAvg"`
	WeeksRemaining int    `field:"weeks_remaining" json:"weeksRemaining"`
}
