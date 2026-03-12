package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"oryoo.com/helper"
)

func ConnectDatabase() {

	const (
		host     = "ep-purple-union-aiia6zpe.c-4.us-east-1.pg.koyeb.app"
		port     = 5432
		user     = "koyeb-adm"
		password = "npg_j4fBv0uEsSkZ"
		dbname   = "koyebdb"
	)

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)
	var err error

	helper.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	if err = helper.DB.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	fmt.Println("Database connection established")

}

func CreateUsersTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			firebase_uid TEXT NOT NULL UNIQUE,
			phone TEXT NOT NULL,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			business_name TEXT,
			brand_image TEXT,
			country TEXT,
			state TEXT,
			city TEXT,
			address TEXT,
			pincode TEXT,
			total_customers INTEGER NOT NULL DEFAULT 0,
			last_login TIMESTAMP,
			last_active TIMESTAMP,
			device_id TEXT,
			device_model TEXT,
			app_version TEXT,
			is_premium BOOLEAN NOT NULL DEFAULT false,
			plan_name TEXT,
			plan_expiry TIMESTAMP,
			rating REAL DEFAULT 5.0,
			account_status TEXT NOT NULL DEFAULT 'active',
			referral_code TEXT,
			referred_by TEXT,
			created_date TIMESTAMP DEFAULT NOW(),
			updated_date TIMESTAMP DEFAULT NOW()
		);
	`

	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating users table: %v", err)
		return err
	}

	fmt.Println("Users table created successfully")
	return nil
}

func CreateClientsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS clients (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			phone TEXT NOT NULL,
			added_by TEXT NOT NULL,
			alternate_phone TEXT,
			avatar TEXT,
			status TEXT NOT NULL DEFAULT 'Active',
			total_orders INTEGER NOT NULL DEFAULT 0,
			total_spent DOUBLE PRECISION NOT NULL DEFAULT 0,
			join_date TIMESTAMP NOT NULL,
			address TEXT NOT NULL,
			tags TEXT[] DEFAULT ARRAY[]::TEXT[],
			last_order_date TIMESTAMP,
			website TEXT,
			notes TEXT,
			company_name TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`

	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating clients table: %v", err)
		return err
	}

	fmt.Println("Clients table created successfully")
	return nil
}

func CreateOrdersTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
			order_number TEXT NOT NULL UNIQUE,

			client_id TEXT NOT NULL,
			client_name TEXT NOT NULL,
			client_avatar TEXT,

			total_amount DOUBLE PRECISION NOT NULL DEFAULT 0,

			status TEXT NOT NULL DEFAULT 'Pending',
			payment_status TEXT NOT NULL DEFAULT 'Pending',

			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP,
			delivery_date TIMESTAMP,

			delivery_address TEXT NOT NULL,
			notes TEXT,
			added_by TEXT
		);
	`

	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating orders table: %v", err)
		return err
	}

	fmt.Println("Orders table created successfully")
	return nil
}

func CreateOrderItemsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS order_items (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
			order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,

			name TEXT NOT NULL,
			description TEXT,
			price DOUBLE PRECISION NOT NULL DEFAULT 0,
			quantity INTEGER NOT NULL DEFAULT 1,

			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`

	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating order_items table: %v", err)
		return err
	}

	fmt.Println("Order items table created successfully")
	return nil
}

func AddCreatedByColumnToOrders() error {
	query := `
		ALTER TABLE orders 
		ADD COLUMN IF NOT EXISTS created_by TEXT;
	`

	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error altering orders table (created_by): %v", err)
		return err
	}

	fmt.Println("Column 'created_by' added to orders table successfully")
	return nil
}

func CreatePaymentsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS payments (
			id TEXT PRIMARY KEY,
			client_id TEXT NOT NULL,
			client_name TEXT NOT NULL,
			client_avatar TEXT,
			amount NUMERIC(10,2) NOT NULL,
			date TIMESTAMP NOT NULL,
			status TEXT NOT NULL,
			order_id TEXT NOT NULL,
			payment_method TEXT NOT NULL,
			paid_amount NUMERIC(10,2),
			notes TEXT,
			created_by TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`

	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating payments table: %v", err)
		return err
	}

	fmt.Println("Table 'payments' created successfully")
	return nil
}

func AddAddedByColumnToPayments() error {
	query := `
		ALTER TABLE payments
		ADD COLUMN IF NOT EXISTS added_by TEXT;
	`

	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error altering payments table (added_by): %v", err)
		return err
	}

	fmt.Println("Column 'added_by' added to payments table successfully")
	return nil
}
func AddtotalCustomersInUsersTable() error {
	query := `
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS total_customers INTEGER NOT NULL DEFAULT 0;
	`

	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error altering users table (total_customers): %v", err)
		return err
	}

	fmt.Println("Column 'total_customers' added to users table successfully")
	return nil
}

func CreateProductsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS products (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			price DOUBLE PRECISION NOT NULL DEFAULT 0,
			sku TEXT,
			added_by TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`

	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating products table: %v", err)
		return err
	}

	fmt.Println("Products table created successfully")
	return nil
}

func CreateBillingTransactionsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS billing_transactions (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
			user_id TEXT NOT NULL,
			plan_id TEXT NOT NULL,
			plan_name TEXT NOT NULL,
			razorpay_order_id TEXT,
			razorpay_payment_id TEXT,
			razorpay_signature TEXT,
			amount INTEGER NOT NULL,
			currency TEXT NOT NULL DEFAULT 'INR',
			status TEXT NOT NULL,
			plan_expiry TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		);
	`

	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating billing_transactions table: %v", err)
		return err
	}

	fmt.Println("Billing transactions table created successfully")
	return nil
}

// --- Oryoo Sites Tables ---

func CreateShopsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS shops (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			subdomain TEXT NOT NULL UNIQUE,
			custom_domain TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			status TEXT NOT NULL DEFAULT 'active'
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating shops table: %v", err)
		return err
	}
	fmt.Println("Shops table created successfully")
	return nil
}

func CreateSiteConfigsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS site_configs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			shop_id UUID NOT NULL UNIQUE REFERENCES shops(id) ON DELETE CASCADE,
			shop_name TEXT,
			tagline TEXT,
			primary_color TEXT,
			gold_color TEXT,
			text_color TEXT,
			text_muted TEXT,
			phone_number TEXT,
			whatsapp_number TEXT,
			store_address TEXT,
			store_address_short TEXT,
			instagram_url TEXT,
			facebook_url TEXT,
			pinterest_url TEXT,
			google_map_url TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating site_configs table: %v", err)
		return err
	}
	fmt.Println("Site_configs table created successfully")
	return nil
}

func CreateSitesCategoriesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			shop_id UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			slug TEXT NOT NULL,
			image_url TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating categories table: %v", err)
		return err
	}
	fmt.Println("Categories table (sites) created successfully")
	return nil
}

func CreateSitesProductsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS products_sites (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			shop_id UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
			category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
			name TEXT NOT NULL,
			description TEXT,
			price DOUBLE PRECISION NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating products_sites table: %v", err)
		return err
	}
	fmt.Println("Products_sites table created successfully")
	return nil
}

func CreateProductImagesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS product_images (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id UUID NOT NULL REFERENCES products_sites(id) ON DELETE CASCADE,
			image_url TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating product_images table: %v", err)
		return err
	}
	fmt.Println("Product_images table created successfully")
	return nil
}

func CreateTestimonialsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS testimonials (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			shop_id UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
			text TEXT NOT NULL,
			author TEXT NOT NULL,
			rating INTEGER NOT NULL DEFAULT 5,
			avatar_url TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating testimonials table: %v", err)
		return err
	}
	fmt.Println("Testimonials table created successfully")
	return nil
}

func CreateAboutPagesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS about_pages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			shop_id UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
			hero_image_url TEXT,
			title TEXT,
			tagline TEXT,
			story_text TEXT,
			story_text_secondary TEXT,
			story_image_url TEXT,
			values JSONB,
			craftsmanship JSONB,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating about_pages table: %v", err)
		return err
	}
	fmt.Println("About_pages table created successfully")
	return nil
}

func CreateContactPagesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS contact_pages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			shop_id UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
			title TEXT,
			subtitle TEXT,
			store_address TEXT,
			phone_number TEXT,
			whatsapp_number TEXT,
			email TEXT,
			google_map_url TEXT,
			store_hours JSONB,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating contact_pages table: %v", err)
		return err
	}
	fmt.Println("Contact_pages table created successfully")
	return nil
}

// CreateSitesTables creates all Oryoo Sites tables (call after ConnectDatabase)
func CreateSitesTables() {
	_ = CreateShopsTable()
	_ = CreateSiteConfigsTable()
	_ = CreateSitesCategoriesTable()
	_ = CreateSitesProductsTable()
	_ = CreateProductImagesTable()
	_ = CreateTestimonialsTable()
	_ = CreateAboutPagesTable()
	_ = CreateContactPagesTable()
}
