package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// StringSlice is a custom type for storing []string as JSONB in PostgreSQL
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	bytes, err := json.Marshal(s)
	return string(bytes), err
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return json.Unmarshal(nil, s)
	}
	return json.Unmarshal(bytes, s)
}

type Circle struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Currency    string         `json:"currency" gorm:"default:'USD'"`
	CreatedBy   string         `json:"created_by" gorm:"type:uuid;not null"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	IsPublic    bool           `json:"is_public" gorm:"default:false"` // Public circles can be joined by anyone
	JoinCode    string         `json:"join_code" gorm:"uniqueIndex"` // Code for joining private circles
	Settings    CircleSettings `json:"settings" gorm:"embedded"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// Relationships
	Members     []CircleMember `json:"members" gorm:"foreignKey:CircleID"`
	Transactions []Transaction  `json:"transactions" gorm:"foreignKey:CircleID"`
}

type CircleMember struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CircleID  string    `json:"circle_id" gorm:"type:uuid;not null;index"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Role      string    `json:"role" gorm:"default:'member'"` // admin, member, viewer
	JoinedAt  time.Time `json:"joined_at" gorm:"autoCreateTime"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	
	// User info (joined)
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

type CircleSettings struct {
	AllowMemberAddTransactions    bool     `json:"allow_member_add_transactions" gorm:"default:true"`
	AllowMemberEditTransactions   bool     `json:"allow_member_edit_transactions" gorm:"default:true"`
	AllowMemberDeleteTransactions bool     `json:"allow_member_delete_transactions" gorm:"default:false"`
	RequireApprovalForAdd         bool     `json:"require_approval_for_add" gorm:"default:false"`
	RequireApprovalForEdit        bool     `json:"require_approval_for_edit" gorm:"default:false"`
	DefaultCategories             StringSlice `json:"default_categories" gorm:"type:jsonb"`
	BudgetAlertsEnabled           bool     `json:"budget_alerts_enabled" gorm:"default:true"`
	MonthlyBudget                float64  `json:"monthly_budget" gorm:"default:0"`
	NotificationPreferences       NotificationPreferences `json:"notification_preferences" gorm:"embedded"`
}

type NotificationPreferences struct {
	EmailNotifications      bool `json:"email_notifications" gorm:"default:true"`
	PushNotifications       bool `json:"push_notifications" gorm:"default:true"`
	WeeklySummary           bool `json:"weekly_summary" gorm:"default:true"`
	BudgetAlerts            bool `json:"budget_alerts" gorm:"default:true"`
	LargeTransactionAlerts  bool `json:"large_transaction_alerts" gorm:"default:true"`
	NewMemberAlerts         bool `json:"new_member_alerts" gorm:"default:true"`
	TransactionApprovalAlerts bool `json:"transaction_approval_alerts" gorm:"default:true"`
}

type CircleSummary struct {
	CircleID           string    `json:"circle_id"`
	TotalMembers       int       `json:"total_members"`
	ActiveMembers      int       `json:"active_members"`
	TotalTransactions  int       `json:"total_transactions"`
	TotalAmount        float64   `json:"total_amount"`
	MonthlyBudget      float64   `json:"monthly_budget"`
	BudgetUsed         float64   `json:"budget_used"`
	BudgetRemaining    float64   `json:"budget_remaining"`
	LastTransactionAt  time.Time `json:"last_transaction_at"`
	CreatedAt          time.Time `json:"created_at"`
}

// CreateCircleRequest represents circle creation data
type CreateCircleRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	Currency    string `json:"currency" validate:"required,iso4217"`
	IsPublic    bool   `json:"is_public"`
	Settings    *CircleSettings `json:"settings"`
}

// UpdateCircleRequest represents circle update data
type UpdateCircleRequest struct {
	Name        string `json:"name" validate:"omitempty,min=3,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Currency    string `json:"currency" validate:"omitempty,iso4217"`
	IsPublic    *bool  `json:"is_public"`
	Settings    *CircleSettings `json:"settings"`
}

// AddMemberRequest represents adding a member to a circle
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required,uuid4"`
	Role   string `json:"role" validate:"required,oneof=admin member viewer"`
}

// JoinCircleRequest represents joining a circle with a code
type JoinCircleRequest struct {
	JoinCode string `json:"join_code" validate:"required,min=6,max=20"`
}

// CircleInvite represents an invitation to join a circle
type CircleInvite struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CircleID  string    `json:"circle_id" gorm:"type:uuid;not null;index"`
	Email     string    `json:"email" gorm:"not null"`
	InvitedBy string    `json:"invited_by" gorm:"type:uuid;not null"`
	Role      string    `json:"role" gorm:"default:'member'"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"index"`
	Accepted  bool      `json:"accepted" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CircleActivity represents activity in a circle
type CircleActivity struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CircleID   string    `json:"circle_id" gorm:"type:uuid;not null;index"`
	UserID     string    `json:"user_id" gorm:"type:uuid;not null"`
	Activity   string    `json:"activity" gorm:"not null"` // transaction_added, member_joined, etc.
	Details    string    `json:"details" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	
	// User info (joined)
	User       User      `json:"user" gorm:"foreignKey:UserID"`
}