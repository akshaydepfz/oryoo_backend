package models

import "time"

type User struct {
	ID            uint      `json:"id"`
	FirebaseUID   string    `json:"firebase_uid"`
	Phone         string    `json:"phone"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	BusinessName  string    `json:"business_name"`
	Country       string    `json:"country"`
	State         string    `json:"state"`
	City          string    `json:"city"`
	Address       string    `json:"address"`
	Pincode       string    `json:"pincode"`
	LastLogin     time.Time `json:"last_login"`
	LastActive    time.Time `json:"last_active"`
	DeviceID      string    `json:"device_id"`
	DeviceModel   string    `json:"device_model"`
	AppVersion    string    `json:"app_version"`
	IsPremium     bool      `json:"is_premium"`
	PlanName      string    `json:"plan_name"`
	PlanExpiry    time.Time `json:"plan_expiry"`
	Rating        float32   `json:"rating"`
	AccountStatus string    `json:"account_status"`
	ReferralCode  string    `json:"referral_code"`
	ReferredBy    string    `json:"referred_by"`
	CreatedDate   time.Time `json:"created_date"`
	UpdatedDate   time.Time `json:"updated_date"`
}

type CheckPhoneRequest struct {
	Phone string `json:"phone"`
}

type CheckPhoneResponse struct {
	IsAlreadyRegistered bool `json:"is_already_registered"`
}
