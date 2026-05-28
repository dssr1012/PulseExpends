package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Timezone     string    `json:"timezone"`
	Currency     string    `json:"currency"` // ISO 4217 currency code
	Language     string    `json:"language"` // BCP 47 language tag
	IsActive     bool      `json:"is_active"`
	IsVerified   bool      `json:"is_verified"`
	LastLoginAt  time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Family represents a family circle/group
type Family struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Currency    string    `json:"currency"` // Primary currency for the family
	CreatedBy   uuid.UUID `json:"created_by"`
	IsActive    bool      `json:"is_active"`
	Settings    FamilySettings `json:"settings"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FamilySettings contains family-specific configuration
type FamilySettings struct {
	MonthlyBudget    float64 `json:"monthly_budget,omitempty"`
	BudgetWarningThreshold float64 `json:"budget_warning_threshold,omitempty"` // Percentage
	DefaultCategory  string  `json:"default_category,omitempty"`
	NotificationSettings NotificationSettings `json:"notification_settings"`
}

// NotificationSettings for family notifications
type NotificationSettings struct {
	EmailNotifications bool `json:"email_notifications"`
	PushNotifications  bool `json:"push_notifications"`
	BudgetAlerts       bool `json:"budget_alerts"`
	LargeExpenseAlerts bool `json:"large_expense_alerts"`
	DailyDigest        bool `json:"daily_digest"`
	WeeklyReport       bool `json:"weekly_report"`
}

// FamilyMember represents a user's membership in a family
type FamilyMember struct {
	FamilyID   uuid.UUID `json:"family_id"`
	UserID     uuid.UUID `json:"user_id"`
	Role       MemberRole `json:"role"`
	JoinedAt   time.Time `json:"joined_at"`
	IsActive   bool      `json:"is_active"`
}

// MemberRole defines user roles within a family
type MemberRole string

const (
	RoleOwner    MemberRole = "owner"
	RoleAdmin    MemberRole = "admin"
	RoleMember   MemberRole = "member"
	RoleViewer   MemberRole = "viewer"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID              uuid.UUID `json:"id"`
	FamilyID        uuid.UUID `json:"family_id"`
	UserID          uuid.UUID `json:"user_id"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	OriginalAmount  float64   `json:"original_amount,omitempty"`
	OriginalCurrency string   `json:"original_currency,omitempty"`
	ExchangeRate    float64   `json:"exchange_rate,omitempty"`
	
	// Transaction details
	Description     string    `json:"description"`
	Category        string    `json:"category"`
	Subcategory     string    `json:"subcategory,omitempty"`
	PaymentMethod   PaymentMethod `json:"payment_method"`
	PaymentDetails  PaymentDetails `json:"payment_details,omitempty"`
	
	// Date and location
	TransactionDate time.Time `json:"transaction_date"`
	RecordedDate    time.Time `json:"recorded_date"`
	Location        string    `json:"location,omitempty"`
	Merchant        string    `json:"merchant,omitempty"`
	
	// Metadata
	Tags            []string  `json:"tags,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	ReceiptURL      string    `json:"receipt_url,omitempty"`
	IsRecurring     bool      `json:"is_recurring"`
	RecurrenceRule  string    `json:"recurrence_rule,omitempty"`
	IsTaxDeductible bool      `json:"is_tax_deductible"`
	
	// Status
	Status          TransactionStatus `json:"status"`
	IsVerified      bool              `json:"is_verified"`
	VerifiedBy      uuid.UUID         `json:"verified_by,omitempty"`
	VerifiedAt      time.Time         `json:"verified_at,omitempty"`
	
	// Timestamps
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       time.Time `json:"deleted_at,omitempty"`
}

// PaymentMethod defines how a transaction was paid
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodDebitCard    PaymentMethod = "debit_card"
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodDigitalWallet PaymentMethod = "digital_wallet"
	PaymentMethodCheck        PaymentMethod = "check"
	PaymentMethodOther        PaymentMethod = "other"
)

// PaymentDetails contains payment-specific information
type PaymentDetails struct {
	CardLastFour string    `json:"card_last_four,omitempty"`
	CardType     string    `json:"card_type,omitempty"`
	BankName     string    `json:"bank_name,omitempty"`
	AccountNumber string   `json:"account_number,omitempty"`
	WalletProvider string  `json:"wallet_provider,omitempty"`
	Installments  int      `json:"installments,omitempty"`
	InstallmentNumber int  `json:"installment_number,omitempty"`
	TotalInstallments int  `json:"total_installments,omitempty"`
}

// TransactionStatus defines the state of a transaction
type TransactionStatus string

const (
	StatusPending   TransactionStatus = "pending"
	StatusCompleted TransactionStatus = "completed"
	StatusCancelled TransactionStatus = "cancelled"
	StatusRefunded  TransactionStatus = "refunded"
	StatusDisputed  TransactionStatus = "disputed"
)

// Category represents a transaction category
type Category struct {
	ID          uuid.UUID `json:"id"`
	FamilyID    uuid.UUID `json:"family_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	Color       string    `json:"color,omitempty"`
	ParentID    uuid.UUID `json:"parent_id,omitempty"`
	IsSystem    bool      `json:"is_system"`
	Budget      float64   `json:"budget,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Budget represents a spending budget
type Budget struct {
	ID          uuid.UUID `json:"id"`
	FamilyID    uuid.UUID `json:"family_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Period      BudgetPeriod `json:"period"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Categories  []uuid.UUID `json:"categories,omitempty"` // Specific categories, empty for all
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BudgetPeriod defines the budget time period
type BudgetPeriod string

const (
	BudgetPeriodDaily    BudgetPeriod = "daily"
	BudgetPeriodWeekly   BudgetPeriod = "weekly"
	BudgetPeriodMonthly  BudgetPeriod = "monthly"
	BudgetPeriodQuarterly BudgetPeriod = "quarterly"
	BudgetPeriodYearly   BudgetPeriod = "yearly"
	BudgetPeriodCustom   BudgetPeriod = "custom"
)

// CreditCardStatement represents a parsed credit card statement
type CreditCardStatement struct {
	ID              uuid.UUID `json:"id"`
	FamilyID        uuid.UUID `json:"family_id"`
	UserID          uuid.UUID `json:"user_id"`
	CardLastFour    string    `json:"card_last_four"`
	Issuer          string    `json:"issuer"`
	StatementDate   time.Time `json:"statement_date"`
	DueDate         time.Time `json:"due_date"`
	TotalAmount     float64   `json:"total_amount"`
	Currency        string    `json:"currency"`
	MinPayment      float64   `json:"min_payment"`
	
	// Parsed transactions from statement
	Transactions    []StatementTransaction `json:"transactions"`
	
	// Processing metadata
	OriginalFilename string    `json:"original_filename"`
	FileURL          string    `json:"file_url"`
	FileHash         string    `json:"file_hash"`
	ParsedAt         time.Time `json:"parsed_at"`
	ParserVersion    string    `json:"parser_version"`
	ConfidenceScore  float64   `json:"confidence_score"` // 0-1 confidence in parsing accuracy
	
	Status          StatementStatus `json:"status"`
	Error           string          `json:"error,omitempty"`
	
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// StatementTransaction represents a transaction from a credit card statement
type StatementTransaction struct {
	ID                  uuid.UUID `json:"id"`
	StatementID         uuid.UUID `json:"statement_id"`
	Date                time.Time `json:"date"`
	Description         string    `json:"description"`
	Amount              float64   `json:"amount"`
	Currency            string    `json:"currency"`
	Merchant            string    `json:"merchant,omitempty"`
	Category            string    `json:"category,omitempty"`
	IsInstallment       bool      `json:"is_installment"`
	InstallmentNumber   int       `json:"installment_number,omitempty"`
	TotalInstallments   int       `json:"total_installments,omitempty"`
	MatchedTransactionID uuid.UUID `json:"matched_transaction_id,omitempty"`
	MatchConfidence     float64   `json:"match_confidence,omitempty"`
	Notes               string    `json:"notes,omitempty"`
}

// StatementStatus defines the processing status of a statement
type StatementStatus string

const (
	StatementStatusPending    StatementStatus = "pending"
	StatementStatusProcessing StatementStatus = "processing"
	StatementStatusParsed     StatementStatus = "parsed"
	StatementStatusMatched    StatementStatus = "matched"
	StatementStatusError      StatementStatus = "error"
)

// Report represents a generated financial report
type Report struct {
	ID          uuid.UUID `json:"id"`
	FamilyID    uuid.UUID `json:"family_id"`
	UserID      uuid.UUID `json:"user_id"`
	Type        ReportType `json:"type"`
	Period      ReportPeriod `json:"period"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	
	// Report data
	Data        ReportData `json:"data"`
	Summary     ReportSummary `json:"summary"`
	
	// Metadata
	Format      ReportFormat `json:"format"`
	FileURL     string      `json:"file_url,omitempty"`
	IsGenerated bool        `json:"is_generated"`
	GeneratedAt time.Time   `json:"generated_at,omitempty"`
	
	CreatedAt   time.Time `json:"created_at"`
}

// ReportType defines the type of report
type ReportType string

const (
	ReportTypeMonthlySummary ReportType = "monthly_summary"
	ReportTypeCategoryBreakdown ReportType = "category_breakdown"
	ReportTypeSpendingTrends ReportType = "spending_trends"
	ReportTypeBudgetVsActual ReportType = "budget_vs_actual"
	ReportTypeTaxReport     ReportType = "tax_report"
)

// ReportPeriod defines the time period for the report
type ReportPeriod struct {
	Type     string    `json:"type"` // "month", "quarter", "year", "custom"
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
}

// ReportData contains the detailed report data
type ReportData struct {
	Transactions []Transaction `json:"transactions,omitempty"`
	Categories   []CategorySummary `json:"categories,omitempty"`
	Budgets      []BudgetSummary `json:"budgets,omitempty"`
	Trends       []TrendData     `json:"trends,omitempty"`
}

// ReportSummary contains high-level summary data
type ReportSummary struct {
	TotalIncome     float64 `json:"total_income"`
	TotalExpenses   float64 `json:"total_expenses"`
	NetBalance      float64 `json:"net_balance"`
	TopCategories   []CategorySummary `json:"top_categories"`
	BudgetStatus    []BudgetStatus    `json:"budget_status"`
	Recommendations []string          `json:"recommendations,omitempty"`
}

// ReportFormat defines the output format
type ReportFormat string

const (
	ReportFormatJSON    ReportFormat = "json"
	ReportFormatPDF     ReportFormat = "pdf"
	ReportFormatCSV     ReportFormat = "csv"
	ReportFormatExcel   ReportFormat = "excel"
)