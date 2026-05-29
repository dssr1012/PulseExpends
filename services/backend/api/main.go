package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/dssr1012/pulse-expends/internal/service"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/dssr1012/pulse-expends/services/backend/api/handlers"
	"github.com/dssr1012/pulse-expends/services/backend/api/middleware"
)

var (
	db                        *gorm.DB
	transactionHandler        *handlers.TransactionHandler
	creditCardHandler         *handlers.CreditCardHandler
	exchangeRateHandler       *handlers.ExchangeRateHandler
	mobileNotificationHandler *handlers.MobileNotificationHandler
	anomalyDetectionHandler   *handlers.AnomalyDetectionHandler
	notificationWhitelistHandler *handlers.NotificationAppWhitelistHandler
)

func initDB() {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://pulseexpends:***@localhost:5432/pulseexpends"
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")
}

func initServices() {
	// Initialize repositories
	transactionRepo := repository.NewPostgresTransactionRepository(db)
	creditCardRepo := repository.NewPostgresCreditCardRepository(db)
	exchangeRateRepo := repository.NewPostgresExchangeRateRepository(db)
	mobileNotificationRepo := repository.NewPostgresMobileNotificationRepository(db)
	anomalyDetectionRepo := repository.NewPostgresAnomalyDetectionRepository(db)
	notificationWhitelistRepo := repository.NewPostgresNotificationAppWhitelistRepository(db)

	// Initialize services
	transactionService := service.NewTransactionService(transactionRepo)
	creditCardService := service.NewCreditCardService(creditCardRepo)
	exchangeRateService := service.NewExchangeRateService(exchangeRateRepo)
	mobileNotificationService := service.NewMobileNotificationService(mobileNotificationRepo, notificationWhitelistRepo)
	anomalyDetectionService := service.NewAnomalyDetectionService(anomalyDetectionRepo, transactionRepo)
	notificationWhitelistService := service.NewNotificationAppWhitelistService(notificationWhitelistRepo)

	// Initialize handlers
	transactionHandler = handlers.NewTransactionHandler(transactionService)
	creditCardHandler = handlers.NewCreditCardHandler(creditCardService)
	exchangeRateHandler = handlers.NewExchangeRateHandler(exchangeRateService)
	mobileNotificationHandler = handlers.NewMobileNotificationHandler(mobileNotificationService)
	anomalyDetectionHandler = handlers.NewAnomalyDetectionHandler(anomalyDetectionService)
	notificationWhitelistHandler = handlers.NewNotificationAppWhitelistHandler(notificationWhitelistService)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from environment
		allowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
		if len(allowedOrigins) == 0 {
			allowedOrigins = []string{"*"}
		}

		origin := r.Header.Get("Origin")
		allowed := false

		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		
		next.ServeHTTP(w, r)
		
		log.Printf("Completed %s %s in %v", r.Method, r.RequestURI, time.Since(start))
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	sqlDB, err := db.DB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusServiceUnavailable)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		http.Error(w, "Database ping failed", http.StatusServiceUnavailable)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"services": map[string]string{
			"database":                 "connected",
			"transaction_service":      "ready",
			"credit_card_service":      "ready",
			"exchange_rate_service":    "ready",
			"mobile_notification_service": "ready",
			"anomaly_detection_service":   "ready",
			"notification_whitelist_service": "ready",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	initDB()

	// Initialize services and handlers
	initServices()

	// Create router
	r := mux.NewRouter()

	// Apply middleware
	r.Use(loggingMiddleware)
	r.Use(enableCORS)

	// Public routes
	public := r.PathPrefix("/api").Subrouter()
	
	// Health check
	public.HandleFunc("/health", healthCheck).Methods("GET")

	// Protected routes (require authentication)
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(db))

	// Transaction routes
	protected.HandleFunc("/transactions", transactionHandler.GetTransactions).Methods("GET")
	protected.HandleFunc("/transactions/{id}", transactionHandler.GetTransaction).Methods("GET")
	protected.HandleFunc("/transactions", transactionHandler.CreateTransaction).Methods("POST")
	protected.HandleFunc("/transactions/{id}", transactionHandler.UpdateTransaction).Methods("PUT")
	protected.HandleFunc("/transactions/{id}", transactionHandler.DeleteTransaction).Methods("DELETE")
	protected.HandleFunc("/transactions/{id}/duplicates", transactionHandler.CheckDuplicates).Methods("GET")
	protected.HandleFunc("/transactions/batch", transactionHandler.CreateBatch).Methods("POST")
	protected.HandleFunc("/transactions/summary", transactionHandler.GetSummary).Methods("GET")

	// Credit Card routes
	protected.HandleFunc("/credit-cards", creditCardHandler.GetCreditCards).Methods("GET")
	protected.HandleFunc("/credit-cards/{id}", creditCardHandler.GetCreditCard).Methods("GET")
	protected.HandleFunc("/credit-cards", creditCardHandler.AddCreditCard).Methods("POST")
	protected.HandleFunc("/credit-cards/{id}", creditCardHandler.UpdateCreditCard).Methods("PUT")
	protected.HandleFunc("/credit-cards/{id}", creditCardHandler.DeleteCreditCard).Methods("DELETE")
	protected.HandleFunc("/credit-cards/{id}/validate", creditCardHandler.ValidateCard).Methods("POST")
	protected.HandleFunc("/credit-cards/upcoming-payments", creditCardHandler.GetUpcomingPayments).Methods("GET")

	// Exchange Rate routes
	protected.HandleFunc("/exchange-rates/latest/{base}/{target}", exchangeRateHandler.GetLatestRate).Methods("GET")
	protected.HandleFunc("/exchange-rates/convert", exchangeRateHandler.ConvertAmount).Methods("POST")
	protected.HandleFunc("/exchange-rates/historical", exchangeRateHandler.GetHistoricalRates).Methods("GET")
	protected.HandleFunc("/exchange-rates/currencies", exchangeRateHandler.GetSupportedCurrencies).Methods("GET")
	protected.HandleFunc("/exchange-rates/update", exchangeRateHandler.UpdateRates).Methods("POST")

	// Mobile Notification routes
	protected.HandleFunc("/notifications", mobileNotificationHandler.GetNotifications).Methods("GET")
	protected.HandleFunc("/notifications/{id}", mobileNotificationHandler.GetNotification).Methods("GET")
	protected.HandleFunc("/notifications", mobileNotificationHandler.CreateNotification).Methods("POST")
	protected.HandleFunc("/notifications/{id}/process", mobileNotificationHandler.ProcessNotification).Methods("POST")
	protected.HandleFunc("/notifications/{id}/ignore", mobileNotificationHandler.IgnoreNotification).Methods("POST")
	protected.HandleFunc("/notifications/stats", mobileNotificationHandler.GetStats).Methods("GET")
	protected.HandleFunc("/notifications/parse", mobileNotificationHandler.ParseNotification).Methods("POST")

	// Anomaly Detection routes
	protected.HandleFunc("/anomaly/rules", anomalyDetectionHandler.GetRules).Methods("GET")
	protected.HandleFunc("/anomaly/rules/{id}", anomalyDetectionHandler.GetRule).Methods("GET")
	protected.HandleFunc("/anomaly/rules", anomalyDetectionHandler.CreateRule).Methods("POST")
	protected.HandleFunc("/anomaly/rules/{id}", anomalyDetectionHandler.UpdateRule).Methods("PUT")
	protected.HandleFunc("/anomaly/rules/{id}", anomalyDetectionHandler.DeleteRule).Methods("DELETE")
	protected.HandleFunc("/anomaly/rules/{id}/toggle", anomalyDetectionHandler.ToggleRule).Methods("POST")
	protected.HandleFunc("/anomaly/detect", anomalyDetectionHandler.DetectAnomalies).Methods("POST")
	protected.HandleFunc("/anomaly/transactions/{id}", anomalyDetectionHandler.CheckTransaction).Methods("POST")
	protected.HandleFunc("/anomaly/stats", anomalyDetectionHandler.GetStats).Methods("GET")

	// Notification App Whitelist routes
	protected.HandleFunc("/whitelist/apps", notificationWhitelistHandler.GetWhitelist).Methods("GET")
	protected.HandleFunc("/whitelist/apps/{id}", notificationWhitelistHandler.GetWhitelistEntry).Methods("GET")
	protected.HandleFunc("/whitelist/apps", notificationWhitelistHandler.AddToWhitelist).Methods("POST")
	protected.HandleFunc("/whitelist/apps/{id}", notificationWhitelistHandler.UpdateWhitelistEntry).Methods("PUT")
	protected.HandleFunc("/whitelist/apps/{id}", notificationWhitelistHandler.RemoveFromWhitelist).Methods("DELETE")
	protected.HandleFunc("/whitelist/apps/check", notificationWhitelistHandler.CheckAppWhitelisted).Methods("POST")
	protected.HandleFunc("/whitelist/apps/auto-whitelist", notificationWhitelistHandler.AutoWhitelistApp).Methods("POST")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🚀 PulseExpends API server started at http://%s:%s", host, port)
	log.Printf("📊 Health check: http://%s:%s/api/health", host, port)
	log.Printf("🔐 Available endpoints:")
	log.Printf("   GET  /api/health - Health check")
	log.Printf("   GET  /api/transactions - List transactions (protected)")
	log.Printf("   POST /api/transactions - Create transaction (protected)")
	log.Printf("   GET  /api/credit-cards - List credit cards (protected)")
	log.Printf("   POST /api/credit-cards - Add credit card (protected)")
	log.Printf("   GET  /api/exchange-rates/latest/{base}/{target} - Get latest rate (protected)")
	log.Printf("   POST /api/exchange-rates/convert - Convert amount (protected)")
	log.Printf("   GET  /api/notifications - List notifications (protected)")
	log.Printf("   POST /api/notifications/parse - Parse notification (protected)")
	log.Printf("   GET  /api/anomaly/rules - List anomaly rules (protected)")
	log.Printf("   POST /api/anomaly/detect - Detect anomalies (protected)")
	log.Printf("   GET  /api/whitelist/apps - List whitelisted apps (protected)")
	log.Printf("   POST /api/whitelist/apps - Whitelist app (protected)")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}