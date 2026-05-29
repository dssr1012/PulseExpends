package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/dssr1012/pulse-expends/internal/service"
	"github.com/dssr1012/pulse-expends/services/backend/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type CreditCardHandler struct {
	creditCardService service.CreditCardService
}

func NewCreditCardHandler(creditCardService service.CreditCardService) *CreditCardHandler {
	return &CreditCardHandler{
		creditCardService: creditCardService,
	}
}

// GetCreditCards retrieves credit cards for the authenticated user
func (h *CreditCardHandler) GetCreditCards(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	circleID := query.Get("circleId")
	activeOnly := query.Get("activeOnly") == "true"

	var cards []model.CreditCard
	var err error

	if circleID != "" {
		circleUUID, err := uuid.Parse(circleID)
		if err != nil {
			http.Error(w, "Invalid circle ID", http.StatusBadRequest)
			return
		}
		cards, err = h.creditCardService.GetFamilyCreditCards(r.Context(), circleUUID)
	} else if activeOnly {
		cards, err = h.creditCardService.GetActiveCreditCards(r.Context(), user.ID)
	} else {
		cards, err = h.creditCardService.GetUserCreditCards(r.Context(), user.ID)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get credit cards: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"creditCards": cards,
		"count":      len(cards),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCreditCard retrieves a specific credit card by ID
func (h *CreditCardHandler) GetCreditCard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	cardID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid credit card ID", http.StatusBadRequest)
		return
	}

	card, err := h.creditCardService.GetCreditCard(r.Context(), cardID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get credit card: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to this card
	if card.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"creditCard": card,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddCreditCard adds a new credit card
func (h *CreditCardHandler) AddCreditCard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var card model.CreditCard
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Set user ID
	card.UserID = user.ID

	// Validate credit card
	if err := h.creditCardService.ValidateCreditCard(r.Context(), &card); err != nil {
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Add credit card
	if err := h.creditCardService.AddCreditCard(r.Context(), &card); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add credit card: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"creditCard": card,
		"message":    "Credit card added successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateCreditCard updates an existing credit card
func (h *CreditCardHandler) UpdateCreditCard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	cardID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid credit card ID", http.StatusBadRequest)
		return
	}

	// Get existing card
	existingCard, err := h.creditCardService.GetCreditCard(r.Context(), cardID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get credit card: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access
	if existingCard.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var updates model.CreditCard
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Preserve immutable fields
	updates.ID = existingCard.ID
	updates.UserID = existingCard.UserID
	updates.CircleID = existingCard.CircleID
	updates.CreatedAt = existingCard.CreatedAt

	// Validate updates
	if err := h.creditCardService.ValidateCreditCard(r.Context(), &updates); err != nil {
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Update credit card
	if err := h.creditCardService.UpdateCreditCard(r.Context(), &updates); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update credit card: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"creditCard": updates,
		"message":    "Credit card updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteCreditCard deletes a credit card
func (h *CreditCardHandler) DeleteCreditCard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	cardID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid credit card ID", http.StatusBadRequest)
		return
	}

	// Get card to check ownership
	card, err := h.creditCardService.GetCreditCard(r.Context(), cardID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get credit card: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access
	if card.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete credit card
	if err := h.creditCardService.DeleteCreditCard(r.Context(), cardID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete credit card: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Credit card deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateCard validates a credit card (Luhn check, etc.)
func (h *CreditCardHandler) ValidateCard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	cardID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid credit card ID", http.StatusBadRequest)
		return
	}

	// Get card to check ownership
	card, err := h.creditCardService.GetCreditCard(r.Context(), cardID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get credit card: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access
	if card.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Validate credit card
	if err := h.creditCardService.ValidateCreditCard(r.Context(), card); err != nil {
		response := map[string]interface{}{
			"success": false,
			"valid":   false,
			"error":   err.Error(),
			"message": "Credit card validation failed",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"valid":   true,
		"message": "Credit card is valid",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUpcomingPayments retrieves upcoming credit card payments
func (h *CreditCardHandler) GetUpcomingPayments(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	daysStr := query.Get("days")
	days := 30 // Default to next 30 days

	if daysStr != "" {
		if val, err := strconv.Atoi(daysStr); err == nil && val > 0 {
			days = val
		}
	}

	cards, err := h.creditCardService.GetUpcomingPayments(r.Context(), user.ID, days)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get upcoming payments: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate payment details
	var upcomingPayments []map[string]interface{}
	var totalDue float64
	var totalMinimumPayment float64

	for _, card := range cards {
		// Calculate days until due
		daysUntilDue := int(card.PaymentDueDate.Sub(time.Now()).Hours() / 24)
		if daysUntilDue < 0 {
			daysUntilDue = 0
		}

		// Calculate minimum payment
		minimumPayment := h.creditCardService.CalculateMinimumPayment(card.CurrentBalance, 0) // Assuming 0 interest for now

		upcomingPayment := map[string]interface{}{
			"card":              card,
			"daysUntilDue":      daysUntilDue,
			"minimumPayment":    minimumPayment,
			"isPaymentDue":      daysUntilDue <= 7, // Payment due within 7 days
			"paymentDueDate":    card.PaymentDueDate,
			"closingDate":       card.ClosingDate,
		}

		upcomingPayments = append(upcomingPayments, upcomingPayment)
		totalDue += card.CurrentBalance
		totalMinimumPayment += minimumPayment
	}

	response := map[string]interface{}{
		"success":              true,
		"upcomingPayments":     upcomingPayments,
		"totalDue":             totalDue,
		"totalMinimumPayment":  totalMinimumPayment,
		"count":                len(upcomingPayments),
		"lookaheadDays":        days,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDefaultCard retrieves the user's default credit card
func (h *CreditCardHandler) GetDefaultCard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	card, err := h.creditCardService.GetDefaultCreditCard(r.Context(), user.ID)
	if err != nil {
		// It's okay if no default card exists
		response := map[string]interface{}{
			"success": true,
			"hasDefaultCard": false,
			"message": "No default credit card found",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success":        true,
		"hasDefaultCard": true,
		"creditCard":     card,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}