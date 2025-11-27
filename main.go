package main

import (
	"net/http"

	"oryoo.com/database"
	"oryoo.com/handler"
)

func main() {
	database.ConnectDatabase()
	http.Handle("/users", http.HandlerFunc(handler.UserHandler))

	http.ListenAndServe(":8080", nil)
}
