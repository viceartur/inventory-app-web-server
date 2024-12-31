package main

import (
	"database/sql"
	"log"
)

type WarehouseJSON struct {
	WarehouseName string `json:"warehouseName"`
	LocationName  string `json:"locationName"`
}

type WarehouseDB struct {
	WarehouseID   int    `field:"warehouse_id"`
	WarehouseName string `field:"name"`
}

func fetchWarehouses(db *sql.DB) ([]WarehouseDB, error) {
	rows, err := db.Query("SELECT * FROM warehouses;")
	if err != nil {
		log.Println("Error fetchWarehouses1: ", err)
		return nil, err
	}
	defer rows.Close()

	var warehouses []WarehouseDB

	for rows.Next() {
		var warehouse WarehouseDB
		if err := rows.Scan(&warehouse.WarehouseID, &warehouse.WarehouseName); err != nil {
			log.Println("Error fetchWarehouses2: ", err)
			return warehouses, err
		}
		warehouses = append(warehouses, warehouse)
	}
	if err = rows.Err(); err != nil {
		return warehouses, err
	}

	return warehouses, nil
}

func createWarehouse(warehouse WarehouseJSON, db *sql.DB) error {
	warehouses, err := fetchWarehouses(db)

	if err != nil {
		return err
	}

	warehousesMap := make(map[string]int)
	for _, warehouse := range warehouses {
		warehousesMap[warehouse.WarehouseName] = warehouse.WarehouseID
	}

	id, ok := warehousesMap[warehouse.WarehouseName]

	var warehouseId int
	if !ok {
		err = db.QueryRow(`
			INSERT INTO warehouses(name) VALUES($1)
			RETURNING warehouse_id;`,
			warehouse.WarehouseName).Scan(&warehouseId)
		if err != nil {
			return err
		}
	} else {
		warehouseId = id
	}

	_, err = db.Exec("INSERT INTO locations(name, warehouse_id) VALUES ($1,$2);",
		warehouse.LocationName, warehouseId)
	if err != nil {
		return err
	}

	return nil
}
