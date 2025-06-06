package customers

import (
	"database/sql"
)

type Customer struct {
	ID        int    `field:"id" json:"customerId"`
	Name      string `field:"name" json:"customerName"`
	Code      string `field:"customer_code" json:"customerCode"`
	AtlasName string `field:"atlas_name" json:"atlasName"`
	IsActive  bool   `field:"is_active" json:"isActive"`
}

func CreateCustomer(db *sql.DB, customer Customer) (Customer, error) {
	var createdCustomer Customer

	err := db.QueryRow(`
		INSERT INTO
			customers (name, customer_code, atlas_name, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING
			customer_id, name, customer_code, atlas_name, is_active;
	`,
		customer.Name,
		customer.Code,
		customer.AtlasName,
		customer.IsActive,
	).Scan(
		&createdCustomer.ID,
		&createdCustomer.Name,
		&createdCustomer.Code,
		&createdCustomer.AtlasName,
		&createdCustomer.IsActive,
	)
	if err != nil {
		return Customer{}, err
	}

	return createdCustomer, nil
}

func FetchCustomers(db *sql.DB) ([]Customer, error) {
	rows, err := db.Query(`
		SELECT
			customer_id,
			name,
			customer_code,
			atlas_name,
			is_active
		FROM
			customers
		ORDER BY
			name;
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []Customer

	for rows.Next() {
		var customer Customer
		var atlasName sql.NullString
		if err := rows.Scan(
			&customer.ID,
			&customer.Name,
			&customer.Code,
			&atlasName,
			&customer.IsActive,
		); err != nil {
			return customers, err
		}

		if atlasName.Valid {
			customer.AtlasName = atlasName.String
		} else {
			customer.AtlasName = ""
		}

		customers = append(customers, customer)
	}
	if err = rows.Err(); err != nil {
		return customers, err
	}

	return customers, nil
}

func FetchCustomer(db *sql.DB, customerId int) (Customer, error) {
	var customer Customer
	var atlasName sql.NullString

	err := db.QueryRow(`
		SELECT
			customer_id,
			name,
			customer_code,
			atlas_name,
			is_active
		FROM
			customers
		WHERE
			customer_id = $1;
	`, customerId).Scan(
		&customer.ID,
		&customer.Name,
		&customer.Code,
		&atlasName,
		&customer.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Customer{}, nil
		}
		return Customer{}, err
	}

	if atlasName.Valid {
		customer.AtlasName = atlasName.String
	} else {
		customer.AtlasName = ""
	}

	return customer, nil
}

func UpdateCustomer(db *sql.DB, customer Customer) (Customer, error) {
	row := db.QueryRow(`
		UPDATE customers
		SET
			name = $1,
			customer_code = $2,
			atlas_name = $3,
			is_active = $4
		WHERE
			customer_id = $5
		RETURNING customer_id, name, customer_code, COALESCE(atlas_name, 'None'), is_active;
	`,
		customer.Name,
		customer.Code,
		customer.AtlasName,
		customer.IsActive,
		customer.ID,
	)

	var updatedCustomer Customer
	err := row.Scan(
		&updatedCustomer.ID,
		&updatedCustomer.Name,
		&updatedCustomer.Code,
		&updatedCustomer.AtlasName,
		&updatedCustomer.IsActive,
	)
	if err != nil {
		return Customer{}, err
	}

	return updatedCustomer, nil
}
