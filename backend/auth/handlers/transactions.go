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

type TransactionHandler struct {
	db *gorm.DB
}

func NewTransactionHandler(db *gorm.DB) *TransactionHandler {
	return &TransactionHandler{db: db}
}

// GetTransactions obtiene las transacciones del usuario
func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	// Parsear parámetros de consulta
	query := r.URL.Query()
	circleID := query.Get("circleId")
	startDate := query.Get("startDate")
	endDate := query.Get("endDate")
	category := query.Get("category")
	transactionType := query.Get("type")
	limit := query.Get("limit")
	offset := query.Get("offset")

	// Construir consulta base
	dbQuery := h.db.Model(&models.Transaction{}).Where("user_id = ? AND is_active = ?", user.ID, true)

	// Filtrar por círculo si se especifica
	if circleID != "" {
		dbQuery = dbQuery.Where("circle_id = ?", circleID)
	}

	// Filtrar por fecha
	if startDate != "" {
		start, err := time.Parse(time.RFC3339, startDate)
		if err == nil {
			dbQuery = dbQuery.Where("date >= ?", start)
		}
	}
	if endDate != "" {
		end, err := time.Parse(time.RFC3339, endDate)
		if err == nil {
			dbQuery = dbQuery.Where("date <= ?", end)
		}
	}

	// Filtrar por categoría
	if category != "" {
		dbQuery = dbQuery.Where("category = ?", category)
	}

	// Filtrar por tipo
	if transactionType != "" {
		dbQuery = dbQuery.Where("type = ?", transactionType)
	}

	// Aplicar paginación
	if limit != "" {
		dbQuery = dbQuery.Limit(parseInt(limit, 50))
	}
	if offset != "" {
		dbQuery = dbQuery.Offset(parseInt(offset, 0))
	}

	// Ordenar por fecha descendente
	dbQuery = dbQuery.Order("date DESC")

	var transactions []models.Transaction
	err := dbQuery.Find(&transactions).Error
	if err != nil {
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	// Obtener total para paginación
	var total int64
	h.db.Model(&models.Transaction{}).Where("user_id = ? AND is_active = ?", user.ID, true).Count(&total)

	response := map[string]interface{}{
		"success": true,
		"transactions": transactions,
		"total": total,
		"limit": parseInt(limit, 50),
		"offset": parseInt(offset, 0),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTransaction obtiene una transacción específica
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	var transaction models.Transaction
	err := h.db.Where("id = ? AND user_id = ? AND is_active = ?", transactionID, user.ID, true).
		Preload("Splits").
		First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found or access denied", http.StatusNotFound)
		return
	}

	// Si la transacción pertenece a un círculo, verificar permisos
	if transaction.CircleID != uuid.Nil {
		var member models.CircleMember
		err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", 
			transaction.CircleID, user.ID, true).First(&member).Error
		if err != nil {
			http.Error(w, "Access denied to circle transaction", http.StatusForbidden)
			return
		}
	}

	response := map[string]interface{}{
		"success": true,
		"transaction": transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateTransaction crea una nueva transacción
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	var req struct {
		CircleID      uuid.UUID   `json:"circleId"`
		Amount        float64     `json:"amount"`
		Currency      string      `json:"currency"`
		Category      string      `json:"category"`
		Description   string      `json:"description"`
		Date          time.Time   `json:"date"`
		Type          string      `json:"type"` // income, expense, transfer
		PaymentMethod string      `json:"paymentMethod"`
		Location      string      `json:"location"`
		Tags          []string    `json:"tags"`
		ReceiptURL    string      `json:"receiptUrl"`
		IsRecurring   bool        `json:"isRecurring"`
		RecurringID   uuid.UUID   `json:"recurringId"`
		Notes         string      `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar campos requeridos
	if req.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}
	if req.Category == "" {
		http.Error(w, "Category is required", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		http.Error(w, "Type is required", http.StatusBadRequest)
		return
	}

	// Validar tipo
	validTypes := map[string]bool{"income": true, "expense": true, "transfer": true}
	if !validTypes[req.Type] {
		http.Error(w, "Invalid transaction type", http.StatusBadRequest)
		return
	}

	// Si es transacción de círculo, verificar permisos
	if req.CircleID != uuid.Nil {
		var member models.CircleMember
		err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", 
			req.CircleID, user.ID, true).First(&member).Error
		if err != nil {
			http.Error(w, "You are not a member of this circle", http.StatusForbidden)
			return
		}
	}

	// Crear transacción
	transaction := models.Transaction{
		UserID:       user.ID,
		CircleID:     req.CircleID,
		Amount:       req.Amount,
		Currency:     req.Currency,
		Category:     req.Category,
		Description:  req.Description,
		Date:         req.Date,
		Type:         req.Type,
		PaymentMethod: req.PaymentMethod,
		Location:     req.Location,
		Tags:         models.JSONB{"tags": req.Tags},
		ReceiptURL:   req.ReceiptURL,
		IsRecurring:  req.IsRecurring,
		RecurringID:  req.RecurringID,
		Status:       "pending", // Por defecto pendiente para transacciones de círculo
		Notes:        req.Notes,
		Metadata:     models.JSONB{},
	}

	// Si no es transacción de círculo, aprobar automáticamente
	if req.CircleID == uuid.Nil {
		transaction.Status = "approved"
		transaction.ApprovedBy = user.ID
		transaction.ApprovedAt = time.Now()
	}

	if err := h.db.Create(&transaction).Error; err != nil {
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Transaction created successfully",
		"transaction": transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateTransaction actualiza una transacción existente
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Obtener transacción existente
	var transaction models.Transaction
	err := h.db.Where("id = ? AND user_id = ? AND is_active = ?", transactionID, user.ID, true).
		First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found or access denied", http.StatusNotFound)
		return
	}

	// Verificar permisos para transacciones de círculo
	if transaction.CircleID != uuid.Nil {
		var member models.CircleMember
		err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", 
			transaction.CircleID, user.ID, true).First(&member).Error
		if err != nil || (member.Role != "admin" && member.Role != "owner") {
			http.Error(w, "Insufficient permissions to update circle transaction", http.StatusForbidden)
			return
		}
	}

	var req struct {
		Amount        *float64   `json:"amount"`
		Currency      string     `json:"currency"`
		Category      string     `json:"category"`
		Description   string     `json:"description"`
		Date          *time.Time `json:"date"`
		Type          string     `json:"type"`
		PaymentMethod string     `json:"paymentMethod"`
		Location      string     `json:"location"`
		Tags          []string   `json:"tags"`
		ReceiptURL    string     `json:"receiptUrl"`
		Notes         string     `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Actualizar campos
	if req.Amount != nil && *req.Amount > 0 {
		transaction.Amount = *req.Amount
	}
	if req.Currency != "" {
		transaction.Currency = req.Currency
	}
	if req.Category != "" {
		transaction.Category = req.Category
	}
	if req.Description != "" {
		transaction.Description = req.Description
	}
	if req.Date != nil {
		transaction.Date = *req.Date
	}
	if req.Type != "" {
		transaction.Type = req.Type
	}
	if req.PaymentMethod != "" {
		transaction.PaymentMethod = req.PaymentMethod
	}
	if req.Location != "" {
		transaction.Location = req.Location
	}
	if req.Tags != nil {
		transaction.Tags = models.JSONB{"tags": req.Tags}
	}
	if req.ReceiptURL != "" {
		transaction.ReceiptURL = req.ReceiptURL
	}
	if req.Notes != "" {
		transaction.Notes = req.Notes
	}

	if err := h.db.Save(&transaction).Error; err != nil {
		http.Error(w, "Failed to update transaction", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Transaction updated successfully",
		"transaction": transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteTransaction elimina una transacción (soft delete)
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Obtener transacción
	var transaction models.Transaction
	err := h.db.Where("id = ? AND user_id = ? AND is_active = ?", transactionID, user.ID, true).
		First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found or access denied", http.StatusNotFound)
		return
	}

	// Verificar permisos para transacciones de círculo
	if transaction.CircleID != uuid.Nil {
		var member models.CircleMember
		err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", 
			transaction.CircleID, user.ID, true).First(&member).Error
		if err != nil || (member.Role != "admin" && member.Role != "owner") {
			http.Error(w, "Insufficient permissions to delete circle transaction", http.StatusForbidden)
			return
		}
	}

	// Soft delete
	if err := h.db.Model(&transaction).Update("is_active", false).Error; err != nil {
		http.Error(w, "Failed to delete transaction", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Transaction deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCircleTransactions obtiene transacciones de un círculo
func (h *TransactionHandler) GetCircleTransactions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["circleId"]

	// Verificar que el usuario sea miembro del círculo
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parsear parámetros de consulta
	query := r.URL.Query()
	startDate := query.Get("startDate")
	endDate := query.Get("endDate")
	category := query.Get("category")
	transactionType := query.Get("type")
	status := query.Get("status")
	limit := query.Get("limit")
	offset := query.Get("offset")

	// Construir consulta
	dbQuery := h.db.Where("circle_id = ? AND is_active = ?", circleID, true)

	// Filtrar por fecha
	if startDate != "" {
		start, err := time.Parse(time.RFC3339, startDate)
		if err == nil {
			dbQuery = dbQuery.Where("date >= ?", start)
		}
	}
	if endDate != "" {
		end, err := time.Parse(time.RFC3339, endDate)
		if err == nil {
			dbQuery = dbQuery.Where("date <= ?", end)
		}
	}

	// Filtrar por categoría
	if category != "" {
		dbQuery = dbQuery.Where("category = ?", category)
	}

	// Filtrar por tipo
	if transactionType != "" {
		dbQuery = dbQuery.Where("type = ?", transactionType)
	}

	// Filtrar por estado
	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	}

	// Aplicar paginación
	if limit != "" {
		dbQuery = dbQuery.Limit(parseInt(limit, 50))
	}
	if offset != "" {
		dbQuery = dbQuery.Offset(parseInt(offset, 0))
	}

	// Ordenar por fecha descendente
	dbQuery = dbQuery.Order("date DESC")

	var transactions []models.Transaction
	err = dbQuery.Preload("User").Find(&transactions).Error
	if err != nil {
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	// Obtener total para paginación
	var total int64
	h.db.Model(&models.Transaction{}).Where("circle_id = ? AND is_active = ?", circleID, true).Count(&total)

	response := map[string]interface{}{
		"success": true,
		"transactions": transactions,
		"total": total,
		"limit": parseInt(limit, 50),
		"offset": parseInt(offset, 0),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SplitTransaction divide una transacción entre múltiples usuarios
func (h *TransactionHandler) SplitTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Obtener transacción
	var transaction models.Transaction
	err := h.db.Where("id = ? AND user_id = ? AND is_active = ?", transactionID, user.ID, true).
		First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found or access denied", http.StatusNotFound)
		return
	}

	// Verificar que sea transacción de círculo
	if transaction.CircleID == uuid.Nil {
		http.Error(w, "Only circle transactions can be split", http.StatusBadRequest)
		return
	}

	var req struct {
		Splits []struct {
			UserID uuid.UUID `json:"userId"`
			Amount float64   `json:"amount"`
			Notes  string    `json:"notes"`
		} `json:"splits"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar splits
	totalAmount := 0.0
	for _, split := range req.Splits {
		if split.Amount <= 0 {
			http.Error(w, "Split amounts must be positive", http.StatusBadRequest)
			return
		}
		totalAmount += split.Amount
	}

	if totalAmount != transaction.Amount {
		http.Error(w, "Split amounts must equal transaction amount", http.StatusBadRequest)
		return
	}

	// Crear splits
	for _, splitReq := range req.Splits {
		split := models.TransactionSplit{
			TransactionID: transaction.ID,
			UserID:        splitReq.UserID,
			Amount:        splitReq.Amount,
			Currency:      transaction.Currency,
			Status:        "pending",
			Notes:         splitReq.Notes,
		}

		if err := h.db.Create(&split).Error; err != nil {
			http.Error(w, "Failed to create split", http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Transaction split successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ApproveTransaction aprueba una transacción de círculo
func (h *TransactionHandler) ApproveTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Obtener transacción
	var transaction models.Transaction
	err := h.db.Where("id = ? AND is_active = ?", transactionID, true).
		First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// Verificar que sea transacción de círculo
	if transaction.CircleID == uuid.Nil {
		http.Error(w, "Only circle transactions can be approved", http.StatusBadRequest)
		return
	}

	// Verificar permisos (solo admin/owner puede aprobar)
	var member models.CircleMember
	err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", 
		transaction.CircleID, user.ID, true).First(&member).Error

	if err != nil || (member.Role != "admin" && member.Role != "owner") {
		http.Error(w, "Insufficient permissions to approve transaction", http.StatusForbidden)
		return
	}

	// Aprobar transacción
	transaction.Status = "approved"
	transaction.ApprovedBy = user.ID
	transaction.ApprovedAt = time.Now()

	if err := h.db.Save(&transaction).Error; err != nil {
		http.Error(w, "Failed to approve transaction", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Transaction approved successfully",
		"transaction": transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RejectTransaction rechaza una transacción de círculo
func (h *TransactionHandler) RejectTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Obtener transacción
	var transaction models.Transaction
	err := h.db.Where("id = ? AND is_active = ?", transactionID, true).
		First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// Verificar que sea transacción de círculo
	if transaction.CircleID == uuid.Nil {
		http.Error(w, "Only circle transactions can be rejected", http.StatusBadRequest)
		return
	}

	// Verificar permisos (solo admin/owner puede rechazar)
	var member models.CircleMember
	err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", 
		transaction.CircleID, user.ID, true).First(&member).Error

	if err != nil || (member.Role != "admin" && member.Role != "owner") {
		http.Error(w, "Insufficient permissions to reject transaction", http.StatusForbidden)
		return
	}

	// Rechazar transacción
	transaction.Status = "rejected"
	transaction.ApprovedBy = user.ID
	transaction.ApprovedAt = time.Now()

	if err := h.db.Save(&transaction).Error; err != nil {
		http.Error(w, "Failed to reject transaction", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Transaction rejected successfully",
		"transaction": transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserStats obtiene estadísticas del usuario
func (h *TransactionHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	// Obtener estadísticas básicas
	var stats struct {
		TotalIncome    float64 `json:"totalIncome"`
		TotalExpenses  float64 `json:"totalExpenses"`
		TotalTransfers float64 `json:"totalTransfers"`
		Balance        float64 `json:"balance"`
		TransactionCount int64  `json:"transactionCount"`
	}

	// Calcular ingresos
	h.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND is_active = ?", user.ID, "income", true).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalIncome)

	// Calcular gastos
	h.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND is_active = ?", user.ID, "expense", true).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalExpenses)

	// Calcular transferencias
	h.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND is_active = ?", user.ID, "transfer", true).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalTransfers)

	// Contar transacciones
	h.db.Model(&models.Transaction{}).
		Where("user_id = ? AND is_active = ?", user.ID, true).
		Count(&stats.TransactionCount)

	// Calcular balance
	stats.Balance = stats.TotalIncome - stats.TotalExpenses

	response := map[string]interface{}{
		"success": true,
		"stats": stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCircleStats obtiene estadísticas de un círculo
func (h *TransactionHandler) GetCircleStats(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["circleId"]

	// Verificar que el usuario sea miembro del círculo
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Obtener estadísticas del círculo
	var stats struct {
		TotalIncome    float64 `json:"totalIncome"`
		TotalExpenses  float64 `json:"totalExpenses"`
		TotalTransfers float64 `json:"totalTransfers"`
		Balance        float64 `json:"balance"`
		TransactionCount int64  `json:"transactionCount"`
		MemberCount    int64   `json:"memberCount"`
	}

	// Calcular ingresos
	h.db.Model(&models.Transaction{}).
		Where("circle_id = ? AND type = ? AND is_active = ? AND status = ?", 
			circleID, "income", true, "approved").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalIncome)

	// Calcular gastos
	h.db.Model(&models.Transaction{}).
		Where("circle_id = ? AND type = ? AND is_active = ? AND status = ?", 
			circleID, "expense", true, "approved").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalExpenses)

	// Calcular transferencias
	h.db.Model(&models.Transaction{}).
		Where("circle_id = ? AND type = ? AND is_active = ? AND status = ?", 
			circleID, "transfer", true, "approved").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalTransfers)

	// Contar transacciones
	h.db.Model(&models.Transaction{}).
		Where("circle_id = ? AND is_active = ? AND status = ?", circleID, true, "approved").
		Count(&stats.TransactionCount)

	// Contar miembros
	h.db.Model(&models.CircleMember{}).
		Where("circle_id = ? AND is_active = ?", circleID, true).
		Count(&stats.MemberCount)

	// Calcular balance
	stats.Balance = stats.TotalIncome - stats.TotalExpenses

	response := map[string]interface{}{
		"success": true,
		"stats": stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function para parsear enteros
func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}