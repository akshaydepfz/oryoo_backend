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

	runMigrations(helper.DB)
	RunSitesMigrations(helper.DB)
	fmt.Println("Database connection established")

}

// runMigrations applies schema migrations to ensure required columns exist.
// Safe to run on every startup; ADD COLUMN IF NOT EXISTS is idempotent.
func runMigrations(db *sql.DB) {
	migrations := []struct {
		name  string
		query string
	}{
		{
			"contact_pages created_at",
			`ALTER TABLE contact_pages ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW();`,
		},
		{
			"contact_pages updated_at",
			`ALTER TABLE contact_pages ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();`,
		},
		{
			"testimonials created_at",
			`ALTER TABLE testimonials ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW();`,
		},
		{
			"testimonials updated_at",
			`ALTER TABLE testimonials ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();`,
		},
		{
			"about_pages created_at",
			`ALTER TABLE about_pages ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW();`,
		},
		{
			"about_pages updated_at",
			`ALTER TABLE about_pages ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();`,
		},
	}
	for _, m := range migrations {
		_, err := db.Exec(m.query)
		if err != nil {
			log.Printf("Migration %s: %v", m.name, err)
		}
	}
}

// RunSitesMigrations adds missing columns for Sites UI. Safe to run on every startup.
func RunSitesMigrations(db *sql.DB) {
	migrations := []struct {
		name  string
		query string
	}{
		{"products_sites updated_at", `ALTER TABLE products_sites ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();`},
		{"products_sites featured", `ALTER TABLE products_sites ADD COLUMN IF NOT EXISTS featured BOOLEAN DEFAULT FALSE;`},
		{"products_sites slug", `ALTER TABLE products_sites ADD COLUMN IF NOT EXISTS slug TEXT;`},
		{"site_configs hero_image_url", `ALTER TABLE site_configs ADD COLUMN IF NOT EXISTS hero_image_url TEXT;`},
		{"site_configs hero_title", `ALTER TABLE site_configs ADD COLUMN IF NOT EXISTS hero_title TEXT;`},
		{"site_configs hero_subtitle", `ALTER TABLE site_configs ADD COLUMN IF NOT EXISTS hero_subtitle TEXT;`},
		{"site_configs logo_url", `ALTER TABLE site_configs ADD COLUMN IF NOT EXISTS logo_url TEXT;`},
		{"site_configs email", `ALTER TABLE site_configs ADD COLUMN IF NOT EXISTS email TEXT;`},
		{"site_configs secondary_color", `ALTER TABLE site_configs ADD COLUMN IF NOT EXISTS secondary_color TEXT;`},
		{"site_configs twitter_url", `ALTER TABLE site_configs ADD COLUMN IF NOT EXISTS twitter_url TEXT;`},
		{"contact_pages latitude", `ALTER TABLE contact_pages ADD COLUMN IF NOT EXISTS latitude NUMERIC;`},
		{"contact_pages longitude", `ALTER TABLE contact_pages ADD COLUMN IF NOT EXISTS longitude NUMERIC;`},
		{"contact_pages map_embed_url", `ALTER TABLE contact_pages ADD COLUMN IF NOT EXISTS map_embed_url TEXT;`},
	}
	for _, m := range migrations {
		_, err := db.Exec(m.query)
		if err != nil {
			log.Printf("RunSitesMigrations %s: %v", m.name, err)
		}
	}
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

// DropSitesTables drops all Oryoo Sites tables in reverse dependency order.
// Call before CreateSitesTables when schema has changed or tables have wrong structure.
func DropSitesTables() error {
	tables := []string{
		"product_images",
		"products_sites",
		"categories",
		"site_configs",
		"testimonials",
		"about_pages",
		"contact_pages",
		"shops",
	}
	for _, t := range tables {
		_, err := helper.DB.ExecContext(context.Background(), "DROP TABLE IF EXISTS "+t+" CASCADE")
		if err != nil {
			log.Printf("Error dropping table %s: %v", t, err)
			return err
		}
	}
	fmt.Println("Oryoo Sites tables dropped")
	return nil
}

func CreateShopsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS shops (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			subdomain TEXT UNIQUE,
			custom_domain TEXT,
			owner_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			status TEXT DEFAULT 'active',
			created_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating shops table: %v", err)
		return err
	}
	return nil
}

// AddOwnerIDToShops adds owner_id column to existing shops table (migration for existing deployments)
func AddOwnerIDToShops() error {
	_, err := helper.DB.ExecContext(context.Background(), `
		ALTER TABLE shops ADD COLUMN IF NOT EXISTS owner_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
	`)
	if err != nil {
		log.Printf("Error adding owner_id to shops: %v", err)
		return err
	}
	return nil
}

func CreateSiteConfigsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS site_configs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			shop_id UUID UNIQUE REFERENCES shops(id) ON DELETE CASCADE,
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
	// Migration: alter color columns from INTEGER to TEXT for existing deployments
	_ = MigrateSiteConfigsColorsToText()
	return nil
}

// MigrateSiteConfigsColorsToText alters color columns from INTEGER to TEXT (for existing tables created with INTEGER).
func MigrateSiteConfigsColorsToText() error {
	migrations := []string{
		`ALTER TABLE site_configs ALTER COLUMN primary_color TYPE TEXT USING primary_color::TEXT`,
		`ALTER TABLE site_configs ALTER COLUMN gold_color TYPE TEXT USING gold_color::TEXT`,
		`ALTER TABLE site_configs ALTER COLUMN text_color TYPE TEXT USING text_color::TEXT`,
		`ALTER TABLE site_configs ALTER COLUMN text_muted TYPE TEXT USING text_muted::TEXT`,
	}
	for _, q := range migrations {
		if _, err := helper.DB.ExecContext(context.Background(), q); err != nil {
			// Column may already be TEXT (new deployments); log and continue
			log.Printf("Migration site_configs colors (may be no-op): %v", err)
		}
	}
	return nil
}

func CreateCategoriesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY,
			shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
			name TEXT,
			slug TEXT,
			image_url TEXT
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating categories table: %v", err)
		return err
	}
	return nil
}

func CreateProductsSitesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS products_sites (
			id UUID PRIMARY KEY,
			shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
			category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
			name TEXT,
			description TEXT,
			price NUMERIC,
			created_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating products_sites table: %v", err)
		return err
	}
	return nil
}

func CreateProductImagesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS product_images (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id UUID REFERENCES products_sites(id) ON DELETE CASCADE,
			image_url TEXT
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating product_images table: %v", err)
		return err
	}
	return nil
}

func CreateTestimonialsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS testimonials (
			id UUID PRIMARY KEY,
			shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
			text TEXT,
			author TEXT,
			rating INTEGER,
			avatar_url TEXT
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating testimonials table: %v", err)
		return err
	}
	return nil
}

func CreateAboutPagesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS about_pages (
			id UUID PRIMARY KEY,
			shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
			hero_image_url TEXT,
			title TEXT,
			tagline TEXT,
			story_text TEXT,
			story_text_secondary TEXT,
			story_image_url TEXT,
			values JSONB,
			craftsmanship JSONB
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating about_pages table: %v", err)
		return err
	}
	return nil
}

func CreateContactPagesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS contact_pages (
			id UUID PRIMARY KEY,
			shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
			title TEXT,
			subtitle TEXT,
			store_address TEXT,
			store_address_short TEXT,
			phone_number TEXT,
			whatsapp_number TEXT,
			email TEXT,
			google_map_url TEXT,
			store_hours JSONB
		);
	`
	_, err := helper.DB.ExecContext(context.Background(), query)
	if err != nil {
		log.Printf("Error creating contact_pages table: %v", err)
		return err
	}
	return nil
}

// CreateSitesIndexes creates indexes for performance
func CreateSitesIndexes() error {
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_shops_subdomain ON shops(subdomain)`,
		`CREATE INDEX IF NOT EXISTS idx_shops_custom_domain ON shops(custom_domain)`,
		`CREATE INDEX IF NOT EXISTS idx_products_sites_shop_id ON products_sites(shop_id)`,
		`CREATE INDEX IF NOT EXISTS idx_categories_shop_id ON categories(shop_id)`,
		`CREATE INDEX IF NOT EXISTS idx_testimonials_shop_id ON testimonials(shop_id)`,
	}
	for _, q := range indexes {
		if _, err := helper.DB.ExecContext(context.Background(), q); err != nil {
			log.Printf("Error creating site index: %v", err)
			return err
		}
	}
	return nil
}

// CreateSitesTables creates all Oryoo Sites tables in dependency order (call after ConnectDatabase).
func CreateSitesTables() {
	_ = CreateShopsTable()
	_ = AddOwnerIDToShops() // Migration: add owner_id to existing shops tables
	_ = CreateSiteConfigsTable()
	_ = CreateCategoriesTable()
	_ = CreateProductsSitesTable()
	_ = CreateProductImagesTable()
	_ = CreateTestimonialsTable()
	_ = CreateAboutPagesTable()
	_ = CreateContactPagesTable()
	_ = CreateSitesIndexes()
}
