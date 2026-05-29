package service

import (
	"context"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/google/uuid"
)

// Service interfaces define the business logic layer that sits between
// HTTP handlers and repository layer.

// ============================================================================
// TRANSACTION SERVICE
// ============================================================================

type TransactionService interface {
	// Basic CRUD operations
	CreateTransaction(ctx context.Context, transaction *model.Transaction) error
	GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *model.Transaction) error
	DeleteTransaction(ctx context.Context, id uuid.UUID) error
	SoftDeleteTransaction(ctx context.Context, id uuid.UUID) error

	// Query operations
	GetTransactionsByFamily(ctx context.Context, familyID uuid.UUID, filters repository.TransactionFilters) ([]model.Transaction, error)
	GetTransactionsByUser(ctx context.Context, userID uuid.UUID, filters repository.TransactionFilters) ([]model.Transaction, error)
	GetTransactionsByDateRange(ctx context.Context, familyID uuid.UUID, start, end time.Time) ([]model.Transaction, error)
	GetTransactionsByCategory(ctx context.Context, familyID uuid.UUID, category string, start, end time.Time) ([]model.Transaction, error)
	GetTransactionsByPaymentMethod(ctx context.Context, familyID uuid.UUID, method model.PaymentMethod, start, end time.Time) ([]model.Transaction, error)
	GetUncategorizedTransactions(ctx context.Context, familyID uuid.UUID, limit int) ([]model.Transaction, error)
	GetUnverifiedTransactions(ctx context.Context, familyID uuid.UUID, limit int) ([]model.Transaction, error)

	// Aggregation operations
	GetMonthlySummary(ctx context.Context, familyID uuid.UUID, year int, month time.Month) (*model.ReportSummary, error)
	GetCategorySummary(ctx context.Context, familyID uuid.UUID, start, end time.Time) ([]repository.CategorySummary, error)
	GetSpendingTrends(ctx context.Context, familyID uuid.UUID, period string, months int) ([]repository.TrendData, error)
	GetBudgetStatus(ctx context.Context, familyID uuid.UUID, budgetID uuid.UUID) (*repository.BudgetStatus, error)

	// Statement matching
	FindSimilarTransactions(ctx context.Context, statementTx model.StatementTransaction) ([]model.Transaction, error)
	MatchStatementTransaction(ctx context.Context, statementTxID, transactionID uuid.UUID, confidence float64) error

	// Bulk operations
	CreateTransactionsBatch(ctx context.Context, transactions []model.Transaction) error
	UpdateTransactionsBatch(ctx context.Context, transactions []model.Transaction) error

	// Business logic operations
	ValidateTransaction(ctx context.Context, transaction *model.Transaction) error
	CalculateTransactionTotals(ctx context.Context, familyID uuid.UUID, start, end time.Time) (*TransactionTotals, error)
	DetectDuplicateTransactions(ctx context.Context, transaction *model.Transaction) ([]model.Transaction, error)
}

// ============================================================================
// CREDIT CARD SERVICE
// ============================================================================

type CreditCardService interface {
	// Basic CRUD operations
	AddCreditCard(ctx context.Context, card *model.CreditCard) error
	GetCreditCard(ctx context.Context, id uuid.UUID) (*model.CreditCard, error)
	UpdateCreditCard(ctx context.Context, card *model.CreditCard) error
	DeleteCreditCard(ctx context.Context, id uuid.UUID) error
	SoftDeleteCreditCard(ctx context.Context, id uuid.UUID) error

	// Query operations
	GetUserCreditCards(ctx context.Context, userID uuid.UUID) ([]model.CreditCard, error)
	GetFamilyCreditCards(ctx context.Context, circleID uuid.UUID) ([]model.CreditCard, error)
	GetCreditCardsByBank(ctx context.Context, bankName string) ([]model.CreditCard, error)
	GetActiveCreditCards(ctx context.Context, userID uuid.UUID) ([]model.CreditCard, error)
	GetDefaultCreditCard(ctx context.Context, userID uuid.UUID) (*model.CreditCard, error)

	// Billing operations
	UpdateCardBalance(ctx context.Context, cardID uuid.UUID, newBalance float64) error
	GetUpcomingPayments(ctx context.Context, userID uuid.UUID, days int) ([]model.CreditCard, error)
	GetStatementPeriod(ctx context.Context, cardID uuid.UUID) (*time.Time, *time.Time, error) // closing date, due date

	// Business logic operations
	ValidateCreditCard(ctx context.Context, card *model.CreditCard) error
	MaskCardNumber(cardNumber string) string
	CalculateMinimumPayment(balance float64, interestRate float64) float64
	CheckPaymentDueDate(cardID uuid.UUID) (bool, time.Time, error) // isDue, dueDate, error
}

// ============================================================================
// EXCHANGE RATE SERVICE
// ============================================================================

type ExchangeRateService interface {
	// Basic CRUD operations
	AddExchangeRate(ctx context.Context, rate *model.ExchangeRate) error
	GetExchangeRate(ctx context.Context, id uuid.UUID) (*model.ExchangeRate, error)
	UpdateExchangeRate(ctx context.Context, rate *model.ExchangeRate) error
	DeleteExchangeRate(ctx context.Context, id uuid.UUID) error

	// Query operations
	GetLatestRate(ctx context.Context, baseCurrency, targetCurrency string) (*model.ExchangeRate, error)
	GetRatesByDate(ctx context.Context, date time.Time) ([]model.ExchangeRate, error)
	GetRatesByCurrency(ctx context.Context, baseCurrency string) ([]model.ExchangeRate, error)
	GetHistoricalRates(ctx context.Context, baseCurrency, targetCurrency string, startDate, endDate time.Time) ([]model.ExchangeRate, error)

	// Conversion operations
	ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string, date time.Time) (float64, error)
	GetSupportedCurrencies(ctx context.Context) ([]string, error)
	UpdateRatesFromAPI(ctx context.Context, rates map[string]float64, date time.Time, source string) error

	// Business logic operations
	ValidateCurrencyCode(currency string) bool
	GetExchangeRateWithFallback(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (*model.ExchangeRate, error)
	CalculateInverseRate(rate float64) float64
	GetCurrencyPairKey(baseCurrency, targetCurrency string) string
}

// ============================================================================
// MOBILE NOTIFICATION SERVICE
// ============================================================================

type MobileNotificationService interface {
	// Basic CRUD operations
	CreateNotification(ctx context.Context, notification *model.MobileNotification) error
	GetNotification(ctx context.Context, id uuid.UUID) (*model.MobileNotification, error)
	UpdateNotification(ctx context.Context, notification *model.MobileNotification) error
	DeleteNotification(ctx context.Context, id uuid.UUID) error

	// Query operations
	GetNotificationsByUser(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]model.MobileNotification, error)
	GetNotificationsByDevice(ctx context.Context, deviceID string, limit int, offset int) ([]model.MobileNotification, error)
	GetPendingNotifications(ctx context.Context, limit int) ([]model.MobileNotification, error)
	UpdateNotificationStatus(ctx context.Context, id uuid.UUID, status model.NotificationStatus) error
	UpdateNotificationConfidence(ctx context.Context, id uuid.UUID, confidence float64) error
	MarkNotificationAsProcessed(ctx context.Context, id uuid.UUID, transactionID *uuid.UUID) error
	GetNotificationStats(ctx context.Context, userID uuid.UUID) (*NotificationStats, error)
	GetAppNotificationStats(ctx context.Context, appPackage string) (*AppNotificationStats, error)

	// Business logic operations
	ParseNotification(rawNotification map[string]interface{}) (*model.MobileNotification, error)
	CalculateConfidence(notification *model.MobileNotification) float64
	IsValidAppPackage(appPackage string) bool
	ExtractTransactionData(notification *model.MobileNotification) (*ExtractedTransactionData, error)
	MatchNotificationToTransaction(ctx context.Context, notification *model.MobileNotification) (*model.Transaction, float64, error)
}

// ============================================================================
// ANOMALY DETECTION SERVICE
// ============================================================================

type AnomalyDetectionService interface {
	// Rule management
	CreateDetectionRule(ctx context.Context, rule *model.AnomalyDetectionRule) error
	GetDetectionRule(ctx context.Context, id uuid.UUID) (*model.AnomalyDetectionRule, error)
	GetDetectionRulesByCircle(ctx context.Context, circleID uuid.UUID) ([]model.AnomalyDetectionRule, error)
	GetActiveDetectionRules(ctx context.Context, circleID uuid.UUID) ([]model.AnomalyDetectionRule, error)
	UpdateDetectionRule(ctx context.Context, rule *model.AnomalyDetectionRule) error
	DeleteDetectionRule(ctx context.Context, id uuid.UUID) error
	ToggleRuleStatus(ctx context.Context, id uuid.UUID, enabled bool) error

	// Anomaly management
	CreateAnomaly(ctx context.Context, anomaly *model.DetectedAnomaly) error
	GetAnomaly(ctx context.Context, id uuid.UUID) (*model.DetectedAnomaly, error)
	GetAnomaliesByCircle(ctx context.Context, circleID uuid.UUID, status string, limit, offset int) ([]model.DetectedAnomaly, error)
	GetAnomaliesByTransaction(ctx context.Context, transactionID uuid.UUID) ([]model.DetectedAnomaly, error)
	GetAnomaliesByRule(ctx context.Context, ruleID uuid.UUID, status string, limit, offset int) ([]model.DetectedAnomaly, error)
	UpdateAnomalyStatus(ctx context.Context, id uuid.UUID, status string, resolvedBy uuid.UUID, resolutionNote string) error
	DeleteAnomaly(ctx context.Context, id uuid.UUID) error
	GetAnomalyStats(ctx context.Context, circleID uuid.UUID) (*repository.AnomalyStats, error)
	GetRuleEffectiveness(ctx context.Context, ruleID uuid.UUID) (*repository.RuleEffectiveness, error)

	// Detection engine
	DetectAnomalies(ctx context.Context, transaction *model.Transaction) ([]*model.DetectedAnomaly, error)
	CheckForDuplicateTransactions(ctx context.Context, transaction *model.Transaction) ([]model.Transaction, error)
	CheckForAmountMismatch(ctx context.Context, transaction *model.Transaction) (bool, float64, error) // isAnomaly, expectedAmount, error
	CheckForOrphanTransaction(ctx context.Context, transaction *model.Transaction) (bool, error)
	CheckForTimeframeAnomaly(ctx context.Context, transaction *model.Transaction) (bool, time.Duration, error) // isAnomaly, timeframe, error
	CalculateAnomalySeverity(anomalyType string, confidence float64, amount float64) string // "warning" or "critical"
}

// ============================================================================
// NOTIFICATION APP WHITELIST SERVICE
// ============================================================================

type NotificationAppWhitelistService interface {
	// Basic CRUD operations
	AddToWhitelist(ctx context.Context, app *model.NotificationAppWhitelist) error
	GetWhitelistEntry(ctx context.Context, id uuid.UUID) (*model.NotificationAppWhitelist, error)
	UpdateWhitelistEntry(ctx context.Context, app *model.NotificationAppWhitelist) error
	RemoveFromWhitelist(ctx context.Context, id uuid.UUID) error

	// Query operations
	GetUserWhitelist(ctx context.Context, userID uuid.UUID) ([]model.NotificationAppWhitelist, error)
	GetCircleWhitelist(ctx context.Context, circleID uuid.UUID) ([]model.NotificationAppWhitelist, error)
	GetWhitelistByAppPackage(ctx context.Context, appPackage string) ([]model.NotificationAppWhitelist, error)
	GetWhitelistEntryByUserAndApp(ctx context.Context, userID uuid.UUID, appPackage string) (*model.NotificationAppWhitelist, error)

	// Validation operations
	IsAppWhitelisted(ctx context.Context, userID uuid.UUID, appPackage string) (bool, error)
	AutoWhitelistApp(ctx context.Context, userID uuid.UUID, appPackage, appName string) error
	GetWhitelistStats(ctx context.Context, userID uuid.UUID) (*repository.WhitelistStats, error)

	// Business logic operations
	ValidateAppPackage(appPackage string) error
	ResolveAppName(appPackage string) (string, error)
	IsTrustedApp(appPackage string) bool
	ShouldAutoWhitelist(appPackage string) bool
}

// ============================================================================
// SUPPORTING TYPES
// ============================================================================

type TransactionTotals struct {
	TotalIncome     float64 `json:"total_income"`
	TotalExpenses   float64 `json:"total_expenses"`
	NetBalance      float64 `json:"net_balance"`
	TransactionCount int    `json:"transaction_count"`
}

type NotificationStats struct {
	TotalNotifications  int     `json:"total_notifications"`
	ProcessedCount     int     `json:"processed_count"`
	PendingCount       int     `json:"pending_count"`
	IgnoredCount       int     `json:"ignored_count"`
	ErrorCount         int     `json:"error_count"`
	AverageConfidence  float64 `json:"average_confidence"`
	MatchRate          float64 `json:"match_rate"` // percentage matched to transactions
}

type AppNotificationStats struct {
	AppPackage        string  `json:"app_package"`
	AppName           string  `json:"app_name"`
	TotalNotifications int    `json:"total_notifications"`
	ProcessedCount    int     `json:"processed_count"`
	AverageConfidence float64 `json:"average_confidence"`
	LastNotification  *time.Time `json:"last_notification,omitempty"`
}

type ExtractedTransactionData struct {
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	Merchant    string    `json:"merchant,omitempty"`
	Category    string    `json:"category,omitempty"`
	Date        time.Time `json:"date"`
	Confidence  float64   `json:"confidence"`
}