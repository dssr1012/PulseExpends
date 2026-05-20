package repository

import (
	"context"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
)

// TransactionRepository defines the interface for transaction data operations
type TransactionRepository interface {
	// Create creates a new transaction
	Create(ctx context.Context, transaction *model.Transaction) error
	
	// FindByID finds a transaction by its ID
	FindByID(ctx context.Context, id string) (*model.Transaction, error)
	
	// FindByCircle finds transactions for a specific circle with optional filters
	FindByCircle(ctx context.Context, circleID string, filter model.TransactionFilter) ([]model.Transaction, error)
	
	// FindByUser finds transactions for a specific user across all circles
	FindByUser(ctx context.Context, userID string, filter model.TransactionFilter) ([]model.Transaction, error)
	
	// Update updates an existing transaction
	Update(ctx context.Context, transaction *model.Transaction) error
	
	// Delete deletes a transaction
	Delete(ctx context.Context, id string) error
	
	// GetSummary gets transaction summary for a circle
	GetSummary(ctx context.Context, circleID string, startDate, endDate time.Time) (*model.TransactionSummary, error)
	
	// GetMonthlyTrend gets monthly transaction trends
	GetMonthlyTrend(ctx context.Context, circleID string, months int) ([]model.MonthlyTrend, error)
	
	// BulkCreate creates multiple transactions
	BulkCreate(ctx context.Context, transactions []model.Transaction) error
	
	// Close closes the repository connection
	Close() error
}

// CircleRepository defines the interface for circle data operations
type CircleRepository interface {
	// Create creates a new circle
	Create(ctx context.Context, circle *model.Circle) error
	
	// FindByID finds a circle by its ID
	FindByID(ctx context.Context, id string) (*model.Circle, error)
	
	// FindByUser finds circles for a specific user
	FindByUser(ctx context.Context, userID string, filter model.CircleFilter) ([]model.Circle, error)
	
	// Update updates an existing circle
	Update(ctx context.Context, circle *model.Circle) error
	
	// Delete deletes a circle
	Delete(ctx context.Context, id string) error
	
	// AddMember adds a member to a circle
	AddMember(ctx context.Context, circleID string, member model.Member) error
	
	// RemoveMember removes a member from a circle
	RemoveMember(ctx context.Context, circleID, userID string) error
	
	// UpdateMember updates a member's role
	UpdateMember(ctx context.Context, circleID, userID, role string) error
	
	// GetSummary gets circle summary
	GetSummary(ctx context.Context, circleID string) (*model.CircleSummary, error)
	
	// Close closes the repository connection
	Close() error
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *model.User) error
	
	// FindByID finds a user by ID
	FindByID(ctx context.Context, id string) (*model.User, error)
	
	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	
	// FindByUsername finds a user by username
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	
	// Update updates an existing user
	Update(ctx context.Context, user *model.User) error
	
	// Delete deletes a user
	Delete(ctx context.Context, id string) error
	
	// CreateSession creates a new user session
	CreateSession(ctx context.Context, session *model.UserSession) error
	
	// FindSession finds a session by ID
	FindSession(ctx context.Context, sessionID string) (*model.UserSession, error)
	
	// UpdateSession updates a session's last used time
	UpdateSession(ctx context.Context, sessionID string) error
	
	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, sessionID string) error
	
	// DeleteExpiredSessions deletes expired sessions
	DeleteExpiredSessions(ctx context.Context) error
	
	// Close closes the repository connection
	Close() error
}

// AuthRepository defines the interface for authentication data operations
type AuthRepository interface {
	// CreateAuth creates authentication data for a user
	CreateAuth(ctx context.Context, auth *model.UserAuth) error
	
	// FindAuthByUserID finds authentication data by user ID
	FindAuthByUserID(ctx context.Context, userID string) (*model.UserAuth, error)
	
	// UpdateAuth updates authentication data
	UpdateAuth(ctx context.Context, auth *model.UserAuth) error
	
	// IncrementFailedAttempts increments failed login attempts
	IncrementFailedAttempts(ctx context.Context, userID string) error
	
	// ResetFailedAttempts resets failed login attempts
	ResetFailedAttempts(ctx context.Context, userID string) error
	
	// LockAccount locks a user account
	LockAccount(ctx context.Context, userID string, until time.Time) error
	
	// UnlockAccount unlocks a user account
	UnlockAccount(ctx context.Context, userID string) error
	
	// Close closes the repository connection
	Close() error
}

// RepositoryManager manages all repositories
type RepositoryManager struct {
	Transaction TransactionRepository
	Circle      CircleRepository
	User        UserRepository
	Auth        AuthRepository
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(
	transaction TransactionRepository,
	circle CircleRepository,
	user UserRepository,
	auth AuthRepository,
) *RepositoryManager {
	return &RepositoryManager{
		Transaction: transaction,
		Circle:      circle,
		User:        user,
		Auth:        auth,
	}
}

// CloseAll closes all repository connections
func (rm *RepositoryManager) CloseAll() error {
	var errs []error
	
	if rm.Transaction != nil {
		if err := rm.Transaction.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	
	if rm.Circle != nil {
		if err := rm.Circle.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	
	if rm.User != nil {
		if err := rm.User.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	
	if rm.Auth != nil {
		if err := rm.Auth.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	
	if len(errs) > 0 {
		// Return first error for simplicity
		return errs[0]
	}
	
	return nil
}