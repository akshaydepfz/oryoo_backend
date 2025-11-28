package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"oryoo.com/helper"
)

func ConnectDatabase() {

	const (
		host     = "ep-dry-feather-a1peu3z3.ap-southeast-1.pg.koyeb.app"
		port     = 5432
		user     = "koyeb-adm"
		password = "npg_PCYTbnK26NxB"
		dbname   = "koyebdb"
	)

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)
	var err error

	helper.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	if err = helper.DB.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	fmt.Println("Database connection established")

}

func CreateClientsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS clients (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			phone TEXT NOT NULL,
			added_by TEXT NOT NULL,
			alternate_phone TEXT,
			avatar TEXT,
			status TEXT NOT NULL DEFAULT 'Active',
			total_orders INTEGER NOT NULL DEFAULT 0,
			total_spent DOUBLE PRECISION NOT NULL DEFAULT 0,
			join_date TIMESTAMP NOT NULL,
			address TEXT NOT NULL,
			tags TEXT[] DEFAULT ARRAY[]::TEXT[],
			last_order_date TIMESTAMP,
			website TEXT,
			notes TEXT,
			company_name TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`

	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating clients table: %v", err)
		return err
	}

	fmt.Println("Clients table created successfully")
	return nil
}
