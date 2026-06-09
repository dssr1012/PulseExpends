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
)

// Simple models for now
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	PasswordHash string `json:"-"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Type        string    `json:"type"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
}

var (
	users = make(map[string]User)
	transactions = make(map[string]Transaction)
	sessions = make(map[string]string) // token -> userID
	mu sync.RWMutex
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version": "2.0.0",
		"service": "MCP Server with Authentication",
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
		"database": "in-memory (for testing)",
		"auth": "integrated",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

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

	mu.Lock()
	defer mu.Unlock()

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

	mu.RLock()
	defer mu.RUnlock()

	// Find user

[Content truncated due to size limit. Click to expand]