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

	var warehouse warehouses.Warehouse
	json.NewDecoder(r.Body).Decode(&warehouse)
	err := warehouses.CreateWarehouse(warehouse, db)

	if err != nil {
		errRes := ErrorResponseJSON{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessResponseJSON{Message: "Warehouse and Location pair created."})
}

func GetWarehouseHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	warehouses, err := warehouses.FetchWarehouses(db)

	if err != nil {
		errRes := ErrorResponseJSON{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(warehouses)
}
