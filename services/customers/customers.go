package customers

import (
	"context"
	"database/sql"
	"time"
)

type Customer struct {
	CustomerID               int       `field:"customer_id" json:"customerId"`
	CustomerName             string    `field:"customer_name" json:"customerName"`
	Emails                   []string  `json:"emails"`
	UserID                   int       `field:"user_id" json:"userId"`
	Username                 string    `field:"username" json:"username"`
	IsConnectedToReports     bool      `field:"is_connected_to_reports" json:"isConnectedToReports"`
	LastReportSentAt         time.Time `field:"last_report_sent_at" json:"lastReportSentAt"`
	LastReportDeliveryStatus string    `field:"last_report_delivery_status" json:"lastReportDeliveryStatus"`
}

type CustomerProgram struct {
	ProgramID    int    `field:"program_id" json:"programId"`
	ProgramName  string `field:"program_name" json:"programName"`
	ProgramCode  string `field:"program_code" json:"programCode"`
	IsActive     bool   `field:"is_active" json:"isActive"`
	CustomerID   int    `field:"customer_id" json:"customerId"`
	CustomerName string `field:"customer_name" json:"customerName"`
}

/* Customers CRUD */

func CreateCustomer(db *sql.DB, customer Customer) (Customer, error) {
	tx, err := db.BeginTx(context.TODO(), nil)
	if err != nil {
		return Customer{}, err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	var customerID int
	err = tx.QueryRow(`
		INSERT INTO
			customers (customer_name, user_id, is_connected_to_reports)
		VALUES
			($1, $2, $3)
		RETURNING
			customer_id;
	`, customer.CustomerName,
		customer.UserID,
		customer.IsConnectedToReports,
	).Scan(&customerID)
	if err != nil {
		tx.Rollback()
		return Customer{}, err
	}

	// Remove old emails if any
	_, err = tx.Exec(
		`DELETE FROM customer_emails WHERE customer_id = $1`,
		customerID,
	)
	if err != nil {
		tx.Rollback()
		return Customer{}, err
	}

	for _, email := range customer.Emails {
		_, err := tx.Exec(`
			INSERT INTO
				customer_emails (customer_id, email)
			VALUES
				($1, $2)
		`, customerID, email)
		if err != nil {
			tx.Rollback()
			return Customer{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Customer{}, err
	}

	customer.CustomerID = customerID
	return customer, nil
}

func GetCustomer(db *sql.DB, customerId int) (Customer, error) {
	var c Customer
	var username sql.NullString
	var lastReportSentAt sql.NullTime
	var lastReportDeliveryStatus sql.NullString
	err := db.QueryRow(`
		SELECT
			c.customer_id,
			c.customer_name,
			c.user_id,
			u.username,
			c.is_connected_to_reports,
			c.last_report_sent_at,
			c.last_report_delivery_status
		FROM customers c
		LEFT JOIN users u ON u.user_id = c.user_id
		WHERE c.customer_id = $1
	`, customerId).Scan(
		&c.CustomerID,
		&c.CustomerName,
		&c.UserID,
		&username,
		&c.IsConnectedToReports,
		&lastReportSentAt,
		&lastReportDeliveryStatus,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Customer{}, nil
		}
		return Customer{}, err
	}

	// Check for NULL values
	if username.Valid {
		c.Username = username.String
	}
	if lastReportSentAt.Valid {
		c.LastReportSentAt = lastReportSentAt.Time
	}
	if lastReportDeliveryStatus.Valid {
		c.LastReportDeliveryStatus = lastReportDeliveryStatus.String
	}

	rows, err := db.Query(`SELECT email FROM customer_emails WHERE customer_id = $1`, customerId)
	if err != nil {
		return Customer{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return Customer{}, err
		}
		c.Emails = append(c.Emails, email)
	}

	return c, nil
}

func GetCustomers(db *sql.DB) ([]Customer, error) {
	rows, err := db.Query(`
		SELECT
			c.customer_id,
			c.customer_name,
			c.user_id,
			u.username,
			c.is_connected_to_reports,
			c.last_report_sent_at,
			c.last_report_delivery_status
		FROM customers c
		LEFT JOIN users u ON u.user_id = c.user_id
		ORDER BY c.customer_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var customers []Customer
	for rows.Next() {
		var c Customer
		var username sql.NullString
		var lastReportSentAt sql.NullTime
		var lastReportDeliveryStatus sql.NullString
		if err := rows.Scan(
			&c.CustomerID,
			&c.CustomerName,
			&c.UserID,
			&username,
			&c.IsConnectedToReports,
			&lastReportSentAt,
			&lastReportDeliveryStatus,
		); err != nil {
			return customers, err
		}

		// Check for NULL values
		if username.Valid {
			c.Username = username.String
		}
		if lastReportSentAt.Valid {
			c.LastReportSentAt = lastReportSentAt.Time
		}
		if lastReportDeliveryStatus.Valid {
			c.LastReportDeliveryStatus = lastReportDeliveryStatus.String
		}

		// Fetch emails
		emailRows, err := db.Query(
			`SELECT email FROM customer_emails WHERE customer_id = $1`,
			c.CustomerID)
		if err == nil {
			for emailRows.Next() {
				var email string
				if err := emailRows.Scan(&email); err == nil {
					c.Emails = append(c.Emails, email)
				}
			}
			emailRows.Close()
		}
		customers = append(customers, c)
	}
	return customers, nil
}

func UpdateCustomer(db *sql.DB, customer Customer) (Customer, error) {
	tx, err := db.BeginTx(context.TODO(), nil)
	if err != nil {
		return Customer{}, err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	_, err = tx.Exec(`
		UPDATE
			customers
		SET
			customer_name = $1,
			user_id = $2,
			is_connected_to_reports = $4
		WHERE
			customer_id = $3
	`,
		customer.CustomerName,
		customer.UserID,
		customer.CustomerID,
		customer.IsConnectedToReports,
	)
	if err != nil {
		tx.Rollback()
		return Customer{}, err
	}

	_, err = tx.Exec(`DELETE FROM customer_emails WHERE customer_id = $1`, customer.CustomerID)
	if err != nil {
		tx.Rollback()
		return Customer{}, err
	}

	for _, email := range customer.Emails {
		_, err := tx.Exec(`
			INSERT INTO customer_emails (customer_id, email)
			VALUES ($1, $2)
			ON CONFLICT (email) DO UPDATE SET customer_id = EXCLUDED.customer_id;
		`, customer.CustomerID, email)
		if err != nil {
			tx.Rollback()
			return Customer{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Customer{}, err
	}
	return customer, nil
}

/* Customer Programs CRUD */

func CreateCustomerProgram(db *sql.DB, cp CustomerProgram) (CustomerProgram, error) {
	tx, err := db.BeginTx(context.TODO(), nil)
	if err != nil {
		return CustomerProgram{}, err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	err = tx.QueryRow(`
		INSERT INTO
			customer_programs (program_name, program_code, customer_id, is_active)
		VALUES
			($1, $2, $3, $4)
		RETURNING
			program_id,
			program_name,
			program_code,
			is_active,
			customer_id;
	`, cp.ProgramName, cp.ProgramCode, cp.CustomerID, cp.IsActive).
		Scan(&cp.ProgramID, &cp.ProgramName, &cp.ProgramCode, &cp.IsActive, &cp.CustomerID)
	if err != nil {
		tx.Rollback()
		return CustomerProgram{}, err
	}

	if err := tx.Commit(); err != nil {
		return CustomerProgram{}, err
	}

	return cp, nil
}

func GetCustomerProgram(db *sql.DB, programID int) (CustomerProgram, error) {
	var cp CustomerProgram
	err := db.QueryRow(`
		SELECT
			cp.program_id,
			cp.program_name,
			cp.program_code,
			cp.is_active,
			COALESCE(cp.customer_id, 0),
			COALESCE(c.customer_name, '')
		FROM
			customer_programs cp
		LEFT JOIN customers c ON c.customer_id = cp.customer_id
		WHERE
			cp.program_id = $1;
	`, programID).Scan(
		&cp.ProgramID,
		&cp.ProgramName,
		&cp.ProgramCode,
		&cp.IsActive,
		&cp.CustomerID,
		&cp.CustomerName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return CustomerProgram{}, nil
		}
		return CustomerProgram{}, err
	}
	return cp, nil
}

func GetCustomerPrograms(db *sql.DB) ([]CustomerProgram, error) {
	rows, err := db.Query(`
		SELECT
			cp.program_id,
			cp.program_name,
			cp.program_code,
			cp.is_active,
			COALESCE(cp.customer_id, 0),
			COALESCE(c.customer_name, '')
		FROM
			customer_programs cp
		LEFT JOIN customers c ON c.customer_id = cp.customer_id
		ORDER BY c.customer_name, cp.program_name;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var programs []CustomerProgram
	for rows.Next() {
		var cp CustomerProgram
		if err := rows.Scan(
			&cp.ProgramID,
			&cp.ProgramName,
			&cp.ProgramCode,
			&cp.IsActive,
			&cp.CustomerID,
			&cp.CustomerName,
		); err != nil {
			return programs, err
		}
		programs = append(programs, cp)
	}

	return programs, nil
}

func UpdateCustomerProgram(db *sql.DB, cp CustomerProgram) (CustomerProgram, error) {
	err := db.QueryRow(`
		UPDATE
			customer_programs
		SET
			program_name = $1,
			program_code = $2,
			is_active = $3,
			customer_id = $4
		WHERE
			program_id = $5
		RETURNING
			program_id, 
			program_name,
			program_code, 
			is_active, 
			customer_id
	`, cp.ProgramName, cp.ProgramCode, cp.IsActive, cp.CustomerID, cp.ProgramID).
		Scan(&cp.ProgramID, &cp.ProgramName, &cp.ProgramCode, &cp.IsActive, &cp.CustomerID)
	if err != nil {
		return CustomerProgram{}, err
	}
	return cp, nil
}
