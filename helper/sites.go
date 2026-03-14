package helper

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"oryoo.com/models"
)

// GetShopByDomain resolves domain to shop. If domain ends with .oryoo.in, extract subdomain; else match custom_domain.
func GetShopByDomain(domain string) (*models.Shop, error) {
	domain = strings.TrimSpace(strings.ToLower(domain))
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	var shop models.Shop
	var subdomain string

	if strings.HasSuffix(domain, ".oryoo.in") {
		// goldpalace.oryoo.in -> goldpalace
		subdomain = strings.TrimSuffix(domain, ".oryoo.in")
		subdomain = strings.TrimSuffix(subdomain, ".")
		if subdomain == "" || subdomain == "oryoo" {
			return nil, fmt.Errorf("invalid subdomain")
		}
		query := `SELECT id, name, subdomain, custom_domain, owner_id, created_at, status FROM shops WHERE subdomain = $1 AND status = 'active'`
		var ownerID *int
		err := DB.QueryRowContext(context.Background(), query, subdomain).Scan(
			&shop.ID, &shop.Name, &shop.Subdomain, &shop.CustomDomain, &ownerID, &shop.CreatedAt, &shop.Status,
		)
		if err == nil && ownerID != nil {
			shop.OwnerID = *ownerID
		}
		if err != nil {
			return nil, err
		}
		return &shop, nil
	}

	// Custom domain: goldpalace.com
	query := `SELECT id, name, subdomain, custom_domain, owner_id, created_at, status FROM shops WHERE LOWER(custom_domain) = $1 AND status = 'active'`
	var ownerID *int
	err := DB.QueryRowContext(context.Background(), query, domain).Scan(
		&shop.ID, &shop.Name, &shop.Subdomain, &shop.CustomDomain, &ownerID, &shop.CreatedAt, &shop.Status,
	)
	if err != nil {
		return nil, err
	}
	if ownerID != nil {
		shop.OwnerID = *ownerID
	}
	return &shop, nil
}

// ValidateOwnerID checks that the given owner_id exists in users table.
// Returns error if not found (caller should return 400).
func ValidateOwnerID(ownerID int) error {
	var id int
	err := DB.QueryRowContext(context.Background(), `SELECT id FROM users WHERE id = $1`, ownerID).Scan(&id)
	if err != nil {
		return fmt.Errorf("owner not found")
	}
	return nil
}

// InsertShop creates a new shop with the given ownerID (user id from users table).
// Caller must call CreateDefaultSiteContent(shop.ID) after successful insert to seed demo data.
func InsertShop(req models.CreateShopRequest, ownerID int) (*models.Shop, error) {
	ctx := context.Background()
	id := uuid.New().String()
	shopQuery := `
		INSERT INTO shops (id, name, subdomain, custom_domain, owner_id, status)
		VALUES ($1, $2, $3, $4, $5, 'active')
		RETURNING id, name, subdomain, custom_domain, owner_id, created_at, status
	`
	var shop models.Shop
	var ownerIDOut *int
	err := DB.QueryRowContext(ctx, shopQuery, id, req.Name, req.Subdomain, req.CustomDomain, ownerID).Scan(
		&shop.ID, &shop.Name, &shop.Subdomain, &shop.CustomDomain, &ownerIDOut, &shop.CreatedAt, &shop.Status,
	)
	if err != nil {
		return nil, err
	}
	if ownerIDOut != nil {
		shop.OwnerID = *ownerIDOut
	}
	return &shop, nil
}

// CreateDefaultSiteContent inserts default demo content for a new shop so the site never loads empty.
// Verifies the shop exists before inserting. Skips if site_configs already has data for the shop.
func CreateDefaultSiteContent(shopID string) error {
	// Verify shop exists before inserting (prevents FK violation)
	var exists bool
	if err := DB.QueryRowContext(context.Background(), `SELECT EXISTS(SELECT 1 FROM shops WHERE id = $1)`, shopID).Scan(&exists); err != nil {
		return fmt.Errorf("check shop exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("shop does not exist")
	}

	// Prevent duplicate demo data
	var count int
	if err := DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM site_configs WHERE shop_id = $1`, shopID).Scan(&count); err != nil {
		return fmt.Errorf("check site_configs: %w", err)
	}
	if count > 0 {
		return nil // data already exists, skip
	}

	tx, err := DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := createDefaultSiteContentTx(context.Background(), tx, shopID); err != nil {
		return err
	}
	return tx.Commit()
}

// createDefaultSiteContentTx inserts default content within a transaction.
// Caller must verify shop exists and no duplicate data before calling.
func createDefaultSiteContentTx(ctx context.Context, tx *sql.Tx, shopID string) error {
	// site_configs - full defaults per spec (payment_enabled false, no Razorpay keys in dummy)
	_, err := tx.ExecContext(ctx, `
		INSERT INTO site_configs (shop_id, shop_name, tagline, hero_image_url, hero_title, hero_subtitle,
			primary_color, secondary_color, phone_number, whatsapp_number, email,
			facebook_url, instagram_url, twitter_url, payment_enabled)
		VALUES ($1, 'Your Store', 'Premium products crafted with care',
			'https://source.unsplash.com/1600x900/?jewelry', 'Discover Timeless Elegance', 'Luxury collections for every occasion',
			'#CBA135', '#111111', '+91 99999 99999', '919999999999', 'hello@example.com',
			'https://facebook.com', 'https://instagram.com', 'https://twitter.com', false)
	`, shopID)
	if err != nil {
		return fmt.Errorf("site_configs: %w", err)
	}

	// categories - 3 categories: Rings, Necklaces, Bracelets
	categories := []struct {
		id   string
		name string
		slug string
		img  string
	}{
		{uuid.New().String(), "Rings", "rings", "https://images.unsplash.com/photo-1605100804763-247f67b3557e?w=800"},
		{uuid.New().String(), "Necklaces", "necklaces", "https://images.unsplash.com/photo-1599643478518-a784e5dc4c8f?w=800"},
		{uuid.New().String(), "Bracelets", "bracelets", "https://images.unsplash.com/photo-1611652022419-a9419f74343a?w=800"},
	}
	for _, cat := range categories {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO categories (id, shop_id, name, slug, image_url)
			VALUES ($1, $2, $3, $4, $5)
		`, cat.id, shopID, cat.name, cat.slug, cat.img)
		if err != nil {
			return fmt.Errorf("categories: %w", err)
		}
	}

	// products_sites - 6 demo products, first 3 featured
	lorem := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	products := []struct {
		id          string
		categoryIdx int
		name        string
		desc        string
		price       float64
		featured    bool
		images      []string
	}{
		{uuid.New().String(), 0, "Eternal Gold Ring", lorem, 24999.00, true, []string{"https://images.unsplash.com/photo-1605100804763-247f67b3557e?w=800", "https://images.unsplash.com/photo-1515562141207-7a88fb7ce338?w=800"}},
		{uuid.New().String(), 1, "Classic Pearl Necklace", lorem, 18999.00, true, []string{"https://images.unsplash.com/photo-1599643478518-a784e5dc4c8f?w=800", "https://images.unsplash.com/photo-1515562141207-7a88fb7ce338?w=800"}},
		{uuid.New().String(), 2, "Elegant Gold Bracelet", lorem, 15999.00, true, []string{"https://images.unsplash.com/photo-1611652022419-a9419f74343a?w=800", "https://images.unsplash.com/photo-1602751584552-8ba73aad10e1?w=800"}},
		{uuid.New().String(), 0, "Diamond Stud Earrings", lorem, 32999.00, false, []string{"https://images.unsplash.com/photo-1573408301185-9146fe634ad0?w=800", "https://images.unsplash.com/photo-1605100804763-247f67b3557e?w=800"}},
		{uuid.New().String(), 1, "Silver Pendant", lorem, 8999.00, false, []string{"https://images.unsplash.com/photo-1599643478518-a784e5dc4c8f?w=800", "https://images.unsplash.com/photo-1611652022419-a9419f74343a?w=800"}},
		{uuid.New().String(), 2, "Rose Gold Bangle", lorem, 12999.00, false, []string{"https://images.unsplash.com/photo-1611652022419-a9419f74343a?w=800", "https://images.unsplash.com/photo-1602751584552-8ba73aad10e1?w=800"}},
	}
	for _, p := range products {
		catID := categories[p.categoryIdx].id
		_, err = tx.ExecContext(ctx, `
			INSERT INTO products_sites (id, shop_id, category_id, name, description, price, featured)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, p.id, shopID, catID, p.name, p.desc, p.price, p.featured)
		if err != nil {
			return fmt.Errorf("products_sites: %w", err)
		}
		for pos, url := range p.images {
			_, err = tx.ExecContext(ctx, `INSERT INTO product_images (product_id, image_url, position) VALUES ($1, $2, $3)`, p.id, url, pos)
			if err != nil {
				return fmt.Errorf("product_images: %w", err)
			}
		}
	}

	// product_variants - dummy S, M, L, XL for first 2 products
	variantSizes := []struct {
		name  string
		price float64
		stock int
		sku   string
	}{
		{"S", 24999.00, 10, "RING-S"},
		{"M", 25999.00, 15, "RING-M"},
		{"L", 26999.00, 12, "RING-L"},
		{"XL", 27999.00, 8, "RING-XL"},
	}
	for _, v := range variantSizes {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO product_variants (id, product_id, name, price, stock, sku)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New().String(), products[0].id, v.name, v.price, v.stock, v.sku)
		if err != nil {
			return fmt.Errorf("product_variants: %w", err)
		}
	}
	necklaceVariants := []struct {
		name  string
		price float64
		stock int
		sku   string
	}{
		{"S", 18999.00, 5, "NECK-S"},
		{"M", 19499.00, 8, "NECK-M"},
		{"L", 19999.00, 6, "NECK-L"},
	}
	for _, v := range necklaceVariants {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO product_variants (id, product_id, name, price, stock, sku)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New().String(), products[1].id, v.name, v.price, v.stock, v.sku)
		if err != nil {
			return fmt.Errorf("product_variants: %w", err)
		}
	}

	// testimonials - 3 testimonials
	testimonials := []struct {
		text   string
		author string
		rating int
		avatar string
	}{
		{"Amazing craftsmanship and excellent quality.", "Happy Customer", 5, "https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=200"},
		{"Amazing craftsmanship and excellent quality.", "Happy Customer", 5, "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=200"},
		{"Amazing craftsmanship and excellent quality.", "Happy Customer", 5, "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=200"},
	}
	for _, t := range testimonials {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO testimonials (id, shop_id, text, author, rating, avatar_url)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New().String(), shopID, t.text, t.author, t.rating, t.avatar)
		if err != nil {
			return fmt.Errorf("testimonials: %w", err)
		}
	}

	// about_pages - lorem ipsum story and hero image
	loremStory := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris."
	loremSecondary := "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
	_, err = tx.ExecContext(ctx, `
		INSERT INTO about_pages (id, shop_id, hero_image_url, title, tagline, story_text, story_text_secondary, story_image_url, values, craftsmanship)
		VALUES ($1, $2, 'https://source.unsplash.com/1600x900/?jewelry', 'About Us', 'Our Story of Craftsmanship',
			$3, $4, 'https://images.unsplash.com/photo-1602751584552-8ba73aad10e1?w=800',
			'{"quality":"Lorem ipsum dolor sit amet.","craftsmanship":"Sed do eiusmod tempor.","heritage":"Ut enim ad minim veniam."}'::jsonb,
			'{"process":"Lorem ipsum dolor sit amet.","materials":"Ut labore et dolore magna aliqua.","finishing":"Duis aute irure dolor."}'::jsonb)
	`, uuid.New().String(), shopID, loremStory, loremSecondary)
	if err != nil {
		return fmt.Errorf("about_pages: %w", err)
	}

	// contact_pages - demo contact details with map_embed_url and store_hours
	storeHours := `{"monday":"9:00 AM - 6:00 PM","tuesday":"9:00 AM - 6:00 PM","wednesday":"9:00 AM - 6:00 PM","thursday":"9:00 AM - 6:00 PM","friday":"9:00 AM - 6:00 PM","saturday":"10:00 AM - 4:00 PM","sunday":"Closed"}`
	_, err = tx.ExecContext(ctx, `
		INSERT INTO contact_pages (id, shop_id, title, subtitle, store_address, phone_number, whatsapp_number, email, map_embed_url, store_hours)
		VALUES ($1, $2, 'Get in Touch', 'We would love to hear from you', '123 Jewelry Lane, Your City', '+91 99999 99999', '919999999999', 'hello@example.com',
			'https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3502.0!2d77.0!3d28.0!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x0%3A0x0!2zMjjCsDAwJzAwLjAiTiA3N8KwMDAnMDAuMCJF!5e0!3m2!1sen!2sin!4v1234567890', $3::jsonb)
	`, uuid.New().String(), shopID, storeHours)
	if err != nil {
		return fmt.Errorf("contact_pages: %w", err)
	}

	return nil
}

// GetAllShops returns all shops (for backward compatibility; prefer GetShopsByOwnerID for admin)
func GetAllShops() ([]models.Shop, error) {
	rows, err := DB.QueryContext(context.Background(),
		`SELECT id, name, subdomain, custom_domain, owner_id, created_at, status FROM shops ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var shops []models.Shop
	for rows.Next() {
		var s models.Shop
		var ownerID *int
		if err := rows.Scan(&s.ID, &s.Name, &s.Subdomain, &s.CustomDomain, &ownerID, &s.CreatedAt, &s.Status); err != nil {
			return nil, err
		}
		if ownerID != nil {
			s.OwnerID = *ownerID
		}
		shops = append(shops, s)
	}
	return shops, nil
}

// GetShopsByOwnerID returns shops owned by the given user (owner_id = users.id)
func GetShopsByOwnerID(ownerID int) ([]models.Shop, error) {
	rows, err := DB.QueryContext(context.Background(),
		`SELECT id, name, subdomain, custom_domain, owner_id, created_at, status FROM shops WHERE owner_id = $1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var shops []models.Shop
	for rows.Next() {
		var s models.Shop
		var ownerIDOut *int
		if err := rows.Scan(&s.ID, &s.Name, &s.Subdomain, &s.CustomDomain, &ownerIDOut, &s.CreatedAt, &s.Status); err != nil {
			return nil, err
		}
		if ownerIDOut != nil {
			s.OwnerID = *ownerIDOut
		}
		shops = append(shops, s)
	}
	return shops, nil
}

// VerifyShopOwnership returns nil if the shop belongs to the given owner (user id), else error
func VerifyShopOwnership(shopID string, ownerID int) error {
	var dummy int
	err := DB.QueryRowContext(context.Background(),
		`SELECT 1 FROM shops WHERE id = $1 AND owner_id = $2`, shopID, ownerID).Scan(&dummy)
	if err != nil {
		return fmt.Errorf("shop not found or access denied")
	}
	return nil
}

// GetShopByID returns a shop by ID
func GetShopByID(id string) (*models.Shop, error) {
	var shop models.Shop
	var ownerID *int
	err := DB.QueryRowContext(context.Background(),
		`SELECT id, name, subdomain, custom_domain, owner_id, created_at, status FROM shops WHERE id = $1`, id,
	).Scan(&shop.ID, &shop.Name, &shop.Subdomain, &shop.CustomDomain, &ownerID, &shop.CreatedAt, &shop.Status)
	if err != nil {
		return nil, err
	}
	if ownerID != nil {
		shop.OwnerID = *ownerID
	}
	return &shop, nil
}

// UpdateShop updates a shop (caller must verify ownership before calling)
func UpdateShop(id string, name, subdomain string, customDomain *string) (*models.Shop, error) {
	query := `
		UPDATE shops SET name = $1, subdomain = $2, custom_domain = $3 WHERE id = $4
		RETURNING id, name, subdomain, custom_domain, owner_id, created_at, status
	`
	var shop models.Shop
	var ownerID *int
	err := DB.QueryRowContext(context.Background(), query, name, subdomain, customDomain, id).Scan(
		&shop.ID, &shop.Name, &shop.Subdomain, &shop.CustomDomain, &ownerID, &shop.CreatedAt, &shop.Status,
	)
	if err != nil {
		return nil, err
	}
	if ownerID != nil {
		shop.OwnerID = *ownerID
	}
	return &shop, nil
}

// GetSiteConfigByShopID returns site config for a shop
func GetSiteConfigByShopID(shopID string) (*models.SiteConfig, error) {
	var c models.SiteConfig
	query := `
		SELECT id, shop_id, shop_name, tagline, hero_image_url, hero_title, hero_subtitle, logo_url,
			primary_color, gold_color, secondary_color, text_color, text_muted,
			phone_number, whatsapp_number, email, store_address, store_address_short,
			instagram_url, facebook_url, pinterest_url, twitter_url, google_map_url, created_at, updated_at
		FROM site_configs WHERE shop_id = $1
	`
	err := DB.QueryRowContext(context.Background(), query, shopID).Scan(
		&c.ID, &c.ShopID, &c.ShopName, &c.Tagline, &c.HeroImageURL, &c.HeroTitle, &c.HeroSubtitle, &c.LogoURL,
		&c.PrimaryColor, &c.GoldColor, &c.SecondaryColor, &c.TextColor, &c.TextMuted,
		&c.PhoneNumber, &c.WhatsappNumber, &c.Email, &c.StoreAddress, &c.StoreAddressShort,
		&c.InstagramURL, &c.FacebookURL, &c.PinterestURL, &c.TwitterURL, &c.GoogleMapURL, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CreateDefaultSiteConfig inserts a default site config for a shop and returns it.
// Returns error if shop does not exist (prevents FK violation).
func CreateDefaultSiteConfig(shopID string) (*models.SiteConfig, error) {
	var exists bool
	if err := DB.QueryRowContext(context.Background(), `SELECT EXISTS(SELECT 1 FROM shops WHERE id = $1)`, shopID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check shop exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("shop does not exist")
	}
	query := `
		INSERT INTO site_configs (shop_id, shop_name, tagline, primary_color, gold_color, text_color, text_muted)
		VALUES ($1, '', '', '#635BFF', '#D4AF37', '#1A1A1A', '#9CA3AF')
		RETURNING id, shop_id, shop_name, tagline, hero_image_url, hero_title, hero_subtitle, logo_url,
			primary_color, gold_color, secondary_color, text_color, text_muted,
			phone_number, whatsapp_number, email, store_address, store_address_short,
			instagram_url, facebook_url, pinterest_url, twitter_url, google_map_url, created_at, updated_at
	`
	var c models.SiteConfig
	err := DB.QueryRowContext(context.Background(), query, shopID).Scan(
		&c.ID, &c.ShopID, &c.ShopName, &c.Tagline, &c.HeroImageURL, &c.HeroTitle, &c.HeroSubtitle, &c.LogoURL,
		&c.PrimaryColor, &c.GoldColor, &c.SecondaryColor, &c.TextColor, &c.TextMuted,
		&c.PhoneNumber, &c.WhatsappNumber, &c.Email, &c.StoreAddress, &c.StoreAddressShort,
		&c.InstagramURL, &c.FacebookURL, &c.PinterestURL, &c.TwitterURL, &c.GoogleMapURL, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpsertSiteConfig creates or updates site config for a shop
func UpsertSiteConfig(req models.CreateSiteConfigRequest) (*models.SiteConfig, error) {
	query := `
		INSERT INTO site_configs (shop_id, shop_name, tagline, primary_color, gold_color, text_color, text_muted,
			phone_number, whatsapp_number, store_address, store_address_short,
			instagram_url, facebook_url, pinterest_url, google_map_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (shop_id) DO UPDATE SET
			shop_name = EXCLUDED.shop_name, tagline = EXCLUDED.tagline,
			primary_color = EXCLUDED.primary_color, gold_color = EXCLUDED.gold_color,
			text_color = EXCLUDED.text_color, text_muted = EXCLUDED.text_muted,
			phone_number = EXCLUDED.phone_number, whatsapp_number = EXCLUDED.whatsapp_number,
			store_address = EXCLUDED.store_address, store_address_short = EXCLUDED.store_address_short,
			instagram_url = EXCLUDED.instagram_url, facebook_url = EXCLUDED.facebook_url,
			pinterest_url = EXCLUDED.pinterest_url, google_map_url = EXCLUDED.google_map_url,
			updated_at = NOW()
		RETURNING id, shop_id, shop_name, tagline, hero_image_url, hero_title, hero_subtitle, logo_url,
			primary_color, gold_color, secondary_color, text_color, text_muted,
			phone_number, whatsapp_number, email, store_address, store_address_short,
			instagram_url, facebook_url, pinterest_url, twitter_url, google_map_url, created_at, updated_at
	`
	var c models.SiteConfig
	err := DB.QueryRowContext(context.Background(), query,
		req.ShopID, req.ShopName, req.Tagline, req.PrimaryColor, req.GoldColor, req.TextColor, req.TextMuted,
		req.PhoneNumber, req.WhatsappNumber, req.StoreAddress, req.StoreAddressShort,
		req.InstagramURL, req.FacebookURL, req.PinterestURL, req.GoogleMapURL,
	).Scan(&c.ID, &c.ShopID, &c.ShopName, &c.Tagline, &c.HeroImageURL, &c.HeroTitle, &c.HeroSubtitle, &c.LogoURL,
		&c.PrimaryColor, &c.GoldColor, &c.SecondaryColor, &c.TextColor, &c.TextMuted,
		&c.PhoneNumber, &c.WhatsappNumber, &c.Email, &c.StoreAddress, &c.StoreAddressShort,
		&c.InstagramURL, &c.FacebookURL, &c.PinterestURL, &c.TwitterURL, &c.GoogleMapURL, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCategoriesByShopID returns categories for a shop
func GetCategoriesByShopID(shopID string) ([]models.SiteCategory, error) {
	rows, err := DB.QueryContext(context.Background(),
		`SELECT id, shop_id, name, slug, image_url, created_at, updated_at FROM categories WHERE shop_id = $1 ORDER BY name`,
		shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.SiteCategory
	for rows.Next() {
		var c models.SiteCategory
		if err := rows.Scan(&c.ID, &c.ShopID, &c.Name, &c.Slug, &c.ImageURL, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

// InsertSiteCategory creates a category
func InsertSiteCategory(req models.CreateCategoryRequest) (*models.SiteCategory, error) {
	id := uuid.New().String()
	query := `INSERT INTO categories (id, shop_id, name, slug, image_url) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, shop_id, name, slug, image_url, created_at, updated_at`
	var c models.SiteCategory
	err := DB.QueryRowContext(context.Background(), query, id, req.ShopID, req.Name, req.Slug, req.ImageURL).Scan(
		&c.ID, &c.ShopID, &c.Name, &c.Slug, &c.ImageURL, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateSiteCategory updates a category (scoped by shop_id)
func UpdateSiteCategory(id, shopID string, req models.UpdateCategoryRequest) (*models.SiteCategory, error) {
	query := `UPDATE categories SET name = $1, slug = $2, image_url = $3, updated_at = NOW()
		WHERE id = $4 AND shop_id = $5 RETURNING id, shop_id, name, slug, image_url, created_at, updated_at`
	var c models.SiteCategory
	err := DB.QueryRowContext(context.Background(), query, req.Name, req.Slug, req.ImageURL, id, shopID).Scan(
		&c.ID, &c.ShopID, &c.Name, &c.Slug, &c.ImageURL, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// DeleteSiteCategory deletes a category (scoped by shop_id)
func DeleteSiteCategory(id, shopID string) error {
	result, err := DB.ExecContext(context.Background(), `DELETE FROM categories WHERE id = $1 AND shop_id = $2`, id, shopID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

// GetSiteProductsByShopID returns products for a shop (with images)
func GetSiteProductsByShopID(shopID string) ([]models.SiteProduct, error) {
	rows, err := DB.QueryContext(context.Background(),
		`SELECT id, shop_id, category_id, name, description, price, slug, featured, created_at, updated_at FROM products_sites WHERE shop_id = $1 ORDER BY created_at DESC`,
		shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.SiteProduct
	for rows.Next() {
		var p models.SiteProduct
		var catID *string
		if err := rows.Scan(&p.ID, &p.ShopID, &catID, &p.Name, &p.Description, &p.Price, &p.Slug, &p.Featured, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.CategoryID = catID
		imgs, _ := GetProductImagesByProductID(p.ID)
		for _, img := range imgs {
			p.Images = append(p.Images, img.ImageURL)
		}
		variants, _ := GetProductVariantsByProduct(p.ID)
		if len(variants) > 0 {
			p.Variants = variants
		}
		list = append(list, p)
	}
	return list, nil
}

// GetSiteProductByID returns a single product with images (scoped by shop_id)
func GetSiteProductByID(id, shopID string) (*models.SiteProduct, error) {
	var p models.SiteProduct
	var catID *string
	err := DB.QueryRowContext(context.Background(),
		`SELECT id, shop_id, category_id, name, description, price, slug, featured, created_at, updated_at FROM products_sites WHERE id = $1 AND shop_id = $2`, id, shopID,
	).Scan(&p.ID, &p.ShopID, &catID, &p.Name, &p.Description, &p.Price, &p.Slug, &p.Featured, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.CategoryID = catID
	imgs, _ := GetProductImagesByProductID(p.ID)
	for _, img := range imgs {
		p.Images = append(p.Images, img.ImageURL)
	}
	variants, _ := GetProductVariantsByProduct(p.ID)
	if len(variants) > 0 {
		p.Variants = variants
	}
	return &p, nil
}

// GetProductImagesByProductID returns images for a product
func GetProductImagesByProductID(productID string) ([]models.ProductImage, error) {
	rows, err := DB.QueryContext(context.Background(),
		`SELECT id, product_id, image_url, COALESCE(position, 0) AS position, created_at FROM product_images WHERE product_id = $1 ORDER BY position, created_at`,
		productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.ProductImage
	for rows.Next() {
		var img models.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.ImageURL, &img.Position, &img.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, img)
	}
	return list, nil
}

// InsertSiteProduct creates a product and its images
func InsertSiteProduct(req models.CreateSiteProductRequest) (*models.SiteProduct, error) {
	id := uuid.New().String()
	query := `INSERT INTO products_sites (id, shop_id, category_id, name, description, price) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, shop_id, category_id, name, description, price, slug, featured, created_at, updated_at`
	var p models.SiteProduct
	err := DB.QueryRowContext(context.Background(), query, id, req.ShopID, req.CategoryID, req.Name, req.Description, req.Price).Scan(
		&p.ID, &p.ShopID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.Slug, &p.Featured, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	for _, url := range req.ImageURLs {
		_ = InsertProductImage(p.ID, url)
	}
	imgs, _ := GetProductImagesByProductID(p.ID)
	for _, img := range imgs {
		p.Images = append(p.Images, img.ImageURL)
	}
	return &p, nil
}

// InsertProductImage adds an image to a product
func InsertProductImage(productID, imageURL string) error {
	query := `INSERT INTO product_images (product_id, image_url) VALUES ($1, $2)`
	_, err := DB.ExecContext(context.Background(), query, productID, imageURL)
	return err
}

// --- Product Variants ---

// CreateProductVariant creates a new product variant
func CreateProductVariant(req models.CreateProductVariantRequest) (*models.ProductVariant, error) {
	id := uuid.New().String()
	query := `
		INSERT INTO product_variants (id, product_id, name, price, stock, sku)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, product_id, name, price, stock, sku, created_at
	`
	var v models.ProductVariant
	err := DB.QueryRowContext(context.Background(), query,
		id, req.ProductID, req.Name, req.Price, req.Stock, req.SKU,
	).Scan(&v.ID, &v.ProductID, &v.Name, &v.Price, &v.Stock, &v.SKU, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// GetProductVariantsByProduct returns variants for a product
func GetProductVariantsByProduct(productID string) ([]models.ProductVariant, error) {
	rows, err := DB.QueryContext(context.Background(),
		`SELECT id, product_id, name, price, stock, sku, created_at FROM product_variants WHERE product_id = $1 ORDER BY created_at`,
		productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.ProductVariant
	for rows.Next() {
		var v models.ProductVariant
		if err := rows.Scan(&v.ID, &v.ProductID, &v.Name, &v.Price, &v.Stock, &v.SKU, &v.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

// GetProductVariantByID returns a single variant by id
func GetProductVariantByID(id string) (*models.ProductVariant, error) {
	var v models.ProductVariant
	err := DB.QueryRowContext(context.Background(),
		`SELECT id, product_id, name, price, stock, sku, created_at FROM product_variants WHERE id = $1`,
		id,
	).Scan(&v.ID, &v.ProductID, &v.Name, &v.Price, &v.Stock, &v.SKU, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// UpdateProductVariant updates an existing product variant
func UpdateProductVariant(id string, req models.UpdateProductVariantRequest) (*models.ProductVariant, error) {
	query := `
		UPDATE product_variants
		SET name = $1, price = $2, stock = $3, sku = $4
		WHERE id = $5
		RETURNING id, product_id, name, price, stock, sku, created_at
	`
	var v models.ProductVariant
	err := DB.QueryRowContext(context.Background(), query,
		req.Name, req.Price, req.Stock, req.SKU, id,
	).Scan(&v.ID, &v.ProductID, &v.Name, &v.Price, &v.Stock, &v.SKU, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// DeleteProductVariant deletes a product variant
func DeleteProductVariant(id string) error {
	result, err := DB.ExecContext(context.Background(),
		`DELETE FROM product_variants WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("variant not found")
	}
	return nil
}

// VerifyProductOwnership ensures the given product belongs to a shop owned by the user
func VerifyProductOwnership(productID string, ownerID int) error {
	var dummy int
	err := DB.QueryRowContext(context.Background(), `
		SELECT 1
		FROM products_sites p
		JOIN shops s ON p.shop_id = s.id
		WHERE p.id = $1 AND s.owner_id = $2
	`, productID, ownerID).Scan(&dummy)
	if err != nil {
		return fmt.Errorf("product not found or access denied")
	}
	return nil
}

// --- Payment Config ---

// GetPaymentConfigByShopID returns payment configuration for a shop
func GetPaymentConfigByShopID(shopID string) (*models.PaymentConfigResponse, error) {
	var paymentEnabled bool
	var razorpayKeyID *string
	err := DB.QueryRowContext(context.Background(),
		`SELECT COALESCE(payment_enabled, false) AS payment_enabled, razorpay_key_id FROM site_configs WHERE shop_id = $1`,
		shopID,
	).Scan(&paymentEnabled, &razorpayKeyID)
	if err != nil {
		return nil, err
	}
	return &models.PaymentConfigResponse{
		PaymentEnabled: paymentEnabled,
		RazorpayKey:    razorpayKeyID,
	}, nil
}

// --- Site Orders ---

// CreateSiteOrder creates an order and its items in a transaction
func CreateSiteOrder(req models.CreateSiteOrderRequest) (*models.CreateSiteOrderResponse, error) {
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}

	if req.PaymentMethod != "COD" && req.PaymentMethod != "RAZORPAY" {
		return nil, fmt.Errorf("invalid payment_method")
	}

	// Validate shop exists
	var shopExists bool
	if err := DB.QueryRowContext(context.Background(), `SELECT EXISTS(SELECT 1 FROM shops WHERE id = $1)`, req.ShopID).Scan(&shopExists); err != nil {
		return nil, fmt.Errorf("check shop exists: %w", err)
	}
	if !shopExists {
		return nil, fmt.Errorf("shop does not exist")
	}

	tx, err := DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAmount := 0.0

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("quantity must be positive")
		}
		var price float64
		// Fetch product price scoped by shop
		var productName string
		err := tx.QueryRowContext(context.Background(), `
			SELECT name, price FROM products_sites WHERE id = $1 AND shop_id = $2
		`, item.ProductID, req.ShopID).Scan(&productName, &price)
		if err != nil {
			return nil, fmt.Errorf("product not found")
		}

		var variantName *string
		if item.VariantID != nil && *item.VariantID != "" {
			var vPrice float64
			err = tx.QueryRowContext(context.Background(), `
				SELECT name, price FROM product_variants WHERE id = $1 AND product_id = $2
			`, *item.VariantID, item.ProductID).Scan(&variantName, &vPrice)
			if err != nil {
				return nil, fmt.Errorf("variant not found")
			}
			price = vPrice
		}

		totalAmount += price * float64(item.Quantity)

		// We'll insert order items after inserting the order (once we have order_id)
		_ = productName
		_ = variantName
	}

	orderID := uuid.New().String()
	paymentStatus := "pending"
	orderStatus := "placed"

	_, err = tx.ExecContext(context.Background(), `
		INSERT INTO site_orders (id, shop_id, customer_name, phone, email, address, city, pincode, total_amount, payment_method, payment_status, order_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, orderID, req.ShopID, req.CustomerName, req.Phone, req.Email, req.Address, req.City, req.Pincode,
		totalAmount, req.PaymentMethod, paymentStatus, orderStatus)
	if err != nil {
		return nil, err
	}

	for _, item := range req.Items {
		var price float64
		var productName string
		err := tx.QueryRowContext(context.Background(), `
			SELECT name, price FROM products_sites WHERE id = $1 AND shop_id = $2
		`, item.ProductID, req.ShopID).Scan(&productName, &price)
		if err != nil {
			return nil, fmt.Errorf("product not found")
		}

		var variantName *string
		if item.VariantID != nil && *item.VariantID != "" {
			var vPrice float64
			err = tx.QueryRowContext(context.Background(), `
				SELECT name, price FROM product_variants WHERE id = $1 AND product_id = $2
			`, *item.VariantID, item.ProductID).Scan(&variantName, &vPrice)
			if err != nil {
				return nil, fmt.Errorf("variant not found")
			}
			price = vPrice
		}

		itemID := uuid.New().String()
		_, err = tx.ExecContext(context.Background(), `
			INSERT INTO site_order_items (id, order_id, product_id, variant_id, product_name, variant_name, price, quantity)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, itemID, orderID, item.ProductID, item.VariantID, productName, variantName, price, item.Quantity)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.CreateSiteOrderResponse{
		OrderID: orderID,
		Amount:  totalAmount,
	}, nil
}

// GetSiteOrdersByShopID returns orders for a shop
func GetSiteOrdersByShopID(shopID string) ([]models.SiteOrder, error) {
	rows, err := DB.QueryContext(context.Background(), `
		SELECT id, shop_id, customer_name, phone, email, address, city, pincode,
		       total_amount, payment_method, payment_status, order_status, created_at
		FROM site_orders
		WHERE shop_id = $1
		ORDER BY created_at DESC
	`, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.SiteOrder
	for rows.Next() {
		var o models.SiteOrder
		if err := rows.Scan(&o.ID, &o.ShopID, &o.CustomerName, &o.Phone, &o.Email, &o.Address, &o.City, &o.Pincode,
			&o.TotalAmount, &o.PaymentMethod, &o.PaymentStatus, &o.OrderStatus, &o.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, nil
}

// GetSiteOrderByID returns a single order with its items
func GetSiteOrderByID(id string) (*models.SiteOrder, error) {
	var o models.SiteOrder
	err := DB.QueryRowContext(context.Background(), `
		SELECT id, shop_id, customer_name, phone, email, address, city, pincode,
		       total_amount, payment_method, payment_status, order_status, created_at
		FROM site_orders
		WHERE id = $1
	`, id).Scan(&o.ID, &o.ShopID, &o.CustomerName, &o.Phone, &o.Email, &o.Address, &o.City, &o.Pincode,
		&o.TotalAmount, &o.PaymentMethod, &o.PaymentStatus, &o.OrderStatus, &o.CreatedAt)
	if err != nil {
		return nil, err
	}

	itemsRows, err := DB.QueryContext(context.Background(), `
		SELECT id, order_id, product_id, variant_id, product_name, variant_name, price, quantity
		FROM site_order_items
		WHERE order_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer itemsRows.Close()

	for itemsRows.Next() {
		var it models.SiteOrderItem
		if err := itemsRows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.VariantID, &it.ProductName, &it.VariantName, &it.Price, &it.Quantity); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, it)
	}

	return &o, nil
}

// UpdateSiteOrderStatus updates the order_status of an order
func UpdateSiteOrderStatus(id string, status string) (*models.SiteOrder, error) {
	allowed := map[string]bool{
		"placed":    true,
		"confirmed": true,
		"shipped":   true,
		"delivered": true,
		"cancelled": true,
	}
	if !allowed[status] {
		return nil, fmt.Errorf("invalid order_status")
	}

	var o models.SiteOrder
	err := DB.QueryRowContext(context.Background(), `
		UPDATE site_orders
		SET order_status = $1
		WHERE id = $2
		RETURNING id, shop_id, customer_name, phone, email, address, city, pincode,
		          total_amount, payment_method, payment_status, order_status, created_at
	`, status, id).Scan(&o.ID, &o.ShopID, &o.CustomerName, &o.Phone, &o.Email, &o.Address, &o.City, &o.Pincode,
		&o.TotalAmount, &o.PaymentMethod, &o.PaymentStatus, &o.OrderStatus, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// UpdateSiteProduct updates a product and replaces images (scoped by shop_id)
func UpdateSiteProduct(id, shopID string, req models.UpdateSiteProductRequest) (*models.SiteProduct, error) {
	query := `UPDATE products_sites SET category_id = $1, name = $2, description = $3, price = $4, updated_at = NOW()
		WHERE id = $5 AND shop_id = $6 RETURNING id, shop_id, category_id, name, description, price, slug, featured, created_at, updated_at`
	var p models.SiteProduct
	err := DB.QueryRowContext(context.Background(), query, req.CategoryID, req.Name, req.Description, req.Price, id, shopID).Scan(
		&p.ID, &p.ShopID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.Slug, &p.Featured, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	// Replace images: delete existing, insert new
	_, _ = DB.ExecContext(context.Background(), `DELETE FROM product_images WHERE product_id = $1`, id)
	for _, url := range req.ImageURLs {
		_ = InsertProductImage(id, url)
	}
	imgs, _ := GetProductImagesByProductID(id)
	for _, img := range imgs {
		p.Images = append(p.Images, img.ImageURL)
	}
	return &p, nil
}

// DeleteSiteProduct deletes a product (cascade deletes images, scoped by shop_id)
func DeleteSiteProduct(id, shopID string) error {
	result, err := DB.ExecContext(context.Background(), `DELETE FROM products_sites WHERE id = $1 AND shop_id = $2`, id, shopID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

// GetTestimonialsByShopID returns testimonials for a shop
func GetTestimonialsByShopID(shopID string) ([]models.Testimonial, error) {
	rows, err := DB.QueryContext(context.Background(),
		`SELECT id, shop_id, text, author, rating, avatar_url, created_at, updated_at FROM testimonials WHERE shop_id = $1 ORDER BY created_at DESC`,
		shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Testimonial
	for rows.Next() {
		var t models.Testimonial
		if err := rows.Scan(&t.ID, &t.ShopID, &t.Text, &t.Author, &t.Rating, &t.AvatarURL, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

// GetAboutPageByShopID returns about page for a shop
func GetAboutPageByShopID(shopID string) (*models.AboutPage, error) {
	var a models.AboutPage
	query := `SELECT id, shop_id, hero_image_url, title, tagline, story_text, story_text_secondary, story_image_url, values, craftsmanship, created_at, updated_at
		FROM about_pages WHERE shop_id = $1`
	err := DB.QueryRowContext(context.Background(), query, shopID).Scan(
		&a.ID, &a.ShopID, &a.HeroImageURL, &a.Title, &a.Tagline, &a.StoryText, &a.StoryTextSecondary, &a.StoryImageURL,
		&a.Values, &a.Craftsmanship, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// CreateDefaultAboutPage inserts a default about page for a shop and returns it.
// Returns error if shop does not exist (prevents FK violation).
func CreateDefaultAboutPage(shopID string) (*models.AboutPage, error) {
	var exists bool
	if err := DB.QueryRowContext(context.Background(), `SELECT EXISTS(SELECT 1 FROM shops WHERE id = $1)`, shopID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check shop exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("shop does not exist")
	}
	query := `
		INSERT INTO about_pages (id, shop_id, title, story_text)
		VALUES ($1, $2, 'About Us', '')
		RETURNING id, shop_id, hero_image_url, title, tagline, story_text, story_text_secondary, story_image_url, values, craftsmanship, created_at, updated_at
	`
	var a models.AboutPage
	err := DB.QueryRowContext(context.Background(), query, uuid.New().String(), shopID).Scan(
		&a.ID, &a.ShopID, &a.HeroImageURL, &a.Title, &a.Tagline, &a.StoryText, &a.StoryTextSecondary, &a.StoryImageURL,
		&a.Values, &a.Craftsmanship, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// GetContactPageByShopID returns contact page for a shop
func GetContactPageByShopID(shopID string) (*models.ContactPage, error) {
	var c models.ContactPage
	query := `SELECT id, shop_id, title, subtitle, store_address, phone_number, whatsapp_number, email, latitude, longitude, google_map_url, map_embed_url, store_hours, created_at, updated_at
		FROM contact_pages WHERE shop_id = $1`
	err := DB.QueryRowContext(context.Background(), query, shopID).Scan(
		&c.ID, &c.ShopID, &c.Title, &c.Subtitle, &c.StoreAddress, &c.PhoneNumber, &c.WhatsappNumber, &c.Email, &c.Latitude, &c.Longitude, &c.GoogleMapURL, &c.MapEmbedURL, &c.StoreHours, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CreateDefaultContactPage inserts a default contact page for a shop and returns it.
// Returns error if shop does not exist (prevents FK violation).
func CreateDefaultContactPage(shopID string) (*models.ContactPage, error) {
	var exists bool
	if err := DB.QueryRowContext(context.Background(), `SELECT EXISTS(SELECT 1 FROM shops WHERE id = $1)`, shopID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check shop exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("shop does not exist")
	}
	query := `
		INSERT INTO contact_pages (id, shop_id, email, phone_number, store_address)
		VALUES ($1, $2, '', '', '')
		RETURNING id, shop_id, title, subtitle, store_address, phone_number, whatsapp_number, email, latitude, longitude, google_map_url, map_embed_url, store_hours, created_at, updated_at
	`
	var c models.ContactPage
	err := DB.QueryRowContext(context.Background(), query, uuid.New().String(), shopID).Scan(
		&c.ID, &c.ShopID, &c.Title, &c.Subtitle, &c.StoreAddress, &c.PhoneNumber, &c.WhatsappNumber, &c.Email, &c.Latitude, &c.Longitude, &c.GoogleMapURL, &c.MapEmbedURL, &c.StoreHours, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

