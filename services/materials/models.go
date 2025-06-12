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
	IsPrimary         *bool  `json:"isPrimary"`
	SerialNumberRange string `json:"serialNumberRange"`
	JobTicket         string `json:"jobTicket"`
	StockID           string `json:"stockId"`
	Description       string `json:"description"`
	Status            string `json:"status"`
	ReasonID          int    `json:"reasonId"`
}

type RequestedMaterialsJSON struct {
	Materials []MaterialJSON `json:"materials"`
	UserID    int            `json:"userId"`
}

type Material struct {
	MaterialID        int       `field:"material_id" json:"materialId"`
	WarehouseName     string    `field:"warehouse_name" json:"warehouseName"`
	StockID           string    `field:"stock_id" json:"stockId"`
	CustomerID        int       `field:"customer_id" json:"customerId"`
	CustomerName      string    `field:"customer_name" json:"customerName"`
	IsActiveCustomer  bool      `field:"is_active_customer" json:"isActiveCustomer"`
	LocationID        int       `field:"location_id" json:"locationId"`
	LocationName      string    `field:"location_name" json:"locationName"`
	MaterialType      string    `field:"material_type" json:"materialType"`
	Description       string    `field:"description" json:"description"`
	Notes             string    `field:"notes" json:"notes"`
	Quantity          int       `field:"quantity" json:"quantity"`
	UpdatedAt         time.Time `field:"updated_at" json:"updatedAt"`
	IsActiveMaterial  bool      `field:"is_active_material" json:"isActiveMaterial"`
	MinQty            int       `field:"min_required_quantity" json:"minQty"`
	MaxQty            int       `field:"max_required_quantity" json:"maxQty"`
	Owner             string    `field:"onwer" json:"owner"`
	IsPrimary         bool      `field:"is_primary" json:"isPrimary"`
	SerialNumberRange string    `field:"serial_number_range" json:"serialNumberRange"`
	RequestID         int       `field:"request_id" json:"requestId"`
	UserName          string    `field:"username" json:"userName"`
	Status            string    `field:"status" json:"status"`
	QtyRequested      int       `field:"quantity_requested" json:"qtyRequested"`
	QtyUsed           int       `field:"quantity_used" json:"qtyUsed"`
	RequestedAt       time.Time `field:"requested_at" json:"requestedAt"`
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
	reasonId          int
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
	reasonId          int       `field:"reason_id"`
}

type MaterialUsageReason struct {
	ReasonID    int    `field:"reason_id" json:"reasonId"`
	ReasonType  string `field:"reason_type" json:"reasonType"`
	Description string `field:"description" json:"description"`
	Code        int    `field:"code" json:"code"`
}
