package model

import (
	"time"
	"github.com/google/uuid"
)

// Transaction represents a financial transaction in the system
// Enhanced with all requirements from functional specification
type Transaction struct {
	ID              string                 `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CircleID        string                 `json:"circle_id" gorm:"type:uuid;index"`
	UserID          string                 `json:"user_id" gorm:"type:uuid;not null;index"`
	
	// Core transaction data
	Amount          float64                `json:"amount" gorm:"type:decimal(15,2);not null"`
	Currency        string                 `json:"currency" gorm:"type:varchar(3);not null;default:'USD'"`
	Category        string                 `json:"category" gorm:"not null"`
	Description     string                 `json:"description"`
	Date            time.Time              `json:"date" gorm:"not null"`
	
	// Payment method details
	PaymentMethod   string                 `json:"payment_method"` // cash, card, transfer, etc.
	CardLastFour    string                 `json:"card_last_four,omitempty" gorm:"type:varchar(4)"`
	BankName        string                 `json:"bank_name,omitempty"`
	
	// Private expense (Gift Mode) support
	IsPrivate       bool                   `json:"is_private" gorm:"default:false"`
	HiddenUntil     *time.Time             `json:"hidden_until,omitempty"`
	PrivateDescription string              `json:"private_description,omitempty"` // Obfuscated description for other users
	
	// Source tracking (manual vs automated)
	Source          string                 `json:"source" gorm:"default:'manual'"` // manual, card_statement, mobile_notification
	SourceID        string                 `json:"source_id,omitempty"` // Reference to original source (statement ID, notification ID)
	
	// Anomaly detection flags
	IsAnomaly       bool                   `json:"is_anomaly" gorm:"default:false"`
	AnomalyType     string                 `json:"anomaly_type,omitempty"` // duplicate, amount_mismatch, orphan
	AnomalySeverity string                 `json:"anomaly_severity,omitempty"` // warning, critical
	MatchedTransactionID string            `json:"matched_transaction_id,omitempty" gorm:"type:uuid"`
	
	// Recurring transactions
	IsRecurring     bool                   `json:"is_recurring" gorm:"default:false"`
	RecurringRule   *RecurringRule         `json:"recurring_rule,omitempty" gorm:"type:jsonb"`
	
	// Metadata
	Tags            []string               `json:"tags,omitempty" gorm:"type:text[]"`
	Location        string                 `json:"location,omitempty"`
	ReceiptURL      string                 `json:"receipt_url,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	
	// Status and approvals
	Status          string                 `json:"status" gorm:"default:'pending'"` // pending, confirmed, rejected
	ApprovedBy      string                 `json:"approved_by,omitempty" gorm:"type:uuid"`
	ApprovedAt      *time.Time             `json:"approved_at,omitempty"`
	
	// Timestamps
	CreatedAt       time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time             `json:"deleted_at,omitempty" gorm:"index"`
}

// RecurringRule defines rules for recurring transactions
type RecurringRule struct {
	Frequency       string    `json:"frequency"` // daily, weekly, monthly, yearly
	Interval        int       `json:"interval"`  // e.g., every 2 weeks
	EndDate         *time.Time `json:"end_date,omitempty"`
	MaxOccurrences  int       `json:"max_occurrences,omitempty"`
	NextOccurrence  time.Time `json:"next_occurrence"`
}

// CreditCard represents a secure credit card record (NO PAN/CVV storage)
type CreditCard struct {
	ID              string    `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID          string    `json:"user_id" gorm:"type:uuid;not null;index"`
	CircleID        string    `json:"circle_id" gorm:"type:uuid;index"`
	
	// Secure storage (NO PAN/CVV per requirements)
	BankName        string    `json:"bank_name" gorm:"not null"` // e.g., "Banco Galicia"
	CardType        string    `json:"card_type" gorm:"not null"` // Visa, Mastercard, Amex
	LastFour        string    `json:"last_four" gorm:"type:varchar(4);not null"`
	CardholderName  string    `json:"cardholder_name"`
	
	// Billing and limits
	CreditLimit     float64   `json:"credit_limit,omitempty" gorm:"type:decimal(15,2)"`
	CurrentBalance  float64   `json:"current_balance" gorm:"type:decimal(15,2);default:0"`
	PaymentDueDate  time.Time `json:"payment_due_date"`
	ClosingDate     time.Time `json:"closing_date"`
	
	// Status
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	IsDefault       bool      `json:"is_default" gorm:"default:false"`
	
	// Timestamps
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// ExchangeRate stores daily currency exchange rates
type ExchangeRate struct {
	ID              string    `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BaseCurrency    string    `json:"base_currency" gorm:"type:varchar(3);not null;index"`
	TargetCurrency  string    `json:"target_currency" gorm:"type:varchar(3);not null;index"`
	Rate            float64   `json:"rate" gorm:"type:decimal(10,6);not null"`
	Date            time.Time `json:"date" gorm:"not null;index"`
	Source          string    `json:"source"` // API source name
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TransactionFilter defines filters for querying transactions
type TransactionFilter struct {
	CircleID        string    `json:"circle_id,omitempty"`
	UserID          string    `json:"user_id,omitempty"`
	StartDate       time.Time `json:"start_date,omitempty"`
	EndDate         time.Time `json:"end_date,omitempty"`
	Category        string    `json:"category,omitempty"`
	PaymentMethod   string    `json:"payment_method,omitempty"`
	Currency        string    `json:"currency,omitempty"`
	MinAmount       float64   `json:"min_amount,omitempty"`
	MaxAmount       float64   `json:"max_amount,omitempty"`
	Tags            []string  `json:"tags,omitempty"`
	IsPrivate       *bool     `json:"is_private,omitempty"`
	IsAnomaly       *bool     `json:"is_anomaly,omitempty"`
	Source          string    `json:"source,omitempty"`
	Status          string    `json:"status,omitempty"`
	Limit           int       `json:"limit,omitempty"`
	Offset          int       `json:"offset,omitempty"`
}

// TransactionSummary provides aggregated transaction data with multi-currency support
type TransactionSummary struct {
	TotalIncome     float64            `json:"total_income"`
	TotalExpenses   float64            `json:"total_expenses"`
	NetBalance      float64            `json:"net_balance"`
	BaseCurrency    string             `json:"base_currency"`
	
	// Currency breakdown
	ByCurrency      map[string]CurrencySummary `json:"by_currency"`
	
	// Category breakdown (converted to base currency)
	ByCategory      map[string]float64 `json:"by_category"`
	
	// Payment method breakdown
	ByPaymentMethod map[string]float64 `json:"by_payment_method"`
	
	// Monthly trends
	MonthlyTrend    []MonthlyTrend     `json:"monthly_trend"`
	
	// Anomaly statistics
	AnomalyCount    int                `json:"anomaly_count"`
	WarningCount    int                `json:"warning_count"`
	CriticalCount   int                `json:"critical_count"`
}

// CurrencySummary provides totals in a specific currency
type CurrencySummary struct {
	Income          float64 `json:"income"`
	Expenses        float64 `json:"expenses"`
	Balance         float64 `json:"balance"`
	ExchangeRate    float64 `json:"exchange_rate"`
	ConvertedAmount float64 `json:"converted_amount"` // Amount in base currency
}

// MonthlyTrend shows monthly transaction trends
type MonthlyTrend struct {
	Month     string  `json:"month"`
	Income    float64 `json:"income"`
	Expenses  float64 `json:"expenses"`
	Balance   float64 `json:"balance"`
	Currency  string  `json:"currency"`
}

// AnomalyDetectionRule defines rules for detecting irregular expenses
type AnomalyDetectionRule struct {
	ID              string    `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CircleID        string    `json:"circle_id" gorm:"type:uuid;index"`
	Name            string    `json:"name" gorm:"not null"`
	Description     string    `json:"description"`
	
	// Rule configuration
	RuleType        string    `json:"rule_type" gorm:"not null"` // duplicate, amount_mismatch, orphan, custom
	Condition       string    `json:"condition" gorm:"type:text"` // JSON or DSL condition
	Severity        string    `json:"severity" gorm:"default:'warning'"` // warning, critical
	
	// Thresholds
	AmountThreshold float64   `json:"amount_threshold,omitempty"`
	DaysThreshold   int       `json:"days_threshold,omitempty"`
	
	// Activation
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
	
	// Timestamps
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DetectedAnomaly represents a detected irregular expense
type DetectedAnomaly struct {
	ID                string    `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TransactionID     string    `json:"transaction_id" gorm:"type:uuid;not null;index"`
	RuleID            string    `json:"rule_id" gorm:"type:uuid;index"`
	CircleID          string    `json:"circle_id" gorm:"type:uuid;index"`
	
	// Anomaly details
	Type              string    `json:"type"` // duplicate, amount_mismatch, orphan
	Severity          string    `json:"severity"` // warning, critical
	Description       string    `json:"description"`
	
	// Matching details (for duplicates)
	MatchedTransactionID string    `json:"matched_transaction_id,omitempty" gorm:"type:uuid"`
	AmountDifference   float64   `json:"amount_difference,omitempty"`
	
	// Resolution
	Status            string    `json:"status" gorm:"default:'pending'"` // pending, reviewed, resolved, ignored
	ResolvedBy        string    `json:"resolved_by,omitempty" gorm:"type:uuid"`
	ResolutionNote    string    `json:"resolution_note,omitempty"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
	
	// Timestamps
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Helper function to validate currency
func IsValidCurrency(currency string) bool {
	validCurrencies := map[string]bool{
		"ARS": true,
		"USD": true,
		"EUR": true,
	}
	return validCurrencies[currency]
}

// Helper function to get obfuscated description for private expenses
func GetObfuscatedDescription(userName string) string {
	return "Gasto Privado de " + userName
}