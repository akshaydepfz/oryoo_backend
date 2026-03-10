package models

import (
	"time"
)

type User struct {
	ID           uint   `json:"id"`
	FirebaseUID  string `json:"firebase_uid"`
	Phone        string `json:"phone"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	BusinessName string `json:"business_name"`

	BrandImage string `json:"brand_image"`

	Country        string `json:"country"`
	State          string `json:"state"`
	City           string `json:"city"`
	Address        string `json:"address"`
	Pincode        string `json:"pincode"`
	TotalCustomers int    `json:"total_customers"`

	LastLogin  time.Time `json:"last_login"`
	LastActive time.Time `json:"last_active"`

	DeviceID    string `json:"device_id"`
	DeviceModel string `json:"device_model"`
	AppVersion  string `json:"app_version"`

	IsPremium  bool      `json:"is_premium"`
	PlanName   string    `json:"plan_name"`
	PlanExpiry time.Time `json:"plan_expiry"`

	Rating        float32 `json:"rating"`
	AccountStatus string  `json:"account_status"`

	ReferralCode string `json:"referral_code"`
	ReferredBy   string `json:"referred_by"`

	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
}

type CheckPhoneRequest struct {
	Phone string `json:"phone"`
}

type CheckPhoneResponse struct {
	IsAlreadyRegistered bool `json:"is_already_registered"`
}

type ClientModel struct {
	ID             string     `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	Email          string     `json:"email" db:"email"`
	Phone          string     `json:"phone" db:"phone"`
	AddedBy        string     `json:"added_by" db:"added_by"`
	AlternatePhone *string    `json:"alternate_phone" db:"alternate_phone"`
	Avatar         *string    `json:"avatar" db:"avatar"`
	Status         string     `json:"status" db:"status"`
	TotalOrders    int        `json:"total_orders" db:"total_orders"`
	TotalSpent     float64    `json:"total_spent" db:"total_spent"`
	JoinDate       time.Time  `json:"join_date" db:"join_date"`
	Address        string     `json:"address" db:"address"`
	Tags           []string   `json:"tags" db:"tags"` // jsonb[] in PostgreSQL
	LastOrderDate  *time.Time `json:"last_order_date" db:"last_order_date"`
	Website        *string    `json:"website" db:"website"`
	Notes          *string    `json:"notes" db:"notes"`
	CompanyName    *string    `json:"company_name" db:"company_name"`
	CreatedAt      *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at" db:"updated_at"`
}

type OrderItemModel struct {
	ID          string  `json:"id"`
	OrderID     string  `json:"order_id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	CreatedAt   string  `json:"created_at"`
}

type OrderModel struct {
	ID              string           `json:"id"`
	OrderNumber     string           `json:"order_number"`
	ClientID        string           `json:"client_id"`
	ClientName      string           `json:"client_name"`
	ClientAvatar    *string          `json:"client_avatar,omitempty"`
	TotalAmount     float64          `json:"total_amount"`
	Status          string           `json:"status"`
	PaymentStatus   string           `json:"payment_status"`
	Items           []OrderItemModel `json:"items"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       *string          `json:"updated_at,omitempty"`
	DeliveryDate    *string          `json:"delivery_date,omitempty"`
	DeliveryAddress string           `json:"delivery_address"`
	Notes           *string          `json:"notes,omitempty"`
	AddedBy         *string          `json:"added_by,omitempty"`
	CreatedBy       *string          `json:"created_by,omitempty"` // NEW FIELD
}

type CreateOrderRequest struct {
	ClientID        string  `json:"client_id"`
	ClientName      string  `json:"client_name"`
	ClientAvatar    *string `json:"client_avatar"`
	TotalAmount     float64 `json:"total_amount"`
	Status          string  `json:"status"`
	PaymentStatus   string  `json:"payment_status"`
	DeliveryAddress string  `json:"delivery_address"`
	DeliveryDate    *string `json:"delivery_date"`
	Notes           *string `json:"notes"`
	AddedBy         *string `json:"added_by"`
	CreatedBy       *string `json:"created_by"` // NEW FIELD

	Items []OrderItemModel `json:"items"`
}

type PaymentModel struct {
	ID            string     `json:"id" db:"id"`
	ClientID      string     `json:"client_id" db:"client_id"`
	ClientName    string     `json:"client_name" db:"client_name"`
	ClientAvatar  *string    `json:"client_avatar" db:"client_avatar"`
	Amount        float64    `json:"amount" db:"amount"`
	Date          time.Time  `json:"date" db:"date"`
	Status        string     `json:"status" db:"status"`
	OrderID       string     `json:"order_id" db:"order_id"`
	PaymentMethod string     `json:"payment_method" db:"payment_method"`
	PaidAmount    *float64   `json:"paid_amount" db:"paid_amount"`
	Notes         *string    `json:"notes" db:"notes"`
	CreatedBy     *string    `json:"created_by" db:"created_by"`
	AddedBy       *string    `json:"added_by" db:"added_by"` // NEW FIELD
	CreatedAt     *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at" db:"updated_at"`
}

type CreatePaymentRequest struct {
	ClientID      string    `json:"client_id"`
	ClientName    string    `json:"client_name"`
	ClientAvatar  *string   `json:"client_avatar"`
	Amount        float64   `json:"amount"`
	Date          time.Time `json:"date"`
	Status        string    `json:"status"`
	OrderID       string    `json:"order_id"`
	PaymentMethod string    `json:"payment_method"`
	PaidAmount    *float64  `json:"paid_amount"`
	Notes         *string   `json:"notes"`
	CreatedBy     *string   `json:"created_by"`
	AddedBy       *string   `json:"added_by"`
}

type DeleteClientRequest struct {
	ID string `json:"id"`
}

type VersionResponse struct {
	Version     string `json:"version"`
	UpdateURL   string `json:"update_url"`
	IsMandatory bool   `json:"is_mandatory"`
}

type BillingTransaction struct {
	ID                string    `json:"id" db:"id"`
	UserID            string    `json:"user_id" db:"user_id"`
	PlanID            string    `json:"plan_id" db:"plan_id"`
	PlanName          string    `json:"plan_name" db:"plan_name"`
	RazorpayOrderID   string    `json:"razorpay_order_id" db:"razorpay_order_id"`
	RazorpayPaymentID string    `json:"razorpay_payment_id" db:"razorpay_payment_id"`
	RazorpaySignature string    `json:"razorpay_signature" db:"razorpay_signature"`
	Amount            int       `json:"amount" db:"amount"`     // 9900 paise
	Currency          string    `json:"currency" db:"currency"` // INR
	Status            string    `json:"status" db:"status"`     // success
	PlanExpiry        time.Time `json:"plan_expiry" db:"plan_expiry"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

type ProductModel struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	Price       float64    `json:"price" db:"price"`
	SKU         *string    `json:"sku,omitempty" db:"sku"`
	AddedBy     string     `json:"added_by" db:"added_by"`
	CreatedAt   *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
}

type CreateProductRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	SKU         *string  `json:"sku,omitempty"`
	AddedBy     string   `json:"added_by"`
}

type DeleteProductRequest struct {
	ID string `json:"id"`
}

type CreateBillingTransactionRequest struct {
	UserID            string    `json:"user_id"` // firebase_uid
	PlanID            string    `json:"plan_id"`
	PlanName          string    `json:"plan_name"` // e.g., "Premium Monthly", "Premium Yearly"
	RazorpayOrderID   string    `json:"razorpay_order_id"`
	RazorpayPaymentID string    `json:"razorpay_payment_id"`
	RazorpaySignature string    `json:"razorpay_signature"`
	Amount            int       `json:"amount"`      // amount in paise
	Currency          string    `json:"currency"`    // INR
	Status            string    `json:"status"`      // success, failed, pending
	PlanExpiry        time.Time `json:"plan_expiry"` // when the plan expires
}
