package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"pulseexpends/backend/auth/models"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
	TokenContextKey contextKey = "token"
)

// AuthMiddleware verifica el token JWT y establece el usuario en el contexto
func AuthMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Obtener token del header Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Intentar obtener token de cookie
				cookie, err := r.Cookie("pulseexpends_session")
				if err != nil {
					http.Error(w, "Missing authorization token", http.StatusUnauthorized)
					return
				}
				authHeader = "Bearer " + cookie.Value
			}

			// Verificar formato Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Parsear y validar token JWT
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Verificar algoritmo de firma
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				
				// Obtener secret desde variables de entorno
				secret := []byte(getJWTSecret())
				return secret, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Extraer claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Verificar expiración
			exp, ok := claims["exp"].(float64)
			if !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}

			// Obtener user ID
			userIDStr, ok := claims["sub"].(string)
			if !ok {
				http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
				return
			}

			// Buscar usuario en la base de datos
			var user models.User
			result := db.Where("id = ? AND is_active = ?", userIDStr, true).First(&user)
			if result.Error != nil {
				http.Error(w, "User not found or inactive", http.StatusUnauthorized)
				return
			}

			// Verificar si el token ha sido revocado
			var session models.UserSession
			result = db.Where("user_id = ? AND token = ? AND is_active = ? AND expires_at > ?", 
				user.ID, tokenString, true, time.Now()).First(&session)
			if result.Error != nil {
				http.Error(w, "Session expired or revoked", http.StatusUnauthorized)
				return
			}

			// Actualizar última actividad de la sesión
			session.LastActivityAt = time.Now()
			db.Save(&session)

			// Establecer usuario y token en el contexto
			ctx := context.WithValue(r.Context(), UserContextKey, &user)
			ctx = context.WithValue(ctx, TokenContextKey, tokenString)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext obtiene el usuario del contexto
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// GetTokenFromContext obtiene el token del contexto
func GetTokenFromContext(ctx context.Context) string {
	token, ok := ctx.Value(TokenContextKey).(string)
	if !ok {
		return ""
	}
	return token
}

// RequireRole verifica que el usuario tenga un rol específico en un círculo
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				http.Error(w, "User not found in context", http.StatusUnauthorized)
				return
			}

			// Obtener circle ID de los parámetros
			vars := mux.Vars(r)
			circleID := vars["circleId"]
			if circleID == "" {
				circleID = vars["id"]
			}

			if circleID == "" {
				http.Error(w, "Circle ID required", http.StatusBadRequest)
				return
			}

			// Verificar rol del usuario en el círculo
			db := r.Context().Value("db").(*gorm.DB)
			var member models.CircleMember
			result := db.Where("circle_id = ? AND user_id = ? AND is_active = ?", 
				circleID, user.ID, true).First(&member)
			
			if result.Error != nil {
				http.Error(w, "User is not a member of this circle", http.StatusForbidden)
				return
			}

			// Verificar rol mínimo requerido
			if !hasRequiredRole(member.Role, role) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasRequiredRole verifica si el rol del usuario cumple con el requerido
func hasRequiredRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		"viewer":  1,
		"member":  2,
		"admin":   3,
		"owner":   4,
	}

	userLevel, userOk := roleHierarchy[userRole]
	requiredLevel, requiredOk := roleHierarchy[requiredRole]

	if !userOk || !requiredOk {
		return false
	}

	return userLevel >= requiredLevel
}

// RateLimitMiddleware implementa rate limiting
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	type clientInfo struct {
		count     int
		lastReset time.Time
	}

	clients := make(map[string]*clientInfo)
	var mu sync.RWMutex

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			
			mu.Lock()
			info, exists := clients[clientIP]
			
			if !exists {
				info = &clientInfo{
					count:     1,
					lastReset: time.Now(),
				}
				clients[clientIP] = info
			} else {
				// Resetear contador si ha pasado 1 minuto
				if time.Since(info.lastReset) > time.Minute {
					info.count = 1
					info.lastReset = time.Now()
				} else {
					info.count++
				}
			}
			
			currentCount := info.count
			mu.Unlock()

			// Establecer headers de rate limit
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", requestsPerMinute-currentCount))
			w.Header().Set("X-RateLimit-Reset", info.lastReset.Add(time.Minute).Format(time.RFC3339))

			if currentCount > requestsPerMinute {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP obtiene la IP real del cliente
func getClientIP(r *http.Request) string {
	// Intentar obtener IP de headers comunes
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}

// getJWTSecret obtiene el secret JWT desde variables de entorno
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-jwt-secret-key-change-in-production"
	}
	return secret
}