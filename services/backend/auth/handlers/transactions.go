package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

// GetTransactions retrieves the authenticated user's transactions with optional filters
func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	// Parse query parameters
	query := r.URL.Query()
	circleID := query.Get("circleId")
	startDate := query.Get("startDate")
	endDate := query.Get("endDate")
	category := query.Get("category")
	transactionType := query.Get("type")
	limit := query.Get("limit")
	offset := query.Get("offset")

	// Build base query (GORM soft-delete filters deleted_at IS NULL automatically)
	dbQuery := h.db.Model(&models.Transaction{}).Where("user_id = ?", user.ID)

	// Filter by circle if specified
	if circleID != "" {
		dbQuery = dbQuery.Where("circle_id = ?", circleID)
	}

	// Filter by date range
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

	// Filter by category
	if category != "" {
		dbQuery = dbQuery.Where("category = ?", category)
	}

	// Filter by type
	if transactionType != "" {
		dbQuery = dbQuery.Where("type = ?", transactionType)
	}

	// Apply pagination
	if limit != "" {
		dbQuery = dbQuery.Limit(parseInt(limit, 50))
	}
	if offset != "" {
		dbQuery = dbQuery.Offset(parseInt(offset, 0))
	}

	// Sort by date descending
	dbQuery = dbQuery.Order("date DESC")

	var transactions []models.Transaction
	err := dbQuery.Find(&transactions).Error
	if err != nil {
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	// Get total count for pagination
	var total int64
	h.db.Model(&models.Transaction{}).Where("user_id = ?", user.ID).Count(&total)

	response := map[string]interface{}{
		"success":      true,
		"transactions": transactions,
		"total":        total,
		"limit":        parseInt(limit, 50),
		"offset":       parseInt(offset, 0),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTransaction retrieves a specific transaction by ID
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	var transaction models.Transaction
	err := h.db.Where("id = ? AND user_id = ?", transactionID, user.ID).
		Preload("Attachments").
		Preload("Comments").
		First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found or access denied", http.StatusNotFound)
		return
	}

	// If the transaction belongs to a circle, verify membership
	if transaction.CircleID != nil && *transaction.CircleID != "" {
		var member models.CircleMember
		err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?",
			*transaction.CircleID, user.ID, true).First(&member).Error
		if err != nil {
			http.Error(w, "Access denied to circle transaction", http.StatusForbidden)
			return
		}
	}

	response := map[string]interface{}{
		"success":     true,
		"transaction": transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateTransaction creates a new transaction
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	var req models.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
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

	// Validate transaction type
	validTypes := map[string]bool{"income": true, "expense": true, "transfer": true}
	if !validTypes[req.Type] {
		http.Error(w, "Invalid transaction type", http.StatusBadRequest)
		return
	}

	// If circle transaction, verify membership
	if req.CircleID != nil && *req.CircleID != "" {
		var member models.CircleMember
		err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?",
			*req.CircleID, user.ID, true).First(&member).Error
		if err != nil {
			http.Error(w, "You are not a member of this circle", http.StatusForbidden)
			return
		}
	}

	// Create transaction
	transaction := models.Transaction{
		UserID:        user.ID,
		CircleID:      req.CircleID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Category:      req.Category,
		Description:   req.Description,
		Date:          req.Date,
		Type:          req.Type,
		PaymentMethod: req.PaymentMethod,
		Location:      req.Location,
		Tags:          req.Tags,
		ReceiptURL:    req.ReceiptURL,
		IsRecurring:   req.IsRecurring,
		Status:        "pending", // Default pending for circle transactions
		Notes:         req.Notes,
		Metadata:      req.Metadata,
	}

	// Auto-approve non-circle transactions
	if req.CircleID == nil || *req.CircleID == "" {
		transaction.Status = "approved"
		transaction.ApprovedBy = &user.ID
		now := time.Now()
		transaction.ApprovedAt = &now
	}

	if err := h.db.Create(&transaction).Error; err != nil {
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"message":     "Transaction created successfully",
		"transaction": transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateTransaction updates an existing transaction
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Fetch existing transaction
	var transaction models.Transaction
	err := h.db.Where("id = ? AND user_id = ?", transactionID, user.ID).
		First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found or access denied", http.StatusNotFound)
		return
	}

	// Check permissions for circle transactions
	if transaction.CircleID != nil && *transaction.CircleID != "" {
		var member models.CircleMember
		err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?",
			*transaction.CircleID, user.ID, true).First(&member).Error
		if err != nil || (member.Role != "admin" && member.Role != "owner") {
			http.Error(w, "Insufficient permissions to update circle transaction", http.StatusForbidden)
			return
		}
	}

	var req models.UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Apply updates
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
		transaction.Tags = req.Tags
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
		"success":     true,
		"message":     "Transaction updated successfully",
		"transaction": transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteTransaction soft-deletes a transaction (GORM sets deleted_at)
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Fetch transaction
	var transaction models.Transaction
	err := h.db.Where("id = ? AND user_id = ?", transactionID, user.ID).
		First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found or access denied", http.StatusNotFound)
		return
	}

	// Check permissions for circle transactions
	if transaction.CircleID != nil && *transaction.CircleID != "" {
		var member models.CircleMember
		err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?",
			*transaction.CircleID, user.ID, true).First(&member).Error
		if err != nil || (member.Role != "admin" && member.Role != "owner") {
			http.Error(w, "Insufficient permissions to delete circle transaction", http.StatusForbidden)
			return
		}
	}

	// GORM soft delete (sets deleted_at timestamp)
	if err := h.db.Delete(&transaction).Error; err != nil {
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

// GetCircleTransactions retrieves all transactions for a circle
func (h *TransactionHandler) GetCircleTransactions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["circleId"]

	// Verify circle membership
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
	transactionType := query.Get("type")
	status := query.Get("status")
	limit := query.Get("limit")
	offset := query.Get("offset")

	// Build query
	dbQuery := h.db.Where("circle_id = ?", circleID)

	// Filter by date range
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

	// Filter by category
	if category != "" {
		dbQuery = dbQuery.Where("category = ?", category)
	}

	// Filter by type
	if transactionType != "" {
		dbQuery = dbQuery.Where("type = ?", transactionType)
	}

	// Filter by status
	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	}

	// Apply pagination
	if limit != "" {
		dbQuery = dbQuery.Limit(parseInt(limit, 50))
	}
	if offset != "" {
		dbQuery = dbQuery.Offset(parseInt(offset, 0))
	}

	// Sort by date descending
	dbQuery = dbQuery.Order("date DESC")

	var transactions []models.Transaction
	err = dbQuery.Preload("User").Find(&transactions).Error
	if err != nil {
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	// Get total count for pagination
	var total int64
	h.db.Model(&models.Transaction{}).Where("circle_id = ?", circleID).Count(&total)

	response := map[string]interface{}{
		"success":      true,
		"transactions": transactions,
		"total":        total,
		"limit":        parseInt(limit, 50),
		"offset":       parseInt(offset, 0),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SplitTransaction divides a transaction among multiple users
func (h *TransactionHandler) SplitTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Fetch transaction
	var transaction models.Transaction
	err := h.db.Where("id = ? AND user_id = ?", transactionID, user.ID).
		First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found or access denied", http.StatusNotFound)
		return
	}

	// Only circle transactions can be split
	if transaction.CircleID == nil || *transaction.CircleID == "" {
		http.Error(w, "Only circle transactions can be split", http.StatusBadRequest)
		return
	}

	var req models.CreateSplitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate split amounts
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

	// Create splits
	for _, splitReq := range req.Splits {
		split := models.SplitTransaction{
			TransactionID: transaction.ID,
			UserID:        splitReq.UserID,
			Amount:        splitReq.Amount,
			Percentage:    splitReq.Percentage,
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

// ApproveTransaction approves a pending circle transaction
func (h *TransactionHandler) ApproveTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Fetch transaction
	var transaction models.Transaction
	err := h.db.Where("id = ?", transactionID).First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// Only circle transactions can be approved
	if transaction.CircleID == nil || *transaction.CircleID == "" {
		http.Error(w, "Only circle transactions can be approved", http.StatusBadRequest)
		return
	}

	// Verify admin/owner permissions
	var member models.CircleMember
	err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?",
		*transaction.CircleID, user.ID, true).First(&member).Error

	if err != nil || (member.Role != "admin" && member.Role != "owner") {
		http.Error(w, "Insufficient permissions to approve transaction", http.StatusForbidden)
		return
	}

	// Approve transaction
	transaction.Status = "approved"
	transaction.ApprovedBy = &user.ID
	now := time.Now()
	transaction.ApprovedAt = &now

	if err := h.db.Save(&transaction).Error; err != nil {
		http.Error(w, "Failed to approve transaction", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"message":     "Transaction approved successfully",
		"transaction": transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RejectTransaction rejects a pending circle transaction
func (h *TransactionHandler) RejectTransaction(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Fetch transaction
	var transaction models.Transaction
	err := h.db.Where("id = ?", transactionID).First(&transaction).Error

	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// Only circle transactions can be rejected
	if transaction.CircleID == nil || *transaction.CircleID == "" {
		http.Error(w, "Only circle transactions can be rejected", http.StatusBadRequest)
		return
	}

	// Verify admin/owner permissions
	var member models.CircleMember
	err = h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?",
		*transaction.CircleID, user.ID, true).First(&member).Error

	if err != nil || (member.Role != "admin" && member.Role != "owner") {
		http.Error(w, "Insufficient permissions to reject transaction", http.StatusForbidden)
		return
	}

	// Reject transaction
	transaction.Status = "rejected"
	transaction.ApprovedBy = &user.ID
	now := time.Now()
	transaction.ApprovedAt = &now

	if err := h.db.Save(&transaction).Error; err != nil {
		http.Error(w, "Failed to reject transaction", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"message":     "Transaction rejected successfully",
		"transaction": transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserStats returns aggregated statistics for the authenticated user
func (h *TransactionHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	var stats struct {
		TotalIncome      float64 `json:"totalIncome"`
		TotalExpenses    float64 `json:"totalExpenses"`
		TotalTransfers   float64 `json:"totalTransfers"`
		Balance          float64 `json:"balance"`
		TransactionCount int64   `json:"transactionCount"`
	}

	// Calculate income
	h.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ?", user.ID, "income").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalIncome)

	// Calculate expenses
	h.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ?", user.ID, "expense").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalExpenses)

	// Calculate transfers
	h.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ?", user.ID, "transfer").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalTransfers)

	// Count transactions
	h.db.Model(&models.Transaction{}).
		Where("user_id = ?", user.ID).
		Count(&stats.TransactionCount)

	// Calculate balance
	stats.Balance = stats.TotalIncome - stats.TotalExpenses

	response := map[string]interface{}{
		"success": true,
		"stats":   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCircleStats returns aggregated statistics for a circle
func (h *TransactionHandler) GetCircleStats(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	circleID := vars["circleId"]

	// Verify circle membership
	var member models.CircleMember
	err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
		First(&member).Error

	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var stats struct {
		TotalIncome      float64 `json:"totalIncome"`
		TotalExpenses    float64 `json:"totalExpenses"`
		TotalTransfers   float64 `json:"totalTransfers"`
		Balance          float64 `json:"balance"`
		TransactionCount int64   `json:"transactionCount"`
		MemberCount      int64   `json:"memberCount"`
	}

	// Calculate income (only approved)
	h.db.Model(&models.Transaction{}).
		Where("circle_id = ? AND type = ? AND status = ?", circleID, "income", "approved").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalIncome)

	// Calculate expenses (only approved)
	h.db.Model(&models.Transaction{}).
		Where("circle_id = ? AND type = ? AND status = ?", circleID, "expense", "approved").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalExpenses)

	// Calculate transfers (only approved)
	h.db.Model(&models.Transaction{}).
		Where("circle_id = ? AND type = ? AND status = ?", circleID, "transfer", "approved").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalTransfers)

	// Count transactions (only approved)
	h.db.Model(&models.Transaction{}).
		Where("circle_id = ? AND status = ?", circleID, "approved").
		Count(&stats.TransactionCount)

	// Count active members
	h.db.Model(&models.CircleMember{}).
		Where("circle_id = ? AND is_active = ?", circleID, true).
		Count(&stats.MemberCount)

	// Calculate balance
	stats.Balance = stats.TotalIncome - stats.TotalExpenses

	response := map[string]interface{}{
		"success": true,
		"stats":   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMonthlyStats returns monthly spending breakdown for dashboard charts
func (h *TransactionHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	query := r.URL.Query()
	year := query.Get("year")
	circleID := query.Get("circleId")

	if year == "" {
		year = fmt.Sprintf("%d", time.Now().Year())
	}

	type MonthlyBreakdown struct {
		Month          string  `json:"month"`
		TotalIncome    float64 `json:"totalIncome"`
		TotalExpenses  float64 `json:"totalExpenses"`
		TotalTransfers float64 `json:"totalTransfers"`
	}

	var breakdown []MonthlyBreakdown

	baseQuery := h.db.Model(&models.Transaction{}).
		Where("user_id = ? AND EXTRACT(YEAR FROM date) = ?", user.ID, year)

	if circleID != "" {
		// Verify membership
		var member models.CircleMember
		err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
			First(&member).Error
		if err != nil {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		baseQuery = baseQuery.Where("circle_id = ?", circleID)
	}

	// Group by month for dashboard chart data
	rows, err := baseQuery.
		Select("TO_CHAR(date, 'YYYY-MM') as month, "+
			"COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as total_income, "+
			"COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as total_expenses, "+
			"COALESCE(SUM(CASE WHEN type = 'transfer' THEN amount ELSE 0 END), 0) as total_transfers").
		Group("TO_CHAR(date, 'YYYY-MM')").
		Order("month ASC").
		Rows()

	if err != nil {
		http.Error(w, "Failed to fetch monthly stats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var m MonthlyBreakdown
		if err := rows.Scan(&m.Month, &m.TotalIncome, &m.TotalExpenses, &m.TotalTransfers); err != nil {
			continue
		}
		breakdown = append(breakdown, m)
	}

	response := map[string]interface{}{
		"success":   true,
		"year":      year,
		"breakdown": breakdown,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCategoryStats returns spending breakdown by category for dashboard charts
func (h *TransactionHandler) GetCategoryStats(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	query := r.URL.Query()
	circleID := query.Get("circleId")
	startDate := query.Get("startDate")
	endDate := query.Get("endDate")

	type CategoryBreakdown struct {
		Category string  `json:"category"`
		Amount   float64 `json:"amount"`
		Count    int64   `json:"count"`
	}

	var breakdown []CategoryBreakdown

	dbQuery := h.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ?", user.ID, "expense")

	if circleID != "" {
		// Verify membership
		var member models.CircleMember
		err := h.db.Where("circle_id = ? AND user_id = ? AND is_active = ?", circleID, user.ID, true).
			First(&member).Error
		if err != nil {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		dbQuery = dbQuery.Where("circle_id = ?", circleID)
	}

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

	// Group by category for dashboard pie/bar charts
	rows, err := dbQuery.
		Select("category, COALESCE(SUM(amount), 0) as amount, COUNT(*) as count").
		Group("category").
		Order("amount DESC").
		Rows()

	if err != nil {
		http.Error(w, "Failed to fetch category stats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var c CategoryBreakdown
		if err := rows.Scan(&c.Category, &c.Amount, &c.Count); err != nil {
			continue
		}
		breakdown = append(breakdown, c)
	}

	response := map[string]interface{}{
		"success":   true,
		"breakdown": breakdown,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// parseInt is a helper to parse a string as int with a default fallback
func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
