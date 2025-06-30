package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"inv_app/database"
	"inv_app/services/email"
	"inv_app/services/reports"

	"github.com/gorilla/mux"
)

type ReqBody struct {
	DateFrom string `json:"dateFrom"`
	DateTo   string `json:"dateTo"`
}

func EmailCustomerReportHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	vars := mux.Vars(r)
	customerId, _ := strconv.Atoi(vars["customerId"])
	var body ReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var reportFilter reports.SearchQuery
	reportFilter.CustomerID = customerId
	reportFilter.DateFrom = body.DateFrom
	reportFilter.DateTo = body.DateTo

	err := email.HandleCustomerReportsEmail(db, reportFilter)
	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusBadRequest)
		return
	}

	res := SuccessResponse{Message: "Email send handled."}
	json.NewEncoder(w).Encode(res)
}

func EmailCustomerReportsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	var body ReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var reportFilter reports.SearchQuery
	reportFilter.DateFrom = body.DateFrom
	reportFilter.DateTo = body.DateTo

	err := email.HandleCustomerReportsEmail(db, reportFilter)
	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusBadRequest)
		return
	}

	res := SuccessResponse{Message: "Emails send handled."}
	json.NewEncoder(w).Encode(res)
}
