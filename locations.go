package main

import (
	"database/sql"
)

type LocationFilter struct {
	stockId string
	owner   string
}

type LocationDB struct {
	ID            int    `field:"location_id"`
	Name          string `field:"name"`
	WarehouseID   int    `field:"warehouse_id"`
	WarehouseName string `field:"warehouse_name"`
}

func fetchLocations(db *sql.DB) ([]LocationDB, error) {
	rows, err := db.Query(`
		SELECT l.location_id, l.name, w.warehouse_id, w.name as "warehouse_name"
		FROM locations l
		LEFT JOIN warehouses w
		ON l.warehouse_id = w.warehouse_id
		ORDER BY l.name ASC;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []LocationDB

	for rows.Next() {
		var location LocationDB
		if err := rows.Scan(&location.ID, &location.Name, &location.WarehouseID, &location.WarehouseName); err != nil {
			return locations, err
		}
		locations = append(locations, location)
	}
	if err = rows.Err(); err != nil {
		return locations, err
	}

	return locations, nil
}

func fetchAvailableLocations(db *sql.DB, opts LocationFilter) ([]LocationDB, error) {
	rows, err := db.Query(`
		SELECT l.location_id, l.name, l.warehouse_id FROM locations l
		LEFT JOIN materials m
		ON l.location_id = m.location_id
		WHERE m.stock_id = $1 AND m.owner = $2 OR m.material_id IS NULL
		ORDER BY l.name ASC;
	`, opts.stockId, opts.owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []LocationDB

	for rows.Next() {
		var location LocationDB
		if err := rows.Scan(&location.ID, &location.Name, &location.WarehouseID); err != nil {
			return locations, err
		}
		locations = append(locations, location)
	}
	if err = rows.Err(); err != nil {
		return locations, err
	}

	return locations, nil
}
