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
	// Configure OAuth2
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

	// Configure cookie store
	cookieSecret := os.Getenv("COOKIE_SECRET")
	if cookieSecret == "" {
		cookieSecret = "your-cookie-secret-key-change-in-production"
	}

	store := sessions.NewCookieStore([]byte(cookieSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   os.Getenv("COOKIE_SECURE") == "true",
		SameSite: http.SameSiteLaxMode,
	}

	// Configure JWT secret
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

// Register handles user registration
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

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	user := models.User{
		Email:       req.Email,
		Username:    req.Username,
		FullName:    req.FullName,
		Password:    string(hashedPassword),
		Provider:    "local",
		IsActive:    true,
		IsVerified:  false,
		LastLoginAt: time.Now(),
		Preferences: models.JSONB{"theme": "light", "currency": "USD", "language": "en"},
	}

	if err := h.db.Create(&user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create JWT token
	token, err := h.createJWTToken(&user)
	if err != nil {
		log.Printf("Error creating JWT token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create session
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

	// Set session cookie
	h.setSessionCookie(w, r, token)

	// Respond
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

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find user
	var user models.User
	if err := h.db.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Verify password
	if user.Password == "" || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Update last login
	user.LastLoginAt = time.Now()
	h.db.Save(&user)

	// Create JWT token
	token, err := h.createJWTToken(&user)
	if err != nil {
		log.Printf("Error creating JWT token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create session
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

	// Set session cookie
	h.setSessionCookie(w, r, token)

	// Respond
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

// GoogleLogin initiates Google OAuth flow
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state to prevent CSRF
	state := generateRandomString(32)
	
	session, _ := h.store.Get(r, "oauth-state")
	session.Values["state"] = state
	session.Save(r, w)

	// Redirect to Google
	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles Google OAuth callback
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state
	session, _ := h.store.Get(r, "oauth-state")
	state := session.Values["state"]
	delete(session.Values, "state")
	session.Save(r, w)

	if r.URL.Query().Get("state") != state {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	code := r.URL.Query().Get("code")
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Error exchanging code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Get user info
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

	// Find or create user
	var user models.User
	err = h.db.Where("provider = ? AND provider_id = ?", "google", googleUser.ID).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			Email:       googleUser.Email,
			FullName:    googleUser.Name,
			AvatarURL:   googleUser.Picture,
			Provider:    "google",
			ProviderID:  googleUser.ID,
			IsActive:    true,
			IsVerified:  true,
			LastLoginAt: time.Now(),
			Preferences: models.JSONB{"theme": "light", "currency": "USD", "language": "en"},
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
		// Update existing user
		user.LastLoginAt = time.Now()
		user.AvatarURL = googleUser.Picture
		h.db.Save(&user)
	}

	// Create JWT token
	jwtToken, err := h.createJWTToken(&user)
	if err != nil {
		log.Printf("Error creating JWT token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create session
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

	// Set session cookie
	h.setSessionCookie(w, r, jwtToken)

	// Redirect to frontend
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, jwtToken)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get token from context or cookie
	token := r.Context().Value("token").(string)
	if token == "" {
		cookie, err := r.Cookie("pulseexpends_session")
		if err == nil {
			token = cookie.Value
		}
	}

	// Invalidate session in database
	if token != "" {
		h.db.Model(&models.UserSession{}).Where("token = ?", token).Update("is_active", false)
	}

	// Clear cookie
	h.clearSessionCookie(w, r)

	response := map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProfile retrieves user profile
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

// UpdateProfile updates user profile
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

	// Update allowed fields
	if req.Username != "" {
		// Check if username is already in use
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

// ChangePassword changes user password
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

	// Verify current password
	if user.Password == "" || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)) != nil {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Hash new password
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

// ForgotPassword handles password reset request
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find user by email
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if user exists or not for security
		response := map[string]interface{}{
			"success": true,
			"message": "If the email exists, a password reset link has been sent",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate reset token (in a real app, you would send an email)
	resetToken := generateRandomString(64)
	// TODO: Store reset token in database with expiration
	// TODO: Send email with reset link

	response := map[string]interface{}{
		"success": true,
		"message": "If the email exists, a password reset link has been sent",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Validate reset token and get user ID
	// For now, just return success
	response := map[string]interface{}{
		"success": true,
		"message": "Password reset successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]

	// TODO: Validate verification token and mark user as verified
	// For now, just return success
	response := map[string]interface{}{
		"success": true,
		"message": "Email verified successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSessions retrieves user sessions
func (h *AuthHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	var sessions []models.UserSession
	if err := h.db.Where("user_id = ? AND is_active = ? AND expires_at > ?", 
		user.ID, true, time.Now()).Find(&sessions).Error; err != nil {
		log.Printf("Error fetching sessions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"sessions": sessions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RevokeSession revokes a specific session
func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	sessionID := mux.Vars(r)["id"]

	// Revoke session
	if err := h.db.Model(&models.UserSession{}).
		Where("id = ? AND user_id = ?", sessionID, user.ID).
		Update("is_active", false).Error; err != nil {
		log.Printf("Error revoking session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Session revoked successfully",
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
		MaxAge:   86400 * 7, // 7 days
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