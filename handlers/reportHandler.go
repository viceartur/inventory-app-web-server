package handlers

import (
	"encoding/json"
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

	trxRep := reports.TransactionReport{Report: reports.Report{DB: db}, TrxFilter: reports.SearchQuery{
		CustomerID:   customerId,
		Owner:        owner,
		MaterialType: materialType,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
	}}
	trxReport, err := trxRep.GetReportList()
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

	balanceRep := reports.BalanceReport{Report: reports.Report{DB: db}, BlcFilter: reports.SearchQuery{
		CustomerID:   customerId,
		Owner:        owner,
		MaterialType: materialType,
		DateAsOf:     dateAsOf,
	}}
	balanceReport, err := balanceRep.GetReportList()
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

	weeklyUsgRep := reports.WeeklyUsageReport{
		Report: reports.Report{DB: db},
		WeeklyUsgFilter: reports.SearchQuery{
			CustomerID:   customerId,
			StockId:      stockId,
			MaterialType: materialType,
			DateAsOf:     dateAsOf,
		}}

	weeklyUsageReport, err := weeklyUsgRep.GetReportList()
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

	trxLogRep := reports.TransactionLogReport{Report: reports.Report{DB: db}, TrxLogFilter: reports.SearchQuery{
		WarehouseID:  warehouseId,
		CustomerID:   customerId,
		Owner:        owner,
		MaterialType: materialType,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
	}}
	trxLogReport, err := trxLogRep.GetReportList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(trxLogReport)
}

func GetVaultReport(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	vaultRep := reports.VaultReport{Report: reports.Report{DB: db}}
	vaultReport, err := vaultRep.GetReportList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(vaultReport)
}
