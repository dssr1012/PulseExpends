package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pulseexpends/backend/auth/handlers"
	"pulseexpends/backend/auth/middleware"
	"pulseexpends/backend/auth/models"
)

var (
	db     *gorm.DB
	auth   *handlers.AuthHandler
	circle *handlers.CircleHandler
	trans  *handlers.TransactionHandler
)

func initDB() {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://pulseexpends:pulseexpends_password@localhost:5432/pulseexpends_auth"
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate models
	err = db.AutoMigrate(
		&models.User{},
		&models.UserAuth{},
		&models.UserPreferences{},
		&models.Circle{},
		&models.CircleMember{},
		&models.Transaction{},
		&models.SplitTransaction{},
		&models.RecurringTransaction{},
		&models.UserSession{},
		&models.CircleInvite{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database connected and migrated successfully")
}

func initHandlers() {
	auth = handlers.NewAuthHandler(db)
	circle = handlers.NewCircleHandler(db)
	trans = handlers.NewTransactionHandler(db)
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
			"database": "connected",
			"auth":     "ready",
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

	// Initialize handlers
	initHandlers()

	// Create router
	r := mux.NewRouter()

	// Apply middleware
	r.Use(loggingMiddleware)
	r.Use(enableCORS)

	// Public routes
	public := r.PathPrefix("/api").Subrouter()
	
	// Auth routes
	public.HandleFunc("/auth/register", auth.Register).Methods("POST")
	public.HandleFunc("/auth/login", auth.Login).Methods("POST")
	public.HandleFunc("/auth/google", auth.GoogleLogin).Methods("GET")
	public.HandleFunc("/auth/google/callback", auth.GoogleCallback).Methods("GET")
	public.HandleFunc("/auth/forgot-password", auth.ForgotPassword).Methods("POST")
	public.HandleFunc("/auth/reset-password", auth.ResetPassword).Methods("POST")
	public.HandleFunc("/auth/verify-email/{token}", auth.VerifyEmail).Methods("GET")
	
	// Health check
	public.HandleFunc("/health", healthCheck).Methods("GET")

	// Protected routes (require authentication)
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(db))

	// Auth protected routes
	protected.HandleFunc("/auth/logout", auth.Logout).Methods("POST")
	protected.HandleFunc("/auth/profile", auth.GetProfile).Methods("GET")
	protected.HandleFunc("/auth/profile", auth.UpdateProfile).Methods("PUT")
	protected.HandleFunc("/auth/change-password", auth.ChangePassword).Methods("POST")
	protected.HandleFunc("/auth/sessions", auth.GetSessions).Methods("GET")
	protected.HandleFunc("/auth/sessions/{id}", auth.RevokeSession).Methods("DELETE")

	// Circles routes
	protected.HandleFunc("/circles", circle.GetCircles).Methods("GET")
	protected.HandleFunc("/circles/{id}", circle.GetCircle).Methods("GET")
	protected.HandleFunc("/circles", circle.CreateCircle).Methods("POST")
	protected.HandleFunc("/circles/{id}", circle.UpdateCircle).Methods("PUT")
	protected.HandleFunc("/circles/{id}", circle.DeleteCircle).Methods("DELETE")
	protected.HandleFunc("/circles/{id}/members", circle.GetCircleMembers).Methods("GET")
	protected.HandleFunc("/circles/{id}/members", circle.AddMember).Methods("POST")
	protected.HandleFunc("/circles/{id}/members/{userId}", circle.RemoveMember).Methods("DELETE")
	protected.HandleFunc("/circles/{id}/members/{userId}/role", circle.UpdateMemberRole).Methods("PUT")
	protected.HandleFunc("/circles/join/{code}", circle.JoinCircle).Methods("POST")
	protected.HandleFunc("/circles/{id}/invite", circle.InviteToCircle).Methods("POST")
	protected.HandleFunc("/circles/{id}/transactions", circle.GetCircleTransactions).Methods("GET")
	protected.HandleFunc("/circles/{id}/activities", circle.GetCircleActivities).Methods("GET")

	// Transactions routes
	protected.HandleFunc("/transactions", trans.GetTransactions).Methods("GET")
	protected.HandleFunc("/transactions/{id}", trans.GetTransaction).Methods("GET")
	protected.HandleFunc("/transactions", trans.CreateTransaction).Methods("POST")
	protected.HandleFunc("/transactions/{id}", trans.UpdateTransaction).Methods("PUT")
	protected.HandleFunc("/transactions/{id}", trans.DeleteTransaction).Methods("DELETE")
	protected.HandleFunc("/transactions/circle/{circleId}", trans.GetCircleTransactions).Methods("GET")
	protected.HandleFunc("/transactions/{id}/split", trans.SplitTransaction).Methods("POST")
	protected.HandleFunc("/transactions/{id}/approve", trans.ApproveTransaction).Methods("POST")
	protected.HandleFunc("/transactions/{id}/reject", trans.RejectTransaction).Methods("POST")

	// Stats routes
	protected.HandleFunc("/stats", trans.GetUserStats).Methods("GET")
	protected.HandleFunc("/stats/circle/{circleId}", trans.GetCircleStats).Methods("GET")
	protected.HandleFunc("/stats/monthly", trans.GetMonthlyStats).Methods("GET")
	protected.HandleFunc("/stats/categories", trans.GetCategoryStats).Methods("GET")

	// Serve static files for frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
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

	log.Printf("🚀 Authentication server started at http://%s:%s", host, port)
	log.Printf("📊 Health check: http://%s:%s/api/health", host, port)
	log.Printf("🔐 Available endpoints:")
	log.Printf("   POST /api/auth/register - Register user")
	log.Printf("   POST /api/auth/login - Login")
	log.Printf("   GET  /api/auth/google - Google OAuth login")
	log.Printf("   GET  /api/auth/profile - User profile (protected)")
	log.Printf("   GET  /api/circles - List circles (protected)")
	log.Printf("   POST /api/circles - Create circle (protected)")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}