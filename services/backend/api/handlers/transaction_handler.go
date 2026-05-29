package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/dssr1012/pulse-expends/internal/service"
	"github.com/dssr1012/pulse-expends/services/backend/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TransactionHandler struct {
	transactionService service.TransactionService
}

func NewTransactionHandler(transactionService service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

// GetTransactions retrieves transactions for the authenticated user
func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	filters := parseTransactionFilters(query, user.ID)

	transactions, err := h.transactionService.GetTransactionsByUser(r.Context(), user.ID, filters)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get transactions: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"transactions": transactions,
		"total":        len(transactions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTransaction retrieves a specific transaction by ID
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	transactionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	transaction, err := h.transactionService.GetTransaction(r.Context(), transactionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to this transaction
	if transaction.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
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
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var transaction model.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Set user ID
	transaction.UserID = user.ID

	// Validate transaction
	if err := h.transactionService.ValidateTransaction(r.Context(), &transaction); err != nil {
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Check for duplicates
	duplicates, err := h.transactionService.DetectDuplicateTransactions(r.Context(), &transaction)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check for duplicates: %v", err), http.StatusInternalServerError)
		return
	}

	if len(duplicates) > 0 {
		response := map[string]interface{}{
			"success":    false,
			"error":      "Duplicate transaction detected",
			"duplicates": duplicates,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create transaction
	if err := h.transactionService.CreateTransaction(r.Context(), &transaction); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create transaction: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"transaction": transaction,
		"message":     "Transaction created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateTransaction updates an existing transaction
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	transactionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	// Get existing transaction
	existingTransaction, err := h.transactionService.GetTransaction(r.Context(), transactionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access
	if existingTransaction.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var updates model.Transaction
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Preserve immutable fields
	updates.ID = existingTransaction.ID
	updates.UserID = existingTransaction.UserID
	updates.CircleID = existingTransaction.CircleID
	updates.CreatedAt = existingTransaction.CreatedAt

	// Validate updates
	if err := h.transactionService.ValidateTransaction(r.Context(), &updates); err != nil {
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Update transaction
	if err := h.transactionService.UpdateTransaction(r.Context(), &updates); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update transaction: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"transaction": updates,
		"message":     "Transaction updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteTransaction deletes a transaction
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	transactionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	// Get transaction to check ownership
	transaction, err := h.transactionService.GetTransaction(r.Context(), transactionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access
	if transaction.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete transaction
	if err := h.transactionService.DeleteTransaction(r.Context(), transactionID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete transaction: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Transaction deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CheckDuplicates checks for duplicate transactions
func (h *TransactionHandler) CheckDuplicates(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	transactionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	// Get transaction
	transaction, err := h.transactionService.GetTransaction(r.Context(), transactionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access
	if transaction.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Check for duplicates
	duplicates, err := h.transactionService.DetectDuplicateTransactions(r.Context(), transaction)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check for duplicates: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"duplicates": duplicates,
		"count":      len(duplicates),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateBatch creates multiple transactions at once
func (h *TransactionHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var transactions []model.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transactions); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Set user ID and validate each transaction
	for i := range transactions {
		transactions[i].UserID = user.ID
		
		if err := h.transactionService.ValidateTransaction(r.Context(), &transactions[i]); err != nil {
			http.Error(w, fmt.Sprintf("Validation failed for transaction %d: %v", i+1, err), http.StatusBadRequest)
			return
		}
	}

	// Create transactions in batch
	if err := h.transactionService.CreateTransactionsBatch(r.Context(), transactions); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create transactions: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"transactions": transactions,
		"count":        len(transactions),
		"message":      fmt.Sprintf("Created %d transactions successfully", len(transactions)),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetSummary gets transaction summary for the user
func (h *TransactionHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	circleID := query.Get("circleId")
	startDateStr := query.Get("startDate")
	endDateStr := query.Get("endDate")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			http.Error(w, "Invalid start date format", http.StatusBadRequest)
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			http.Error(w, "Invalid end date format", http.StatusBadRequest)
			return
		}
	}

	// If dates not provided, use last 30 days
	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, 0, -30)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	var circleUUID uuid.UUID
	if circleID != "" {
		circleUUID, err = uuid.Parse(circleID)
		if err != nil {
			http.Error(w, "Invalid circle ID", http.StatusBadRequest)
			return
		}
	}

	var summary *model.TransactionSummary
	if circleID != "" {
		// Get summary for circle
		summary, err = h.transactionService.GetMonthlySummary(r.Context(), circleUUID, endDate.Year(), endDate.Month())
	} else {
		// Get summary for user
		// Note: This would need a new method in the service
		// For now, we'll return a placeholder
		summary = &model.TransactionSummary{
			TotalIncome:     0,
			TotalExpenses:   0,
			NetBalance:      0,
			BaseCurrency:    "USD",
			ByCurrency:      make(map[string]model.CurrencySummary),
			ByCategory:      make(map[string]float64),
			ByPaymentMethod: make(map[string]float64),
			MonthlyTrend:    []model.MonthlyTrend{},
			AnomalyCount:    0,
			WarningCount:    0,
			CriticalCount:   0,
		}
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get summary: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"summary": summary,
		"period": map[string]interface{}{
			"startDate": startDate,
			"endDate":   endDate,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to parse transaction filters from query parameters
func parseTransactionFilters(query map[string][]string, userID uuid.UUID) repository.TransactionFilters {
	filters := repository.TransactionFilters{
		UserID: userID,
	}

	if circleID := query.Get("circleId"); circleID != "" {
		if circleUUID, err := uuid.Parse(circleID); err == nil {
			filters.CircleID = &circleUUID
		}
	}

	if startDate := query.Get("startDate"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filters.StartDate = &t
		}
	}

	if endDate := query.Get("endDate"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filters.EndDate = &t
		}
	}

	if category := query.Get("category"); category != "" {
		filters.Category = category
	}

	if paymentMethod := query.Get("paymentMethod"); paymentMethod != "" {
		filters.PaymentMethod = model.PaymentMethod(paymentMethod)
	}

	if currency := query.Get("currency"); currency != "" {
		filters.Currency = currency
	}

	if minAmount := query.Get("minAmount"); minAmount != "" {
		if val, err := strconv.ParseFloat(minAmount, 64); err == nil {
			filters.MinAmount = &val
		}
	}

	if maxAmount := query.Get("maxAmount"); maxAmount != "" {
		if val, err := strconv.ParseFloat(maxAmount, 64); err == nil {
			filters.MaxAmount = &val
		}
	}

	if isPrivate := query.Get("isPrivate"); isPrivate != "" {
		if val, err := strconv.ParseBool(isPrivate); err == nil {
			filters.IsPrivate = &val
		}
	}

	if isAnomaly := query.Get("isAnomaly"); isAnomaly != "" {
		if val, err := strconv.ParseBool(isAnomaly); err == nil {
			filters.IsAnomaly = &val
		}
	}

	if source := query.Get("source"); source != "" {
		filters.Source = source
	}

	if status := query.Get("status"); status != "" {
		filters.Status = status
	}

	if limit := query.Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil && val > 0 {
			filters.Limit = val
		}
	} else {
		filters.Limit = 50 // Default limit
	}

	if offset := query.Get("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil && val >= 0 {
			filters.Offset = val
		}
	}

	if sortBy := query.Get("sortBy"); sortBy != "" {
		filters.SortBy = sortBy
	} else {
		filters.SortBy = "date" // Default sort by date
	}

	if sortOrder := query.Get("sortOrder"); sortOrder != "" {
		filters.SortOrder = sortOrder
	} else {
		filters.SortOrder = "desc" // Default descending order
	}

	return filters
}