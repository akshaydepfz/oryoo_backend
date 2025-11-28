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
