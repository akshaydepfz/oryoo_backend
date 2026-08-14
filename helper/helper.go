package helper

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"oryoo.com/models"
)

var DB *sql.DB

func scanUserRow(scanner interface{ Scan(dest ...any) error }, user *models.User) error {
	var fcm sql.NullString
	var lastOpen sql.NullTime
	var purchasedAt sql.NullTime
	err := scanner.Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.Phone,
		&user.Name,
		&user.Email,
		&user.BusinessName,
		&user.BrandImage,
		&user.Country,
		&user.State,
		&user.City,
		&user.Address,
		&user.Pincode,
		&user.TotalCustomers,
		&user.ProductCount,
		&user.LastLogin,
		&user.LastActive,
		&user.DeviceID,
		&user.DeviceModel,
		&user.AppVersion,
		&fcm,
		&lastOpen,
		&user.IsPremium,
		&user.PlanName,
		&user.PlanExpiry,
		&user.Rating,
		&user.AccountStatus,
		&user.ReferralCode,
		&user.ReferredBy,
		&user.CreatedDate,
		&user.UpdatedDate,
		&purchasedAt,
	)
	if err != nil {
		return err
	}
	if fcm.Valid {
		s := fcm.String
		user.FCMToken = &s
	}
	if lastOpen.Valid {
		t := lastOpen.Time
		user.LastOpen = &t
	}
	if purchasedAt.Valid {
		t := purchasedAt.Time
		user.PurchasedAt = &t
	}
	return nil
}

func GetUsers() ([]models.User, error) {
	rows, err := DB.Query(`
		SELECT 
			id, firebase_uid, phone, name, email, business_name,
			brand_image,
			country, state, city, address, pincode, total_customers, product_count,
			last_login, last_active,
			device_id, device_model, app_version,
			fcm_token, last_open,
			is_premium, plan_name, plan_expiry,
			rating, account_status,
			referral_code, referred_by,
			created_date, updated_date,
			purchased_at
		FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usersList []models.User

	for rows.Next() {
		var user models.User
		if err := scanUserRow(rows, &user); err != nil {
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
			created_date, updated_date,
			fcm_token, last_open
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
			$24, $25,
			$26, $27
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
		user.FCMToken, user.LastOpen,
	)

	return err
}

func UpdateUserByFirebaseUID(user models.User) error {
	query := `
		UPDATE users SET
			phone = $2,
			name = $3,
			email = $4,
			business_name = $5,
			brand_image = $6,
			country = $7,
			state = $8,
			city = $9,
			address = $10,
			pincode = $11,
			device_id = $12,
			device_model = $13,
			app_version = $14,
			fcm_token = $15,
			last_open = $16,
			updated_date = $17
		WHERE firebase_uid = $1
	`

	_, err := DB.Exec(
		query,
		user.FirebaseUID,
		user.Phone,
		user.Name,
		user.Email,
		user.BusinessName,
		user.BrandImage,
		user.Country,
		user.State,
		user.City,
		user.Address,
		user.Pincode,
		user.DeviceID,
		user.DeviceModel,
		user.AppVersion,
		user.FCMToken,
		user.LastOpen,
		user.UpdatedDate,
	)

	return err
}

func IncrementUserTotalCustomers(firebaseUID string) error {
	query := `
		UPDATE users SET
			total_customers = total_customers + 1
		WHERE firebase_uid = $1
	`

	_, err := DB.Exec(query, firebaseUID)
	return err
}

func IncrementUserProductCount(firebaseUID string) error {
	query := `
		UPDATE users SET
			product_count = product_count + 1
		WHERE firebase_uid = $1
	`

	_, err := DB.Exec(query, firebaseUID)
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

	if err != nil {
		return err
	}

	// Update total_customers in users table using added_by (firebase_uid)
	if client.AddedBy != "" {
		err = IncrementUserTotalCustomers(client.AddedBy)
		if err != nil {
			log.Printf("Failed to increment total_customers for user %s: %v", client.AddedBy, err)
			return fmt.Errorf("client created but failed to update user total_customers: %w", err)
		}
	}

	return nil
}

func UpdateClient(client *models.ClientModel) error {
	now := time.Now()
	client.JoinDate = now
	client.UpdatedAt = &now

	query := `
		UPDATE clients SET
			name = $1,
			email = $2,
			phone = $3,
			alternate_phone = $4,
			avatar = $5,
			status = $6,
			total_orders = $7,
			total_spent = $8,
			join_date = $9,
			address = $10,
			tags = $11,
			last_order_date = $12,
			website = $13,
			notes = $14,
			company_name = $15,
			updated_at = $16
		WHERE id = $17
	`

	_, err := DB.Exec(
		query,
		client.Name,
		client.Email,
		client.Phone,
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
		client.UpdatedAt,
		client.ID,
	)

	return err
}

func DeleteClient(clientID string) error {
	query := `DELETE FROM clients WHERE id = $1`

	_, err := DB.ExecContext(
		context.Background(),
		query,
		clientID,
	)

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

// GetUserIDByFirebaseUID returns the numeric users.id for the given firebase_uid.
// Use this for owner_id in shops - owner_id must always be users.id, never firebase_uid.
func GetUserIDByFirebaseUID(firebaseUID string) (int, error) {
	var userID int
	err := DB.QueryRow(`SELECT id FROM users WHERE firebase_uid = $1`, firebaseUID).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func GetUserByFirebaseUID(firebaseUID string) (models.User, error) {
	rows, err := DB.Query(`
		SELECT 
			id, firebase_uid, phone, name, email, business_name,
			brand_image,
			country, state, city, address, pincode, total_customers, product_count,
			last_login, last_active,
			device_id, device_model, app_version,
			fcm_token, last_open,
			is_premium, plan_name, plan_expiry,
			rating, account_status,
			referral_code, referred_by,
			created_date, updated_date,
			purchased_at
		FROM users WHERE firebase_uid = $1
	`, firebaseUID)
	if err != nil {
		return models.User{}, err
	}
	defer rows.Close()

	var user models.User
	for rows.Next() {
		if err := scanUserRow(rows, &user); err != nil {
			return models.User{}, err
		}
	}

	return user, nil
}

// UpdateUserActivity sets last_open to now and optionally updates fcm_token when non-empty.
// userID is firebase_uid. Returns rows affected (0 if no such user).
func UpdateUserActivity(ctx context.Context, firebaseUID string, fcmToken *string) (int64, error) {
	now := time.Now()
	if fcmToken != nil {
		if tok := strings.TrimSpace(*fcmToken); tok != "" {
			res, err := DB.ExecContext(ctx, `
				UPDATE users SET last_open = $1, fcm_token = $2, updated_date = $3
				WHERE firebase_uid = $4`,
				now, tok, now, firebaseUID)
			if err != nil {
				return 0, err
			}
			return res.RowsAffected()
		}
	}
	res, err := DB.ExecContext(ctx, `
		UPDATE users SET last_open = $1, updated_date = $2
		WHERE firebase_uid = $3`,
		now, now, firebaseUID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func InsertOrder(orderID, orderNumber string, req models.CreateOrderRequest) error {
	query := `
		INSERT INTO orders (
			id,
			order_number,
			client_id,
			client_name,
			client_avatar,
			total_amount,
			status,
			payment_status,
			created_at,
			delivery_date,
			delivery_address,
			notes,
			added_by,
			created_by,
			gst_number,
			discount,
			delivery_fee
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			NOW(),
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16
		);
	`

	_, err := DB.ExecContext(
		context.Background(),
		query,
		orderID,
		orderNumber,
		req.ClientID,
		req.ClientName,
		req.ClientAvatar,
		req.TotalAmount,
		req.Status,
		req.PaymentStatus,
		req.DeliveryDate,
		req.DeliveryAddress,
		req.Notes,
		req.AddedBy,
		req.CreatedBy,
		req.GstNumber,
		models.AmountOrZero(req.Discount),
		models.AmountOrZero(req.DeliveryFee),
	)

	return err
}

func UpdateClientStatsOnOrder(clientID string, orderAmount float64) error {
	query := `
		UPDATE clients SET
			total_orders = total_orders + 1,
			total_spent = total_spent + $1,
			updated_at = NOW()
		WHERE id = $2
	`

	_, err := DB.ExecContext(
		context.Background(),
		query,
		orderAmount,
		clientID,
	)

	return err
}

func InsertOrderItem(orderID string, itemID string, item models.OrderItemModel) error {
	query := `
		INSERT INTO order_items (
			id,
			order_id,
			name,
			description,
			price,
			quantity,
			created_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, NOW()
		);
	`

	_, err := DB.ExecContext(
		context.Background(),
		query,
		itemID,
		orderID,
		item.Name,
		item.Description,
		item.Price,
		item.Quantity,
	)

	if err != nil {
		log.Printf("InsertOrderItem error: %v", err)
		return err
	}

	return nil
}

func DeleteOrderItems(orderID string) error {
	query := `DELETE FROM order_items WHERE order_id = $1`

	_, err := DB.ExecContext(
		context.Background(),
		query,
		orderID,
	)

	return err
}

// GetOrderClientAndAmount fetches client_id and total_amount for an order (used before delete to update client stats).
func GetOrderClientAndAmount(orderID string) (clientID string, totalAmount float64, err error) {
	query := `SELECT client_id, total_amount FROM orders WHERE id = $1`
	err = DB.QueryRowContext(context.Background(), query, orderID).Scan(&clientID, &totalAmount)
	return
}

// UpdateClientStatsOnOrderDelete decrements client stats when an order is deleted.
func UpdateClientStatsOnOrderDelete(clientID string, orderAmount float64) error {
	query := `
		UPDATE clients SET
			total_orders = GREATEST(0, total_orders - 1),
			total_spent = GREATEST(0, total_spent - $1),
			updated_at = NOW()
		WHERE id = $2
	`

	_, err := DB.ExecContext(
		context.Background(),
		query,
		orderAmount,
		clientID,
	)

	return err
}

func DeleteOrder(orderID string) error {
	query := `DELETE FROM orders WHERE id = $1`

	result, err := DB.ExecContext(
		context.Background(),
		query,
		orderID,
	)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

func UpdateOrder(req models.OrderModel, bill *models.BillFieldUpdate, statusSet, paymentStatusSet bool) error {
	query := `UPDATE orders SET updated_at = NOW()`
	args := []interface{}{}
	n := 1

	if statusSet {
		query += fmt.Sprintf(", status = $%d", n)
		args = append(args, req.Status)
		n++
	}
	if paymentStatusSet {
		query += fmt.Sprintf(", payment_status = $%d", n)
		args = append(args, req.PaymentStatus)
		n++
	}

	if bill != nil {
		if bill.GstNumberSet {
			query += fmt.Sprintf(", gst_number = $%d", n)
			args = append(args, bill.GstNumber)
			n++
		}
		if bill.DiscountSet {
			query += fmt.Sprintf(", discount = $%d", n)
			args = append(args, bill.Discount)
			n++
		}
		if bill.DeliveryFeeSet {
			query += fmt.Sprintf(", delivery_fee = $%d", n)
			args = append(args, bill.DeliveryFee)
			n++
		}
		if bill.TotalAmountSet {
			query += fmt.Sprintf(", total_amount = $%d", n)
			args = append(args, bill.TotalAmount)
			n++
		}
	}

	query += fmt.Sprintf(" WHERE id = $%d", n)
	args = append(args, req.ID)

	_, err := DB.ExecContext(context.Background(), query, args...)
	return err
}

const orderSelectColumns = `
			id,
			order_number,
			client_id,
			client_name,
			client_avatar,
			total_amount,
			status,
			payment_status,
			created_at,
			updated_at,
			delivery_date,
			delivery_address,
			notes,
			added_by,
			created_by,
			gst_number,
			discount,
			delivery_fee`

func scanOrderRow(scanner interface{ Scan(dest ...any) error }, o *models.OrderModel) error {
	return scanner.Scan(
		&o.ID,
		&o.OrderNumber,
		&o.ClientID,
		&o.ClientName,
		&o.ClientAvatar,
		&o.TotalAmount,
		&o.Status,
		&o.PaymentStatus,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.DeliveryDate,
		&o.DeliveryAddress,
		&o.Notes,
		&o.AddedBy,
		&o.CreatedBy,
		&o.GstNumber,
		&o.Discount,
		&o.DeliveryFee,
	)
}

func attachOrderItems(o *models.OrderModel) error {
	items, err := FetchOrderItems(o.ID)
	if err != nil {
		return err
	}
	o.Items = items
	return nil
}

func FetchOrderByID(orderID string) (*models.OrderModel, error) {
	query := `SELECT ` + orderSelectColumns + ` FROM orders WHERE id = $1`
	var o models.OrderModel
	err := scanOrderRow(DB.QueryRowContext(context.Background(), query, orderID), &o)
	if err != nil {
		return nil, err
	}
	if err := attachOrderItems(&o); err != nil {
		return nil, err
	}
	return &o, nil
}

// MarkOrderPaid sets payment_status to paid and status to completed for an order_number.
func MarkOrderPaid(orderNumber string) error {
	query := `
		UPDATE orders
		SET payment_status = 'Paid',
		    status = 'Completed',
		    updated_at = NOW()
		WHERE order_number = $1
	`

	_, err := DB.ExecContext(context.Background(), query, orderNumber)
	return err
}

func GetAllOrders() ([]models.OrderModel, error) {
	query := `SELECT ` + orderSelectColumns + `
		FROM orders
		ORDER BY created_at DESC;
	`

	rows, err := DB.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.OrderModel
	for rows.Next() {
		var o models.OrderModel
		if err := scanOrderRow(rows, &o); err != nil {
			return nil, err
		}
		if err := attachOrderItems(&o); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func FetchOrdersByCreatedBy(createdBy string) ([]models.OrderModel, error) {
	query := `SELECT ` + orderSelectColumns + `
		FROM orders
		WHERE created_by = $1
		ORDER BY created_at DESC;
	`

	rows, err := DB.QueryContext(context.Background(), query, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.OrderModel

	for rows.Next() {
		var o models.OrderModel
		if err := scanOrderRow(rows, &o); err != nil {
			return nil, err
		}
		if err := attachOrderItems(&o); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}
func FetchOrderItems(orderID string) ([]models.OrderItemModel, error) {
	query := `
		SELECT 
			id,
			order_id,
			name,
			description,
			price,
			quantity,
			created_at
		FROM order_items
		WHERE order_id = $1;
	`

	rows, err := DB.QueryContext(context.Background(), query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItemModel

	for rows.Next() {
		var it models.OrderItemModel

		err := rows.Scan(
			&it.ID,
			&it.OrderID,
			&it.Name,
			&it.Description,
			&it.Price,
			&it.Quantity,
			&it.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		items = append(items, it)
	}

	return items, nil
}

func InsertPayment(paymentID string, req models.CreatePaymentRequest) error {
	query := `
		INSERT INTO payments (
			id,
			client_id,
			client_name,
			client_avatar,
			amount,
			date,
			status,
			order_id,
			payment_method,
			paid_amount,
			notes,
			created_by,
			added_by,        -- NEW COLUMN
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11, $12,
			$13,             -- NEW VALUE
			NOW(), NOW()
		);
	`

	_, err := DB.ExecContext(
		context.Background(),
		query,
		paymentID,
		req.ClientID,
		req.ClientName,
		req.ClientAvatar,
		req.Amount,
		req.Date,
		req.Status,
		req.OrderID,
		req.PaymentMethod,
		req.PaidAmount,
		req.Notes,
		req.CreatedBy,
		req.AddedBy, // NEW FIELD
	)

	return err
}

func GetAllPayments() ([]models.PaymentModel, error) {
	rows, err := DB.Query(`
		SELECT 
			id,
			client_id,
			client_name,
			client_avatar,
			amount,
			date,
			status,
			order_id,
			payment_method,
			paid_amount,
			notes,
			created_by,
			added_by,
			created_at,
			updated_at
		FROM payments
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.PaymentModel
	for rows.Next() {
		var p models.PaymentModel
		err := rows.Scan(
			&p.ID,
			&p.ClientID,
			&p.ClientName,
			&p.ClientAvatar,
			&p.Amount,
			&p.Date,
			&p.Status,
			&p.OrderID,
			&p.PaymentMethod,
			&p.PaidAmount,
			&p.Notes,
			&p.CreatedBy,
			&p.AddedBy,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}

func GetPaymentsByCreatedBy(createdBy string) ([]models.PaymentModel, error) {
	rows, err := DB.Query(`
		SELECT 
			id,
			client_id,
			client_name,
			client_avatar,
			amount,
			date,
			status,
			order_id,
			payment_method,
			paid_amount,
			notes,
			created_by,
			added_by,
			created_at,
			updated_at
		FROM payments
		WHERE created_by = $1
		ORDER BY created_at DESC
	`, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.PaymentModel

	for rows.Next() {
		var p models.PaymentModel

		err := rows.Scan(
			&p.ID,
			&p.ClientID,
			&p.ClientName,
			&p.ClientAvatar,
			&p.Amount,
			&p.Date,
			&p.Status,
			&p.OrderID,
			&p.PaymentMethod,
			&p.PaidAmount,
			&p.Notes,
			&p.CreatedBy,
			&p.AddedBy,
			&p.CreatedAt,
			&p.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		payments = append(payments, p)
	}

	return payments, nil
}

func DeletePayment(paymentID string) error {
	query := `DELETE FROM payments WHERE id = $1`

	result, err := DB.ExecContext(
		context.Background(),
		query,
		paymentID,
	)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("payment not found")
	}

	return nil
}

func InsertBillingTransaction(req models.CreateBillingTransactionRequest) (string, error) {
	transactionID := uuid.New().String()

	query := `
		INSERT INTO billing_transactions (
			id,
			user_id,
			plan_id,
			plan_name,
			razorpay_order_id,
			razorpay_payment_id,
			razorpay_signature,
			amount,
			currency,
			status,
			plan_expiry,
			created_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW()
		)
		RETURNING id
	`

	err := DB.QueryRow(
		query,
		transactionID,
		req.UserID,
		req.PlanID,
		req.PlanName,
		req.RazorpayOrderID,
		req.RazorpayPaymentID,
		req.RazorpaySignature,
		req.Amount,
		req.Currency,
		req.Status,
		req.PlanExpiry,
	).Scan(&transactionID)

	if err != nil {
		return "", err
	}

	return transactionID, nil
}

func GetAllBillingTransactions() ([]models.BillingTransaction, error) {
	rows, err := DB.Query(`
		SELECT 
			id,
			user_id,
			plan_id,
			plan_name,
			razorpay_order_id,
			razorpay_payment_id,
			razorpay_signature,
			amount,
			currency,
			status,
			plan_expiry,
			created_at
		FROM billing_transactions
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.BillingTransaction
	for rows.Next() {
		var t models.BillingTransaction
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.PlanID,
			&t.PlanName,
			&t.RazorpayOrderID,
			&t.RazorpayPaymentID,
			&t.RazorpaySignature,
			&t.Amount,
			&t.Currency,
			&t.Status,
			&t.PlanExpiry,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}

func InsertProduct(product *models.ProductModel) error {
	query := `
		INSERT INTO products (
			name, description, price, profit, sku, image_url, added_by,
			created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9
		)
		RETURNING id
	`

	err := DB.QueryRow(
		query,
		product.Name,
		product.Description,
		product.Price,
		product.Profit,
		product.SKU,
		product.ImageURL,
		product.AddedBy,
		product.CreatedAt,
		product.UpdatedAt,
	).Scan(&product.ID)

	if err != nil {
		return err
	}

	if product.AddedBy != "" {
		if err := IncrementUserProductCount(product.AddedBy); err != nil {
			log.Printf("Failed to increment product_count for user %s: %v", product.AddedBy, err)
			return fmt.Errorf("product created but failed to update user product_count: %w", err)
		}
	}

	return nil
}

func UpdateProduct(product *models.ProductModel) error {
	now := time.Now()
	product.UpdatedAt = &now

	query := `
		UPDATE products SET
			name = $1,
			description = $2,
			price = $3,
			profit = $4,
			sku = $5,
			updated_at = $6,
			image_url = COALESCE($7, image_url)
		WHERE id = $8
	`

	_, err := DB.ExecContext(
		context.Background(),
		query,
		product.Name,
		product.Description,
		product.Price,
		product.Profit,
		product.SKU,
		product.UpdatedAt,
		product.ImageURL,
		product.ID,
	)
	return err
}

// GetProductByID returns a CRM catalog product by id.
func GetProductByID(id string) (*models.ProductModel, error) {
	var p models.ProductModel
	err := DB.QueryRow(`
		SELECT id, name, description, price, profit, sku, image_url, added_by, created_at, updated_at
		FROM products WHERE id = $1
	`, id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.Price,
		&p.Profit,
		&p.SKU,
		&p.ImageURL,
		&p.AddedBy,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product not found")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func DeleteProduct(productID string) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := DB.ExecContext(
		context.Background(),
		query,
		productID,
	)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func GetAllProducts() ([]models.ProductModel, error) {
	rows, err := DB.Query(`
		SELECT id, name, description, price, profit, sku, image_url, added_by, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.ProductModel
	for rows.Next() {
		var p models.ProductModel
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Profit,
			&p.SKU,
			&p.ImageURL,
			&p.AddedBy,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func GetProductsByCreatedBy(createdBy string) ([]models.ProductModel, error) {
	rows, err := DB.Query(`
		SELECT id, name, description, price, profit, sku, image_url, added_by, created_at, updated_at
		FROM products
		WHERE added_by = $1
		ORDER BY created_at DESC
	`, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.ProductModel
	for rows.Next() {
		var p models.ProductModel
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Profit,
			&p.SKU,
			&p.ImageURL,
			&p.AddedBy,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func UpdateUserPlanData(firebaseUID string, planName string, planExpiry time.Time) error {
	query := `
		UPDATE users SET
			is_premium = $1,
			plan_name = $2,
			plan_expiry = $3,
			updated_date = NOW()
		WHERE firebase_uid = $4
	`

	_, err := DB.Exec(
		query,
		true,        // is_premium = true
		planName,    // plan_name
		planExpiry,  // plan_expiry
		firebaseUID, // user_id (firebase_uid)
	)

	if err != nil {
		return err
	}

	return nil
}
