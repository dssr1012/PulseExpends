package repository

import (
	"context"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/google/uuid"
)

// TransactionRepository defines the interface for transaction data access
type TransactionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, transaction *model.Transaction) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	FindByFamily(ctx context.Context, familyID uuid.UUID, filters TransactionFilters) ([]model.Transaction, error)
	FindByUser(ctx context.Context, userID uuid.UUID, filters TransactionFilters) ([]model.Transaction, error)
	Update(ctx context.Context, transaction *model.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Query operations
	FindByDateRange(ctx context.Context, familyID uuid.UUID, start, end time.Time) ([]model.Transaction, error)
	FindByCategory(ctx context.Context, familyID uuid.UUID, category string, start, end time.Time) ([]model.Transaction, error)
	FindByPaymentMethod(ctx context.Context, familyID uuid.UUID, method model.PaymentMethod, start, end time.Time) ([]model.Transaction, error)
	FindUncategorized(ctx context.Context, familyID uuid.UUID, limit int) ([]model.Transaction, error)
	FindUnverified(ctx context.Context, familyID uuid.UUID, limit int) ([]model.Transaction, error)

	// Aggregation operations
	GetMonthlySummary(ctx context.Context, familyID uuid.UUID, year int, month time.Month) (*model.ReportSummary, error)
	GetCategorySummary(ctx context.Context, familyID uuid.UUID, start, end time.Time) ([]model.CategorySummary, error)
	GetSpendingTrends(ctx context.Context, familyID uuid.UUID, period string, months int) ([]model.TrendData, error)
	GetBudgetStatus(ctx context.Context, familyID uuid.UUID, budgetID uuid.UUID) (*model.BudgetStatus, error)

	// Statement matching
	FindSimilarTransactions(ctx context.Context, statementTx model.StatementTransaction) ([]model.Transaction, error)
	MatchStatementTransaction(ctx context.Context, statementTxID, transactionID uuid.UUID, confidence float64) error

	// Bulk operations
	CreateBatch(ctx context.Context, transactions []model.Transaction) error
	UpdateBatch(ctx context.Context, transactions []model.Transaction) error
}

// FamilyRepository defines the interface for family data access
type FamilyRepository interface {
	Create(ctx context.Context, family *model.Family) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Family, error)
	FindByUser(ctx context.Context, userID uuid.UUID) ([]model.Family, error)
	Update(ctx context.Context, family *model.Family) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Member management
	AddMember(ctx context.Context, familyID, userID uuid.UUID, role model.MemberRole) error
	RemoveMember(ctx context.Context, familyID, userID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, familyID, userID uuid.UUID, role model.MemberRole) error
	GetMembers(ctx context.Context, familyID uuid.UUID) ([]model.FamilyMember, error)
	IsMember(ctx context.Context, familyID, userID uuid.UUID) (bool, error)
	GetMemberRole(ctx context.Context, familyID, userID uuid.UUID) (model.MemberRole, error)

	// Settings
	UpdateSettings(ctx context.Context, familyID uuid.UUID, settings model.FamilySettings) error
	GetSettings(ctx context.Context, familyID uuid.UUID) (*model.FamilySettings, error)
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Authentication
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	VerifyUser(ctx context.Context, userID uuid.UUID) error

	// Preferences
	UpdatePreferences(ctx context.Context, userID uuid.UUID, preferences map[string]interface{}) error
	GetPreferences(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
}

// CategoryRepository defines the interface for category data access
type CategoryRepository interface {
	Create(ctx context.Context, category *model.Category) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error)
	FindByFamily(ctx context.Context, familyID uuid.UUID, includeSystem bool) ([]model.Category, error)
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Hierarchy operations
	FindByParent(ctx context.Context, parentID uuid.UUID) ([]model.Category, error)
	GetCategoryTree(ctx context.Context, familyID uuid.UUID) ([]model.Category, error)

	// Budget operations
	UpdateBudget(ctx context.Context, categoryID uuid.UUID, budget float64) error
	GetCategorySpending(ctx context.Context, categoryID uuid.UUID, start, end time.Time) (float64, error)
}

// BudgetRepository defines the interface for budget data access
type BudgetRepository interface {
	Create(ctx context.Context, budget *model.Budget) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Budget, error)
	FindByFamily(ctx context.Context, familyID uuid.UUID, activeOnly bool) ([]model.Budget, error)
	Update(ctx context.Context, budget *model.Budget) error
	Delete(ctx context.Context, id uuid.UUID) error
	Archive(ctx context.Context, id uuid.UUID) error

	// Budget tracking
	GetCurrentSpending(ctx context.Context, budgetID uuid.UUID) (float64, error)
	GetBudgetStatus(ctx context.Context, budgetID uuid.UUID) (*model.BudgetStatus, error)
	GetBudgetAlerts(ctx context.Context, familyID uuid.UUID, threshold float64) ([]model.BudgetAlert, error)

	// Period operations
	GetActiveBudgets(ctx context.Context, familyID uuid.UUID, date time.Time) ([]model.Budget, error)
	CreateNextPeriod(ctx context.Context, budgetID uuid.UUID) (*model.Budget, error)
}

// StatementRepository defines the interface for credit card statement data access
type StatementRepository interface {
	Create(ctx context.Context, statement *model.CreditCardStatement) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.CreditCardStatement, error)
	FindByFamily(ctx context.Context, familyID uuid.UUID, filters StatementFilters) ([]model.CreditCardStatement, error)
	Update(ctx context.Context, statement *model.CreditCardStatement) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Processing operations
	UpdateStatus(ctx context.Context, statementID uuid.UUID, status model.StatementStatus, errorMsg string) error
	AddTransactions(ctx context.Context, statementID uuid.UUID, transactions []model.StatementTransaction) error
	GetUnprocessed(ctx context.Context, limit int) ([]model.CreditCardStatement, error)
	GetRecentByCard(ctx context.Context, cardLastFour string, limit int) ([]model.CreditCardStatement, error)

	// Matching operations
	FindUnmatchedTransactions(ctx context.Context, statementID uuid.UUID) ([]model.StatementTransaction, error)
	UpdateTransactionMatch(ctx context.Context, statementTxID, transactionID uuid.UUID, confidence float64) error
}

// ReportRepository defines the interface for report data access
type ReportRepository interface {
	Create(ctx context.Context, report *model.Report) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Report, error)
	FindByFamily(ctx context.Context, familyID uuid.UUID, filters ReportFilters) ([]model.Report, error)
	Update(ctx context.Context, report *model.Report) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Generation operations
	GenerateMonthlyReport(ctx context.Context, familyID uuid.UUID, year int, month time.Month) (*model.Report, error)
	GenerateCategoryReport(ctx context.Context, familyID uuid.UUID, start, end time.Time) (*model.Report, error)
	GenerateTrendReport(ctx context.Context, familyID uuid.UUID, period string, months int) (*model.Report, error)
	GenerateBudgetReport(ctx context.Context, familyID uuid.UUID, budgetID uuid.UUID) (*model.Report, error)

	// Cache operations
	GetCachedReport(ctx context.Context, cacheKey string) (*model.Report, error)
	CacheReport(ctx context.Context, cacheKey string, report *model.Report, ttl time.Duration) error
	InvalidateReportCache(ctx context.Context, familyID uuid.UUID, reportType model.ReportType) error
}

// Filter structures for querying
type TransactionFilters struct {
	StartDate      *time.Time
	EndDate        *time.Time
	Category       string
	PaymentMethod  model.PaymentMethod
	MinAmount      *float64
	MaxAmount      *float64
	Status         model.TransactionStatus
	IsVerified     *bool
	IsRecurring    *bool
	Search         string
	Limit          int
	Offset         int
	SortBy         string
	SortOrder      string // "asc" or "desc"
}

type StatementFilters struct {
	StartDate    *time.Time
	EndDate      *time.Time
	CardLastFour string
	Issuer       string
	Status       model.StatementStatus
	Limit        int
	Offset       int
}

type ReportFilters struct {
	Type      model.ReportType
	StartDate *time.Time
	EndDate   *time.Time
	Format    model.ReportFormat
	Limit     int
	Offset    int
}

// Helper types for aggregation results
type CategorySummary struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Amount       float64   `json:"amount"`
	Percentage   float64   `json:"percentage"`
	Count        int       `json:"count"`
	Budget       float64   `json:"budget,omitempty"`
	Variance     float64   `json:"variance,omitempty"`
}

type BudgetSummary struct {
	BudgetID     uuid.UUID `json:"budget_id"`
	BudgetName   string    `json:"budget_name"`
	Planned      float64   `json:"planned"`
	Actual       float64   `json:"actual"`
	Remaining    float64   `json:"remaining"`
	Percentage   float64   `json:"percentage"`
	Status       string    `json:"status"` // "on_track", "warning", "exceeded"
}

type TrendData struct {
	Period     string  `json:"period"`
	Income     float64 `json:"income"`
	Expenses   float64 `json:"expenses"`
	Savings    float64 `json:"savings"`
	Categories map[string]float64 `json:"categories,omitempty"`
}

type BudgetStatus struct {
	BudgetID       uuid.UUID `json:"budget_id"`
	BudgetName     string    `json:"budget_name"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	PlannedAmount  float64   `json:"planned_amount"`
	ActualAmount   float64   `json:"actual_amount"`
	Remaining      float64   `json:"remaining"`
	PercentageUsed float64   `json:"percentage_used"`
	Status         string    `json:"status"`
	DaysRemaining  int       `json:"days_remaining"`
	DailyAverage   float64   `json:"daily_average"`
	ProjectedEnd   float64   `json:"projected_end"`
}

type BudgetAlert struct {
	BudgetID   uuid.UUID `json:"budget_id"`
	BudgetName string    `json:"budget_name"`
	Type       string    `json:"type"` // "warning", "exceeded"
	Message    string    `json:"message"`
	Threshold  float64   `json:"threshold"`
	Current    float64   `json:"current"`
	CreatedAt  time.Time `json:"created_at"`
}

// ============================================================================
// NEW REPOSITORY INTERFACES FOR REQUIREMENTS
// ============================================================================

// CreditCardRepository defines the interface for secure credit card data access
type CreditCardRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, card *model.CreditCard) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.CreditCard, error)
	FindByUser(ctx context.Context, userID uuid.UUID) ([]model.CreditCard, error)
	FindByCircle(ctx context.Context, circleID uuid.UUID) ([]model.CreditCard, error)
	Update(ctx context.Context, card *model.CreditCard) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Query operations
	FindByBank(ctx context.Context, bankName string) ([]model.CreditCard, error)
	FindByLastFour(ctx context.Context, lastFour string) ([]model.CreditCard, error)
	FindActiveCards(ctx context.Context, userID uuid.UUID) ([]model.CreditCard, error)
	FindDefaultCard(ctx context.Context, userID uuid.UUID) (*model.CreditCard, error)

	// Billing operations
	UpdateBalance(ctx context.Context, cardID uuid.UUID, newBalance float64) error
	GetUpcomingPayments(ctx context.Context, userID uuid.UUID, days int) ([]model.CreditCard, error)
	GetStatementPeriod(ctx context.Context, cardID uuid.UUID) (*time.Time, *time.Time, error) // closing date, due date
}

// ExchangeRateRepository defines the interface for currency exchange rate data access
type ExchangeRateRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, rate *model.ExchangeRate) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.ExchangeRate, error)
	Update(ctx context.Context, rate *model.ExchangeRate) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations
	FindLatestRate(ctx context.Context, baseCurrency, targetCurrency string) (*model.ExchangeRate, error)
	FindRatesByDate(ctx context.Context, date time.Time) ([]model.ExchangeRate, error)
	FindRatesByCurrency(ctx context.Context, baseCurrency string) ([]model.ExchangeRate, error)
	FindHistoricalRates(ctx context.Context, baseCurrency, targetCurrency string, startDate, endDate time.Time) ([]model.ExchangeRate, error)

	// Conversion operations
	ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string, date time.Time) (float64, error)
	GetSupportedCurrencies(ctx context.Context) ([]string, error)
	UpdateRatesFromAPI(ctx context.Context, rates map[string]float64, date time.Time, source string) error
}

// MobileNotificationRepository defines the interface for mobile notification data access
type MobileNotificationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, notification *model.MobileNotification) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.MobileNotification, error)
	FindByUser(ctx context.Context, userID uuid.UUID, filters NotificationFilters) ([]model.MobileNotification, error)
	Update(ctx context.Context, notification *model.MobileNotification) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Processing operations
	FindPending(ctx context.Context, limit int) ([]model.MobileNotification, error)
	UpdateStatus(ctx context.Context, notificationID uuid.UUID, status string, transactionID *uuid.UUID, errorMsg string) error
	MarkAsProcessed(ctx context.Context, notificationID uuid.UUID, transactionID uuid.UUID) error
	MarkAsIgnored(ctx context.Context, notificationID uuid.UUID) error

	// Analytics operations
	GetProcessingStats(ctx context.Context, userID uuid.UUID, days int) (*NotificationStats, error)
	GetAppStats(ctx context.Context, userID uuid.UUID) ([]AppNotificationStats, error)
	GetConfidenceDistribution(ctx context.Context, userID uuid.UUID) (map[string]int, error)
}

// AnomalyDetectionRepository defines the interface for anomaly detection data access
type AnomalyDetectionRepository interface {
	// Rule management
	CreateRule(ctx context.Context, rule *model.AnomalyDetectionRule) error
	FindRuleByID(ctx context.Context, id uuid.UUID) (*model.AnomalyDetectionRule, error)
	FindRulesByCircle(ctx context.Context, circleID uuid.UUID, activeOnly bool) ([]model.AnomalyDetectionRule, error)
	UpdateRule(ctx context.Context, rule *model.AnomalyDetectionRule) error
	DeleteRule(ctx context.Context, id uuid.UUID) error
	ToggleRule(ctx context.Context, id uuid.UUID, active bool) error

	// Anomaly detection
	CreateAnomaly(ctx context.Context, anomaly *model.DetectedAnomaly) error
	FindAnomalyByID(ctx context.Context, id uuid.UUID) (*model.DetectedAnomaly, error)
	FindAnomaliesByCircle(ctx context.Context, circleID uuid.UUID, filters AnomalyFilters) ([]model.DetectedAnomaly, error)
	FindAnomaliesByTransaction(ctx context.Context, transactionID uuid.UUID) ([]model.DetectedAnomaly, error)
	UpdateAnomaly(ctx context.Context, anomaly *model.DetectedAnomaly) error
	ResolveAnomaly(ctx context.Context, anomalyID uuid.UUID, resolvedBy uuid.UUID, note string) error
	IgnoreAnomaly(ctx context.Context, anomalyID uuid.UUID, ignoredBy uuid.UUID) error

	// Detection operations
	CheckForDuplicates(ctx context.Context, transaction *model.Transaction, windowDays int) ([]model.DetectedAnomaly, error)
	CheckForAmountMismatch(ctx context.Context, transaction *model.Transaction, percentageThreshold float64) (*model.DetectedAnomaly, error)
	CheckForOrphanTransactions(ctx context.Context, circleID uuid.UUID, days int) ([]model.DetectedAnomaly, error)
	RunAnomalyDetection(ctx context.Context, circleID uuid.UUID, ruleID *uuid.UUID) ([]model.DetectedAnomaly, error)

	// Analytics
	GetAnomalyStats(ctx context.Context, circleID uuid.UUID, startDate, endDate time.Time) (*AnomalyStats, error)
	GetRuleEffectiveness(ctx context.Context, circleID uuid.UUID) ([]RuleEffectiveness, error)
}

// NotificationAppWhitelistRepository defines the interface for notification app whitelist data access
type NotificationAppWhitelistRepository interface {
	// Basic CRUD operations
	AddApp(ctx context.Context, whitelist *model.NotificationAppWhitelist) error
	RemoveApp(ctx context.Context, id uuid.UUID) error
	UpdateApp(ctx context.Context, whitelist *model.NotificationAppWhitelist) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.NotificationAppWhitelist, error)

	// Query operations
	FindByUser(ctx context.Context, userID uuid.UUID) ([]model.NotificationAppWhitelist, error)
	FindByAppPackage(ctx context.Context, userID uuid.UUID, appPackage string) (*model.NotificationAppWhitelist, error)
	IsAppWhitelisted(ctx context.Context, userID uuid.UUID, appPackage string) (bool, error)
	GetEnabledApps(ctx context.Context, userID uuid.UUID) ([]string, error)

	// Bulk operations
	AddApps(ctx context.Context, userID uuid.UUID, apps []string) error
	RemoveApps(ctx context.Context, userID uuid.UUID, apps []string) error
	ToggleApps(ctx context.Context, userID uuid.UUID, apps []string, enabled bool) error
}

// ============================================================================
// FILTER STRUCTURES FOR NEW MODELS
// ============================================================================

type CreditCardFilters struct {
	BankName    string
	CardType    string
	IsActive    *bool
	IsDefault   *bool
	Limit       int
	Offset      int
}

type ExchangeRateFilters struct {
	BaseCurrency   string
	TargetCurrency string
	StartDate      *time.Time
	EndDate        *time.Time
	Source         string
	Limit          int
	Offset         int
}

type NotificationFilters struct {
	AppPackage   string
	Status       string
	StartDate    *time.Time
	EndDate      *time.Time
	MinConfidence *float64
	MaxConfidence *float64
	Limit        int
	Offset       int
	SortBy       string
	SortOrder    string // "asc" or "desc"
}

type AnomalyFilters struct {
	Type         string
	Severity     string
	Status       string
	StartDate    *time.Time
	EndDate      *time.Time
	RuleID       *uuid.UUID
	Limit        int
	Offset       int
	SortBy       string
	SortOrder    string // "asc" or "desc"
}

// ============================================================================
// HELPER TYPES FOR NEW MODELS
// ============================================================================

type NotificationStats struct {
	Total          int     `json:"total"`
	Processed      int     `json:"processed"`
	Pending        int     `json:"pending"`
	Ignored        int     `json:"ignored"`
	Error          int     `json:"error"`
	AvgConfidence  float64 `json:"avg_confidence"`
	SuccessRate    float64 `json:"success_rate"`
}

type AppNotificationStats struct {
	AppPackage     string  `json:"app_package"`
	AppName        string  `json:"app_name"`
	Total          int     `json:"total"`
	Processed      int     `json:"processed"`
	AvgConfidence  float64 `json:"avg_confidence"`
	LastProcessed  *time.Time `json:"last_processed"`
}

type AnomalyStats struct {
	Total          int            `json:"total"`
	ByType         map[string]int `json:"by_type"`
	BySeverity     map[string]int `json:"by_severity"`
	ByStatus       map[string]int `json:"by_status"`
	ResolvedRate   float64        `json:"resolved_rate"`
	AvgResolutionTime time.Duration `json:"avg_resolution_time"`
}

type RuleEffectiveness struct {
	RuleID        uuid.UUID `json:"rule_id"`
	RuleName      string    `json:"rule_name"`
	RuleType      string    `json:"rule_type"`
	TotalDetected int       `json:"total_detected"`
	TruePositives int       `json:"true_positives"`
	FalsePositives int      `json:"false_positives"`
	Effectiveness float64   `json:"effectiveness"`
	LastTriggered *time.Time `json:"last_triggered"`
}

// ============================================================================
// UPDATED TRANSACTION REPOSITORY INTERFACE
// ============================================================================

// Add new methods to TransactionRepository for requirements
type TransactionRepository interface {
	// ... existing methods ...

	// New methods for requirements
	FindPrivateExpenses(ctx context.Context, circleID uuid.UUID, viewerID uuid.UUID, filters TransactionFilters) ([]model.Transaction, error)
	FindAnomalies(ctx context.Context, circleID uuid.UUID, filters TransactionFilters) ([]model.Transaction, error)
	FindBySource(ctx context.Context, circleID uuid.UUID, source model.TransactionSource, filters TransactionFilters) ([]model.Transaction, error)
	
	// Currency conversion
	ConvertTransactionAmount(ctx context.Context, transaction *model.Transaction, targetCurrency string) (float64, error)
	GetCurrencySummary(ctx context.Context, circleID uuid.UUID, startDate, endDate time.Time) (map[string]CurrencySummary, error)
	
	// Mobile notification integration
	FindByNotificationID(ctx context.Context, notificationID string) (*model.Transaction, error)
	BulkCreateFromNotifications(ctx context.Context, notifications []model.MobileNotification) ([]model.Transaction, error)
	
	// Anomaly detection helpers
	FindSimilarAmounts(ctx context.Context, circleID uuid.UUID, amount float64, currency string, windowDays int, threshold float64) ([]model.Transaction, error)
	FindRecentByCategory(ctx context.Context, circleID uuid.UUID, category string, days int) ([]model.Transaction, error)
	FindOrphanTransactions(ctx context.Context, circleID uuid.UUID, days int) ([]model.Transaction, error)
}

type CurrencySummary struct {
	Currency        string  `json:"currency"`
	TotalIncome     float64 `json:"total_income"`
	TotalExpenses   float64 `json:"total_expenses"`
	NetBalance      float64 `json:"net_balance"`
	TransactionCount int    `json:"transaction_count"`
	ExchangeRate    float64 `json:"exchange_rate,omitempty"`
	ConvertedAmount float64 `json:"converted_amount,omitempty"`
}

// Update TransactionFilters to include new fields
type TransactionFilters struct {
	StartDate      *time.Time
	EndDate        *time.Time
	Category       string
	PaymentMethod  model.PaymentMethod
	MinAmount      *float64
	MaxAmount      *float64
	Status         model.TransactionStatus
	IsVerified     *bool
	IsRecurring    *bool
	Search         string
	Limit          int
	Offset         int
	SortBy         string
	SortOrder      string // "asc" or "desc"
	
	// New fields for requirements
	IsPrivate      *bool
	IsAnomaly      *bool
	AnomalyType    string
	AnomalySeverity string
	Source         model.TransactionSource
	Currency       string
	BaseCurrency   string
	IncludeHidden  bool // Include private expenses that are still hidden
}
