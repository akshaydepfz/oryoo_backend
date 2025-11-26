package database

import (
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
