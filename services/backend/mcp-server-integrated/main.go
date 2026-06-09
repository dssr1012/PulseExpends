package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Simple models for now - we'll integrate the full models later
type User struct {
	ID        string `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email     string `json:"email" gorm:"uniqueIndex;not null"`
	Username  string `json:"username" gorm:"uniqueIndex"`
	FullName  string `json:"full_name"`
	PasswordHash string `json:"-" gorm:"type:varchar(255)"`
	IsActive  bool   `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Transaction struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Type        string    `json:"type"` // "income" or "expense"
	UserID      string    `json:"user_id"`
	FamilyID    *string   `json:"family_id,omitempty"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

var (
	db *gorm.DB
	mu sync.RWMutex
)

func initDB() {
	var err error
	
	// Get database connection string from environment
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Default to Huawei Cloud RDS PostgreSQL
		dsn = "host=10.0.101.182 user=pulseexpends password=your_password dbname=pulseexpends_auth port=5432 sslmode=disable"
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		log.Println("Starting with in-memory storage (for testing)")
		db = nil
		return
	}

	// Auto migrate models
	err = db.AutoMigrate(&User{}, &Transaction{})
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
		log.Println("Starting with in-memory storage")
		db = nil
		return
	}

	log.Println("Database connected and migrated successfully")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version": "2.0.0",
		"endpoints": []string{
			"GET /api/health",
			"POST /api/auth/login",
			"POST /api/auth/register",
			"POST /api/auth/logout",
			"GET /api/auth/profile",
			"GET /api/transactions",
			"POST /api/transactions",
			"GET /api/transactions/{id}",
			"GET /api/summary",
		},
	}

	// Check database connection if available
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			if err := sqlDB.Ping(); err == nil {
				response["services"] = map[string]string{
					"database": "connected",
					"auth":     "integrated",
				}
			} else {
				response["services"] = map[string]string{
					"database": "disconnected",
					"auth":     "in-memory",
				}
			}
		}
	} else {
		response["services"] = map[string]string{
			"database": "in-memory",
			"auth":     "in-memory",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Simple in-memory storage for testing
var (
	users = make(map[string]User)
	transactions = make(map[string]Transaction)
	sessions = make(map[string]string) // token -> userID
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
		FullName string `json:"fullName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"Email and password are required"}`, http.StatusBadRequest)
		return
	}

	// Check if user already exists
	for _, user := range users {
		if user.Email == req.Email {
			http.Error(w, `{"error":"User with this email already exists"}`, http.StatusConflict)
			return
		}
	}

	// Create user
	user := User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Username:  req.Username,
		FullName:  req.FullName,
		PasswordHash: req.Password, // In production, hash this!
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	users[user.ID] = user

	// Create session token
	token := uuid.New().String()
	sessions[token] = user.ID

	response := map[string]interface{}{
		"success": true,
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
			"fullName": user.FullName,
		},
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Find user
	var foundUser User
	for _, user := range users {
		if user.Email == req.Email && user.PasswordHash == req.Password {
			foundUser = user
			break
		}
	}

	if foundUser.ID == "" {
		http.Error(w, `{"error":"Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	// Create session token
	token := uuid.New().String()
	sessions[token] = foundUser.ID

	response := map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":       foundUser.ID,
			"email":    foundUser.Email,
			"username": foundUser.Username,
			"fullName": foundUser.FullName,
		},
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Missing authorization token"}`, http.StatusUnauthorized)
			return
		}

		// Verify Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"Invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]
		userID, exists := sessions[token]
		if !exists {
			http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Find user
		user, exists := users[userID]
		if !exists {
			http.Error(w, `{"error":"User not found"}`, http.StatusUnauthorized)
			return
		}

		// Add user to context and continue
		r.Header.Set("X-User-ID", user.ID)
		next.ServeHTTP(w, r)
	})
}

func getProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	user, exists := users[userID]
	if !exists {
		http.Error(w, `{"error":"User not found"}`, http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"user": map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
			"fullName": user.FullName,
			"isActive": user.IsActive,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	
	var userTransactions []Transaction
	for _, tx := range transactions {
		if tx.UserID == userID {
			userTransactions = append(userTransactions, tx)
		}
	}

	response := map[string]interface{}{
		"status": "success",
		"data":   userTransactions,
		"count":  len(userTransactions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createTransactionHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	var req struct {
		Amount      float64   `json:"amount"`
		Currency    string    `json:"currency"`
		Category    string    `json:"category"`
		Description string    `json:"description"`
		Date        time.Time `json:"date"`
		Type        string    `json:"type"` // "income" or "expense"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, `{"error":"Amount must be positive"}`, http.StatusBadRequest)
		return
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}
	if req.Type != "income" && req.Type != "expense" {
		http.Error(w, `{"error":"Type must be 'income' or 'expense'"}`, http.StatusBadRequest)
		return
	}
	if req.Date.IsZero() {
		req.Date = time.Now()
	}

	transaction := Transaction{
		ID:          uuid.New().String(),
		Amount:      req.Amount,
		Currency:    req.Currency,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
		Type:        req.Type,
		UserID:      userID,
		CreatedAt:   time.Now(),
	}

	transactions[transaction.ID] = transaction

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Transaction created successfully",
		"data":    transaction,
	})
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	initDB()

	// Create router
	r := mux.NewRouter()
	
	// Public routes
	r.HandleFunc("/api/health", healthCheck).Methods("GET")
	r.HandleFunc("/api/auth/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/auth/register", registerHandler).Methods("POST")
	
	// Protected routes (require authentication)
	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware)
	
	// Auth routes
	api.HandleFunc("/auth/profile", getProfileHandler).Methods("GET")
	
	// Transaction routes
	api.HandleFunc("/transactions", getTransactionsHandler).Methods("GET")
	api.HandleFunc("/transactions", createTransactionHandler).Methods("POST")
	api.HandleFunc("/summary", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		
		var totalExpenses, totalIncome float64
		categories := make(map[string]float64)
		
		for _, tx := range transactions {
			if tx.UserID == userID {
				if tx.Type == "expense" {
					totalExpenses += tx.Amount
					categories[tx.Category] += tx.Amount
				} else if tx.Type == "income" {
					totalIncome += tx.Amount
				}
			}
		}
		
		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"total_expenses":   totalExpenses,
				"total_income":     totalIncome,
				"net_balance":      totalIncome - totalExpenses,
				"categories":       categories,
				"transaction_count": len(transactions),
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")
	
	// Default route for unknown endpoints
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Endpoint not found",
			"path":  r.URL.Path,
		})
	})
	
	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
	
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsMiddleware(r),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	log.Printf("Enhanced MCP Server with Authentication starting on :%s", port)
	log.Println("Available endpoints:")
	log.Println("  GET  /api/health              - Health check")
	log.Println("  POST /api/auth/login          - User login")
	log.Println("  POST /api/auth/register       - User registration")
	log.Println("  GET  /api/auth/profile        - Get user profile (protected)")
	log.Println("  GET  /api/transactions        - List transactions (protected)")
	log.Println("  POST /api/transactions        - Create transaction (protected)")
	log.Println("  GET  /api/summary             - Get expense summary (protected)")
	
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}