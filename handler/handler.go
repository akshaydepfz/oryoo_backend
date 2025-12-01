package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("brand_image")
	if err != nil {
		http.Error(w, "Brand image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}

	contentType := http.DetectContentType(fileBytes)
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "Only image files are allowed", http.StatusBadRequest)
		return
	}

	const (
		s3Bucket  = "oryoo-bucket"
		s3Region  = "ap-southeast-2"
		awsKey    = "AKIAYLWS7S6WYP6WYSIA"
		awsSecret = "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"
	)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(s3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsKey, awsSecret, "")),
	)
	if err != nil {
		http.Error(w, "AWS config error", http.StatusInternalServerError)
		return
	}

	client := s3.NewFromConfig(cfg)

	key := fmt.Sprintf("UserBrandImages/%d_%s", time.Now().UnixNano(), fileHeader.Filename)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		http.Error(w, "Image upload failed", http.StatusInternalServerError)
		return
	}

	imageURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3Bucket, s3Region, key)

	user := models.User{
		FirebaseUID:   r.FormValue("firebase_uid"),
		Phone:         r.FormValue("phone"),
		Name:          r.FormValue("name"),
		Email:         r.FormValue("email"),
		BusinessName:  r.FormValue("business_name"),
		BrandImage:    imageURL,
		Country:       r.FormValue("country"),
		State:         r.FormValue("state"),
		City:          r.FormValue("city"),
		Address:       r.FormValue("address"),
		Pincode:       r.FormValue("pincode"),
		LastLogin:     time.Now(),
		LastActive:    time.Now(),
		DeviceID:      r.FormValue("device_id"),
		DeviceModel:   r.FormValue("device_model"),
		AppVersion:    r.FormValue("app_version"),
		IsPremium:     false,
		PlanName:      r.FormValue("plan_name"),
		PlanExpiry:    time.Now(),
		Rating:        5,
		AccountStatus: "active",
		ReferralCode:  r.FormValue("referral_code"),
		ReferredBy:    r.FormValue("referred_by"),
		CreatedDate:   time.Now(),
		UpdatedDate:   time.Now(),
	}

	err = helper.InsertUser(user)
	if err != nil {
		http.Error(w, "Failed to save user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User created successfully",
		"user":    user,
	})

}

func GetUserByFirebaseUID(w http.ResponseWriter, r *http.Request) {
	firebaseUID := r.URL.Query().Get("firebase_uid")
	user, err := helper.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func ClientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		CreateClient(w, r)

	} else if r.Method == http.MethodGet {
		GetClients(w, r)

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}

}

func CreateClient(w http.ResponseWriter, r *http.Request) {
	var c models.ClientModel

	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	now := time.Now()
	c.CreatedAt = &now
	c.UpdatedAt = &now
	err := helper.InsertClient(&c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func GetClients(w http.ResponseWriter, r *http.Request) {
	clients, err := helper.GetClients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}
func GetClientByCreatedBy(w http.ResponseWriter, r *http.Request) {
	createdBy := r.URL.Query().Get("created_by")
	clients, err := helper.GetClientsByCreatedBy(createdBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest

	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Generate IDs
	orderID := uuid.New().String()
	orderNumber := "ORD-" + time.Now().Format("20060102-150405")

	// Insert order
	err := helper.InsertOrder(orderID, orderNumber, req)
	if err != nil {
		http.Error(w, "Failed to insert order", http.StatusInternalServerError)
		return
	}

	// Insert items
	for _, item := range req.Items {
		itemID := uuid.New().String()
		err := helper.InsertOrderItem(orderID, itemID, item)
		if err != nil {
			http.Error(w, "Failed to insert order items", http.StatusInternalServerError)
			return
		}
	}

	// Response
	response := map[string]interface{}{
		"success":      true,
		"message":      "Order created successfully",
		"order_id":     orderID,
		"order_number": orderNumber,
		"created_by":   req.CreatedBy, // NEW
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetOrdersByCreatedBy(w http.ResponseWriter, r *http.Request) {
	createdBy := r.URL.Query().Get("created_by")

	if createdBy == "" {
		http.Error(w, "created_by is required", http.StatusBadRequest)
		return
	}

	orders, err := helper.FetchOrdersByCreatedBy(createdBy)
	if err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePaymentRequest

	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Generate ID
	paymentID := uuid.New().String()

	// Insert payment
	err := helper.InsertPayment(paymentID, req)
	if err != nil {
		http.Error(w, "Failed to insert payment", http.StatusInternalServerError)
		return
	}

	// Response
	response := map[string]interface{}{
		"success":    true,
		"message":    "Payment created successfully",
		"payment_id": paymentID,
		"created_by": req.CreatedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetPaymentsByCreatedByHandler(w http.ResponseWriter, r *http.Request) {
	createdBy := r.URL.Query().Get("created_by")

	if createdBy == "" {
		http.Error(w, "created_by is required", http.StatusBadRequest)
		return
	}

	payments, err := helper.GetPaymentsByCreatedBy(createdBy)
	if err != nil {
		http.Error(w, "Failed to fetch payments", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"count":    len(payments),
		"payments": payments,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
