package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Shop model for Oryoo Sites
type Shop struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Subdomain    string    `json:"subdomain"`
	CustomDomain *string   `json:"custom_domain,omitempty"`
	OwnerID      int       `json:"owner_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// SiteConfig model
type SiteConfig struct {
	ID                string    `json:"id"`
	ShopID            string    `json:"shop_id"`
	ShopName          *string   `json:"shop_name,omitempty"`
	Tagline           *string   `json:"tagline,omitempty"`
	HeroImageURL      *string   `json:"hero_image_url,omitempty"`
	HeroTitle         *string   `json:"hero_title,omitempty"`
	HeroSubtitle      *string   `json:"hero_subtitle,omitempty"`
	LogoURL           *string   `json:"logo_url,omitempty"`
	PrimaryColor      *string   `json:"primary_color,omitempty"`
	GoldColor         *string   `json:"gold_color,omitempty"`
	SecondaryColor    *string   `json:"secondary_color,omitempty"`
	TextColor         *string   `json:"text_color,omitempty"`
	TextMuted         *string   `json:"text_muted,omitempty"`
	PhoneNumber       *string   `json:"phone_number,omitempty"`
	WhatsappNumber    *string   `json:"whatsapp_number,omitempty"`
	Email             *string   `json:"email,omitempty"`
	StoreAddress      *string   `json:"store_address,omitempty"`
	StoreAddressShort *string   `json:"store_address_short,omitempty"`
	InstagramURL      *string   `json:"instagram_url,omitempty"`
	FacebookURL       *string   `json:"facebook_url,omitempty"`
	PinterestURL      *string   `json:"pinterest_url,omitempty"`
	TwitterURL        *string   `json:"twitter_url,omitempty"`
	GoogleMapURL      *string   `json:"google_map_url,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// SiteCategory model
type SiteCategory struct {
	ID        string     `json:"id"`
	ShopID    string     `json:"shop_id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	ImageURL  *string    `json:"image_url,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// SiteProduct model (for Oryoo Sites - distinct from CRM products)
type SiteProduct struct {
	ID         string    `json:"id"`
	ShopID     string    `json:"shop_id"`
	CategoryID *string   `json:"category_id,omitempty"`
	Name       string    `json:"name"`
	Description *string  `json:"description,omitempty"`
	Price      float64   `json:"price"`
	Slug       *string   `json:"slug,omitempty"`
	Featured   bool      `json:"featured"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Images     []string  `json:"images,omitempty"`
	Variants   []ProductVariant `json:"variants,omitempty"`
}

// ProductImage model
type ProductImage struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	ImageURL  string    `json:"image_url"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

// ProductVariant model
type ProductVariant struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	SKU       *string   `json:"sku,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Testimonial model
type Testimonial struct {
	ID        string     `json:"id"`
	ShopID    string     `json:"shop_id"`
	Text      string     `json:"text"`
	Author    string     `json:"author"`
	Rating    int        `json:"rating"`
	AvatarURL *string    `json:"avatar_url,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// AboutPage model
type AboutPage struct {
	ID                string          `json:"id"`
	ShopID            string          `json:"shop_id"`
	HeroImageURL      *string         `json:"hero_image_url,omitempty"`
	Title             *string         `json:"title,omitempty"`
	Tagline           *string         `json:"tagline,omitempty"`
	StoryText         *string         `json:"story_text,omitempty"`
	StoryTextSecondary *string        `json:"story_text_secondary,omitempty"`
	StoryImageURL     *string         `json:"story_image_url,omitempty"`
	Values            JSONB          `json:"values,omitempty"`
	Craftsmanship     JSONB          `json:"craftsmanship,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// ContactPage model
type ContactPage struct {
	ID             string    `json:"id"`
	ShopID         string    `json:"shop_id"`
	Title          *string   `json:"title,omitempty"`
	Subtitle       *string   `json:"subtitle,omitempty"`
	StoreAddress   *string   `json:"store_address,omitempty"`
	PhoneNumber    *string   `json:"phone_number,omitempty"`
	WhatsappNumber *string   `json:"whatsapp_number,omitempty"`
	Email          *string   `json:"email,omitempty"`
	Latitude       *float64  `json:"latitude,omitempty"`
	Longitude      *float64  `json:"longitude,omitempty"`
	GoogleMapURL   *string   `json:"google_map_url,omitempty"`
	MapEmbedURL    *string   `json:"map_embed_url,omitempty"`
	StoreHours     JSONB     `json:"store_hours,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SiteOrder model
type SiteOrder struct {
	ID            string           `json:"id"`
	ShopID        string           `json:"shop_id"`
	CustomerName  string           `json:"customer_name"`
	Phone         string           `json:"phone"`
	Email         *string          `json:"email,omitempty"`
	Address       string           `json:"address"`
	City          *string          `json:"city,omitempty"`
	Pincode       *string          `json:"pincode,omitempty"`
	TotalAmount   float64          `json:"total_amount"`
	PaymentMethod string           `json:"payment_method"`
	PaymentStatus string           `json:"payment_status"`
	OrderStatus   string           `json:"order_status"`
	CreatedAt     time.Time        `json:"created_at"`
	Items         []SiteOrderItem  `json:"items,omitempty"`
}

// SiteOrderItem model
type SiteOrderItem struct {
	ID          string   `json:"id"`
	OrderID     string   `json:"order_id"`
	ProductID   string   `json:"product_id"`
	VariantID   *string  `json:"variant_id,omitempty"`
	ProductName string   `json:"product_name"`
	VariantName *string  `json:"variant_name,omitempty"`
	Price       float64  `json:"price"`
	Quantity    int      `json:"quantity"`
}

// JSONB type for PostgreSQL jsonb columns
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// --- Request/Response structs ---

// ShopByDomainResponse for GET /sites/shop/by-domain
type ShopByDomainResponse struct {
	ShopID   string `json:"shop_id"`
	Shop     *Shop  `json:"shop,omitempty"`
	Found    bool   `json:"found"`
}

// CreateShopRequest for POST /sites/admin/shops
type CreateShopRequest struct {
	OwnerID      int     `json:"owner_id"`
	Name         string  `json:"name"`
	Subdomain    string  `json:"subdomain"`
	CustomDomain *string `json:"custom_domain,omitempty"`
}

// CreateSiteConfigRequest for site config (upsert with shop)
type CreateSiteConfigRequest struct {
	ShopID            string  `json:"shop_id"`
	ShopName          *string `json:"shop_name,omitempty"`
	Tagline           *string `json:"tagline,omitempty"`
	PrimaryColor      *string `json:"primary_color,omitempty"`
	GoldColor         *string `json:"gold_color,omitempty"`
	TextColor         *string `json:"text_color,omitempty"`
	TextMuted         *string `json:"text_muted,omitempty"`
	PhoneNumber       *string `json:"phone_number,omitempty"`
	WhatsappNumber    *string `json:"whatsapp_number,omitempty"`
	StoreAddress      *string `json:"store_address,omitempty"`
	StoreAddressShort *string `json:"store_address_short,omitempty"`
	InstagramURL      *string `json:"instagram_url,omitempty"`
	FacebookURL       *string `json:"facebook_url,omitempty"`
	PinterestURL      *string `json:"pinterest_url,omitempty"`
	GoogleMapURL      *string `json:"google_map_url,omitempty"`
}

// CreateCategoryRequest for POST /sites/admin/categories
type CreateCategoryRequest struct {
	ShopID   string  `json:"shop_id"`
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	ImageURL *string `json:"image_url,omitempty"`
}

// UpdateCategoryRequest for PUT /sites/admin/categories/{id}
type UpdateCategoryRequest struct {
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	ImageURL *string `json:"image_url,omitempty"`
}

// CreateSiteProductRequest for POST /sites/admin/products
type CreateSiteProductRequest struct {
	ShopID      string   `json:"shop_id"`
	CategoryID  *string  `json:"category_id,omitempty"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Price       float64  `json:"price"`
	ImageURLs   []string `json:"image_urls,omitempty"`
}

// UpdateSiteProductRequest for PUT /sites/admin/products/{id}
type UpdateSiteProductRequest struct {
	CategoryID  *string  `json:"category_id,omitempty"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Price       float64  `json:"price"`
	ImageURLs   []string `json:"image_urls,omitempty"`
}

// CreateProductVariantRequest for POST /sites/admin/product-variants
type CreateProductVariantRequest struct {
	ProductID string   `json:"product_id"`
	Name      string   `json:"name"`
	Price     float64  `json:"price"`
	Stock     int      `json:"stock"`
	SKU       *string  `json:"sku,omitempty"`
}

// UpdateProductVariantRequest for PUT /sites/admin/product-variants/{id}
type UpdateProductVariantRequest struct {
	Name  string   `json:"name"`
	Price float64  `json:"price"`
	Stock int      `json:"stock"`
	SKU   *string  `json:"sku,omitempty"`
}

// CreateSiteOrderItemRequest for POST /sites/orders
type CreateSiteOrderItemRequest struct {
	ProductID string  `json:"product_id"`
	VariantID *string `json:"variant_id,omitempty"`
	Quantity  int     `json:"quantity"`
}

// CreateSiteOrderRequest for POST /sites/orders
type CreateSiteOrderRequest struct {
	ShopID        string                      `json:"shop_id"`
	CustomerName  string                      `json:"customer_name"`
	Phone         string                      `json:"phone"`
	Email         *string                     `json:"email,omitempty"`
	Address       string                      `json:"address"`
	City          *string                     `json:"city,omitempty"`
	Pincode       *string                     `json:"pincode,omitempty"`
	Items         []CreateSiteOrderItemRequest `json:"items"`
	PaymentMethod string                      `json:"payment_method"`
}

// CreateSiteOrderResponse returned from POST /sites/orders
type CreateSiteOrderResponse struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

// UpdateSiteOrderStatusRequest for PUT /sites/admin/orders/{id}/status
type UpdateSiteOrderStatusRequest struct {
	OrderStatus string `json:"order_status"`
}

// PaymentConfigResponse for GET /sites/payment-config
type PaymentConfigResponse struct {
	PaymentEnabled bool    `json:"payment_enabled"`
	RazorpayKey    *string `json:"razorpay_key,omitempty"`
}

// UploadResponse for POST /sites/admin/upload
type UploadResponse struct {
	URL string `json:"url"`
}
