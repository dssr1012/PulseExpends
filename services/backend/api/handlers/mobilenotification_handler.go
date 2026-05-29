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

type MobileNotificationHandler struct {
	mobileNotificationService service.MobileNotificationService
}

func NewMobileNotificationHandler(mobileNotificationService service.MobileNotificationService) *MobileNotificationHandler {
	return &MobileNotificationHandler{
		mobileNotificationService: mobileNotificationService,
	}
}

// GetNotifications retrieves mobile notifications for the authenticated user
func (h *MobileNotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")
	status := query.Get("status")
	appPackage := query.Get("appPackage")
	startDateStr := query.Get("startDate")
	endDateStr := query.Get("endDate")

	// Parse pagination
	limit := 50
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 1000 {
			limit = val
		}
	}

	offset := 0
	if offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}

	// Parse date range
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

	notifications, err := h.mobileNotificationService.GetNotificationsByUser(r.Context(), user.ID, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get notifications: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter by status if provided
	var filteredNotifications []model.MobileNotification
	if status != "" {
		for _, notification := range notifications {
			if notification.Status == status {
				filteredNotifications = append(filteredNotifications, notification)
			}
		}
	} else {
		filteredNotifications = notifications
	}

	// Filter by app package if provided
	if appPackage != "" {
		var appFiltered []model.MobileNotification
		for _, notification := range filteredNotifications {
			if notification.AppPackage == appPackage {
				appFiltered = append(appFiltered, notification)
			}
		}
		filteredNotifications = appFiltered
	}

	// Filter by date range if provided
	if !startDate.IsZero() || !endDate.IsZero() {
		var dateFiltered []model.MobileNotification
		for _, notification := range filteredNotifications {
			notificationTime := notification.Timestamp
			if (!startDate.IsZero() && notificationTime.Before(startDate)) ||
				(!endDate.IsZero() && notificationTime.After(endDate)) {
				continue
			}
			dateFiltered = append(dateFiltered, notification)
		}
		filteredNotifications = dateFiltered
	}

	// Get statistics
	stats, err := h.mobileNotificationService.GetNotificationStats(r.Context(), user.ID)
	if err != nil {
		// Continue without stats if there's an error
		stats = &service.NotificationStats{}
	}

	response := map[string]interface{}{
		"success":       true,
		"notifications": filteredNotifications,
		"total":         len(filteredNotifications),
		"limit":         limit,
		"offset":        offset,
		"stats":         stats,
		"filters": map[string]interface{}{
			"status":     status,
			"appPackage": appPackage,
			"startDate":  startDateStr,
			"endDate":    endDateStr,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetNotification retrieves a specific notification by ID
func (h *MobileNotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	notification, err := h.mobileNotificationService.GetNotification(r.Context(), notificationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get notification: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to this notification
	if notification.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"notification": notification,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateNotification creates a new mobile notification
func (h *MobileNotificationHandler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var notification model.MobileNotification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Set user ID and timestamps
	notification.UserID = user.ID
	if notification.Timestamp.IsZero() {
		notification.Timestamp = time.Now()
	}
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	// Validate app package
	if !h.mobileNotificationService.IsValidAppPackage(notification.AppPackage) {
		http.Error(w, fmt.Sprintf("Invalid app package: %s", notification.AppPackage), http.StatusBadRequest)
		return
	}

	// Calculate confidence score
	notification.Confidence = h.mobileNotificationService.CalculateConfidence(&notification)

	// Create notification
	if err := h.mobileNotificationService.CreateNotification(r.Context(), &notification); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create notification: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"notification": notification,
		"message":      "Notification created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ProcessNotification processes a notification (extracts transaction data)
func (h *MobileNotificationHandler) ProcessNotification(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	// Get notification
	notification, err := h.mobileNotificationService.GetNotification(r.Context(), notificationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get notification: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access
	if notification.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Check if already processed
	if notification.Status == "processed" {
		response := map[string]interface{}{
			"success": true,
			"message": "Notification already processed",
			"notification": notification,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse notification text to extract transaction data
	extractedData, err := h.mobileNotificationService.ExtractTransactionData(notification)
	if err != nil {
		// Update notification status to error
		h.mobileNotificationService.UpdateNotificationStatus(r.Context(), notificationID, "error")
		http.Error(w, fmt.Sprintf("Failed to parse notification: %v", err), http.StatusBadRequest)
		return
	}

	// Try to match with existing transaction
	matchedTransaction, confidence, err := h.mobileNotificationService.MatchNotificationToTransaction(r.Context(), notification)
	if err != nil {
		// Continue processing even if matching fails
		matchedTransaction = nil
		confidence = 0.0
	}

	// Update notification with extracted data and status
	notification.ParsedAmount = extractedData.Amount
	notification.ParsedCurrency = extractedData.Currency
	notification.ParsedMerchant = extractedData.Merchant
	notification.ParsedCategory = extractedData.Category
	notification.Confidence = extractedData.Confidence

	if matchedTransaction != nil {
		notification.TransactionID = &matchedTransaction.ID
		notification.Status = "processed"
	} else {
		notification.Status = "pending"
	}

	// Update notification
	if err := h.mobileNotificationService.UpdateNotification(r.Context(), notification); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update notification: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"message":      "Notification processed successfully",
		"notification": notification,
		"extractedData": extractedData,
		"matched":      matchedTransaction != nil,
		"confidence":   confidence,
	}

	if matchedTransaction != nil {
		response["transaction"] = matchedTransaction
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// IgnoreNotification marks a notification as ignored
func (h *MobileNotificationHandler) IgnoreNotification(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	// Get notification
	notification, err := h.mobileNotificationService.GetNotification(r.Context(), notificationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get notification: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access
	if notification.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Mark as ignored
	if err := h.mobileNotificationService.UpdateNotificationStatus(r.Context(), notificationID, "ignored"); err != nil {
		http.Error(w, fmt.Sprintf("Failed to ignore notification: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Notification ignored successfully",
		"notificationId": notificationID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetStats retrieves notification statistics for the user
func (h *MobileNotificationHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	daysStr := query.Get("days")
	days := 30 // Default to last 30 days

	if daysStr != "" {
		if val, err := strconv.Atoi(daysStr); err == nil && val > 0 && val <= 365 {
			days = val
		}
	}

	stats, err := h.mobileNotificationService.GetNotificationStats(r.Context(), user.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get notification stats: %v", err), http.StatusInternalServerError)
		return
	}

	// Get app-specific stats
	appStats, err := h.mobileNotificationService.GetAppNotificationStats(r.Context(), user.ID)
	if err != nil {
		// Continue without app stats if there's an error
		appStats = []service.AppNotificationStats{}
	}

	response := map[string]interface{}{
		"success": true,
		"stats":   stats,
		"appStats": appStats,
		"periodDays": days,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ParseNotification parses raw notification data
func (h *MobileNotificationHandler) ParseNotification(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		RawData map[string]interface{} `json:"rawData"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse notification from raw data
	notification, err := h.mobileNotificationService.ParseNotification(request.RawData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse notification: %v", err), http.StatusBadRequest)
		return
	}

	// Set user ID for the parsed notification
	notification.UserID = user.ID

	// Calculate confidence
	notification.Confidence = h.mobileNotificationService.CalculateConfidence(notification)

	// Extract transaction data
	extractedData, extractErr := h.mobileNotificationService.ExtractTransactionData(notification)
	if extractErr != nil {
		// Continue even if extraction fails
		extractedData = &service.ExtractedTransactionData{}
	}

	response := map[string]interface{}{
		"success":        true,
		"parsedNotification": notification,
		"extractedData":  extractedData,
		"confidence":     notification.Confidence,
		"isValidApp":     h.mobileNotificationService.IsValidAppPackage(notification.AppPackage),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPendingNotifications retrieves pending notifications for processing
func (h *MobileNotificationHandler) GetPendingNotifications(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	limitStr := query.Get("limit")
	limit := 100

	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 1000 {
			limit = val
		}
	}

	notifications, err := h.mobileNotificationService.GetPendingNotifications(r.Context(), limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get pending notifications: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter to only user's notifications
	var userNotifications []model.MobileNotification
	for _, notification := range notifications {
		if notification.UserID == user.ID {
			userNotifications = append(userNotifications, notification)
		}
	}

	response := map[string]interface{}{
		"success":       true,
		"notifications": userNotifications,
		"total":         len(userNotifications),
		"limit":         limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}