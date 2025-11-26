package helper

import (
	"database/sql"
)

var DB *sql.DB

func InsertCustomer(name, email string) (uint, error) {
	var id uint
	query := `INSERT INTO customers (name, email, created_date) 
              VALUES ($1, $2, NOW()) RETURNING id`
	err := DB.QueryRow(query, name, email).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
