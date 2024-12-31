package main

import (
	"database/sql"
	"log"
)

type CustomerJSON struct {
	Name string `json:"customerName"`
	Code string `json:"customerCode"`
}

type CustomerDB struct {
	ID   int    `field:"id"`
	Name string `field:"name"`
	Code string `field:"customer_code"`
}

func createCustomer(customer CustomerJSON, db *sql.DB) error {
	_, err := db.Exec("INSERT INTO customers (name, customer_code) VALUES ($1,$2)",
		customer.Name, customer.Code)

	if err != nil {
		return err
	}
	return nil
}

func fetchCustomers(db *sql.DB) ([]CustomerDB, error) {
	rows, err := db.Query("SELECT * FROM customers ORDER BY name ASC;")
	if err != nil {
		log.Println("Error fetchCustomers1: ", err)
		return nil, err
	}
	defer rows.Close()

	var customers []CustomerDB

	for rows.Next() {
		var customer CustomerDB
		if err := rows.Scan(&customer.ID, &customer.Name, &customer.Code); err != nil {
			log.Println("Error fetchCustomers2: ", err)
			return customers, err
		}
		customers = append(customers, customer)
	}
	if err = rows.Err(); err != nil {
		return customers, err
	}

	return customers, nil
}
