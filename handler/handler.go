package handler

import (
	"encoding/json"
	"net/http"

	"oryoo.com/helper"
	"oryoo.com/models"
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

func CheckPhoneHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CheckPhoneRequest

	// Decode JSON
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Phone == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check in database
	exists, err := helper.CheckPhoneExists(req.Phone)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Create response
	resp := models.CheckPhoneResponse{
		IsAlreadyRegistered: exists,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
