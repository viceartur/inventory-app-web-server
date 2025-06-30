package handlers

import (
	"encoding/json"
	"fmt"
	"inv_app/database"
	"inv_app/services/reports"
	"net/http"
	"strconv"
)

func GetTransactionsReport(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	customerIdStr := r.URL.Query().Get("customerId")
	customerId, _ := strconv.Atoi(customerIdStr)
	owner := r.URL.Query().Get("owner")
	materialType := r.URL.Query().Get("materialType")
	dateFrom := r.URL.Query().Get("dateFrom")
	dateTo := r.URL.Query().Get("dateTo")

	report := reports.Report{DB: db, TrxFilter: reports.SearchQuery{
		ProgramID:    customerId,
		Owner:        owner,
		MaterialType: materialType,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
	}}

	trxReport, err := report.GetTransactions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(trxReport)
}

func GetBalanceReport(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	customerIdStr := r.URL.Query().Get("customerId")
	customerId, _ := strconv.Atoi(customerIdStr)
	owner := r.URL.Query().Get("owner")
	materialType := r.URL.Query().Get("materialType")
	dateAsOf := r.URL.Query().Get("dateAsOf")

	report := reports.Report{DB: db, BlcFilter: reports.SearchQuery{
		ProgramID:    customerId,
		Owner:        owner,
		MaterialType: materialType,
		DateAsOf:     dateAsOf,
	}}
	balanceReport, err := report.GetBalance()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(balanceReport)
}

func GetWeeklyUsageReport(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	customerIdStr := r.URL.Query().Get("customerId")
	customerId, _ := strconv.Atoi(customerIdStr)
	stockId := r.URL.Query().Get("stockId")
	materialType := r.URL.Query().Get("materialType")
	dateAsOf := r.URL.Query().Get("dateAsOf")

	report := reports.Report{
		DB: db,
		WeeklyUsgFilter: reports.SearchQuery{
			ProgramID:    customerId,
			StockId:      stockId,
			MaterialType: materialType,
			DateAsOf:     dateAsOf,
		}}

	weeklyUsageReport, err := report.GetWeeklyUsage()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weeklyUsageReport)
}

func GetTransactionsLogReport(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	warehouseIdStr := r.URL.Query().Get("warehouseId")
	warehouseId, _ := strconv.Atoi(warehouseIdStr)
	customerIdStr := r.URL.Query().Get("customerId")
	customerId, _ := strconv.Atoi(customerIdStr)
	owner := r.URL.Query().Get("owner")
	materialType := r.URL.Query().Get("materialType")
	dateFrom := r.URL.Query().Get("dateFrom")
	dateTo := r.URL.Query().Get("dateTo")

	report := reports.Report{DB: db, TrxLogFilter: reports.SearchQuery{
		WarehouseID:  warehouseId,
		ProgramID:    customerId,
		Owner:        owner,
		MaterialType: materialType,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
	}}
	trxLogReport, err := report.GetTransactionsLog()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(trxLogReport)
}

func GetVaultReport(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	report := reports.Report{DB: db}
	vaultReport, err := report.GetVault()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(vaultReport)
}

func GetCustomerUsageReport(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	customerIdStr := r.URL.Query().Get("customerId")
	customerId, _ := strconv.Atoi(customerIdStr)
	dateFrom := r.URL.Query().Get("dateFrom")
	dateTo := r.URL.Query().Get("dateTo")

	report := reports.Report{DB: db, CustomerUsageFilter: reports.SearchQuery{
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		CustomerID: customerId,
	}}

	customerUsgReport, err := report.GetCustomerUsage()

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{
		Message: fmt.Sprintf(
			"Customer Usage Report for the period from %s to %s. Customer ID: %s.",
			dateFrom,
			dateTo,
			customerIdStr,
		),
		Data: customerUsgReport,
	}
	json.NewEncoder(w).Encode(res)
}
