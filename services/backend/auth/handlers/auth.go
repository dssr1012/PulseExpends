package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	db          *gorm.DB
	oauthConfig *oauth2.Config
	store       *sessions.CookieStore
	jwtSecret   []byte
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	oauthConfig := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		endpoint: google.endpoint,
	}

	cookieSecret := os.Getenv("COOKIE_SECRET")
	if cookieSecret == "" {
		cookieSecret = "change-this-cookie-secret-in-production"
	}

	store := sessions.NewCookieStore([]byte(cookieSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   os.Getenv("COOKIE_SECURE") == "true",
		SameSite: http.SameSiteLaxMode,
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-this-jwt-secret-in-production"
	}

	return &AuthHandler{
		db:          db,
		oauthConfig: oauthConfig,
		store:       store,
		jwtSecret:   []byte(jwtSecret),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
		FullName string `json:"fullName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).error; err == nil {
		http.error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing password: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Auto-generate username from email if not provided
	username := req.Username
	if username == "" {
		// Use the local part of the email as username
		if idx := strings.Index(req.Email, "@"); idx > 0 {
			username = req.Email[:idx]
		} else {
			username = req.Email
		}
	}

	// Create user
	now := time.Now()
	user := models.User{
		Email:      req.Email,
		Username:   username,
		FullName:   req.FullName,
		Provider:   "local",
		IsActive:   true,
		IsVerified: false,
		LastLoginAt: &now,
	}

	if err := h.db.Create(&user).error; err != nil {
		log.Printf("error creating user: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create auth record with hashed password
	userAuth := models.UserAuth{
		UserID:       user.ID,
		PasswordHash: string(hashedPassword),
	}
	if err := h.db.Create(&userAuth).error; err != nil {
		log.Printf("error creating user auth: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create JWT token
	token, err := h.createJWTToken(&user)
	if err != nil {
		log.Printf("error creating JWT token: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create session
	session := models.UserSession{
		IsActive: true,
		UserID:    user.ID,
		Token:     token,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.db.Create(&session).error; err != nil {
		log.Printf("error creating session: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.setSessionCookie(w, r, token)

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
		http.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find user with auth
	var user models.User
	if err := h.db.Where("email = ? AND is_active = ?", req.Email, true).
		Preload("Auth").First(&user).error; err != nil {
		http.error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Verify password against UserAuth.PasswordHash
	if user.Auth.PasswordHash == "" ||
		bcrypt.CompareHashAndPassword([]byte(user.Auth.PasswordHash), []byte(req.Password)) != nil {
		http.error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	h.db.Save(&user)

	// Create JWT token
	token, err := h.createJWTToken(&user)
	if err != nil {
		log.Printf("error creating JWT token: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create session
	session := models.UserSession{
		IsActive: true,
		UserID:    user.ID,
		Token:     token,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.db.Create(&session).error; err != nil {
		log.Printf("error creating session: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.setSessionCookie(w, r, token)

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
	state := generateOAuthState()

	sess, _ := h.store.Get(r, "oauth-state")
	sess.Values["state"] = state
	sess.Save(r, w)

	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles Google OAuth callback
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state
	sess, _ := h.store.Get(r, "oauth-state")
	state := sess.Values["state"]
	delete(sess.Values, "state")
	sess.Save(r, w)

	if r.URL.Query().Get("state") != state {
		http.error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	code := r.URL.Query().Get("code")
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("error exchanging code: %v", err)
		http.error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	client := h.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("error getting user info: %v", err)
		http.error(w, "Failed to get user info", http.StatusInternalServerError)
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
		log.Printf("error decoding user info: %v", err)
		http.error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Find or create user
	var user models.User
	err = h.db.Where("provider = ? AND provider_id = ?", "google", googleUser.ID).First(&user).error

	if err == gorm.ErrRecordNotFound {
		now := time.Now()
		user = models.User{
			Email:      googleUser.Email,
			FullName:   googleUser.Name,
			AvatarURL:  googleUser.Picture,
			Provider:   "google",
			ProviderID: googleUser.ID,
			IsActive:   true,
			IsVerified: true,
			LastLoginAt: &now,
		}

		if err := h.db.Create(&user).error; err != nil {
			log.Printf("error creating user: %v", err)
			http.error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.Printf("error finding user: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else {
		now := time.Now()
		user.LastLoginAt = &now
		user.AvatarURL = googleUser.Picture
		h.db.Save(&user)
	}

	// Create JWT token
	jwtToken, err := h.createJWTToken(&user)
	if err != nil {
		log.Printf("error creating JWT token: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create session
	session := models.UserSession{
		IsActive: true,
		UserID:    user.ID,
		Token:     jwtToken,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.db.Create(&session).error; err != nil {
		log.Printf("error creating session: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

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
	token := ""
	if ctxToken := r.Context().Value("token"); ctxToken != nil {
		token = ctxToken.(string)
	}
	if token == "" {
		cookie, err := r.Cookie("pulseexpends_session")
		if err == nil {
			token = cookie.Value
		}
	}

	// Delete session from database
	if token != "" {
		h.db.Where("token = ?", token).Delete(&models.UserSession{})
	}

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
			"id":          user.ID,
			"email":       user.Email,
			"username":    user.Username,
			"fullName":    user.FullName,
			"avatarURL":   user.AvatarURL,
			"provider":    user.Provider,
			"isVerified":  user.IsVerified,
			"lastLogin":   user.LastLoginAt,
			"preferences": user.Preferences,
			"createdAt":   user.CreatedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProfile updates user profile
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	var req struct {
		Username  string `json:"username"`
		FullName  string `json:"fullName"`
		AvatarURL string `json:"avatarURL"`
		Theme     string `json:"theme"`
		Language  string `json:"language"`
		Currency  string `json:"currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username != "" {
		var existingUser models.User
		if err := h.db.Where("username = ? AND id != ?", req.Username, user.ID).First(&existingUser).error; err == nil {
			http.error(w, "Username already in use", http.StatusConflict)
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
	if req.Theme != "" {
		user.Preferences.Theme = req.Theme
	}
	if req.Language != "" {
		user.Preferences.Language = req.Language
	}
	if req.Currency != "" {
		user.Preferences.DefaultCurrency = req.Currency
	}

	if err := h.db.Save(&user).error; err != nil {
		log.Printf("error updating user: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Profile updated successfully",
		"user": map[string]interface{}{
			"id":          user.ID,
			"email":       user.Email,
			"username":    user.Username,
			"fullName":    user.FullName,
			"avatarURL":   user.AvatarURL,
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
		http.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Fetch auth record
	var userAuth models.UserAuth
	if err := h.db.Where("user_id = ?", user.ID).First(&userAuth).error; err != nil {
		http.error(w, "Auth record not found", http.StatusInternalServerError)
		return
	}

	// Verify current password
	if userAuth.PasswordHash == "" ||
		bcrypt.CompareHashAndPassword([]byte(userAuth.PasswordHash), []byte(req.CurrentPassword)) != nil {
		http.error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing password: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userAuth.PasswordHash = string(hashedPassword)
	if err := h.db.Save(&userAuth).error; err != nil {
		log.Printf("error updating password: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
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
		http.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Always return the same response to prevent email enumeration
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
		http.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Validate reset token and update password
	response := map[string]interface{}{
		"success": true,
		"message": "Password reset successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	// TODO: Validate verification token and mark user as verified
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
	if err := h.db.Where("user_id = ? AND expires_at > ?", user.ID, time.Now()).
		Find(&sessions).error; err != nil {
		log.Printf("error fetching sessions: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"sessions": sessions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RevokeSession revokes a specific session
func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	sessionID := mux.Vars(r)["id"]

	if err := h.db.Where("id = ? AND user_id = ?", sessionID, user.ID).
		Delete(&models.UserSession{}).error; err != nil {
		log.Printf("error revoking session: %v", err)
		http.error(w, "Internal server error", http.StatusInternalServerError)
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
		"sub":   user.ID,
		"email": user.Email,
		"exp":   expiry.Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	cookie := &http.Cookie{
		Name:     "pulseexpends_session",
		Value:    token,
		Path:     "/",
		MaxAge:   86400 * 7,
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

// generateOAuthState generates a random OAuth state string
func generateOAuthState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
