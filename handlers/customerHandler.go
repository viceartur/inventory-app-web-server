package handlers

import (
	"encoding/json"
	"inv_app/database"
	"inv_app/services/customers"
	"net/http"
)

func CreateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	var customer customers.CustomerJSON
	json.NewDecoder(r.Body).Decode(&customer)
	err := customers.CreateCustomer(customer, db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(customer)
}

func GetCustomersHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	customers, err := customers.FetchCustomers(db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(customers)
}
