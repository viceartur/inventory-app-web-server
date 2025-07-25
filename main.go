package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	routeHandlers "inv_app/handlers"
	"inv_app/services/websocket"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	router := mux.NewRouter()
	origins := handlers.AllowedOrigins([]string{"*"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	headers := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	// Auth
	router.HandleFunc("/users/auth", routeHandlers.AuthUsersHandler).Methods("POST")
	router.HandleFunc("/users", routeHandlers.CreateUserHandler).Methods("POST")

	// WebSocket
	router.HandleFunc("/ws/{userRole}", websocket.WsEndpoint)

	// Routes
	router.HandleFunc("/customers", routeHandlers.CreateCustomerHandler).Methods("POST")
	router.HandleFunc("/customers", routeHandlers.GetCustomersHandler).Methods("GET")
	router.HandleFunc("/customers/{customerId}", routeHandlers.GetCustomerHandler).Methods("GET")
	router.HandleFunc("/customers/{customerId}", routeHandlers.UpdateCustomerHandler).Methods("PATCH")

	router.HandleFunc("/customer_programs", routeHandlers.CreateCustomerProgramHandler).Methods("POST")
	router.HandleFunc("/customer_programs", routeHandlers.GetCustomerProgramsHandler).Methods("GET")
	router.HandleFunc("/customer_programs/{programId}", routeHandlers.GetCustomerProgramHandler).Methods("GET")
	router.HandleFunc("/customer_programs/{programId}", routeHandlers.UpdateCustomerProgramHandler).Methods("PATCH")

	router.HandleFunc("/materials", routeHandlers.CreateMaterialHandler).Methods("POST")
	router.HandleFunc("/materials/like", routeHandlers.GetMaterialsLikeHandler).Methods("GET")
	router.HandleFunc("/materials/exact", routeHandlers.GetMaterialsExactHandler).Methods("GET")
	router.HandleFunc("/materials", routeHandlers.UpdateMaterialHandler).Methods("PATCH")
	router.HandleFunc("/material_types", routeHandlers.GetMaterialTypesHandler).Methods("GET")
	router.HandleFunc("/material_usage_reasons", routeHandlers.GetMaterialUsageReasonsHandler).Methods("GET")
	router.HandleFunc("/materials/move-to-location", routeHandlers.MoveMaterialHandler).Methods("PATCH")
	router.HandleFunc("/materials/remove-from-location", routeHandlers.RemoveMaterialHandler).Methods("PATCH")
	router.HandleFunc("/materials/description", routeHandlers.GetMaterialDescriptionHandler).Methods("GET")
	router.HandleFunc("/materials/transactions", routeHandlers.GetMaterialTransactionsHandler).Methods("GET")
	router.HandleFunc("/materials/{stockId}/status", routeHandlers.UpdateMaterialsStatusHandler).Methods("PATCH")
	router.HandleFunc("/materials/grouped-by-stock", routeHandlers.GetMaterialsGroupedByStockIDHandler).Methods("GET")

	router.HandleFunc("/requested_materials", routeHandlers.RequestMaterialsHandler).Methods("POST")
	router.HandleFunc("/requested_materials", routeHandlers.GetRequestedMaterialsHandler).Methods("GET")
	router.HandleFunc("/requested_materials/count", routeHandlers.GetRequestedMaterialsCountHandler).Methods("GET")
	router.HandleFunc("/requested_materials", routeHandlers.UpdateRequestedMaterialHandler).Methods("PATCH")

	router.HandleFunc("/incoming_materials", routeHandlers.SendMaterialHandler).Methods("POST")
	router.HandleFunc("/incoming_materials", routeHandlers.GetIncomingMaterialsHandler).Methods("GET")
	router.HandleFunc("/incoming_materials", routeHandlers.UpdateIncomingMaterialHandler).Methods("PUT")
	router.HandleFunc("/incoming_materials", routeHandlers.DeleteIncomingMaterialHandler).Methods("DELETE")

	router.HandleFunc("/warehouses", routeHandlers.CreateWarehouseHandler).Methods("POST")
	router.HandleFunc("/warehouses", routeHandlers.GetWarehouseHandler).Methods("GET")
	router.HandleFunc("/locations", routeHandlers.GetLocationsHandler).Methods("GET")
	router.HandleFunc("/available_locations", routeHandlers.GetAvailableLocationsHandler).Methods("GET")

	router.HandleFunc("/reports/transactions", routeHandlers.GetTransactionsReport).Methods("GET")
	router.HandleFunc("/reports/balance", routeHandlers.GetBalanceReport).Methods("GET")
	router.HandleFunc("/reports/weekly_usage", routeHandlers.GetWeeklyUsageReport).Methods("GET")
	router.HandleFunc("/reports/transactions_log", routeHandlers.GetTransactionsLogReport).Methods("GET")
	router.HandleFunc("/reports/vault", routeHandlers.GetVaultReport).Methods("GET")
	router.HandleFunc("/reports/customer_usage", routeHandlers.GetCustomerUsageReport).Methods("GET")

	router.HandleFunc("/email_customer_report/{customerId}", routeHandlers.EmailCustomerReportHandler).Methods("POST")
	router.HandleFunc("/email_customer_reports", routeHandlers.EmailCustomerReportsHandler).Methods("POST")

	// Env loading
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	port := os.Getenv("PORT")

	fmt.Println("Server running on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(origins, methods, headers)(router)))
}
