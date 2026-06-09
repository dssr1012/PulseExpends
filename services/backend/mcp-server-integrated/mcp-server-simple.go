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
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid request body",
			})
			return
		}
		
		// Simple validation
		if req.Email == "" || req.Password == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Email and password are required",
			})
			return
		}
		
		// Mock authentication - in production, validate against database
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			"user": map[string]interface{}{
				"id": "123e4567-e89b-12d3-a456-426614174000",
				"email": req.Email,
				"firstName": "John",
				"lastName": "Doe",
				"displayName": "John Doe",
			},
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
			"token": "test-j...-456",
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