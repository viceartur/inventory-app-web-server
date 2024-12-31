package main

import (
	"database/sql"
	"encoding/csv"
	"os"
	"strconv"
)

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
	UnitCost      string
	ERR_REASON    string
}

type ImportResponse struct {
	Records              int
	Imported_Records     int
	Not_Imported_Records int
	ErrLocations         []string
	Not_Imported_Data    []ImportData
}

func importDataToDB(db *sql.DB) (ImportResponse, error) {
	file, err := os.Open("./import_data.csv")
	if err != nil {
		return ImportResponse{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	if err != nil {
		return ImportResponse{}, err
	}

	db.Query(`
		DELETE FROM transactions_log;
		DELETE FROM prices;
		DELETE FROM materials;
		DELETE FROM locations;
		DELETE FROM customers;
		DELETE FROM warehouses;
	`)

	materialsCounter := 0
	notImportedData := []ImportData{}
	locations := []string{}

	for _, record := range records {
		customerName := record[0]
		customerCode := record[1]
		warehouseName := record[2]
		locationName := record[3]
		stockID := record[4]
		materialType := record[5]
		description := record[6]
		notes := record[7]
		qty, _ := strconv.Atoi(record[8])
		minQty, _ := strconv.Atoi(record[9])
		maxQty, _ := strconv.Atoi(record[10])
		isActive, _ := strconv.ParseBool(record[11])
		owner := record[12]
		unitCost := record[13]

		importData := ImportData{
			CustomerName:  customerName,
			CustomerCode:  customerCode,
			WarehouseName: warehouseName,
			LocationName:  locationName,
			StockID:       stockID,
			MaterialType:  materialType,
			Description:   description,
			Notes:         notes,
			Qty:           qty,
			MinQty:        minQty,
			MaxQty:        maxQty,
			IsActive:      isActive,
			Owner:         owner,
			UnitCost:      unitCost,
		}

		if customerName == "" {
			importData.ERR_REASON = "No Customer Name"
			notImportedData = append(notImportedData, importData)
			continue
		}

		if qty == 0 {
			importData.ERR_REASON = "No Qty"
			notImportedData = append(notImportedData, importData)
			continue
		}

		// Check for a customer
		var customerId int
		err = db.QueryRow(`SELECT customer_id FROM customers
						WHERE name = $1
						AND customer_code = $2`,
			customerName, customerCode).
			Scan(&customerId)

		if customerId == 0 {
			err = db.QueryRow(`
			INSERT INTO customers(name,customer_code) VALUES($1,$2) RETURNING customer_id`,
				customerName, customerCode).
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
			warehouseName).
			Scan(&warehouseId)

		if warehouseId == 0 {
			err = db.QueryRow(`
					INSERT INTO warehouses(name) VALUES($1) RETURNING warehouse_id`,
				warehouseName).
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
			locationName, warehouseId).
			Scan(&locationId)

		if locationId == 0 {
			err = db.QueryRow(`
			INSERT INTO locations(name,warehouse_id) VALUES($1,$2) RETURNING location_id`,
				locationName, warehouseId).
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
					max_required_quantity,is_active,owner,updated_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW())
			RETURNING material_id`,
			stockID, locationId, customerId, materialType,
			description, notes, qty, minQty, maxQty, isActive, owner).
			Scan(&materialId)
		if err != nil {
			importData.ERR_REASON = err.Error()
			notImportedData = append(notImportedData, importData)
			locations = append(locations, locationName)
			continue
		}

		var priceId int
		err = db.QueryRow(`
		INSERT INTO prices(material_id,quantity,cost)
		VALUES($1,$2,$3)
		RETURNING price_id`,
			materialId, qty, unitCost,
		).Scan(&priceId)
		if err != nil {
			importData.ERR_REASON = err.Error()
			notImportedData = append(notImportedData, importData)
			continue
		}

		_, err = db.Query(`
			INSERT INTO transactions_log(price_id, quantity_change, notes, job_ticket, updated_at)
			VALUES($1,$2,$3,$4,NOW())`,
			priceId, qty, notes, "Imported",
		)
		if err != nil {
			importData.ERR_REASON = err.Error()
			notImportedData = append(notImportedData, importData)
			continue
		}

		materialsCounter++

	}

	res := ImportResponse{
		Records:              len(records),
		Imported_Records:     materialsCounter,
		Not_Imported_Records: len(notImportedData),
		ErrLocations:         locations,
		Not_Imported_Data:    notImportedData}
	return res, nil
}
