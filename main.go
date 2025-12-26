package main

import (
	"net/http"

	"oryoo.com/database"
	"oryoo.com/handler"
)

func main() {
	database.ConnectDatabase()
	http.Handle("/users", http.HandlerFunc(handler.UserHandler))
	http.HandleFunc("/auth/check-phone", handler.CheckPhoneHandler)
	http.HandleFunc("/auth/register", handler.CreateUserHandler)
	http.HandleFunc("/clients", handler.ClientHandler)
	http.HandleFunc("/clients/created-by", handler.GetClientByCreatedBy)
	http.HandleFunc("/users/firebase-uid", handler.GetUserByFirebaseUID)
	http.HandleFunc("/orders/create", handler.CreateOrder)
	http.HandleFunc("/orders/update", handler.UpdateOrder)
	http.HandleFunc("/orders/by-created", handler.GetOrdersByCreatedBy)
	http.HandleFunc("/payments/create", handler.CreatePayment)
	http.HandleFunc("/payments/by-created", handler.GetPaymentsByCreatedByHandler)
	http.HandleFunc("/billing/transaction", handler.CreateBillingTransactionHandler)
	http.HandleFunc("/app/latest-version", handler.GetLatestVersionHandler)
	http.ListenAndServe(":8080", nil)
}
