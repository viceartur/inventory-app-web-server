package materials

import "time"

type IncomingMaterialJSON struct {
	ShippingId   string `json:"shippingId"`
	CustomerID   string `json:"customerId"`
	StockID      string `json:"stockId"`
	MaterialType string `json:"type"`
	Qty          string `json:"quantity"`
	Cost         string `json:"cost"`
	MinQty       string `json:"minQuantity"`
	MaxQty       string `json:"maxQuantity"`
	Description  string `json:"description"`
	Owner        string `json:"owner"`
	IsActive     bool   `json:"isActive"`
	UserID       string `json:"userId"`
}

type IncomingMaterialDB struct {
	ShippingID   string  `field:"shipping_id"`
	CustomerName string  `field:"customer_name"`
	CustomerID   int     `field:"customer_id"`
	StockID      string  `field:"stock_id"`
	Cost         float64 `field:"cost"`
	Quantity     int     `field:"quantity"`
	MinQty       int     `field:"min_required_quantity"`
	MaxQty       int     `field:"max_required_quantity"`
	Description  string  `field:"description"`
	IsActive     bool    `field:"is_active"`
	MaterialType string  `field:"material_type"`
	Owner        string  `field:"owner"`
	UserID       int     `field:"user_id"`
	UserName     string  `field:"username"`
}

type IncomingMaterial struct {
	shippingId int
	qty        int
}

type MaterialJSON struct {
	MaterialID        string `json:"materialId"`
	LocationID        string `json:"locationId"`
	Qty               string `json:"quantity"`
	Notes             string `json:"notes"`
	IsPrimary         *bool  `json:"isPrimary,omitempty"`
	SerialNumberRange string `json:"serialNumberRange"`
	JobTicket         string `json:"jobTicket"`
	StockID           string `json:"stockId"`
	Description       string `json:"description"`
	Status            string `json:"status"`
}

type RequestedMaterialsJSON struct {
	Materials []MaterialJSON `json:"materials"`
	UserID    int            `json:"userId"`
}

type MaterialDB struct {
	MaterialID        int       `field:"material_id"`
	WarehouseName     string    `field:"warehouse_name"`
	StockID           string    `field:"stock_id"`
	CustomerID        int       `field:"customer_id"`
	CustomerName      string    `field:"customer_name"`
	LocationID        int       `field:"location_id"`
	LocationName      string    `field:"location_name"`
	MaterialType      string    `field:"material_type"`
	Description       string    `field:"description"`
	Notes             string    `field:"notes"`
	Quantity          int       `field:"quantity"`
	UpdatedAt         time.Time `field:"updated_at"`
	IsActive          bool      `field:"is_active"`
	MinQty            int       `field:"min_required_quantity"`
	MaxQty            int       `field:"max_required_quantity"`
	Owner             string    `field:"onwer"`
	IsPrimary         bool      `field:"is_primary"`
	SerialNumberRange string    `field:"serial_number_range"`
	RequestID         int       `field:"request_id"`
	UserName          string    `field:"username"`
	Status            string    `field:"status"`
	QtyRequested      int       `field:"quantity_requested"`
	QtyUsed           int       `field:"quantity_used"`
	RequestedAt       time.Time `field:"requested_at"`
}

type Transaction struct {
	MaterialID    int    `field:"material_id" json:"materialId"`
	StockID       string `field:"stock_id" json:"stockId"`
	LocationID    int    `field:"location_id" json:"locationId"`
	LocationName  string `field:"location_name" json:"locationName"`
	WarehouseID   int    `field:"warehouse_id" json:"warehouseId"`
	WarehouseName string `field:"warehouse_name" json:"warehouseName"`
	Quantity      int    `field:"quantity" json:"quantity"`
	JobTicket     string `field:"job_ticket" json:"jobTicket"`
}

type MaterialFilter struct {
	MaterialId   int
	StockId      string
	CustomerName string
	Description  string
	LocationName string
	Status       string
	RequestId    int
	RequestedAt  string
}

type Price struct {
	priceId    int
	materialId int
	qty        int
	cost       float64
}

type PriceToRemove struct {
	materialId        int
	qty               int
	notes             string
	jobTicket         string
	serialNumberRange string
}

type PriceDB struct {
	PriceID    int     `field:"price_id"`
	MaterialID int     `field:"material_id"`
	Qty        int     `field:"quantity"`
	Cost       float64 `field:"cost"`
}

type TransactionInfo struct {
	priceId           int       `field:"price_id"`
	qty               int       `field:"quantity_change"`
	notes             string    `field:"notes"`
	jobTicket         string    `field:"job_ticket"`
	updatedAt         time.Time `field:"updated_at"`
	serialNumberRange string    `field:"serial_number_range"`
}
