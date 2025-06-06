package warehouses

import (
	"database/sql"
)

type Warehouse struct {
	WarehouseID   int    `field:"id" json:"warehouseId"`
	WarehouseName string `field:"warehouse_name" json:"warehouseName"`
	LocationName  string `field:"location_name" json:"locationName"`
}

func FetchWarehouses(db *sql.DB) ([]Warehouse, error) {
	rows, err := db.Query("SELECT * FROM warehouses;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warehouses []Warehouse

	for rows.Next() {
		var warehouse Warehouse
		if err := rows.Scan(&warehouse.WarehouseID, &warehouse.WarehouseName); err != nil {
			return warehouses, err
		}
		warehouses = append(warehouses, warehouse)
	}
	if err = rows.Err(); err != nil {
		return warehouses, err
	}

	return warehouses, nil
}

func CreateWarehouse(warehouse Warehouse, db *sql.DB) error {
	warehouses, err := FetchWarehouses(db)

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
