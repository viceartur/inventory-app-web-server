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
	api := router.PathPrefix("/api/v1").Subrouter()

	origins := handlers.AllowedOrigins([]string{"*"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	headers := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	// Auth
	api.HandleFunc("/users/auth", routeHandlers.AuthUsersHandler).Methods("POST")
	api.HandleFunc("/users", routeHandlers.CreateUserHandler).Methods("POST")

	// WebSocket
	api.HandleFunc("/ws/{userRole}", websocket.WsEndpoint)

	// Routes
	api.HandleFunc("/customers", routeHandlers.CreateCustomerHandler).Methods("POST")
	api.HandleFunc("/customers", routeHandlers.GetCustomersHandler).Methods("GET")
	api.HandleFunc("/customers/{customerId}", routeHandlers.GetCustomerHandler).Methods("GET")
	api.HandleFunc("/customers/{customerId}", routeHandlers.UpdateCustomerHandler).Methods("PATCH")

	api.HandleFunc("/customer_programs", routeHandlers.CreateCustomerProgramHandler).Methods("POST")
	api.HandleFunc("/customer_programs", routeHandlers.GetCustomerProgramsHandler).Methods("GET")
	api.HandleFunc("/customer_programs/{programId}", routeHandlers.GetCustomerProgramHandler).Methods("GET")
	api.HandleFunc("/customer_programs/{programId}", routeHandlers.UpdateCustomerProgramHandler).Methods("PATCH")

	api.HandleFunc("/materials", routeHandlers.CreateMaterialHandler).Methods("POST")
	api.HandleFunc("/materials/like", routeHandlers.GetMaterialsLikeHandler).Methods("GET")
	api.HandleFunc("/materials/exact", routeHandlers.GetMaterialsExactHandler).Methods("GET")
	api.HandleFunc("/materials", routeHandlers.UpdateMaterialHandler).Methods("PATCH")
	api.HandleFunc("/material_types", routeHandlers.GetMaterialTypesHandler).Methods("GET")
	api.HandleFunc("/material_usage_reasons", routeHandlers.GetMaterialUsageReasonsHandler).Methods("GET")
	api.HandleFunc("/materials/move-to-location", routeHandlers.MoveMaterialHandler).Methods("PATCH")
	api.HandleFunc("/materials/remove-from-location", routeHandlers.RemoveMaterialHandler).Methods("PATCH")
	api.HandleFunc("/materials/description", routeHandlers.GetMaterialDescriptionHandler).Methods("GET")
	api.HandleFunc("/materials/transactions", routeHandlers.GetMaterialTransactionsHandler).Methods("GET")
	api.HandleFunc("/materials/{stockId}/status", routeHandlers.UpdateMaterialsStatusHandler).Methods("PATCH")
	api.HandleFunc("/materials/grouped-by-stock", routeHandlers.GetMaterialsGroupedByStockIDHandler).Methods("GET")

	api.HandleFunc("/requested_materials", routeHandlers.RequestMaterialsHandler).Methods("POST")
	api.HandleFunc("/requested_materials", routeHandlers.GetRequestedMaterialsHandler).Methods("GET")
	api.HandleFunc("/requested_materials/count", routeHandlers.GetRequestedMaterialsCountHandler).Methods("GET")
	api.HandleFunc("/requested_materials", routeHandlers.UpdateRequestedMaterialHandler).Methods("PATCH")

	api.HandleFunc("/incoming_materials", routeHandlers.SendMaterialHandler).Methods("POST")
	api.HandleFunc("/incoming_materials", routeHandlers.GetIncomingMaterialsHandler).Methods("GET")
	api.HandleFunc("/incoming_materials", routeHandlers.UpdateIncomingMaterialHandler).Methods("PUT")
	api.HandleFunc("/incoming_materials", routeHandlers.DeleteIncomingMaterialHandler).Methods("DELETE")

	api.HandleFunc("/warehouses", routeHandlers.CreateWarehouseHandler).Methods("POST")
	api.HandleFunc("/warehouses", routeHandlers.GetWarehouseHandler).Methods("GET")
	api.HandleFunc("/locations", routeHandlers.GetLocationsHandler).Methods("GET")
	api.HandleFunc("/available_locations", routeHandlers.GetAvailableLocationsHandler).Methods("GET")

	api.HandleFunc("/reports/transactions", routeHandlers.GetTransactionsReport).Methods("GET")
	api.HandleFunc("/reports/balance", routeHandlers.GetBalanceReport).Methods("GET")
	api.HandleFunc("/reports/weekly_usage", routeHandlers.GetWeeklyUsageReport).Methods("GET")
	api.HandleFunc("/reports/transactions_log", routeHandlers.GetTransactionsLogReport).Methods("GET")
	api.HandleFunc("/reports/vault", routeHandlers.GetVaultReport).Methods("GET")
	api.HandleFunc("/reports/customer_usage", routeHandlers.GetCustomerUsageReport).Methods("GET")

	api.HandleFunc("/email_customer_report/{customerId}", routeHandlers.EmailCustomerReportHandler).Methods("POST")
	api.HandleFunc("/email_customer_reports", routeHandlers.EmailCustomerReportsHandler).Methods("POST")

	// Env loading
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	port := os.Getenv("PORT")

	fmt.Println("Server running on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(origins, methods, headers)(router)))
}
