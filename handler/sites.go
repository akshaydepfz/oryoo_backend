package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
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

// SitesAdminUploadHandler POST /sites/admin/upload - multipart file upload to S3
func SitesAdminUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
		firebaseUID := r.Header.Get("X-Firebase-UID")
		if firebaseUID == "" {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				firebaseUID = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			}
		}
		if firebaseUID == "" {
			http.Error(w, "authentication required: provide X-Firebase-UID header or Authorization: Bearer <firebase_uid>", http.StatusUnauthorized)
			return
		}
		userID, err := helper.GetUserIDByFirebaseUID(firebaseUID)
		if err != nil {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}
		var req models.CreateShopRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.Subdomain == "" {
			http.Error(w, "name and subdomain are required", http.StatusBadRequest)
			return
		}
		shop, err := helper.InsertShop(req, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(shop)
		return
	}
	if r.Method == http.MethodGet {
		firebaseUID := r.Header.Get("X-Firebase-UID")
		if firebaseUID == "" {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				firebaseUID = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			}
		}
		if firebaseUID == "" {
			http.Error(w, "authentication required: provide X-Firebase-UID header or Authorization: Bearer <firebase_uid>", http.StatusUnauthorized)
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
	user, err := getAuthenticatedUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/sites/admin/shops/")
	if id == "" || !isValidUUID(id) {
		http.Error(w, "Invalid shop ID", http.StatusBadRequest)
		return
	}
	if err := helper.VerifyShopOwnership(id, int(user.ID)); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
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
	user, err := getAuthenticatedUser(r)
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
	if err := helper.VerifyShopOwnership(req.ShopID, int(user.ID)); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
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
	user, err := getAuthenticatedUser(r)
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
	if err := helper.VerifyShopOwnership(shopID, int(user.ID)); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
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
	user, err := getAuthenticatedUser(r)
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
	if err := helper.VerifyShopOwnership(req.ShopID, int(user.ID)); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
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
	user, err := getAuthenticatedUser(r)
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
	if err := helper.VerifyShopOwnership(shopID, int(user.ID)); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
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

func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
