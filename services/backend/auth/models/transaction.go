package models

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string         `json:"user_id" gorm:"type:uuid;not null;index"`
	CircleID    *string        `json:"circle_id,omitempty" gorm:"type:uuid;index"` // Optional: if transaction belongs to a circle
	Amount      float64        `json:"amount" gorm:"not null"`
	Currency    string         `json:"currency" gorm:"default:'USD'"`
	Category    string         `json:"category" gorm:"not null"`
	Description string         `json:"description"`
	Date        time.Time      `json:"date" gorm:"not null;index"`
	Type        string         `json:"type" gorm:"not null;index"` // income, expense, transfer
	PaymentMethod string       `json:"payment_method"`
	Location    string         `json:"location"`
	Tags        []string       `json:"tags" gorm:"type:jsonb"`
	ReceiptURL  string         `json:"receipt_url"`
	IsRecurring bool           `json:"is_recurring" gorm:"default:false"`
	RecurringID *string        `json:"recurring_id,omitempty" gorm:"type:uuid;index"`
	Status      string         `json:"status" gorm:"default:'pending'"` // pending, approved, rejected
	ApprovedBy  *string        `json:"approved_by,omitempty" gorm:"type:uuid"`
	ApprovedAt  *time.Time     `json:"approved_at"`
	Notes       string         `json:"notes"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// Relationships
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	Circle      *Circle        `json:"circle,omitempty" gorm:"foreignKey:CircleID"`
	Attachments []Attachment   `json:"attachments" gorm:"foreignKey:TransactionID"`
	Comments    []Comment      `json:"comments" gorm:"foreignKey:TransactionID"`
}

type Attachment struct {
	ID            string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TransactionID string    `json:"transaction_id" gorm:"type:uuid;not null;index"`
	UserID        string    `json:"user_id" gorm:"type:uuid;not null"`
	FileName      string    `json:"file_name" gorm:"not null"`
	FileURL       string    `json:"file_url" gorm:"not null"`
	FileType      string    `json:"file_type"`
	FileSize      int64     `json:"file_size"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	
	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

type Comment struct {
	ID            string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TransactionID string    `json:"transaction_id" gorm:"type:uuid;not null;index"`
	UserID        string    `json:"user_id" gorm:"type:uuid;not null"`
	Content       string    `json:"content" gorm:"not null"`
	IsEdited      bool      `json:"is_edited" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

type RecurringTransaction struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string         `json:"user_id" gorm:"type:uuid;not null;index"`
	CircleID    *string        `json:"circle_id,omitempty" gorm:"type:uuid;index"`
	Amount      float64        `json:"amount" gorm:"not null"`
	Currency    string         `json:"currency" gorm:"default:'USD'"`
	Category    string         `json:"category" gorm:"not null"`
	Description string         `json:"description"`
	Frequency   string         `json:"frequency" gorm:"not null"` // daily, weekly, monthly, yearly
	DayOfMonth  *int           `json:"day_of_month"` // For monthly frequency
	DayOfWeek   *string        `json:"day_of_week"`  // For weekly frequency
	StartDate   time.Time      `json:"start_date" gorm:"not null"`
	EndDate     *time.Time     `json:"end_date"`
	LastRun     *time.Time     `json:"last_run"`
	NextRun     time.Time      `json:"next_run" gorm:"index"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	User   User    `json:"user" gorm:"foreignKey:UserID"`
	Circle *Circle `json:"circle,omitempty" gorm:"foreignKey:CircleID"`
}

// CreateTransactionRequest represents transaction creation data
type CreateTransactionRequest struct {
	Amount        float64                `json:"amount" validate:"required"`
	Currency      string                 `json:"currency" validate:"required,iso4217"`
	Category      string                 `json:"category" validate:"required"`
	Description   string                 `json:"description"`
	Date          time.Time              `json:"date" validate:"required"`
	Type          string                 `json:"type" validate:"required,oneof=income expense transfer"`
	PaymentMethod string                 `json:"payment_method"`
	Location      string                 `json:"location"`
	Tags          []string               `json:"tags"`
	ReceiptURL    string                 `json:"receipt_url" validate:"omitempty,url"`
	IsRecurring   bool                   `json:"is_recurring"`
	CircleID      *string                `json:"circle_id" validate:"omitempty,uuid4"`
	Notes         string                 `json:"notes"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// UpdateTransactionRequest represents transaction update data
type UpdateTransactionRequest struct {
	Amount        *float64               `json:"amount"`
	Currency      string                 `json:"currency" validate:"omitempty,iso4217"`
	Category      string                 `json:"category"`
	Description   string                 `json:"description"`
	Date          *time.Time             `json:"date"`
	Type          string                 `json:"type" validate:"omitempty,oneof=income expense transfer"`
	PaymentMethod string                 `json:"payment_method"`
	Location      string                 `json:"location"`
	Tags          []string               `json:"tags"`
	ReceiptURL    string                 `json:"receipt_url" validate:"omitempty,url"`
	Status        string                 `json:"status" validate:"omitempty,oneof=pending approved rejected"`
	Notes         string                 `json:"notes"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// TransactionFilter represents filters for querying transactions
type TransactionFilter struct {
	UserID     string    `json:"user_id"`
	CircleID   *string   `json:"circle_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	MinAmount  *float64  `json:"min_amount"`
	MaxAmount  *float64  `json:"max_amount"`
	Category   string    `json:"category"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Search     string    `json:"search"`
	Limit      int       `json:"limit" validate:"min=1,max=100"`
	Offset     int       `json:"offset" validate:"min=0"`
	SortBy     string    `json:"sort_by" validate:"oneof=date amount category"`
	SortOrder  string    `json:"sort_order" validate:"oneof=asc desc"`
}

// TransactionSummary represents aggregated transaction data
type TransactionSummary struct {
	TotalIncome     float64            `json:"total_income"`
	TotalExpenses   float64            `json:"total_expenses"`
	NetBalance      float64            `json:"net_balance"`
	ByCategory      map[string]float64 `json:"by_category"`
	ByPaymentMethod map[string]float64 `json:"by_payment_method"`
	ByMonth         map[string]float64 `json:"by_month"`
	TransactionCount int               `json:"transaction_count"`
	AverageAmount   float64            `json:"average_amount"`
	LargestExpense  float64            `json:"largest_expense"`
	LargestIncome   float64            `json:"largest_income"`
}

// SplitTransaction represents a transaction split among multiple users
type SplitTransaction struct {
	ID            string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TransactionID string    `json:"transaction_id" gorm:"type:uuid;not null;index"`
	UserID        string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Amount        float64   `json:"amount" gorm:"not null"`
	Percentage    float64   `json:"percentage" gorm:"not null"`
	IsPaid        bool      `json:"is_paid" gorm:"default:false"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	User        User        `json:"user" gorm:"foreignKey:UserID"`
	Transaction Transaction `json:"transaction" gorm:"foreignKey:TransactionID"`
}

// CreateSplitRequest represents split transaction creation data
type CreateSplitRequest struct {
	TransactionID string               `json:"transaction_id" validate:"required,uuid4"`
	Splits        []SplitDistribution `json:"splits" validate:"required,min=2"`
}

type SplitDistribution struct {
	UserID     string  `json:"user_id" validate:"required,uuid4"`
	Amount     float64 `json:"amount" validate:"required,min=0"`
	Percentage float64 `json:"percentage" validate:"required,min=0,max=100"`
}