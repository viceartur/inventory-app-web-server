package handlers

import (
	"encoding/json"
	"inv_app/database"
	"inv_app/services/import_data"
	"net/http"
)

func ImportData(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	var dataToImport import_data.ImportJSON
	json.NewDecoder(r.Body).Decode(&dataToImport)
	importRes, err := import_data.ImportDataToDB(db, dataToImport)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response := SuccessResponseJSON{Message: "Data Imported to the DB", Data: importRes}
	json.NewEncoder(w).Encode(response)
}
