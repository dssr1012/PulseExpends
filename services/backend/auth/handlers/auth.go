package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"

	"pulseexpends/backend/auth/models"
)

type AuthHandler struct {
	db           *gorm.DB
	oauthConfig  *oauth2.Config
	store        *sessions.CookieStore
	jwtSecret    []byte
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	// Configurar OAuth2
	oauthConfig := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Configurar cookie store
	cookieSecret := os.Getenv("COOKIE_SECRET")
	if cookieSecret == "" {
		cookieSecret = "your-cookie-secret-key-change-in-production"
	}

	store := sessions.NewCookieStore([]byte(cookieSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 días
		HttpOnly: true,
		Secure:   os.Getenv("COOKIE_SECURE") == "true",
		SameSite: http.SameSiteLaxMode,
	}

	// Configurar JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-jwt-secret-key-change-in-production"
	}

	return &AuthHandler{
		db:           db,
		oauthConfig:  oauthConfig,
		store:        store,
		jwtSecret:    []byte(jwtSecret),
	}
}

// Register maneja el registro de nuevos usuarios
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Username  string `json:"username"`
		FullName  string `json:"fullName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar campos requeridos
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Verificar si el usuario ya existe
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	// Hashear contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Crear usuario
	user := models.User{
		Email:       req.Email,
		Username:    req.Username,
		FullName:    req.FullName,
		Password:    string(hashedPassword),
		Provider:    "local",
		IsActive:    true,
		IsVerified:  false,
		LastLoginAt: time.Now(),
		Preferences: models.JSONB{"theme": "light", "currency": "USD", "language": "es"},
	}

	if err := h.db.Create(&user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Crear token JWT
	token, err := h.createJWTToken(&user)
	if err != nil {
		log.Printf("Error creating JWT token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Crear sesión
	session := models.UserSession{
		UserID:         user.ID,
		Token:          token,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		IsActive:       true,
		LastActivityAt: time.Now(),
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.db.Create(&session).Error; err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Configurar cookie
	h.setSessionCookie(w, r, token)

	// Responder
	response := map[string]interface{}{
		"success": true,
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id":        user.ID,
			"email":     user.Email,
			"username":  user.Username,
			"fullName":  user.FullName,
			"createdAt": user.CreatedAt,
		},
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Login maneja el inicio de sesión
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Buscar usuario
	var user models.User
	if err := h.db.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Verificar contraseña
	if user.Password == "" || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Actualizar último login
	user.LastLoginAt = time.Now()
	h.db.Save(&user)

	// Crear token JWT
	token, err := h.createJWTToken(&user)
	if err != nil {
		log.Printf("Error creating JWT token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Crear sesión
	session := models.UserSession{
		UserID:         user.ID,
		Token:          token,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		IsActive:       true,
		LastActivityAt: time.Now(),
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.db.Create(&session).Error; err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Configurar cookie
	h.setSessionCookie(w, r, token)

	// Responder
	response := map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":        user.ID,
			"email":     user.Email,
			"username":  user.Username,
			"fullName":  user.FullName,
			"avatarURL": user.AvatarURL,
		},
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GoogleLogin inicia el flujo de OAuth con Google
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Generar state para prevenir CSRF
	state := generateRandomString(32)
	
	session, _ := h.store.Get(r, "oauth-state")
	session.Values["state"] = state
	session.Save(r, w)

	// Redirigir a Google
	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback maneja el callback de Google OAuth
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verificar state
	session, _ := h.store.Get(r, "oauth-state")
	state := session.Values["state"]
	delete(session.Values, "state")
	session.Save(r, w)

	if r.URL.Query().Get("state") != state {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Intercambiar código por token
	code := r.URL.Query().Get("code")
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Error exchanging code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Obtener información del usuario
	client := h.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		log.Printf("Error decoding user info: %v", err)
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Buscar o crear usuario
	var user models.User
	err = h.db.Where("provider = ? AND provider_id = ?", "google", googleUser.ID).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// Crear nuevo usuario
		user = models.User{
			Email:       googleUser.Email,
			FullName:    googleUser.Name,
			AvatarURL:   googleUser.Picture,
			Provider:    "google",
			ProviderID:  googleUser.ID,
			IsActive:    true,
			IsVerified:  true,
			LastLoginAt: time.Now(),
			Preferences: models.JSONB{"theme": "light", "currency": "USD", "language": "es"},
		}

		if err := h.db.Create(&user).Error; err != nil {
			log.Printf("Error creating user: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.Printf("Error finding user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else {
		// Actualizar usuario existente
		user.LastLoginAt = time.Now()
		user.AvatarURL = googleUser.Picture
		h.db.Save(&user)
	}

	// Crear token JWT
	jwtToken, err := h.createJWTToken(&user)
	if err != nil {
		log.Printf("Error creating JWT token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Crear sesión
	session := models.UserSession{
		UserID:         user.ID,
		Token:          jwtToken,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		IsActive:       true,
		LastActivityAt: time.Now(),
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.db.Create(&session).Error; err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Configurar cookie
	h.setSessionCookie(w, r, jwtToken)

	// Redirigir al frontend
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, jwtToken)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// Logout cierra la sesión del usuario
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Obtener token del contexto o cookie
	token := r.Context().Value("token").(string)
	if token == "" {
		cookie, err := r.Cookie("pulseexpends_session")
		if err == nil {
			token = cookie.Value
		}
	}

	// Invalidar sesión en la base de datos
	if token != "" {
		h.db.Model(&models.UserSession{}).Where("token = ?", token).Update("is_active", false)
	}

	// Limpiar cookie
	h.clearSessionCookie(w, r)

	response := map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProfile obtiene el perfil del usuario
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	response := map[string]interface{}{
		"success": true,
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"fullName":   user.FullName,
			"avatarURL":  user.AvatarURL,
			"provider":   user.Provider,
			"isVerified": user.IsVerified,
			"lastLogin":  user.LastLoginAt,
			"preferences": user.Preferences,
			"createdAt":  user.CreatedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProfile actualiza el perfil del usuario
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	var req struct {
		Username  string          `json:"username"`
		FullName  string          `json:"fullName"`
		AvatarURL string          `json:"avatarURL"`
		Preferences models.JSONB  `json:"preferences"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Actualizar campos permitidos
	if req.Username != "" {
		// Verificar que el username no esté en uso
		var existingUser models.User
		if err := h.db.Where("username = ? AND id != ?", req.Username, user.ID).First(&existingUser).Error; err == nil {
			http.Error(w, "Username already in use", http.StatusConflict)
			return
		}
		user.Username = req.Username
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}

	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if req.Preferences != nil {
		user.Preferences = req.Preferences
	}

	if err := h.db.Save(&user).Error; err != nil {
		log.Printf("Error updating user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Profile updated successfully",
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"fullName":   user.FullName,
			"avatarURL":  user.AvatarURL,
			"preferences": user.Preferences,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ChangePassword cambia la contraseña del usuario
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verificar contraseña actual
	if user.Password == "" || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)) != nil {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Hashear nueva contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)
	if err := h.db.Save(&user).Error; err != nil {
		log.Printf("Error updating password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Password changed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions
func (h *AuthHandler) createJWTToken(user *models.User) (string, error) {
	expiry := time.Now().Add(7 * 24 * time.Hour)
	
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"email": user.Email,
		"exp": expiry.Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	cookie := &http.Cookie{
		Name:     "pulseexpends_session",
		Value:    token,
		Path:     "/",
		MaxAge:   86400 * 7, // 7 días
		HttpOnly: true,
		Secure:   os.Getenv("COOKIE_SECURE") == "true",
		SameSite: http.SameSiteLaxMode,
	}
	
	http.SetCookie(w, cookie)
}

func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "pulseexpends_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   os.Getenv("COOKIE_SECURE") == "true",
		SameSite: http.SameSiteLaxMode,
	}
	
	http.SetCookie(w, cookie)
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}