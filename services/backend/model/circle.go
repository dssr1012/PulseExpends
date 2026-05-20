package model

import (
	"time"
)

// Circle represents a family circle for expense tracking
type Circle struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Currency    string                 `json:"currency"` // Default currency for the circle
	Members     []Member               `json:"members"`
	Settings    CircleSettings         `json:"settings"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Member represents a user within a circle
type Member struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"` // admin, member, viewer
	JoinedAt  time.Time `json:"joined_at"`
	IsActive  bool      `json:"is_active"`
}

// CircleSettings contains configurable settings for a circle
type CircleSettings struct {
	AllowMemberAddTransactions bool     `json:"allow_member_add_transactions"`
	AllowMemberEditTransactions bool    `json:"allow_member_edit_transactions"`
	AllowMemberDeleteTransactions bool  `json:"allow_member_delete_transactions"`
	DefaultCategories          []string `json:"default_categories"`
	BudgetAlertsEnabled        bool     `json:"budget_alerts_enabled"`
	MonthlyBudget             float64  `json:"monthly_budget,omitempty"`
	NotificationPreferences   NotificationPreferences `json:"notification_preferences"`
}

// NotificationPreferences defines notification settings
type NotificationPreferences struct {
	EmailNotifications bool `json:"email_notifications"`
	PushNotifications  bool `json:"push_notifications"`
	WeeklySummary      bool `json:"weekly_summary"`
	BudgetAlerts       bool `json:"budget_alerts"`
	LargeTransactionAlerts bool `json:"large_transaction_alerts"`
}

// CircleFilter defines filters for querying circles
type CircleFilter struct {
	UserID    string `json:"user_id,omitempty"`
	Name      string `json:"name,omitempty"`
	IsActive  *bool  `json:"is_active,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// CircleSummary provides aggregated circle data
type CircleSummary struct {
	CircleID          string  `json:"circle_id"`
	TotalMembers      int     `json:"total_members"`
	ActiveMembers     int     `json:"active_members"`
	TotalTransactions int     `json:"total_transactions"`
	TotalAmount       float64 `json:"total_amount"`
	LastTransactionAt time.Time `json:"last_transaction_at,omitempty"`
}