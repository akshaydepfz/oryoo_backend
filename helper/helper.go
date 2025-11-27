package helper

import (
	"database/sql"
	"fmt"

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
