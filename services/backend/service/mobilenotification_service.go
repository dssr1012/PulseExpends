package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/google/uuid"
)

// MobileNotificationServiceImpl implements MobileNotificationService
type MobileNotificationServiceImpl struct {
	notificationRepo repository.MobileNotificationRepository
	whitelistService NotificationAppWhitelistService
}

// NewMobileNotificationService creates a new MobileNotificationService
func NewMobileNotificationService(
	notificationRepo repository.MobileNotificationRepository,
	whitelistService NotificationAppWhitelistService,
) MobileNotificationService {
	return &MobileNotificationServiceImpl{
		notificationRepo: notificationRepo,
		whitelistService: whitelistService,
	}
}

// CreateNotification creates a new mobile notification
func (s *MobileNotificationServiceImpl) CreateNotification(ctx context.Context, notification *model.MobileNotification) error {
	// Validate notification
	if err := s.validateNotification(notification); err != nil {
		return fmt.Errorf("notification validation failed: %w", err)
	}

	// Check if app is whitelisted
	if notification.AppPackage != "" {
		whitelisted, err := s.whitelistService.IsAppWhitelisted(ctx, notification.UserID, notification.AppPackage)
		if err != nil {
			return fmt.Errorf("failed to check app whitelist: %w", err)
		}
		if !whitelisted {
			notification.Status = model.NotificationStatusIgnored
			notification.IgnoredReason = "App not whitelisted"
		}
	}

	// Parse and extract data if raw notification is provided
	if notification.RawData != nil && len(notification.RawData) > 0 {
		parsedNotification, err := s.ParseNotification(notification.RawData)
		if err != nil {
			return fmt.Errorf("failed to parse notification: %w", err)
		}

		// Merge parsed data
		if parsedNotification.Amount > 0 {
			notification.Amount = parsedNotification.Amount
		}
		if parsedNotification.Currency != "" {
			notification.Currency = parsedNotification.Currency
		}
		if parsedNotification.Description != "" {
			notification.Description = parsedNotification.Description
		}
		if parsedNotification.Merchant != "" {
			notification.Merchant = parsedNotification.Merchant
		}
		if parsedNotification.Category != "" {
			notification.Category = parsedNotification.Category
		}
		if !parsedNotification.Date.IsZero() {
			notification.Date = parsedNotification.Date
		}
	}

	// Calculate confidence score
	if notification.Confidence == 0 {
		notification.Confidence = s.CalculateConfidence(notification)
	}

	// Set timestamps if not set
	now := time.Now().UTC()
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = now
	}
	notification.UpdatedAt = now

	// Generate ID if not set
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}

	// Set default status if not set
	if notification.Status == "" {
		notification.Status = model.NotificationStatusPending
	}

	return s.notificationRepo.Create(ctx, notification)
}

// GetNotification retrieves a notification by ID
func (s *MobileNotificationServiceImpl) GetNotification(ctx context.Context, id uuid.UUID) (*model.MobileNotification, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid notification ID")
	}

	return s.notificationRepo.FindByID(ctx, id)
}

// UpdateNotification updates an existing notification
func (s *MobileNotificationServiceImpl) UpdateNotification(ctx context.Context, notification *model.MobileNotification) error {
	if notification.ID == uuid.Nil {
		return fmt.Errorf("notification ID is required")
	}

	// Validate notification
	if err := s.validateNotification(notification); err != nil {
		return fmt.Errorf("notification validation failed: %w", err)
	}

	// Update timestamp
	notification.UpdatedAt = time.Now().UTC()

	return s.notificationRepo.Update(ctx, notification)
}

// DeleteNotification deletes a notification
func (s *MobileNotificationServiceImpl) DeleteNotification(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid notification ID")
	}

	return s.notificationRepo.Delete(ctx, id)
}

// GetNotificationsByUser retrieves notifications for a user
func (s *MobileNotificationServiceImpl) GetNotificationsByUser(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]model.MobileNotification, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Apply default limit if not set
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	// Validate offset
	if offset < 0 {
		offset = 0
	}

	return s.notificationRepo.FindByUser(ctx, userID, limit, offset)
}

// GetNotificationsByDevice retrieves notifications for a device
func (s *MobileNotificationServiceImpl) GetNotificationsByDevice(ctx context.Context, deviceID string, limit int, offset int) ([]model.MobileNotification, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	// Apply default limit if not set
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	// Validate offset
	if offset < 0 {
		offset = 0
	}

	return s.notificationRepo.FindByDevice(ctx, deviceID, limit, offset)
}

// GetPendingNotifications retrieves pending notifications
func (s *MobileNotificationServiceImpl) GetPendingNotifications(ctx context.Context, limit int) ([]model.MobileNotification, error) {
	// Apply default limit if not set
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	return s.notificationRepo.FindPending(ctx, limit)
}

// UpdateNotificationStatus updates a notification's status
func (s *MobileNotificationServiceImpl) UpdateNotificationStatus(ctx context.Context, id uuid.UUID, status model.NotificationStatus) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid notification ID")
	}

	// Validate status
	validStatuses := map[model.NotificationStatus]bool{
		model.NotificationStatusPending:   true,
		model.NotificationStatusProcessed: true,
		model.NotificationStatusIgnored:   true,
		model.NotificationStatusError:     true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid notification status: %s", status)
	}

	return s.notificationRepo.UpdateStatus(ctx, id, status)
}

// UpdateNotificationConfidence updates a notification's confidence score
func (s *MobileNotificationServiceImpl) UpdateNotificationConfidence(ctx context.Context, id uuid.UUID, confidence float64) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid notification ID")
	}

	// Validate confidence
	if confidence < 0 || confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}

	return s.notificationRepo.UpdateConfidence(ctx, id, confidence)
}

// MarkNotificationAsProcessed marks a notification as processed
func (s *MobileNotificationServiceImpl) MarkNotificationAsProcessed(ctx context.Context, id uuid.UUID, transactionID *uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid notification ID")
	}

	return s.notificationRepo.MarkAsProcessed(ctx, id, transactionID)
}

// GetNotificationStats retrieves notification statistics for a user
func (s *MobileNotificationServiceImpl) GetNotificationStats(ctx context.Context, userID uuid.UUID) (*NotificationStats, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	stats, err := s.notificationRepo.GetNotificationStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert repository stats to service stats
	notificationStats := &NotificationStats{
		TotalNotifications:  stats.Total,
		ProcessedCount:     stats.Processed,
		PendingCount:       stats.Pending,
		IgnoredCount:       stats.Ignored,
		ErrorCount:         stats.Error,
		AverageConfidence:  stats.AverageConfidence,
		MatchRate:          stats.MatchRate,
	}

	return notificationStats, nil
}

// GetAppNotificationStats retrieves notification statistics for an app
func (s *MobileNotificationServiceImpl) GetAppNotificationStats(ctx context.Context, appPackage string) (*AppNotificationStats, error) {
	if appPackage == "" {
		return nil, fmt.Errorf("app package is required")
	}

	stats, err := s.notificationRepo.GetAppNotificationStats(ctx, appPackage)
	if err != nil {
		return nil, err
	}

	// Convert repository stats to service stats
	appStats := &AppNotificationStats{
		AppPackage:         appPackage,
		AppName:            stats.AppName,
		TotalNotifications: stats.Total,
		ProcessedCount:    stats.Processed,
		AverageConfidence: stats.AverageConfidence,
	}

	if stats.LastNotification.Valid {
		appStats.LastNotification = &stats.LastNotification.Time
	}

	return appStats, nil
}

// ParseNotification parses raw notification data
func (s *MobileNotificationServiceImpl) ParseNotification(rawNotification map[string]interface{}) (*model.MobileNotification, error) {
	if rawNotification == nil || len(rawNotification) == 0 {
		return nil, fmt.Errorf("raw notification data is empty")
	}

	notification := &model.MobileNotification{
		RawData: rawNotification,
	}

	// Extract amount
	if amount, ok := rawNotification["amount"].(float64); ok {
		notification.Amount = amount
	} else if amountStr, ok := rawNotification["amount"].(string); ok {
		// Try to parse string amount
		var amount float64
		_, err := fmt.Sscanf(strings.Trim(amountStr, "$€£¥, "), "%f", &amount)
		if err == nil {
			notification.Amount = amount
		}
	}

	// Extract currency
	if currency, ok := rawNotification["currency"].(string); ok {
		notification.Currency = strings.ToUpper(currency)
	}

	// Extract description
	if description, ok := rawNotification["description"].(string); ok {
		notification.Description = description
	} else if title, ok := rawNotification["title"].(string); ok {
		notification.Description = title
	}

	// Extract merchant
	if merchant, ok := rawNotification["merchant"].(string); ok {
		notification.Merchant = merchant
	}

	// Extract category
	if category, ok := rawNotification["category"].(string); ok {
		notification.Category = category
	}

	// Extract date
	if dateStr, ok := rawNotification["date"].(string); ok {
		// Try common date formats
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04:05",
			"2006-01-02",
			time.RFC1123,
		}

		for _, format := range formats {
			if parsed, err := time.Parse(format, dateStr); err == nil {
				notification.Date = parsed
				break
			}
		}
	}

	// Extract app package
	if appPackage, ok := rawNotification["package"].(string); ok {
		notification.AppPackage = appPackage
	} else if appName, ok := rawNotification["app"].(string); ok {
		// Try to map common app names to packages
		appPackages := map[string]string{
			"mercadopago": "com.mercadopago",
			"uala":        "com.uala",
			"brubank":     "com.brubank",
			"naranjax":    "com.naranjax",
			"personalpay": "com.personalpay",
			// Add more mappings as needed
		}
		if pkg, exists := appPackages[strings.ToLower(appName)]; exists {
			notification.AppPackage = pkg
		}
	}

	return notification, nil
}

// CalculateConfidence calculates confidence score for a notification
func (s *MobileNotificationServiceImpl) CalculateConfidence(notification *model.MobileNotification) float64 {
	confidence := 0.0

	// Amount presence adds confidence
	if notification.Amount > 0 {
		confidence += 0.3
	}

	// Currency presence adds confidence
	if notification.Currency != "" && s.isValidCurrency(notification.Currency) {
		confidence += 0.2
	}

	// Description presence adds confidence
	if notification.Description != "" {
		confidence += 0.2
	}

	// Date presence adds confidence
	if !notification.Date.IsZero() {
		confidence += 0.1
	}

	// App package validation adds confidence
	if notification.AppPackage != "" && s.IsValidAppPackage(notification.AppPackage) {
		confidence += 0.2
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// IsValidAppPackage validates an Android app package name
func (s *MobileNotificationServiceImpl) IsValidAppPackage(appPackage string) bool {
	if appPackage == "" {
		return false
	}

	// Android package pattern: com.example.app
	pattern := `^[a-zA-Z][a-zA-Z0-9_]*(\.[a-zA-Z][a-zA-Z0-9_]*)+$`
	matched, _ := regexp.MatchString(pattern, appPackage)
	return matched
}

// ExtractTransactionData extracts transaction data from notification
func (s *MobileNotificationServiceImpl) ExtractTransactionData(notification *model.MobileNotification) (*ExtractedTransactionData, error) {
	if notification == nil {
		return nil, fmt.Errorf("notification cannot be nil")
	}

	data := &ExtractedTransactionData{
		Amount:      notification.Amount,
		Currency:    notification.Currency,
		Description: notification.Description,
		Merchant:    notification.Merchant,
		Category:    notification.Category,
		Date:        notification.Date,
		Confidence:  notification.Confidence,
	}

	// Validate extracted data
	if data.Amount <= 0 {
		return nil, fmt.Errorf("invalid amount: %f", data.Amount)
	}

	if data.Currency == "" || !s.isValidCurrency(data.Currency) {
		return nil, fmt.Errorf("invalid currency: %s", data.Currency)
	}

	if data.Description == "" {
		return nil, fmt.Errorf("description is required")
	}

	if data.Date.IsZero() {
		data.Date = time.Now().UTC()
	}

	return data, nil
}

// MatchNotificationToTransaction matches a notification to an existing transaction
func (s *MobileNotificationServiceImpl) MatchNotificationToTransaction(ctx context.Context, notification *model.MobileNotification) (*model.Transaction, float64, error) {
	// This would typically involve more complex matching logic
	// For now, return nil as this requires integration with transaction service
	return nil, 0.0, nil
}

// validateNotification validates a mobile notification
func (s *MobileNotificationServiceImpl) validateNotification(notification *model.MobileNotification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	// Validate user ID
	if notification.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	// Validate device ID
	if notification.DeviceID == "" {
		return fmt.Errorf("device ID is required")
	}
	if len(notification.DeviceID) > 100 {
		return fmt.Errorf("device ID cannot exceed 100 characters")
	}

	// Validate app package if provided
	if notification.AppPackage != "" && !s.IsValidAppPackage(notification.AppPackage) {
		return fmt.Errorf("invalid app package format: %s", notification.AppPackage)
	}

	// Validate amount if provided
	if notification.Amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	// Validate currency if provided
	if notification.Currency != "" && !s.isValidCurrency(notification.Currency) {
		return fmt.Errorf("invalid currency: %s", notification.Currency)
	}

	// Validate confidence
	if notification.Confidence < 0 || notification.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}

	// Validate status
	if notification.Status != "" {
		validStatuses := map[model.NotificationStatus]bool{
			model.NotificationStatusPending:   true,
			model.NotificationStatusProcessed: true,
			model.NotificationStatusIgnored:   true,
			model.NotificationStatusError:     true,
		}
		if !validStatuses[notification.Status] {
			return fmt.Errorf("invalid notification status: %s", notification.Status)
		}
	}

	return nil
}

// isValidCurrency validates a currency code
func (s *MobileNotificationServiceImpl) isValidCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}

	// Check if all characters are uppercase letters
	for _, ch := range currency {
		if ch < 'A' || ch > 'Z' {
			return false
		}
	}

	return true
}