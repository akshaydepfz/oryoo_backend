package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"oryoo.com/helper"
	"oryoo.com/models"
)

const adminSecret = "8f3k29df0sdf89sdf98sd7f98sd7f9sd87f"

// isAdminAuth returns true if the request has the admin Bearer token.
func isAdminAuth(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	return auth == "Bearer "+adminSecret
}

// getAuthenticatedUser extracts firebase_uid from Authorization: Bearer <firebase_uid> or X-Firebase-UID header,
// looks up the user, and returns it. Returns error if not authenticated.
func getAuthenticatedUser(r *http.Request) (models.User, error) {
	firebaseUID := r.Header.Get("X-Firebase-UID")
	if firebaseUID == "" {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			firebaseUID = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		}
	}
	if firebaseUID == "" {
		return models.User{}, fmt.Errorf("authentication required: provide Authorization: Bearer <firebase_uid> or X-Firebase-UID header")
	}
	user, err := helper.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return models.User{}, fmt.Errorf("user not found")
	}
	return user, nil
}

// getAuthenticatedUserOrAdmin returns (user, isAdmin, err). If admin Bearer token is used, isAdmin=true and user is zero.
// If Firebase UID is used, isAdmin=false and user is set. Otherwise returns error.
func getAuthenticatedUserOrAdmin(r *http.Request) (models.User, bool, error) {
	if isAdminAuth(r) {
		return models.User{}, true, nil
	}
	user, err := getAuthenticatedUser(r)
	if err != nil {
		return models.User{}, false, err
	}
	return user, false, nil
}

const (
	s3Bucket  = "oryoo-bucket"
	s3Region  = "ap-southeast-2"
	awsKey    = "AKIAYLWS7S6WYP6WYSIA"
	awsSecret = "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"
)

// uploadToS3 uploads file bytes to S3 and returns the public URL (same logic as CreateUserHandler)
func uploadToS3(fileBytes []byte, filename, contentType, folder string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(s3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsKey, awsSecret, "")),
	)
	if err != nil {
		return "", err
	}
	client := s3.NewFromConfig(cfg)
	key := fmt.Sprintf("%s/%d_%s", folder, time.Now().UnixNano(), filename)
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3Bucket, s3Region, key), nil
}

// SitesShopByDomainHandler GET /sites/shop/by-domain?domain=example.com
func SitesShopByDomainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "domain query parameter is required", http.StatusBadRequest)
		return
	}
	shop, err := helper.GetShopByDomain(domain)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ShopByDomainResponse{Found: false})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ShopByDomainResponse{
		ShopID: shop.ID,
		Shop:   shop,
		Found:  true,
	})
}

// SitesProductsHandler GET /sites/products?shop_id=
func SitesProductsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	shopID := r.URL.Query().Get("shop_id")
	if shopID == "" {
		http.Error(w, "shop_id is required", http.StatusBadRequest)
		return
	}
	products, err := helper.GetSiteProductsByShopID(shopID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// SitesProductByIDHandler GET /sites/products/{id}?shop_id=
func SitesProductByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/sites/products/")
	if id == "" {
		http.Error(w, "product id is required", http.StatusBadRequest)
		return
	}
	shopID := r.URL.Query().Get("shop_id")
	if shopID == "" {
		http.Error(w, "shop_id is required", http.StatusBadRequest)
		return
	}
	product, err := helper.GetSiteProductByID(id, shopID)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// SitesCategoriesHandler GET /sites/categories?shop_id=
func SitesCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	shopID := r.URL.Query().Get("shop_id")
	if shopID == "" {
		http.Error(w, "shop_id is required", http.StatusBadRequest)
		return
	}
	categories, err := helper.GetCategoriesByShopID(shopID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// SitesSiteConfigHandler GET /sites/site-config?shop_id=
func SitesSiteConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	shopID := r.URL.Query().Get("shop_id")
	if shopID == "" {
		http.Error(w, "shop_id is required", http.StatusBadRequest)
		return
	}
	config, err := helper.GetSiteConfigByShopID(shopID)
	if err != nil {
		if err == sql.ErrNoRows {
			config, err = helper.CreateDefaultSiteConfig(shopID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// SitesAboutHandler GET /sites/about?shop_id=
func SitesAboutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	shopID := r.URL.Query().Get("shop_id")
	if shopID == "" {
		http.Error(w, "shop_id is required", http.StatusBadRequest)
		return
	}
	about, err := helper.GetAboutPageByShopID(shopID)
	if err != nil {
		if err == sql.ErrNoRows {
			about, err = helper.CreateDefaultAboutPage(shopID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(about)
}

// SitesContactHandler GET /sites/contact?shop_id=
func SitesContactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	shopID := r.URL.Query().Get("shop_id")
	if shopID == "" {
		http.Error(w, "shop_id is required", http.StatusBadRequest)
		return
	}
	contact, err := helper.GetContactPageByShopID(shopID)
	if err != nil {
		if err == sql.ErrNoRows {
			contact, err = helper.CreateDefaultContactPage(shopID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contact)
}

// SitesTestimonialsHandler GET /sites/testimonials?shop_id=
func SitesTestimonialsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	shopID := r.URL.Query().Get("shop_id")
	if shopID == "" {
		http.Error(w, "shop_id is required", http.StatusBadRequest)
		return
	}
	testimonials, err := helper.GetTestimonialsByShopID(shopID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testimonials)
}

// SitesPaymentConfigHandler GET /sites/payment-config?shop_id=
func SitesPaymentConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	shopID := r.URL.Query().Get("shop_id")
	if shopID == "" {
		http.Error(w, "shop_id is required", http.StatusBadRequest)
		return
	}
	cfg, err := helper.GetPaymentConfigByShopID(shopID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No config yet: return disabled
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.PaymentConfigResponse{
				PaymentEnabled: false,
				RazorpayKey:    nil,
			})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

// SitesAdminUploadHandler POST /sites/admin/upload - multipart file upload to S3
func SitesAdminUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, _, err := getAuthenticatedUserOrAdmin(r); err != nil {
		http.Error(w, "authentication required: provide admin Bearer token or Authorization: Bearer <firebase_uid>", http.StatusUnauthorized)
		return
	}
	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		file, fileHeader, err = r.FormFile("image")
	}
	if err != nil {
		http.Error(w, "file or image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}

	contentType := http.DetectContentType(fileBytes)
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "Only image files are allowed", http.StatusBadRequest)
		return
	}

	url, err := uploadToS3(fileBytes, fileHeader.Filename, contentType, "sites")
	if err != nil {
		http.Error(w, "Image upload failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.UploadResponse{URL: url})
}

// SitesAdminShopsHandler POST /sites/admin/shops, GET /sites/admin/shops
func SitesAdminShopsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		user, isAdmin, err := getAuthenticatedUserOrAdmin(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		var req models.CreateShopRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if req.OwnerID <= 0 {
			http.Error(w, "owner_id is required and must be a positive number", http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.Subdomain == "" {
			http.Error(w, "name and subdomain are required", http.StatusBadRequest)
			return
		}
		if !isAdmin {
			if req.OwnerID != int(user.ID) {
				http.Error(w, "owner_id must match your user id", http.StatusForbidden)
				return
			}
		}
		if err := helper.ValidateOwnerID(req.OwnerID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// 1. Insert the shop first; confirm insert succeeded
		shop, err := helper.InsertShop(req, req.OwnerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 2. Only AFTER successful insert: seed default site content
		if err := helper.CreateDefaultSiteContent(shop.ID); err != nil {
			log.Println("default site content failed:", err)
		}
		// 3. Return the shop response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(shop)
		return
	}
	if r.Method == http.MethodGet {
		if isAdminAuth(r) {
			shops, err := helper.GetAllShops()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(shops)
			return
		}
		firebaseUID := r.Header.Get("X-Firebase-UID")
		if firebaseUID == "" {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				firebaseUID = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			}
		}
		if firebaseUID == "" {
			http.Error(w, "authentication required: provide X-Firebase-UID header or Authorization: Bearer <firebase_uid> or admin Bearer token", http.StatusUnauthorized)
			return
		}
		userID, err := helper.GetUserIDByFirebaseUID(firebaseUID)
		if err != nil {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}
		shops, err := helper.GetShopsByOwnerID(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(shops)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// SitesAdminShopsByIDHandler PUT /sites/admin/shops/{id}
func SitesAdminShopsByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, isAdmin, err := getAuthenticatedUserOrAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/sites/admin/shops/")
	if id == "" || !isValidUUID(id) {
		http.Error(w, "Invalid shop ID", http.StatusBadRequest)
		return
	}
	if !isAdmin {
		if err := helper.VerifyShopOwnership(id, int(user.ID)); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
	var req struct {
		Name         string  `json:"name"`
		Subdomain    string  `json:"subdomain"`
		CustomDomain *string `json:"custom_domain,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Subdomain == "" {
		http.Error(w, "name and subdomain are required", http.StatusBadRequest)
		return
	}
	shop, err := helper.UpdateShop(id, req.Name, req.Subdomain, req.CustomDomain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shop)
}

// SitesAdminProductsHandler POST /sites/admin/products
func SitesAdminProductsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, isAdmin, err := getAuthenticatedUserOrAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var req models.CreateSiteProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.ShopID == "" || req.Name == "" {
		http.Error(w, "shop_id and name are required", http.StatusBadRequest)
		return
	}
	if !isAdmin {
		if err := helper.VerifyShopOwnership(req.ShopID, int(user.ID)); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
	product, err := helper.InsertSiteProduct(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// SitesAdminProductsByIDHandler PUT /sites/admin/products/{id}?shop_id=, DELETE /sites/admin/products/{id}?shop_id=
func SitesAdminProductsByIDHandler(w http.ResponseWriter, r *http.Request) {
	user, isAdmin, err := getAuthenticatedUserOrAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/sites/admin/products/")
	if id == "" || !isValidUUID(id) {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	shopID := r.URL.Query().Get("shop_id")
	if shopID == "" {
		http.Error(w, "shop_id is required", http.StatusBadRequest)
		return
	}
	if !isAdmin {
		if err := helper.VerifyShopOwnership(shopID, int(user.ID)); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
	if r.Method == http.MethodPut {
		var req models.UpdateSiteProductRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		product, err := helper.UpdateSiteProduct(id, shopID, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
		return
	}
	if r.Method == http.MethodDelete {
		if err := helper.DeleteSiteProduct(id, shopID); err != nil {
			if err.Error() == "product not found" {
				http.Error(w, "Product not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Product deleted successfully", "id": id})
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// SitesAdminCategoriesHandler POST /sites/admin/categories
func SitesAdminCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, isAdmin, err := getAuthenticatedUserOrAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.ShopID == "" || req.Name == "" || req.Slug == "" {
		http.Error(w, "shop_id, name and slug are required", http.StatusBadRequest)
		return
	}
	if !isAdmin {
		if err := helper.VerifyShopOwnership(req.ShopID, int(user.ID)); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
	category, err := helper.InsertSiteCategory(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// SitesAdminCategoriesByIDHandler PUT /sites/admin/categories/{id}?shop_id=, DELETE /sites/admin/categories/{id}?shop_id=
func SitesAdminCategoriesByIDHandler(w http.ResponseWriter, r *http.Request) {
	user, isAdmin, err := getAuthenticatedUserOrAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/sites/admin/categories/")
	if id == "" || !isValidUUID(id) {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	shopID := r.URL.Query().Get("shop_id")
	if shopID == "" {
		http.Error(w, "shop_id is required", http.StatusBadRequest)
		return
	}
	if !isAdmin {
		if err := helper.VerifyShopOwnership(shopID, int(user.ID)); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
	if r.Method == http.MethodPut {
		var req models.UpdateCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.Slug == "" {
			http.Error(w, "name and slug are required", http.StatusBadRequest)
			return
		}
		category, err := helper.UpdateSiteCategory(id, shopID, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(category)
		return
	}
	if r.Method == http.MethodDelete {
		if err := helper.DeleteSiteCategory(id, shopID); err != nil {
			if err.Error() == "category not found" {
				http.Error(w, "Category not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Category deleted successfully", "id": id})
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// SitesAdminProductVariantsHandler POST /sites/admin/product-variants, GET /sites/admin/product-variants?product_id=
func SitesAdminProductVariantsHandler(w http.ResponseWriter, r *http.Request) {
	user, isAdmin, err := getAuthenticatedUserOrAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost {
		var req models.CreateProductVariantRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if req.ProductID == "" || req.Name == "" {
			http.Error(w, "product_id and name are required", http.StatusBadRequest)
			return
		}
		if !isAdmin {
			if err := helper.VerifyProductOwnership(req.ProductID, int(user.ID)); err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
		}
		variant, err := helper.CreateProductVariant(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(variant)
		return
	}

	if r.Method == http.MethodGet {
		productID := r.URL.Query().Get("product_id")
		if productID == "" {
			http.Error(w, "product_id is required", http.StatusBadRequest)
			return
		}
		if !isAdmin {
			if err := helper.VerifyProductOwnership(productID, int(user.ID)); err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
		}
		variants, err := helper.GetProductVariantsByProduct(productID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(variants)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// SitesAdminProductVariantsByIDHandler PUT /sites/admin/product-variants/{id}, DELETE /sites/admin/product-variants/{id}
func SitesAdminProductVariantsByIDHandler(w http.ResponseWriter, r *http.Request) {
	user, isAdmin, err := getAuthenticatedUserOrAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/sites/admin/product-variants/")
	if id == "" || !isValidUUID(id) {
		http.Error(w, "Invalid variant ID", http.StatusBadRequest)
		return
	}

	// Verify ownership via product
	var productID string
	v, err := helper.GetProductVariantByID(id)
	if err != nil {
		http.Error(w, "Variant not found", http.StatusNotFound)
		return
	}
	productID = v.ProductID
	if !isAdmin {
		if err := helper.VerifyProductOwnership(productID, int(user.ID)); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}

	if r.Method == http.MethodPut {
		var req models.UpdateProductVariantRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		updated, err := helper.UpdateProductVariant(id, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
		return
	}

	if r.Method == http.MethodDelete {
		if err := helper.DeleteProductVariant(id); err != nil {
			if err.Error() == "variant not found" {
				http.Error(w, "Variant not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Variant deleted successfully", "id": id})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// SitesOrdersHandler POST /sites/orders
func SitesOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req models.CreateSiteOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.ShopID == "" || req.CustomerName == "" || req.Phone == "" || req.Address == "" {
		http.Error(w, "shop_id, customer_name, phone and address are required", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "items are required", http.StatusBadRequest)
		return
	}

	order, err := helper.CreateSiteOrder(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// SitesAdminOrdersHandler GET /sites/admin/orders?shop_id=
func SitesAdminOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, isAdmin, err := getAuthenticatedUserOrAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	shopID := r.URL.Query().Get("shop_id")
	if shopID == "" {
		http.Error(w, "shop_id is required", http.StatusBadRequest)
		return
	}
	if !isAdmin {
		if err := helper.VerifyShopOwnership(shopID, int(user.ID)); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
	orders, err := helper.GetSiteOrdersByShopID(shopID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// SitesAdminOrdersByIDHandler GET /sites/admin/orders/{id}, PUT /sites/admin/orders/{id}/status
func SitesAdminOrdersByIDHandler(w http.ResponseWriter, r *http.Request) {
	user, isAdmin, err := getAuthenticatedUserOrAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	if strings.HasSuffix(path, "/status") && r.Method == http.MethodPut {
		id := strings.TrimPrefix(strings.TrimSuffix(path, "/status"), "/sites/admin/orders/")
		if id == "" || !isValidUUID(id) {
			http.Error(w, "Invalid order ID", http.StatusBadRequest)
			return
		}
		order, err := helper.GetSiteOrderByID(id)
		if err != nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		if !isAdmin {
			if err := helper.VerifyShopOwnership(order.ShopID, int(user.ID)); err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
		}
		var req models.UpdateSiteOrderStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		updated, err := helper.UpdateSiteOrderStatus(id, req.OrderStatus)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
		return
	}

	if r.Method == http.MethodGet {
		id := strings.TrimPrefix(path, "/sites/admin/orders/")
		if id == "" || !isValidUUID(id) {
			http.Error(w, "Invalid order ID", http.StatusBadRequest)
			return
		}
		order, err := helper.GetSiteOrderByID(id)
		if err != nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		if !isAdmin {
			if err := helper.VerifyShopOwnership(order.ShopID, int(user.ID)); err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
