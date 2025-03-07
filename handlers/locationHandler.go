package handlers

import (
	"encoding/json"
	"inv_app/database"
	"inv_app/services/locations"
	"net/http"
)

func GetLocationsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	locations, err := locations.FetchLocations(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(locations)
}

func GetAvailableLocationsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	stockId := r.URL.Query().Get("stockId")
	owner := r.URL.Query().Get("owner")

	locations, err := locations.FetchAvailableLocations(db, locations.LocationFilter{StockId: stockId, Owner: owner})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(locations)
}
