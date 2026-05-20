package model

import (
	"time"
)

// User represents a system user
type User struct {
	ID           string                 `json:"id"`
	Email        string                 `json:"email"`
	Username     string                 `json:"username"`
	FullName     string                 `json:"full_name"`
	AvatarURL    string                 `json:"avatar_url,omitempty"`
	Preferences  UserPreferences        `json:"preferences"`
	IsActive     bool                   `json:"is_active"`
	LastLoginAt  time.Time              `json:"last_login_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UserPreferences contains user-specific settings
type UserPreferences struct {
	DefaultCurrency string            `json:"default_currency"`
	Language        string            `json:"language"`
	TimeZone        string            `json:"time_zone"`
	DateFormat      string            `json:"date_format"`
	CurrencyFormat  string            `json:"currency_format"`
	Theme           string            `json:"theme"` // light, dark, system
	NotificationSettings NotificationSettings `json:"notification_settings"`
}

// NotificationSettings defines user notification preferences
type NotificationSettings struct {
	EmailFrequency string `json:"email_frequency"` // daily, weekly, monthly, never
	PushEnabled    bool   `json:"push_enabled"`
	DesktopAlerts  bool   `json:"desktop_alerts"`
	SoundEnabled   bool   `json:"sound_enabled"`
}

// UserSession represents an active user session
type UserSession struct {
	SessionID   string    `json:"session_id"`
	UserID      string    `json:"user_id"`
	DeviceInfo  string    `json:"device_info,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// UserFilter defines filters for querying users
type UserFilter struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// UserAuth represents authentication data
type UserAuth struct {
	UserID       string    `json:"user_id"`
	PasswordHash string    `json:"password_hash"`
	Salt         string    `json:"salt"`
	MFAEnabled   bool      `json:"mfa_enabled"`
	MFASecret    string    `json:"mfa_secret,omitempty"`
	LastPasswordChange time.Time `json:"last_password_change"`
	FailedAttempts int     `json:"failed_attempts"`
	LockedUntil   time.Time `json:"locked_until,omitempty"`
}