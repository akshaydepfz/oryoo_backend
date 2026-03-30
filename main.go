package main

import (
	"log"
	"net/http"
	"strings"

	"oryoo.com/database"
	"oryoo.com/handler"
)

func main() {

	database.ConnectDatabase()

	http.Handle("/users", http.HandlerFunc(handler.UserHandler))
	http.HandleFunc("/api/user/update-activity", handler.UpdateUserActivityHandler)
	http.HandleFunc("/auth/check-phone", handler.CheckPhoneHandler)
	http.HandleFunc("/auth/register", handler.CreateUserHandler)
	http.HandleFunc("/clients", handler.ClientHandler)
	http.HandleFunc("/clients/created-by", handler.GetClientByCreatedBy)
	http.HandleFunc("/users/firebase-uid", handler.GetUserByFirebaseUID)
	http.HandleFunc("/orders/create", handler.CreateOrder)
	http.HandleFunc("/orders/update", handler.UpdateOrder)
	http.HandleFunc("/orders/by-created", handler.GetOrdersByCreatedBy)
	http.HandleFunc("/orders/", handler.DeleteOrderHandler)
	http.HandleFunc("/payments/create", handler.CreatePayment)
	http.HandleFunc("/payments/by-created", handler.GetPaymentsByCreatedByHandler)
	http.HandleFunc("/payments/", handler.DeletePaymentHandler)
	http.HandleFunc("/billing/transaction", handler.CreateBillingTransactionHandler)
	http.HandleFunc("/app/latest-version", handler.GetLatestVersionHandler)

	// Oryoo Sites - register more specific paths first
	http.HandleFunc("/sites/shop/by-domain", handler.SitesShopByDomainHandler)
	http.HandleFunc("/sites/products/", handler.SitesProductByIDHandler)
	http.HandleFunc("/sites/products", handler.SitesProductsHandler)
	http.HandleFunc("/sites/categories", handler.SitesCategoriesHandler)
	http.HandleFunc("/sites/site-config", handler.SitesSiteConfigHandler)
	http.HandleFunc("/sites/about", handler.SitesAboutHandler)
	http.HandleFunc("/sites/contact", handler.SitesContactHandler)
	http.HandleFunc("/sites/testimonials", handler.SitesTestimonialsHandler)
	http.HandleFunc("/sites/admin/upload", handler.SitesAdminUploadHandler)
	http.HandleFunc("/sites/admin/shops/", handler.SitesAdminShopsByIDHandler)
	http.HandleFunc("/sites/admin/shops", handler.SitesAdminShopsHandler)
	http.HandleFunc("/sites/admin/products/", handler.SitesAdminProductsByIDHandler)
	http.HandleFunc("/sites/admin/products", handler.SitesAdminProductsHandler)
	http.HandleFunc("/sites/admin/categories/", handler.SitesAdminCategoriesByIDHandler)
	http.HandleFunc("/sites/admin/categories", handler.SitesAdminCategoriesHandler)

	// Admin - list endpoints
	http.HandleFunc("/admin/customers", handler.AdminGetCustomers)
	http.HandleFunc("/admin/clients", handler.AdminGetClients)
	http.HandleFunc("/admin/payments", handler.AdminGetPayments)
	http.HandleFunc("/admin/orders", handler.AdminGetOrders)
	http.HandleFunc("/admin/products", handler.AdminGetProducts)
	http.HandleFunc("/admin/billing-transactions", handler.AdminGetBillingTransactions)

	// Products - register specific routes before generic
	http.HandleFunc("/products/created-by", handler.GetProductsByCreatedBy)
	http.HandleFunc("/products/", handler.DeleteProductHandler)
	http.HandleFunc("/products", handler.ProductHandler)

	//mailService := mailer.NewMailService()
	// if err := mailService.SendTestEmail(mailer.DefaultFrom, "hi@oryoo.in", "Oryoo backend Mailgun test"); err != nil {
	// 	log.Fatalf("mail send: %v", err)
	// }
	//log.Printf("Test email sent to hi@oryoo.in (from %s)", mailer.DefaultFrom)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", enableCors(adminAuthMiddleware(http.DefaultServeMux))); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// adminAuthMiddleware requires Authorization: Bearer <static secret> for /admin/* routes.
// All other routes (including /sites/admin/*) pass through and use their existing auth (e.g. Firebase UID).
func adminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/admin/") {
			authHeader := r.Header.Get("Authorization")
			expected := "Bearer 8f3k29df0sdf89sdf98sd7f98sd7f9sd87f"

			if authHeader != expected {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow subdomains (*.oryoo.in), main domain, and deployed frontend
		if origin != "" &&
			(strings.HasSuffix(origin, ".oryoo.in") ||
				origin == "https://oryoo.in" ||
				origin == "https://oryoo-app.web.app") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Firebase-UID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
