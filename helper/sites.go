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

// InsertShop creates a new shop with the given ownerID (user id from users table)
// and automatically seeds default website content via seedShopData.
func InsertShop(req models.CreateShopRequest, ownerID int) (*models.Shop, error) {
	ctx := context.Background()
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	id := uuid.New().String()
	shopQuery := `
		INSERT INTO shops (id, name, subdomain, custom_domain, owner_id, status)
		VALUES ($1, $2, $3, $4, $5, 'active')
		RETURNING id, name, subdomain, custom_domain, owner_id, created_at, status
	`
	var shop models.Shop
	var ownerIDOut *int
	err = tx.QueryRowContext(ctx, shopQuery, id, req.Name, req.Subdomain, req.CustomDomain, ownerID).Scan(
		&shop.ID, &shop.Name, &shop.Subdomain, &shop.CustomDomain, &ownerIDOut, &shop.CreatedAt, &shop.Status,
	)
	if err != nil {
		return nil, err
	}
	if ownerIDOut != nil {
		shop.OwnerID = *ownerIDOut
	}

	if err := seedShopData(ctx, tx, id, req.Name); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &shop, nil
}

// seedShopData inserts default website content for a new shop so the site (shopname.oryoo.in)
// has a complete layout with demo content. Customers can edit or delete from Sites Admin.
// Uses Unsplash images, lorem ipsum text, and varied product prices for a realistic demo.
func seedShopData(ctx context.Context, tx *sql.Tx, shopID, shopName string) error {
	// site_configs - full defaults with lorem ipsum tagline
	_, err := tx.ExecContext(ctx, `
		INSERT INTO site_configs (shop_id, shop_name, tagline, primary_color, gold_color, text_color, text_muted,
			store_address, store_address_short, phone_number, whatsapp_number, instagram_url, facebook_url, pinterest_url)
		VALUES ($1, $2, 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Crafted with passion.', '#635BFF', '#D4AF37', '#1A1A1A', '#9CA3AF',
			'123 Jewelry Lane, Your City', '123 Jewelry Lane', '+1 234 567 8900', '+1 234 567 8900',
			'https://instagram.com', 'https://facebook.com', 'https://pinterest.com')
	`, shopID, shopName)
	if err != nil {
		return fmt.Errorf("site_configs: %w", err)
	}

	// categories - multiple with Unsplash images
	categories := []struct {
		id   string
		name string
		slug string
		img  string
	}{
		{uuid.New().String(), "Rings", "rings", "https://images.unsplash.com/photo-1605100804763-247f67b3557e?w=800"},
		{uuid.New().String(), "Necklaces", "necklaces", "https://images.unsplash.com/photo-1599643478518-a784e5dc4c8f?w=800"},
		{uuid.New().String(), "Bracelets", "bracelets", "https://images.unsplash.com/photo-1611652022419-a9419f74343a?w=800"},
		{uuid.New().String(), "Earrings", "earrings", "https://images.unsplash.com/photo-1573408301185-9146fe634ad0?w=800"},
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

	// products_sites - multiple products with varied prices (earnings) and lorem ipsum
	products := []struct {
		id          string
		categoryIdx int
		name        string
		desc        string
		price       float64
		images      []string
	}{
		{
			uuid.New().String(), 0,
			"Eternal Gold Ring",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. A timeless piece crafted with precision.",
			24999.00,
			[]string{"https://images.unsplash.com/photo-1605100804763-247f67b3557e?w=800", "https://images.unsplash.com/photo-1515562141207-7a88fb7ce338?w=800"},
		},
		{
			uuid.New().String(), 1,
			"Classic Pearl Necklace",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
			18999.00,
			[]string{"https://images.unsplash.com/photo-1599643478518-a784e5dc4c8f?w=800", "https://images.unsplash.com/photo-1515562141207-7a88fb7ce338?w=800"},
		},
		{
			uuid.New().String(), 2,
			"Elegant Gold Bracelet",
			"Lorem ipsum dolor sit amet, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Duis aute irure dolor in reprehenderit in voluptate.",
			15999.00,
			[]string{"https://images.unsplash.com/photo-1611652022419-a9419f74343a?w=800", "https://images.unsplash.com/photo-1602751584552-8ba73aad10e1?w=800"},
		},
		{
			uuid.New().String(), 3,
			"Diamond Stud Earrings",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			32999.00,
			[]string{"https://images.unsplash.com/photo-1573408301185-9146fe634ad0?w=800", "https://images.unsplash.com/photo-1605100804763-247f67b3557e?w=800"},
		},
	}
	for _, p := range products {
		catID := categories[p.categoryIdx].id
		_, err = tx.ExecContext(ctx, `
			INSERT INTO products_sites (id, shop_id, category_id, name, description, price)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, p.id, shopID, catID, p.name, p.desc, p.price)
		if err != nil {
			return fmt.Errorf("products_sites: %w", err)
		}
		for _, url := range p.images {
			_, err = tx.ExecContext(ctx, `INSERT INTO product_images (product_id, image_url) VALUES ($1, $2)`, p.id, url)
			if err != nil {
				return fmt.Errorf("product_images: %w", err)
			}
		}
	}

	// about_pages - full lorem ipsum content with Unsplash hero and story images
	loremStory := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris."
	loremSecondary := "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
	_, err = tx.ExecContext(ctx, `
		INSERT INTO about_pages (id, shop_id, hero_image_url, title, tagline, story_text, story_text_secondary, story_image_url, values, craftsmanship)
		VALUES ($1, $2, 'https://images.unsplash.com/photo-1515562141207-7a88fb7ce338?w=1200', 'About Us', 'Lorem ipsum dolor sit amet - Our Story of Craftsmanship',
			$3, $4, 'https://images.unsplash.com/photo-1602751584552-8ba73aad10e1?w=800',
			'{"quality":"Lorem ipsum dolor sit amet, consectetur adipiscing elit.","craftsmanship":"Sed do eiusmod tempor incididunt ut labore.","heritage":"Ut enim ad minim veniam, quis nostrud."}'::jsonb,
			'{"process":"Lorem ipsum dolor sit amet, sed do eiusmod tempor.","materials":"Ut labore et dolore magna aliqua.","finishing":"Duis aute irure dolor in reprehenderit."}'::jsonb)
	`, uuid.New().String(), shopID, loremStory, loremSecondary)
	if err != nil {
		return fmt.Errorf("about_pages: %w", err)
	}

	// contact_pages - lorem ipsum subtitle
	_, err = tx.ExecContext(ctx, `
		INSERT INTO contact_pages (id, shop_id, title, subtitle, store_address, phone_number, email)
		VALUES ($1, $2, 'Get in Touch', 'Lorem ipsum dolor sit amet - We would love to hear from you', '123 Jewelry Lane, Your City', '+1 234 567 8900', 'hello@example.com')
	`, uuid.New().String(), shopID)
	if err != nil {
		return fmt.Errorf("contact_pages: %w", err)
	}

	// testimonials - lorem ipsum with Unsplash avatars
	testimonials := []struct {
		text   string
		author string
		rating int
		avatar string
	}{
		{"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Exceptional quality and service!", "Priya S.", 5, "https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=200"},
		{"Ut enim ad minim veniam, quis nostrud exercitation. Beautiful craftsmanship.", "Rahul M.", 5, "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=200"},
		{"Lorem ipsum dolor sit amet, sed do eiusmod tempor incididunt. Highly recommend!", "Anita K.", 5, "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=200"},
		{"Duis aute irure dolor in reprehenderit. Stunning pieces.", "Vikram R.", 5, "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=200"},
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
		SELECT id, shop_id, shop_name, tagline, primary_color, gold_color, text_color, text_muted,
			phone_number, whatsapp_number, store_address, store_address_short,
			instagram_url, facebook_url, pinterest_url, google_map_url, created_at, updated_at
		FROM site_configs WHERE shop_id = $1
	`
	err := DB.QueryRowContext(context.Background(), query, shopID).Scan(
		&c.ID, &c.ShopID, &c.ShopName, &c.Tagline, &c.PrimaryColor, &c.GoldColor, &c.TextColor, &c.TextMuted,
		&c.PhoneNumber, &c.WhatsappNumber, &c.StoreAddress, &c.StoreAddressShort,
		&c.InstagramURL, &c.FacebookURL, &c.PinterestURL, &c.GoogleMapURL, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CreateDefaultSiteConfig inserts a default site config for a shop and returns it
func CreateDefaultSiteConfig(shopID string) (*models.SiteConfig, error) {
	query := `
		INSERT INTO site_configs (shop_id, shop_name, tagline, primary_color, gold_color, text_color, text_muted)
		VALUES ($1, '', '', '#635BFF', '#D4AF37', '#1A1A1A', '#9CA3AF')
		RETURNING id, shop_id, shop_name, tagline, primary_color, gold_color, text_color, text_muted,
			phone_number, whatsapp_number, store_address, store_address_short,
			instagram_url, facebook_url, pinterest_url, google_map_url, created_at, updated_at
	`
	var c models.SiteConfig
	err := DB.QueryRowContext(context.Background(), query, shopID).Scan(
		&c.ID, &c.ShopID, &c.ShopName, &c.Tagline, &c.PrimaryColor, &c.GoldColor, &c.TextColor, &c.TextMuted,
		&c.PhoneNumber, &c.WhatsappNumber, &c.StoreAddress, &c.StoreAddressShort,
		&c.InstagramURL, &c.FacebookURL, &c.PinterestURL, &c.GoogleMapURL, &c.CreatedAt, &c.UpdatedAt,
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
		RETURNING id, shop_id, shop_name, tagline, primary_color, gold_color, text_color, text_muted,
			phone_number, whatsapp_number, store_address, store_address_short,
			instagram_url, facebook_url, pinterest_url, google_map_url, created_at, updated_at
	`
	var c models.SiteConfig
	err := DB.QueryRowContext(context.Background(), query,
		req.ShopID, req.ShopName, req.Tagline, req.PrimaryColor, req.GoldColor, req.TextColor, req.TextMuted,
		req.PhoneNumber, req.WhatsappNumber, req.StoreAddress, req.StoreAddressShort,
		req.InstagramURL, req.FacebookURL, req.PinterestURL, req.GoogleMapURL,
	).Scan(&c.ID, &c.ShopID, &c.ShopName, &c.Tagline, &c.PrimaryColor, &c.GoldColor, &c.TextColor, &c.TextMuted,
		&c.PhoneNumber, &c.WhatsappNumber, &c.StoreAddress, &c.StoreAddressShort,
		&c.InstagramURL, &c.FacebookURL, &c.PinterestURL, &c.GoogleMapURL, &c.CreatedAt, &c.UpdatedAt,
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
		`SELECT id, shop_id, category_id, name, description, price, created_at, updated_at FROM products_sites WHERE shop_id = $1 ORDER BY created_at DESC`,
		shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.SiteProduct
	for rows.Next() {
		var p models.SiteProduct
		var catID *string
		if err := rows.Scan(&p.ID, &p.ShopID, &catID, &p.Name, &p.Description, &p.Price, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.CategoryID = catID
		imgs, _ := GetProductImagesByProductID(p.ID)
		for _, img := range imgs {
			p.Images = append(p.Images, img.ImageURL)
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
		`SELECT id, shop_id, category_id, name, description, price, created_at, updated_at FROM products_sites WHERE id = $1 AND shop_id = $2`, id, shopID,
	).Scan(&p.ID, &p.ShopID, &catID, &p.Name, &p.Description, &p.Price, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.CategoryID = catID
	imgs, _ := GetProductImagesByProductID(p.ID)
	for _, img := range imgs {
		p.Images = append(p.Images, img.ImageURL)
	}
	return &p, nil
}

// GetProductImagesByProductID returns images for a product
func GetProductImagesByProductID(productID string) ([]models.ProductImage, error) {
	rows, err := DB.QueryContext(context.Background(),
		`SELECT id, product_id, image_url, created_at FROM product_images WHERE product_id = $1 ORDER BY created_at`,
		productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.ProductImage
	for rows.Next() {
		var img models.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.ImageURL, &img.CreatedAt); err != nil {
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
		RETURNING id, shop_id, category_id, name, description, price, created_at, updated_at`
	var p models.SiteProduct
	err := DB.QueryRowContext(context.Background(), query, id, req.ShopID, req.CategoryID, req.Name, req.Description, req.Price).Scan(
		&p.ID, &p.ShopID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.CreatedAt, &p.UpdatedAt,
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

// UpdateSiteProduct updates a product and replaces images (scoped by shop_id)
func UpdateSiteProduct(id, shopID string, req models.UpdateSiteProductRequest) (*models.SiteProduct, error) {
	query := `UPDATE products_sites SET category_id = $1, name = $2, description = $3, price = $4, updated_at = NOW()
		WHERE id = $5 AND shop_id = $6 RETURNING id, shop_id, category_id, name, description, price, created_at, updated_at`
	var p models.SiteProduct
	err := DB.QueryRowContext(context.Background(), query, req.CategoryID, req.Name, req.Description, req.Price, id, shopID).Scan(
		&p.ID, &p.ShopID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.CreatedAt, &p.UpdatedAt,
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

// CreateDefaultAboutPage inserts a default about page for a shop and returns it
func CreateDefaultAboutPage(shopID string) (*models.AboutPage, error) {
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
	query := `SELECT id, shop_id, title, subtitle, store_address, phone_number, whatsapp_number, email, google_map_url, store_hours, created_at, updated_at
		FROM contact_pages WHERE shop_id = $1`
	err := DB.QueryRowContext(context.Background(), query, shopID).Scan(
		&c.ID, &c.ShopID, &c.Title, &c.Subtitle, &c.StoreAddress, &c.PhoneNumber, &c.WhatsappNumber, &c.Email, &c.GoogleMapURL, &c.StoreHours, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CreateDefaultContactPage inserts a default contact page for a shop and returns it
func CreateDefaultContactPage(shopID string) (*models.ContactPage, error) {
	query := `
		INSERT INTO contact_pages (id, shop_id, email, phone_number, store_address)
		VALUES ($1, $2, '', '', '')
		RETURNING id, shop_id, title, subtitle, store_address, phone_number, whatsapp_number, email, google_map_url, store_hours, created_at, updated_at
	`
	var c models.ContactPage
	err := DB.QueryRowContext(context.Background(), query, uuid.New().String(), shopID).Scan(
		&c.ID, &c.ShopID, &c.Title, &c.Subtitle, &c.StoreAddress, &c.PhoneNumber, &c.WhatsappNumber, &c.Email, &c.GoogleMapURL, &c.StoreHours, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

