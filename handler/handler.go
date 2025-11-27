package handler

import (
	"encoding/json"
	"net/http"

	"oryoo.com/helper"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

	} else if r.Method == http.MethodGet {

	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}

}

func GetCustomers(w http.ResponseWriter, r *http.Request) {

	customers, err := helper.GetUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)

}
