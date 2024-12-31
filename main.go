package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type SuccessResponseJSON struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorResponseJSON struct {
	Message string `json:"message"`
}

func main() {
	router := mux.NewRouter()
	origins := handlers.AllowedOrigins([]string{"*"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	headers := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	// Routes
	router.HandleFunc("/customers", createCustomerHandler).Methods("POST")
	router.HandleFunc("/customers", getCustomersHandler).Methods("GET")

	router.HandleFunc("/materials", createMaterialHandler).Methods("POST")
	router.HandleFunc("/materials", getMaterialsHandler).Methods("GET")
	router.HandleFunc("/material_types", getMaterialTypesHandler).Methods("GET")
	router.HandleFunc("/materials/move-to-location", moveMaterialHandler).Methods("PATCH")
	router.HandleFunc("/materials/remove-from-location", removeMaterialHandler).Methods("PATCH")

	router.HandleFunc("/incoming_materials", sendMaterialHandler).Methods("POST")
	router.HandleFunc("/incoming_materials", getIncomingMaterialsHandler).Methods("GET")

	router.HandleFunc("/warehouses", createWarehouseHandler).Methods("POST")
	router.HandleFunc("/warehouses", getWarehouseHandler).Methods("GET")
	router.HandleFunc("/locations", getLocationsHandler).Methods("GET")
	router.HandleFunc("/available_locations", getAvailableLocationsHandler).Methods("GET")

	router.HandleFunc("/reports/transactions", getTransactionsReport).Methods("GET")
	router.HandleFunc("/reports/balance", getBalanceReport).Methods("GET")

	router.HandleFunc("/import_data", importData).Methods("POST")

	// Env loading
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	port := os.Getenv("PORT")

	fmt.Println("Server running on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(origins, methods, headers)(router)))
}

// Controllers
func createCustomerHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()
	var customer CustomerJSON
	json.NewDecoder(r.Body).Decode(&customer)
	err := createCustomer(customer, db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(customer)
}

func getCustomersHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()
	customers, err := fetchCustomers(db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(customers)
}

func getMaterialTypesHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()
	materialTypes, err := fetchMaterialTypes(db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(materialTypes)
}

func sendMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()
	var material IncomingMaterialJSON
	json.NewDecoder(r.Body).Decode(&material)
	err := sendMaterial(material, db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(material)
}

func getIncomingMaterialsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()
	materials, err := getIncomingMaterials(db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(materials)
}

func createMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()

	var material MaterialJSON
	json.NewDecoder(r.Body).Decode(&material)

	ctx := context.TODO()
	materialId, err := createMaterial(ctx, db, material)

	if err != nil {
		errRes := ErrorResponseJSON{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponseJSON{Message: "Material ID created", Data: materialId}
	json.NewEncoder(w).Encode(res)
}

func getMaterialsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()
	materials, err := getMaterials(db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(materials)
}

func moveMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()
	var material MaterialJSON
	json.NewDecoder(r.Body).Decode(&material)

	ctx := context.TODO()
	err := moveMaterial(ctx, db, material)

	if err != nil {
		errRes := ErrorResponseJSON{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponseJSON{Message: "Material Moved", Data: material}
	json.NewEncoder(w).Encode(res)
}

func removeMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()
	var material MaterialToRemoveJSON
	json.NewDecoder(r.Body).Decode(&material)

	ctx := context.TODO()
	err := removeMaterial(ctx, db, material)

	if err != nil {
		errRes := ErrorResponseJSON{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponseJSON{Message: "Material Quantity Removed", Data: material}
	json.NewEncoder(w).Encode(res)
}

func createWarehouseHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()

	var warehouse WarehouseJSON
	json.NewDecoder(r.Body).Decode(&warehouse)
	err := createWarehouse(warehouse, db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(warehouse)
}

func getWarehouseHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()

	warehouses, err := fetchWarehouses(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(warehouses)
}

func getLocationsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()

	locations, err := fetchLocations(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(locations)
}

func getAvailableLocationsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()
	stockId := r.URL.Query().Get("stockId")
	owner := r.URL.Query().Get("owner")

	locations, err := fetchAvailableLocations(db, LocationFilter{stockId: stockId, owner: owner})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(locations)
}

func getTransactionsReport(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()

	customerIdStr := r.URL.Query().Get("customerId")
	customerId, _ := strconv.Atoi(customerIdStr)
	materialType := r.URL.Query().Get("materialType")
	dateFrom := r.URL.Query().Get("dateFrom")
	dateTo := r.URL.Query().Get("dateTo")

	trxRep := TransactionReport{Report: Report{db: db}, trxFilter: SearchQuery{
		customerId:   customerId,
		materialType: materialType,
		dateFrom:     dateFrom,
		dateTo:       dateTo,
	}}
	trxReport, err := trxRep.getReportList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(trxReport)
}

func getBalanceReport(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()

	customerIdStr := r.URL.Query().Get("customerId")
	customerId, _ := strconv.Atoi(customerIdStr)
	materialType := r.URL.Query().Get("materialType")
	dateAsOf := r.URL.Query().Get("dateAsOf")

	balanceRep := BalanceReport{Report: Report{db: db}, blcFilter: SearchQuery{
		customerId:   customerId,
		materialType: materialType,
		dateAsOf:     dateAsOf,
	}}
	balanceReport, err := balanceRep.getReportList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(balanceReport)
}

func importData(w http.ResponseWriter, r *http.Request) {
	db, _ := connectToDB()
	defer db.Close()

	importRes, err := importDataToDB(db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response := SuccessResponseJSON{Message: "Data Imported to the DB", Data: importRes}
	json.NewEncoder(w).Encode(response)
}
