package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Shop model for Oryoo Sites
type Shop struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Subdomain    string     `json:"subdomain"`
	CustomDomain *string    `json:"custom_domain,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	Status       string     `json:"status"`
}

// SiteConfig model
type SiteConfig struct {
	ID                 string  `json:"id"`
	ShopID             string  `json:"shop_id"`
	ShopName           *string `json:"shop_name,omitempty"`
	Tagline            *string `json:"tagline,omitempty"`
	PrimaryColor       *string `json:"primary_color,omitempty"`
	GoldColor          *string `json:"gold_color,omitempty"`
	TextColor          *string `json:"text_color,omitempty"`
	TextMuted          *string `json:"text_muted,omitempty"`
	PhoneNumber        *string `json:"phone_number,omitempty"`
	WhatsappNumber     *string `json:"whatsapp_number,omitempty"`
	StoreAddress       *string `json:"store_address,omitempty"`
	StoreAddressShort  *string `json:"store_address_short,omitempty"`
	InstagramURL       *string `json:"instagram_url,omitempty"`
	FacebookURL        *string `json:"facebook_url,omitempty"`
	PinterestURL       *string `json:"pinterest_url,omitempty"`
	GoogleMapURL       *string `json:"google_map_url,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
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
	ID          string     `json:"id"`
	ShopID      string     `json:"shop_id"`
	CategoryID  *string    `json:"category_id,omitempty"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Price       float64    `json:"price"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Images      []string   `json:"images,omitempty"`
}

// ProductImage model
type ProductImage struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	ImageURL  string    `json:"image_url"`
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
	GoogleMapURL   *string   `json:"google_map_url,omitempty"`
	StoreHours     JSONB     `json:"store_hours,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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

// UploadResponse for POST /sites/admin/upload
type UploadResponse struct {
	URL string `json:"url"`
}
