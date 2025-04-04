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
	router.HandleFunc("/ws", websocket.WsEndpoint)

	// Routes
	router.HandleFunc("/customers", routeHandlers.CreateCustomerHandler).Methods("POST")
	router.HandleFunc("/customers", routeHandlers.GetCustomersHandler).Methods("GET")

	router.HandleFunc("/materials", routeHandlers.CreateMaterialHandler).Methods("POST")
	router.HandleFunc("/materials", routeHandlers.GetMaterialsHandler).Methods("GET")
	router.HandleFunc("/materials", routeHandlers.UpdateMaterialHandler).Methods("PATCH")
	router.HandleFunc("/material_types", routeHandlers.GetMaterialTypesHandler).Methods("GET")
	router.HandleFunc("/materials/move-to-location", routeHandlers.MoveMaterialHandler).Methods("PATCH")
	router.HandleFunc("/materials/remove-from-location", routeHandlers.RemoveMaterialHandler).Methods("PATCH")
	router.HandleFunc("/materials/description", routeHandlers.GetMaterialDescriptionHandler).Methods("GET")

	router.HandleFunc("/requested_materials", routeHandlers.RequestMaterialsHandler).Methods("POST")
	router.HandleFunc("/requested_materials", routeHandlers.GetRequestedMaterialsHandler).Methods("GET")
	router.HandleFunc("/requested_materials", routeHandlers.UpdateRequestedMaterialHandler).Methods("PATCH")

	router.HandleFunc("/incoming_materials", routeHandlers.SendMaterialHandler).Methods("POST")
	router.HandleFunc("/incoming_materials", routeHandlers.GetIncomingMaterialsHandler).Methods("GET")
	router.HandleFunc("/incoming_materials", routeHandlers.UpdateIncomingMaterialHandler).Methods("PUT")

	router.HandleFunc("/warehouses", routeHandlers.CreateWarehouseHandler).Methods("POST")
	router.HandleFunc("/warehouses", routeHandlers.GetWarehouseHandler).Methods("GET")
	router.HandleFunc("/locations", routeHandlers.GetLocationsHandler).Methods("GET")
	router.HandleFunc("/available_locations", routeHandlers.GetAvailableLocationsHandler).Methods("GET")

	router.HandleFunc("/reports/transactions", routeHandlers.GetTransactionsReport).Methods("GET")
	router.HandleFunc("/reports/balance", routeHandlers.GetBalanceReport).Methods("GET")

	router.HandleFunc("/import_data", routeHandlers.ImportData).Methods("POST")

	// Env loading
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	port := os.Getenv("PORT")

	fmt.Println("Server running on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(origins, methods, headers)(router)))
}
