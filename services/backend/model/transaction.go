package model

import (
	"time"
)

// Transaction represents a financial transaction in the system
type Transaction struct {
	ID            string                 `json:"id"`
	CircleID      string                 `json:"circle_id"`
	UserID        string                 `json:"user_id"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	Category      string                 `json:"category"`
	PaymentMethod string                 `json:"payment_method"`
	Description   string                 `json:"description"`
	Date          time.Time              `json:"date"`
	IsRecurring   bool                   `json:"is_recurring"`
	RecurringRule *RecurringRule         `json:"recurring_rule,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// RecurringRule defines rules for recurring transactions
type RecurringRule struct {
	Frequency   string    `json:"frequency"` // daily, weekly, monthly, yearly
	Interval    int       `json:"interval"`  // e.g., every 2 weeks
	EndDate     time.Time `json:"end_date,omitempty"`
	MaxOccurrences int    `json:"max_occurrences,omitempty"`
}

// TransactionFilter defines filters for querying transactions
type TransactionFilter struct {
	CircleID      string    `json:"circle_id,omitempty"`
	UserID        string    `json:"user_id,omitempty"`
	StartDate     time.Time `json:"start_date,omitempty"`
	EndDate       time.Time `json:"end_date,omitempty"`
	Category      string    `json:"category,omitempty"`
	PaymentMethod string    `json:"payment_method,omitempty"`
	MinAmount     float64   `json:"min_amount,omitempty"`
	MaxAmount     float64   `json:"max_amount,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
	Limit         int       `json:"limit,omitempty"`
	Offset        int       `json:"offset,omitempty"`
}

// TransactionSummary provides aggregated transaction data
type TransactionSummary struct {
	TotalIncome    float64            `json:"total_income"`
	TotalExpenses  float64            `json:"total_expenses"`
	NetBalance     float64            `json:"net_balance"`
	ByCategory     map[string]float64 `json:"by_category"`
	ByPaymentMethod map[string]float64 `json:"by_payment_method"`
	MonthlyTrend   []MonthlyTrend     `json:"monthly_trend"`
}

// MonthlyTrend shows monthly transaction trends
type MonthlyTrend struct {
	Month     string  `json:"month"`
	Income    float64 `json:"income"`
	Expenses  float64 `json:"expenses"`
	Balance   float64 `json:"balance"`
}