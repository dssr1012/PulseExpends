package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string         `json:"email" gorm:"uniqueIndex;not null"`
	Username     string         `json:"username" gorm:"uniqueIndex"`
	FullName     string         `json:"full_name"`
	AvatarURL    string         `json:"avatar_url"`
	Provider     string         `json:"provider" gorm:"default:'local'"` // local, google
	ProviderID   string         `json:"provider_id" gorm:"index"` // ID from OAuth provider
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	IsVerified   bool           `json:"is_verified" gorm:"default:false"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// Relationships
	Preferences  UserPreferences `json:"preferences" gorm:"embedded"`
	Auth         UserAuth        `json:"-" gorm:"foreignKey:UserID"`
	Sessions     []UserSession   `json:"-" gorm:"foreignKey:UserID"`
	CircleMembers []CircleMember `json:"-" gorm:"foreignKey:UserID"`
	Transactions []Transaction   `json:"-" gorm:"foreignKey:UserID"`
}

type UserPreferences struct {
	DefaultCurrency string            `json:"default_currency" gorm:"default:'USD'"`
	Language        string            `json:"language" gorm:"default:'es'"`
	TimeZone        string            `json:"time_zone" gorm:"default:'UTC'"`
	DateFormat      string            `json:"date_format" gorm:"default:'YYYY-MM-DD'"`
	CurrencyFormat  string            `json:"currency_format" gorm:"default:'$0,0.00'"`
	Theme           string            `json:"theme" gorm:"default:'light'"` // light, dark, system
	NotificationSettings NotificationSettings `json:"notification_settings" gorm:"embedded"`
}

type NotificationSettings struct {
	EmailFrequency string `json:"email_frequency" gorm:"default:'weekly'"` // daily, weekly, monthly, never
	PushEnabled    bool   `json:"push_enabled" gorm:"default:true"`
	DesktopAlerts  bool   `json:"desktop_alerts" gorm:"default:true"`
	SoundEnabled   bool   `json:"sound_enabled" gorm:"default:true"`
}

type UserAuth struct {
	ID                 string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID             string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	PasswordHash       string    `json:"-" gorm:"type:varchar(255)"`
	Salt               string    `json:"-" gorm:"type:varchar(255)"`
	MFAEnabled         bool      `json:"mfa_enabled" gorm:"default:false"`
	MFASecret          string    `json:"-" gorm:"type:varchar(255)"`
	LastPasswordChange time.Time `json:"last_password_change"`
	FailedAttempts     int       `json:"failed_attempts" gorm:"default:0"`
	LockedUntil        *time.Time `json:"locked_until"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type UserSession struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Token      string    `json:"token" gorm:"type:text;not null;index"`
	DeviceInfo string    `json:"device_info"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	ExpiresAt  time.Time `json:"expires_at" gorm:"index"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	LastUsedAt time.Time `json:"last_used_at" gorm:"autoUpdateTime"`
}

// GoogleUserInfo represents user info from Google OAuth
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	FullName string `json:"full_name" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest represents user login data
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateProfileRequest represents profile update data
type UpdateProfileRequest struct {
	Username string `json:"username" validate:"omitempty,min=3,max=50"`
	FullName string `json:"full_name" validate:"omitempty"`
	AvatarURL string `json:"avatar_url" validate:"omitempty,url"`
}

// UpdatePreferencesRequest represents preferences update data
type UpdatePreferencesRequest struct {
	DefaultCurrency string `json:"default_currency"`
	Language        string `json:"language"`
	TimeZone        string `json:"time_zone"`
	DateFormat      string `json:"date_format"`
	CurrencyFormat  string `json:"currency_format"`
	Theme           string `json:"theme"`
	NotificationSettings *NotificationSettings `json:"notification_settings"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User        *User      `json:"user"`
	Token       string     `json:"token"`
	ExpiresAt   time.Time  `json:"expires_at"`
	SessionID   string     `json:"session_id"`
}

// ErrorResponse represents API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}