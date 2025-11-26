package main

import (
	"net/http"

	"oryoo.com/database"
)

func main() {
	database.ConnectDatabase()

	http.ListenAndServe(":8080", nil)
}
