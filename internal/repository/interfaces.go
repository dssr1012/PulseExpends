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