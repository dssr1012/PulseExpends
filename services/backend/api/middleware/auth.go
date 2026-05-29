package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"gorm.io/gorm"
)

type contextKey string

const userContextKey = contextKey("user")

// AuthMiddleware validates JWT tokens and adds user to context
func AuthMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check if it's a Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Bearer token required", http.StatusUnauthorized)
				return
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token and get user
			user, err := validateTokenAndGetUser(db, token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the authenticated user from context
func GetUserFromContext(ctx context.Context) *model.User {
	if user, ok := ctx.Value(userContextKey).(*model.User); ok {
		return user
	}
	return nil
}

// validateTokenAndGetUser validates JWT token and returns user
func validateTokenAndGetUser(db *gorm.DB, token string) (*model.User, error) {
	// TODO: Implement JWT validation
	// For now, we'll use a simple mock implementation
	// In production, validate JWT token and fetch user from database
	
	// Mock user for development
	user := &model.User{
		ID:       "550e8400-e29b-41d4-a716-446655440000", // Mock UUID
		Email:    "user@example.com",
		Username: "testuser",
		FullName: "Test User",
		IsActive: true,
		IsVerified: true,
	}
	
	return user, nil
}

// RequireAuth is a helper middleware that ensures user is authenticated
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// RequireCircleMember ensures user is a member of the specified circle
func RequireCircleMember(db *gorm.DB, circleID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user is member of circle
			var count int64
			err := db.Model(&model.FamilyMember{}).
				Where("family_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
				Count(&count).Error
			
			if err != nil || count == 0 {
				http.Error(w, "User is not a member of this circle", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// RequireCircleAdmin ensures user is an admin of the specified circle
func RequireCircleAdmin(db *gorm.DB, circleID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user is admin or owner of circle
			var member model.FamilyMember
			err := db.Where("family_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
				First(&member).Error
			
			if err != nil {
				http.Error(w, "User is not a member of this circle", http.StatusForbidden)
				return
			}

			if member.Role != "admin" && member.Role != "owner" {
				http.Error(w, "Admin privileges required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// RequireCircleOwner ensures user is the owner of the specified circle
func RequireCircleOwner(db *gorm.DB, circleID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user is owner of circle
			var member model.FamilyMember
			err := db.Where("family_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
				First(&member).Error
			
			if err != nil {
				http.Error(w, "User is not a member of this circle", http.StatusForbidden)
				return
			}

			if member.Role != "owner" {
				http.Error(w, "Owner privileges required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// ValidateContentType ensures request has correct content type
func ValidateContentType(contentType string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") != contentType {
				http.Error(w, "Content-Type must be "+contentType, http.StatusUnsupportedMediaType)
				return
			}
			next.ServeHTTP(w, r)
		}
	}
}

// RateLimitMiddleware implements basic rate limiting
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	// Simple in-memory rate limiter
	// In production, use a distributed rate limiter like Redis
	type clientInfo struct {
		count    int
		lastTime int64
	}
	
	clients := make(map[string]*clientInfo)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr
			now := time.Now().Unix()
			
			// Clean up old entries (optional)
			if len(clients) > 10000 {
				clients = make(map[string]*clientInfo)
			}
			
			info, exists := clients[clientIP]
			if !exists {
				info = &clientInfo{count: 1, lastTime: now}
				clients[clientIP] = info
			} else {
				// Reset counter if more than a minute has passed
				if now-info.lastTime > 60 {
					info.count = 1
					info.lastTime = now
				} else {
					info.count++
				}
			}
			
			if info.count > requestsPerMinute {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs request details
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create response writer wrapper to capture status code
		rw := &responseWriter{w, http.StatusOK}
		
		next.ServeHTTP(rw, r)
		
		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, rw.status, duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// ErrorHandlerMiddleware handles panics and returns JSON errors
func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Internal server error",
					"message": "An unexpected error occurred",
				})
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// JSONMiddleware sets Content-Type to application/json
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}