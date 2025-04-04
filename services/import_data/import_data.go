package import_data

import (
	"database/sql"
	"log"
	"strconv"
)

type ImportDataJSON struct {
	CustomerName  string  `json:"Customer Name"`
	CustomerCode  string  `json:"Customer Code"`
	WarehouseName string  `json:"Warehouse Name"`
	LocationName  string  `json:"Location Name"`
	StockID       string  `json:"Stock ID"`
	MaterialType  string  `json:"Material Type"`
	Description   string  `json:"Description"`
	Notes         string  `json:"Notes"`
	Qty           int     `json:"Qty"`
	MinQty        int     `json:"Min Qty"`
	MaxQty        int     `json:"Max Qty"`
	IsActive      bool    `json:"Is Active"`
	Owner         string  `json:"Owner"`
	UnitCost      float64 `json:"Unit Cost"`
}

type ImportJSON struct {
	Data []ImportDataJSON `json:"data"`
}

type ImportData struct {
	CustomerName  string
	CustomerCode  string
	WarehouseName string
	LocationName  string
	StockID       string
	MaterialType  string
	Description   string
	Notes         string
	Qty           int
	MinQty        int
	MaxQty        int
	IsActive      bool
	Owner         string
	UnitCost      float64
	ERR_REASON    string
}

type ImportResponse struct {
	Records              int
	Imported_Records     int
	Not_Imported_Records int
	ErrLocations         []string
	Not_Imported_Data    []ImportData
}

func removeLeadingZeros(s string) string {
	num, err := strconv.Atoi(s)
	if err != nil {
		return s
	}
	return strconv.Itoa(num)
}

func ImportDataToDB(db *sql.DB, data ImportJSON) (ImportResponse, error) {
	materialsCounter := 0
	notImportedData := []ImportData{}
	locations := []string{}

	for i := range data.Data {
		record := data.Data[i]
		importData := ImportData{
			CustomerName:  record.CustomerName,
			CustomerCode:  removeLeadingZeros(record.CustomerCode),
			WarehouseName: record.WarehouseName,
			LocationName:  record.LocationName,
			StockID:       record.StockID,
			MaterialType:  record.MaterialType,
			Description:   record.Description,
			Notes:         record.Notes,
			Qty:           record.Qty,
			MinQty:        record.MinQty,
			MaxQty:        record.MaxQty,
			IsActive:      record.IsActive,
			Owner:         record.Owner,
			UnitCost:      record.UnitCost,
		}

		if importData.StockID == "" {
			log.Println(importData)
			importData.ERR_REASON = "No Stock ID provided"
			notImportedData = append(notImportedData, importData)
			continue
		}

		if importData.CustomerName == "" {
			importData.ERR_REASON = "No Customer Name"
			notImportedData = append(notImportedData, importData)
			continue
		}

		if importData.LocationName == "" {
			importData.ERR_REASON = "No Location Name"
			notImportedData = append(notImportedData, importData)
			continue
		}

		if importData.Qty == 0 {
			importData.ERR_REASON = "No Qty"
			notImportedData = append(notImportedData, importData)
			continue
		}

		// Check for a customer
		var customerId int
		err := db.QueryRow(`
			SELECT customer_id FROM customers
			WHERE LOWER(name) = LOWER($1) AND customer_code = $2`,
			importData.CustomerName, importData.CustomerCode).
			Scan(&customerId)

		if customerId == 0 {
			err = db.QueryRow(`
			INSERT INTO customers(name,customer_code) VALUES($1,$2) RETURNING customer_id`,
				importData.CustomerName, importData.CustomerCode).
				Scan(&customerId)
		}
		if err != nil {
			importData.ERR_REASON = err.Error()
			notImportedData = append(notImportedData, importData)
			continue
		}

		// Check for a warehouse
		var warehouseId int
		err = db.QueryRow(`SELECT warehouse_id FROM warehouses
						WHERE name = $1
						`,
			importData.WarehouseName).
			Scan(&warehouseId)

		if warehouseId == 0 {
			err = db.QueryRow(`
					INSERT INTO warehouses(name) VALUES($1) RETURNING warehouse_id`,
				importData.WarehouseName).
				Scan(&warehouseId)
		}
		if err != nil {
			importData.ERR_REASON = err.Error()
			notImportedData = append(notImportedData, importData)
			continue
		}

		// Check for a location
		var locationId int
		err = db.QueryRow(`SELECT location_id FROM locations
						WHERE name = $1
						AND warehouse_id = $2
						`,
			importData.LocationName, warehouseId).
			Scan(&locationId)

		if locationId == 0 {
			err = db.QueryRow(`
			INSERT INTO locations(name,warehouse_id) VALUES($1,$2) RETURNING location_id`,
				importData.LocationName, warehouseId).
				Scan(&locationId)
		}
		if err != nil {
			importData.ERR_REASON = err.Error()
			notImportedData = append(notImportedData, importData)
			continue
		}

		var materialId int
		err = db.QueryRow(`
			INSERT INTO materials(
					stock_id,location_id,customer_id,material_type,
					description,notes,quantity,min_required_quantity,
					max_required_quantity,is_active,owner,updated_at,is_primary)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW(),false)
			RETURNING material_id`,
			importData.StockID, locationId, customerId, importData.MaterialType,
			importData.Description, importData.Notes, importData.Qty, importData.MinQty, importData.MaxQty, importData.IsActive, importData.Owner).
			Scan(&materialId)
		if err != nil {
			importData.ERR_REASON = err.Error()
			notImportedData = append(notImportedData, importData)
			locations = append(locations, importData.LocationName)
			continue
		}

		var priceId int
		err = db.QueryRow(`
		INSERT INTO prices(material_id,quantity,cost)
		VALUES($1,$2,$3)
		RETURNING price_id`,
			materialId, importData.Qty, importData.UnitCost,
		).Scan(&priceId)
		if err != nil {
			importData.ERR_REASON = err.Error()
			notImportedData = append(notImportedData, importData)
			continue
		}

		_, err = db.Query(`
			INSERT INTO transactions_log(price_id, quantity_change, notes, job_ticket, updated_at)
			VALUES($1,$2,$3,$4,NOW())`,
			priceId, importData.Qty, importData.Notes, "Imported",
		)
		if err != nil {
			importData.ERR_REASON = err.Error()
			notImportedData = append(notImportedData, importData)
			continue
		}

		materialsCounter++

	}

	res := ImportResponse{
		Records:              len(data.Data),
		Imported_Records:     materialsCounter,
		Not_Imported_Records: len(notImportedData),
		ErrLocations:         locations,
		Not_Imported_Data:    notImportedData}
	return res, nil
}
