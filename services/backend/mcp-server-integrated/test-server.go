package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Create router
	r := mux.NewRouter()
	
	// Health check endpoint
	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version": "2.0.0",
			"service": "MCP Server with Authentication",
			"endpoints": []string{
				"GET /api/health",
				"POST /api/auth/login",
				"POST /api/auth/register",
				"GET /api/auth/profile",
				"GET /api/transactions",
				"POST /api/transactions",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")
	
	// Auth endpoints
	r.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Login endpoint ready",
			"token": "test-jwt-token-123",
		})
	}).Methods("POST")
	
	r.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Registration endpoint ready",
			"user": map[string]string{
				"id": "test-user-123",
				"email": "test@test.com",
			},
			"token": "test-jwt-token-456",
		})
	}).Methods("POST")
	
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
	
	log.Printf("MCP Server with Authentication starting on :%s", port)
	log.Println("Available endpoints:")
	log.Println("  GET  /api/health              - Health check")
	log.Println("  POST /api/auth/login          - User login")
	log.Println("  POST /api/auth/register       - User registration")
	
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}