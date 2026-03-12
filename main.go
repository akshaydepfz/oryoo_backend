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
	database.CreateSitesTables()
	http.Handle("/users", http.HandlerFunc(handler.UserHandler))
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

	// mailService := initMailService()

	// err := mailService.SendWelcomeEmail(
	// 	"akshaypk.dev@gmail.com",
	// 	"Akshay",
	// )

	// if err != nil {
	// 	log.Fatalf("Failed to send welcome email: %v", err)
	// }

	// log.Println("Welcome email sent successfully")

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", enableCors(http.DefaultServeMux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
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

// func initMailService() *mailer.MailService {
// 	mailService, err := mailer.NewMailService()
// 	if err != nil {
// 		log.Fatalf("Failed to initialize mail service: %v", err)
// 	}

// 	return mailService
// }
