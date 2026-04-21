package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"oryoo.com/helper"
)

type adminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createAdminRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsActive *bool  `json:"is_active,omitempty"`
}

type adminLoginData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}

type adminLoginSuccessResponse struct {
	Success bool           `json:"success"`
	Data    adminLoginData `json:"data"`
}

type createAdminData struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

type createAdminSuccessResponse struct {
	Success bool            `json:"success"`
	Data    createAdminData `json:"data"`
}

type adminErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func writeAdminError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(adminErrorResponse{
		Success: false,
		Error:   message,
	})
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// AdminLoginHandler handles POST /admin/login
func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAdminError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req adminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		writeAdminError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	var id, name, dbEmail, passwordHash string
	var isActive bool
	err := helper.DB.QueryRow(
		`SELECT id, name, email, password_hash, is_active FROM admins WHERE email = $1`,
		email,
	).Scan(&id, &name, &dbEmail, &passwordHash, &isActive)

	if err != nil {
		if err == sql.ErrNoRows {
			writeAdminError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		writeAdminError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		writeAdminError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !isActive {
		writeAdminError(w, http.StatusForbidden, "Account disabled")
		return
	}

	token, err := generateSessionToken()
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(adminLoginSuccessResponse{
		Success: true,
		Data: adminLoginData{
			ID:    id,
			Name:  name,
			Email: dbEmail,
			Token: token,
		},
	})
}

// CreateAdminHandler handles POST /admin/create-admin
func CreateAdminHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAdminError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req createAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)
	if name == "" || email == "" || password == "" {
		writeAdminError(w, http.StatusBadRequest, "Name, email and password are required")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	var id, createdName, createdEmail string
	var createdIsActive bool
	err = helper.DB.QueryRow(
		`INSERT INTO admins (name, email, password_hash, is_active) VALUES ($1, $2, $3, $4)
		 RETURNING id, name, email, is_active`,
		name, email, string(passwordHash), isActive,
	).Scan(&id, &createdName, &createdEmail, &createdIsActive)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeAdminError(w, http.StatusConflict, "Admin with this email already exists")
			return
		}
		writeAdminError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createAdminSuccessResponse{
		Success: true,
		Data: createAdminData{
			ID:       id,
			Name:     createdName,
			Email:    createdEmail,
			IsActive: createdIsActive,
		},
	})
}

// AdminGetCustomers returns all users (app customers)
func AdminGetCustomers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	customers, err := helper.GetUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

// AdminGetClients returns all clients across the platform
func AdminGetClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clients, err := helper.GetClients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

// AdminGetPayments returns all payments
func AdminGetPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payments, err := helper.GetAllPayments()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

// AdminGetOrders returns all orders
func AdminGetOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orders, err := helper.GetAllOrders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// AdminGetProducts returns all products
func AdminGetProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	products, err := helper.GetAllProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// AdminGetBillingTransactions returns all billing transactions
func AdminGetBillingTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	transactions, err := helper.GetAllBillingTransactions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}
