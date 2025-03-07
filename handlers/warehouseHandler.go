package handlers

import (
	"encoding/json"
	"inv_app/database"
	"inv_app/services/warehouses"
	"net/http"
)

func CreateWarehouseHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	var warehouse warehouses.WarehouseJSON
	json.NewDecoder(r.Body).Decode(&warehouse)
	err := warehouses.CreateWarehouse(warehouse, db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(warehouse)
}

func GetWarehouseHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	warehouses, err := warehouses.FetchWarehouses(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(warehouses)
}
