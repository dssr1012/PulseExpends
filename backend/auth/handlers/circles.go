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

// GetCircles obtiene todos los círculos del usuario
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

// GetCircle obtiene un círculo específico
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

// CreateCircle crea un nuevo círculo
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

	// Validar campos requeridos
	if req.Name == "" {
		http.Error(w, "Circle name is required", http.StatusBadRequest)
		return
	}

	// Crear círculo
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

	// Agregar creador como admin
	member := models.CircleMember{
		CircleID: circle.ID,
		UserID:   user.ID,
		Role:     "owner",
		JoinedAt: time.Now(),
		IsActive: true,
	}

	if err := h.db.Create(&member).Error; err != nil {
		// Revertir creación del círculo
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

// UpdateCircle actualiza un círculo existente
func (h *CircleHandler) UpdateCircle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Verificar permisos (solo admin/owner puede actualizar)
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil || (member.Role != "admin" && member.Role != "owner") {
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

	var circle models.Circle
	if err := h.db.Where("id = ? AND is_active = ?", circleID, true).First(&circle).Error; err != nil {
		http.Error(w, "Circle not found", http.StatusNotFound)
		return
	}

	// Actualizar campos
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

// DeleteCircle elimina un círculo (soft delete)
func (h *CircleHandler) DeleteCircle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Verificar permisos (solo owner puede eliminar)
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil || member.Role != "owner" {
		http.Error(w, "Only circle owner can delete the circle", http.StatusForbidden)
		return
	}

	// Soft delete del círculo
	if err := h.db.Model(&models.Circle{}).Where("id = ?", circleID).Update("is_active", false).Error; err != nil {
		http.Error(w, "Failed to delete circle", http.StatusInternalServerError)
		return
	}

	// Desactivar todos los miembros
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

// GetCircleMembers obtiene los miembros de un círculo
func (h *CircleHandler) GetCircleMembers(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Verificar que el usuario sea miembro del círculo
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
		http.Error(w, "Failed to fetch members", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"members": members,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddMember agrega un miembro al círculo
func (h *CircleHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Verificar permisos (solo admin/owner puede agregar miembros)
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil || (member.Role != "admin" && member.Role != "owner") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		UserID uuid.UUID `json:"userId"`
		Role   string    `json:"role"`
		Email  string    `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Buscar usuario por ID o email
	var targetUser models.User
	if req.UserID != uuid.Nil {
		err = h.db.Where("id = ? AND is_active = ?", req.UserID, true).First(&targetUser).Error
	} else if req.Email != "" {
		err = h.db.Where("email = ? AND is_active = ?", req.Email, true).First(&targetUser).Error
	} else {
		http.Error(w, "Either userId or email is required", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Verificar que el usuario no sea ya miembro
	var existingMember models.CircleMember
	err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, targetUser.ID, true).
		First(&existingMember).Error

	if err == nil {
		http.Error(w, "User is already a member of this circle", http.StatusConflict)
		return
	}

	// Determinar rol (default: member)
	role := "member"
	if req.Role != "" {
		role = req.Role
	}

	// Agregar miembro
	newMember := models.CircleMember{
		CircleID: uuid.MustParse(circleID),
		UserID:   targetUser.ID,
		Role:     role,
		JoinedAt: time.Now(),
		IsActive: true,
	}

	if err := h.db.Create(&newMember).Error; err != nil {
		http.Error(w, "Failed to add member", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Member added successfully",
		"member":  newMember,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RemoveMember elimina un miembro del círculo
func (h *CircleHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]
	userID := vars["userId"]

	// Verificar permisos (solo admin/owner puede eliminar miembros, y no puede eliminarse a sí mismo)
	var requesterMember models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&requesterMember).Error

	if err != nil || (requesterMember.Role != "admin" && requesterMember.Role != "owner") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// Verificar que no se está intentando eliminar al owner
	if userID == user.ID.String() {
		http.Error(w, "Cannot remove yourself as owner", http.StatusBadRequest)
		return
	}

	// Verificar que el miembro objetivo existe
	var targetMember models.CircleMember
	err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, userID, true).
		First(&targetMember).Error

	if err != nil {
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}

	// Soft delete del miembro
	if err := h.db.Model(&targetMember).Update("is_active", false).Error; err != nil {
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

// UpdateMemberRole actualiza el rol de un miembro
func (h *CircleHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]
	userID := vars["userId"]

	// Verificar permisos (solo owner puede cambiar roles)
	var requesterMember models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&requesterMember).Error

	if err != nil || requesterMember.Role != "owner" {
		http.Error(w, "Only circle owner can change roles", http.StatusForbidden)
		return
	}

	var req struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar rol
	validRoles := map[string]bool{"viewer": true, "member": true, "admin": true}
	if !validRoles[req.Role] {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Verificar que no se está intentando cambiar el rol del owner
	if userID == user.ID.String() {
		http.Error(w, "Cannot change your own role as owner", http.StatusBadRequest)
		return
	}

	// Actualizar rol
	if err := h.db.Model(&models.CircleMember{}).
		Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, userID, true).
		Update("role", req.Role).Error; err != nil {
		http.Error(w, "Failed to update role", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Role updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// JoinCircle permite unirse a un círculo con código de invitación
func (h *CircleHandler) JoinCircle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	joinCode := vars["code"]

	// Buscar círculo por código de invitación
	var circle models.Circle
	err := h.db.Where("join_code = ? AND is_active = ? AND is_public = ?", joinCode, true, true).
		First(&circle).Error

	if err != nil {
		http.Error(w, "Invalid or expired invitation code", http.StatusNotFound)
		return
	}

	// Verificar que el usuario no sea ya miembro
	var existingMember models.CircleMember
	err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circle.ID, user.ID, true).
		First(&existingMember).Error

	if err == nil {
		http.Error(w, "You are already a member of this circle", http.StatusConflict)
		return
	}

	// Agregar como miembro con rol viewer
	member := models.CircleMember{
		CircleID: circle.ID,
		UserID:   user.ID,
		Role:     "viewer",
		JoinedAt: time.Now(),
		IsActive: true,
	}

	if err := h.db.Create(&member).Error; err != nil {
		http.Error(w, "Failed to join circle", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Successfully joined circle",
		"circle":  circle,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// InviteToCircle invita a un usuario por email
func (h *CircleHandler) InviteToCircle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["id"]

	// Verificar permisos (solo admin/owner puede invitar)
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil || (member.Role != "admin" && member.Role != "owner") {
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

	// Buscar usuario por email
	var targetUser models.User
	err = h.db.Where("email = ? AND is_active = ?", req.Email, true).First(&targetUser).Error

	if err != nil {
		// Crear invitación pendiente
		invite := models.CircleInvite{
			CircleID:   uuid.MustParse(circleID),
			Email:      req.Email,
			Role:       req.Role,
			InvitedBy:  user.ID,
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
			IsAccepted: false,
		}

		if err := h.db.Create(&invite).Error; err != nil {
			http.Error(w, "Failed to create invitation", http.StatusInternalServerError)
			return
		}

		// TODO: Enviar email de invitación
		response := map[string]interface{}{
			"success": true,
			"message": "Invitation sent to email",
			"invite":  invite,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Usuario existe, agregar directamente
	newMember := models.CircleMember{
		CircleID: uuid.MustParse(circleID),
		UserID:   targetUser.ID,
		Role:     req.Role,
		JoinedAt: time.Now(),
		IsActive: true,
	}

	if err := h.db.Create(&newMember).Error; err != nil {
		http.Error(w, "Failed to add member", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "User added to circle",
		"member":  newMember,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function para generar código de invitación
func generateJoinCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(b)
}