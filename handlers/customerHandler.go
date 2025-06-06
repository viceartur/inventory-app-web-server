package handlers

import (
	"encoding/json"
	"inv_app/database"
	"inv_app/services/customers"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func CreateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	var customer customers.Customer
	json.NewDecoder(r.Body).Decode(&customer)
	createdCustomer, err := customers.CreateCustomer(db, customer)

	if err != nil {
		errRes := ErrorResponseJSON{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponseJSON{Message: "Customer created.", Data: createdCustomer}
	json.NewEncoder(w).Encode(res)
}

func GetCustomersHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	customers, err := customers.FetchCustomers(db)

	if err != nil {
		errRes := ErrorResponseJSON{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

func GetCustomerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerId, _ := strconv.Atoi(vars["customerId"])

	db, _ := database.ConnectToDB()
	defer db.Close()
	customer, err := customers.FetchCustomer(db, customerId)

	if err != nil {
		errRes := ErrorResponseJSON{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

func UpdateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	var customer customers.Customer
	json.NewDecoder(r.Body).Decode(&customer)
	updatedCustomer, err := customers.UpdateCustomer(db, customer)

	if err != nil {
		errRes := ErrorResponseJSON{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponseJSON{Message: "Customer updated.", Data: updatedCustomer}
	json.NewEncoder(w).Encode(res)
}
