package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"pulseexpends/backend/auth/models"
)

type CircleHandler struct {
	db *gorm.DB
}

func NewCircleHandler(db *gorm.DB) *CircleHandler {
	return &CircleHandler{db: db}
}

// GetCircles retrieves all circles for the user
func (h *CircleHandler) GetCircles(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	var circles []models.Circle
	err := h.db.Joins("JOIN circle_members ON circle_members.circle_id = circles.id").
		Where("circle_members.user_id = ? AND circle_members.is_active = ? AND circles.is_active = ?", 
			user.ID, true, true).
		Preload("Members").
		Find(&circles).Error

	if err != nil {
		http.Error(w, "Failed to fetch circles", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"circles": circles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCircle retrieves a specific circle
func (h *CircleHandler) GetCircle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	var circle models.Circle
	err := h.db.Joins("JOIN circle_members ON circle_members.circle_id = circles.id").
		Where("circles.id = ? AND circle_members.user_id = ? AND circle_members.is_active = ? AND circles.is_active = ?",
			circleID, user.ID, true, true).
		Preload("Members").
		Preload("Transactions").
		First(&circle).Error

	if err != nil {
		http.Error(w, "Circle not found or access denied", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"circle":  circle,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateCircle creates a new circle
func (h *CircleHandler) CreateCircle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Currency    string `json:"currency"`
		IsPublic    bool   `json:"isPublic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "Circle name is required", http.StatusBadRequest)
		return
	}

	// Create circle
	circle := models.Circle{
		Name:        req.Name,
		Description: req.Description,
		Currency:    req.Currency,
		CreatedBy:   user.ID,
		IsActive:    true,
		IsPublic:    req.IsPublic,
		JoinCode:    generateJoinCode(),
		Settings:    models.JSONB{"notifications": true, "default_currency": req.Currency},
	}

	if err := h.db.Create(&circle).Error; err != nil {
		http.Error(w, "Failed to create circle", http.StatusInternalServerError)
		return
	}

	// Add creator as owner
	member := models.CircleMember{
		CircleID: circle.ID,
		UserID:   user.ID,
		Role:     "owner",
		JoinedAt: time.Now(),
		IsActive: true,
	}

	if err := h.db.Create(&member).Error; err != nil {
		// Rollback circle creation
		h.db.Delete(&circle)
		http.Error(w, "Failed to add creator to circle", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Circle created successfully",
		"circle":  circle,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateCircle updates an existing circle
func (h *CircleHandler) UpdateCircle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Check permissions (only admin/owner can update)
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Currency    string `json:"currency"`
		IsPublic    *bool  `json:"isPublic"`
		Settings    models.JSONB `json:"settings"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find circle
	var circle models.Circle
	if err := h.db.Where("id = ? AND is_active = ?", circleID, true).First(&circle).Error; err != nil {
		http.Error(w, "Circle not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.Name != "" {
		circle.Name = req.Name
	}
	if req.Description != "" {
		circle.Description = req.Description
	}
	if req.Currency != "" {
		circle.Currency = req.Currency
	}
	if req.IsPublic != nil {
		circle.IsPublic = *req.IsPublic
	}
	if req.Settings != nil {
		circle.Settings = req.Settings
	}

	if err := h.db.Save(&circle).Error; err != nil {
		http.Error(w, "Failed to update circle", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Circle updated successfully",
		"circle":  circle,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteCircle deletes a circle (soft delete)
func (h *CircleHandler) DeleteCircle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Check permissions (only owner can delete)
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil || member.Role != "owner" {
		http.Error(w, "Only the owner can delete the circle", http.StatusForbidden)
		return
	}

	// Soft delete circle
	if err := h.db.Model(&models.Circle{}).Where("id = ?", circleID).Update("is_active", false).Error; err != nil {
		http.Error(w, "Failed to delete circle", http.StatusInternalServerError)
		return
	}

	// Deactivate all members
	if err := h.db.Model(&models.CircleMember{}).Where("circle_id = ?", circleID).Update("is_active", false).Error; err != nil {
		http.Error(w, "Failed to deactivate circle members", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Circle deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCircleMembers retrieves all members of a circle
func (h *CircleHandler) GetCircleMembers(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Verify user is a member of the circle
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var members []models.CircleMember
	err = h.db.Where("circle_id = ? AND is_active = ?", circleID, true).
		Preload("User").
		Find(&members).Error

	if err != nil {
		http.Error(w, "Failed to fetch circle members", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"members": members,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddMember adds a new member to a circle
func (h *CircleHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Check permissions (only admin/owner can add members)
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		UserID string `json:"userId"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate role
	if req.Role != "member" && req.Role != "admin" && req.Role != "viewer" {
		http.Error(w, "Invalid role. Must be 'member', 'admin', or 'viewer'", http.StatusBadRequest)
		return
	}

	// Check if user exists
	var targetUser models.User
	if err := h.db.Where("id = ? AND is_active = ?", req.UserID, true).First(&targetUser).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if user is already a member
	var existingMember models.CircleMember
	err = h.db.Where("circle_id = ? AND user_id = ?", circleID, req.UserID).First(&existingMember).Error
	if err == nil {
		if existingMember.IsActive {
			http.Error(w, "User is already a member of this circle", http.StatusConflict)
			return
		}
		// Reactivate existing membership
		existingMember.IsActive = true
		existingMember.Role = req.Role
		if err := h.db.Save(&existingMember).Error; err != nil {
			http.Error(w, "Failed to add member", http.StatusInternalServerError)
			return
		}
	} else {
		// Create new membership
		newMember := models.CircleMember{
			CircleID: circleID,
			UserID:   req.UserID,
			Role:     req.Role,
			JoinedAt: time.Now(),
			IsActive: true,
		}
		if err := h.db.Create(&newMember).Error; err != nil {
			http.Error(w, "Failed to add member", http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Member added successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RemoveMember removes a member from a circle
func (h *CircleHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]
	memberID := vars["userId"]

	// Check permissions (only admin/owner can remove members, owners cannot remove themselves)
	var currentMember models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&currentMember).Error

	if err != nil || (currentMember.Role != "owner" && currentMember.Role != "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// Check if trying to remove self (owners cannot remove themselves)
	if memberID == user.ID && currentMember.Role == "owner" {
		http.Error(w, "Owners cannot remove themselves from the circle", http.StatusBadRequest)
		return
	}

	// Deactivate membership
	if err := h.db.Model(&models.CircleMember{}).
		Where("circle_id = ? AND user_id = ?", circleID, memberID).
		Update("is_active", false).Error; err != nil {
		http.Error(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Member removed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateMemberRole updates a member's role in a circle
func (h *CircleHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]
	memberID := vars["userId"]

	// Check permissions (only owner can change roles)
	var currentMember models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&currentMember).Error

	if err != nil || currentMember.Role != "owner" {
		http.Error(w, "Only the owner can change member roles", http.StatusForbidden)
		return
	}

	var req struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate role
	if req.Role != "owner" && req.Role != "admin" && req.Role != "member" && req.Role != "viewer" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Update role
	if err := h.db.Model(&models.CircleMember{}).
		Where("circle_id = ? AND user_id = ?", circleID, memberID).
		Update("role", req.Role).Error; err != nil {
		http.Error(w, "Failed to update member role", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Member role updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// JoinCircle allows a user to join a circle using a join code
func (h *CircleHandler) JoinCircle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	joinCode := vars["code"]

	// Find circle by join code
	var circle models.Circle
	err := h.db.Where("join_code = ? AND is_active = ? AND is_public = ?", joinCode, true, true).
		First(&circle).Error

	if err != nil {
		http.Error(w, "Invalid or expired join code", http.StatusNotFound)
		return
	}

	// Check if user is already a member
	var existingMember models.CircleMember
	err = h.db.Where("circle_id = ? AND user_id = ?", circle.ID, user.ID).First(&existingMember).Error
	if err == nil {
		if existingMember.IsActive {
			http.Error(w, "You are already a member of this circle", http.StatusConflict)
			return
		}
		// Reactivate membership
		existingMember.IsActive = true
		if err := h.db.Save(&existingMember).Error; err != nil {
			http.Error(w, "Failed to join circle", http.StatusInternalServerError)
			return
		}
	} else {
		// Create new membership as regular member
		newMember := models.CircleMember{
			CircleID: circle.ID,
			UserID:   user.ID,
			Role:     "member",
			JoinedAt: time.Now(),
			IsActive: true,
		}
		if err := h.db.Create(&newMember).Error; err != nil {
			http.Error(w, "Failed to join circle", http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Successfully joined circle",
		"circle":  circle,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// InviteToCircle creates an invitation for a user to join a circle
func (h *CircleHandler) InviteToCircle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Check permissions (only admin/owner can invite)
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate role
	if req.Role != "member" && req.Role != "admin" && req.Role != "viewer" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Find user by email
	var targetUser models.User
	if err := h.db.Where("email = ? AND is_active = ?", req.Email, true).First(&targetUser).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if user is already a member
	var existingMember models.CircleMember
	err = h.db.Where("circle_id = ? AND user_id = ?", circleID, targetUser.ID).First(&existingMember).Error
	if err == nil {
		http.Error(w, "User is already a member of this circle", http.StatusConflict)
		return
	}

	// Create invitation
	invitation := models.CircleInvite{
		CircleID:  circleID,
		Email:     req.Email,
		InvitedBy: user.ID,
		Role:      req.Role,
		Token:     generateRandomString(32),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		Accepted:  false,
	}

	if err := h.db.Create(&invitation).Error; err != nil {
		http.Error(w, "Failed to create invitation", http.StatusInternalServerError)
		return
	}

	// TODO: Send invitation email

	response := map[string]interface{}{
		"success": true,
		"message": "Invitation sent successfully",
		"invitation": map[string]interface{}{
			"id":        invitation.ID,
			"email":     invitation.Email,
			"role":      invitation.Role,
			"expiresAt": invitation.ExpiresAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCircleTransactions retrieves transactions for a circle
func (h *CircleHandler) GetCircleTransactions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Verify user is a member of the circle
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	startDate := query.Get("startDate")
	endDate := query.Get("endDate")
	category := query.Get("category")
	limit := query.Get("limit")
	offset := query.Get("offset")

	// Build query
	dbQuery := h.db.Where("circle_id = ? AND deleted_at IS NULL", circleID)

	if startDate != "" {
		dbQuery = dbQuery.Where("date >= ?", startDate)
	}
	if endDate != "" {
		dbQuery = dbQuery.Where("date <= ?", endDate)
	}
	if category != "" {
		dbQuery = dbQuery.Where("category = ?", category)
	}

	// Apply pagination
	if limit != "" {
		dbQuery = dbQuery.Limit(limit)
	} else {
		dbQuery = dbQuery.Limit(50) // Default limit
	}
	if offset != "" {
		dbQuery = dbQuery.Offset(offset)
	}

	var transactions []models.Transaction
	err = dbQuery.Order("date DESC").Find(&transactions).Error
	if err != nil {
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"transactions": transactions,
		"count":        len(transactions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCircleActivities retrieves recent activities for a circle
func (h *CircleHandler) GetCircleActivities(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Verify user is a member of the circle
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	limit := query.Get("limit")
	if limit == "" {
		limit = "50"
	}

	var activities []models.CircleActivity
	err = h.db.Where("circle_id = ?", circleID).
		Order("created_at DESC").
		Limit(limit).
		Preload("User").
		Find(&activities).Error

	if err != nil {
		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"activities": activities,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCircleStats retrieves statistics for a circle
func (h *CircleHandler) GetCircleStats(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Verify user is a member of the circle
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	startDate := query.Get("startDate")
	endDate := query.Get("endDate")

	// Build base query
	dbQuery := h.db.Model(&models.Transaction{}).
		Where("circle_id = ? AND deleted_at IS NULL", circleID)

	if startDate != "" {
		dbQuery = dbQuery.Where("date >= ?", startDate)
	}
	if endDate != "" {
		dbQuery = dbQuery.Where("date <= ?", endDate)
	}

	// Get total income
	var totalIncome float64
	if err := dbQuery.Where("type = ?", "income").Select("COALESCE(SUM(amount), 0)").Scan(&totalIncome).Error; err != nil {
		http.Error(w, "Failed to calculate income", http.StatusInternalServerError)
		return
	}

	// Get total expenses
	var totalExpenses float64
	if err := dbQuery.Where("type = ?", "expense").Select("COALESCE(SUM(amount), 0)").Scan(&totalExpenses).Error; err != nil {
		http.Error(w, "Failed to calculate expenses", http.StatusInternalServerError)
		return
	}

	// Get category breakdown
	type CategorySummary struct {
		Category string  `json:"category"`
		Total    float64 `json:"total"`
		Count    int     `json:"count"`
	}
	var categoryBreakdown []CategorySummary
	err = dbQuery.Where("type = ?", "expense").
		Select("category, SUM(amount) as total, COUNT(*) as count").
		Group("category").
		Order("total DESC").
		Scan(&categoryBreakdown).Error

	if err != nil {
		http.Error(w, "Failed to calculate category breakdown", http.StatusInternalServerError)
		return
	}

	// Get member count
	var memberCount int64
	if err := h.db.Model(&models.CircleMember{}).
		Where("circle_id = ? AND is_active = ?", circleID, true).
		Count(&memberCount).Error; err != nil {
		http.Error(w, "Failed to count members", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"stats": map[string]interface{}{
			"totalIncome":   totalIncome,
			"totalExpenses": totalExpenses,
			"netBalance":    totalIncome - totalExpenses,
			"memberCount":   memberCount,
			"categories":    categoryBreakdown,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to generate join code
func generateJoinCode() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// Helper function to generate random string
func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}