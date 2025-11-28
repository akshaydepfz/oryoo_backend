package helper

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"oryoo.com/models"
)

var DB *sql.DB

func GetUsers() ([]models.User, error) {
	rows, err := DB.Query(`
		SELECT 
			id, firebase_uid, phone, name, email, business_name,
			country, state, city, address, pincode,
			last_login, last_active,
			device_id, device_model, app_version,
			is_premium, plan_name, plan_expiry,
			rating, account_status,
			referral_code, referred_by,
			created_date, updated_date
		FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usersList []models.User

	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.FirebaseUID,
			&user.Phone,
			&user.Name,
			&user.Email,
			&user.BusinessName,
			&user.Country,
			&user.State,
			&user.City,
			&user.Address,
			&user.Pincode,
			&user.LastLogin,
			&user.LastActive,
			&user.DeviceID,
			&user.DeviceModel,
			&user.AppVersion,
			&user.IsPremium,
			&user.PlanName,
			&user.PlanExpiry,
			&user.Rating,
			&user.AccountStatus,
			&user.ReferralCode,
			&user.ReferredBy,
			&user.CreatedDate,
			&user.UpdatedDate,
		)
		if err != nil {
			return nil, err
		}
		usersList = append(usersList, user)
	}

	fmt.Println("Get Users Successful")
	return usersList, nil
}

func CheckPhoneExists(phone string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS (
			SELECT 1 FROM users WHERE phone = $1
		)
	`

	err := DB.QueryRow(query, phone).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// helper/insert_user.go

func InsertUser(user models.User) error {
	query := `
		INSERT INTO users (
			firebase_uid, phone, name, email, business_name,
			brand_image,
			country, state, city, address, pincode,
			last_login, last_active,
			device_id, device_model, app_version,
			is_premium, plan_name, plan_expiry,
			rating, account_status,
			referral_code, referred_by,
			created_date, updated_date
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6,
			$7, $8, $9, $10, $11,
			$12, $13,
			$14, $15, $16,
			$17, $18, $19,
			$20, $21,
			$22, $23,
			$24, $25
		)
	`

	_, err := DB.Exec(
		query,
		user.FirebaseUID, user.Phone, user.Name, user.Email, user.BusinessName,
		user.BrandImage,
		user.Country, user.State, user.City, user.Address, user.Pincode,
		user.LastLogin, user.LastActive,
		user.DeviceID, user.DeviceModel, user.AppVersion,
		user.IsPremium, user.PlanName, user.PlanExpiry,
		user.Rating, user.AccountStatus,
		user.ReferralCode, user.ReferredBy,
		user.CreatedDate, user.UpdatedDate,
	)

	return err
}

func InsertClient(client *models.ClientModel) error {
	query := `
		INSERT INTO clients (
			name, email, phone, added_by,
			alternate_phone, avatar, status,
			total_orders, total_spent, join_date,
			address, tags,
			last_order_date, website, notes, company_name,
			created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9, $10,
			$11, $12,
			$13, $14, $15, $16,
			$17, $18
		)
		RETURNING id
	`

	err := DB.QueryRow(
		query,
		client.Name,
		client.Email,
		client.Phone,
		client.AddedBy,
		client.AlternatePhone,
		client.Avatar,
		client.Status,
		client.TotalOrders,
		client.TotalSpent,
		client.JoinDate,
		client.Address,
		pq.Array(client.Tags),
		client.LastOrderDate,
		client.Website,
		client.Notes,
		client.CompanyName,
		client.CreatedAt,
		client.UpdatedAt,
	).Scan(&client.ID)

	return err
}

func GetClients() ([]models.ClientModel, error) {
	rows, err := DB.Query(`
		SELECT * FROM clients
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clientsList []models.ClientModel
	for rows.Next() {
		var client models.ClientModel
		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.Email,
			&client.Phone,
			&client.AddedBy,
			&client.AlternatePhone,
			&client.Avatar,
			&client.Status,
			&client.TotalOrders,
			&client.TotalSpent,
			&client.JoinDate,
			&client.Address,
			pq.Array(&client.Tags),
			&client.LastOrderDate,
			&client.Website,
			&client.Notes,
			&client.CompanyName,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		clientsList = append(clientsList, client)
	}

	fmt.Println("Get Clients Successful")
	return clientsList, nil
}

func GetClientsByCreatedBy(createdBy string) ([]models.ClientModel, error) {
	rows, err := DB.Query(`
		SELECT * FROM clients WHERE added_by = $1
	`, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clientsList []models.ClientModel
	for rows.Next() {
		var client models.ClientModel
		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.Email,
			&client.Phone,
			&client.AddedBy,
			&client.AlternatePhone,
			&client.Avatar,
			&client.Status,
			&client.TotalOrders,
			&client.TotalSpent,
			&client.JoinDate,
			&client.Address,
			pq.Array(&client.Tags),
			&client.LastOrderDate,
			&client.Website,
			&client.Notes,
			&client.CompanyName,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		clientsList = append(clientsList, client)
	}

	fmt.Println("Get Clients by Created By Successful")
	return clientsList, nil
}
