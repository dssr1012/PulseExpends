package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/dssr1012/pulse-expends/internal/repository"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresRepository implements repository interfaces using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository instance
func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// PostgresTransactionRepository implements TransactionRepository using PostgreSQL
type PostgresTransactionRepository struct {
	db *sql.DB
}

// NewPostgresTransactionRepository creates a new transaction repository
func NewPostgresTransactionRepository(db *sql.DB) repository.TransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, circle_id, user_id, amount, currency, category, 
			payment_method, description, date, is_recurring, 
			recurring_rule, tags, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	recurringRuleJSON, _ := json.Marshal(transaction.RecurringRule)
	tagsJSON, _ := json.Marshal(transaction.Tags)
	metadataJSON, _ := json.Marshal(transaction.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		transaction.ID,
		transaction.CircleID,
		transaction.UserID,
		transaction.Amount,
		transaction.Currency,
		transaction.Category,
		transaction.PaymentMethod,
		transaction.Description,
		transaction.Date,
		transaction.IsRecurring,
		recurringRuleJSON,
		tagsJSON,
		metadataJSON,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)

	return err
}

func (r *PostgresTransactionRepository) FindByID(ctx context.Context, id string) (*model.Transaction, error) {
	query := `
		SELECT id, circle_id, user_id, amount, currency, category, 
		       payment_method, description, date, is_recurring, 
		       recurring_rule, tags, metadata, created_at, updated_at
		FROM transactions WHERE id = $1
	`

	var transaction model.Transaction
	var recurringRuleJSON, tagsJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.CircleID,
		&transaction.UserID,
		&transaction.Amount,
		&transaction.Currency,
		&transaction.Category,
		&transaction.PaymentMethod,
		&transaction.Description,
		&transaction.Date,
		&transaction.IsRecurring,
		&recurringRuleJSON,
		&tagsJSON,
		&metadataJSON,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found: %s", id)
		}
		return nil, err
	}

	// Parse JSON fields
	if len(recurringRuleJSON) > 0 {
		json.Unmarshal(recurringRuleJSON, &transaction.RecurringRule)
	}
	if len(tagsJSON) > 0 {
		json.Unmarshal(tagsJSON, &transaction.Tags)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &transaction.Metadata)
	}

	return &transaction, nil
}

func (r *PostgresTransactionRepository) FindByCircle(ctx context.Context, circleID string, filter model.TransactionFilter) ([]model.Transaction, error) {
	query := `
		SELECT id, circle_id, user_id, amount, currency, category, 
		       payment_method, description, date, is_recurring, 
		       recurring_rule, tags, metadata, created_at, updated_at
		FROM transactions 
		WHERE circle_id = $1
	`
	args := []interface{}{circleID}
	argIndex := 2

	// Build WHERE clause based on filters
	if !filter.StartDate.IsZero() {
		query += fmt.Sprintf(" AND date >= $%d", argIndex)
		args = append(args, filter.StartDate)
		argIndex++
	}
	if !filter.EndDate.IsZero() {
		query += fmt.Sprintf(" AND date <= $%d", argIndex)
		args = append(args, filter.EndDate)
		argIndex++
	}
	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, filter.Category)
		argIndex++
	}
	if filter.PaymentMethod != "" {
		query += fmt.Sprintf(" AND payment_method = $%d", argIndex)
		args = append(args, filter.PaymentMethod)
		argIndex++
	}
	if filter.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, filter.UserID)
		argIndex++
	}
	if filter.MinAmount != 0 {
		query += fmt.Sprintf(" AND amount >= $%d", argIndex)
		args = append(args, filter.MinAmount)
		argIndex++
	}
	if filter.MaxAmount != 0 {
		query += fmt.Sprintf(" AND amount <= $%d", argIndex)
		args = append(args, filter.MaxAmount)
		argIndex++
	}

	// Add ordering
	query += " ORDER BY date DESC, created_at DESC"

	// Add limit/offset
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var transaction model.Transaction
		var recurringRuleJSON, tagsJSON, metadataJSON []byte

		err := rows.Scan(
			&transaction.ID,
			&transaction.CircleID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.Currency,
			&transaction.Category,
			&transaction.PaymentMethod,
			&transaction.Description,
			&transaction.Date,
			&transaction.IsRecurring,
			&recurringRuleJSON,
			&tagsJSON,
			&metadataJSON,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if len(recurringRuleJSON) > 0 {
			json.Unmarshal(recurringRuleJSON, &transaction.RecurringRule)
		}
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &transaction.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &transaction.Metadata)
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (r *PostgresTransactionRepository) FindByUser(ctx context.Context, userID string, filter model.TransactionFilter) ([]model.Transaction, error) {
	query := `
		SELECT t.id, t.circle_id, t.user_id, t.amount, t.currency, t.category, 
		       t.payment_method, t.description, t.date, t.is_recurring, 
		       t.recurring_rule, t.tags, t.metadata, t.created_at, t.updated_at
		FROM transactions t
		INNER JOIN circle_members cm ON t.circle_id = cm.circle_id
		WHERE cm.user_id = $1 AND cm.is_active = true
	`
	args := []interface{}{userID}
	argIndex := 2

	// Build WHERE clause based on filters
	if filter.CircleID != "" {
		query += fmt.Sprintf(" AND t.circle_id = $%d", argIndex)
		args = append(args, filter.CircleID)
		argIndex++
	}
	if !filter.StartDate.IsZero() {
		query += fmt.Sprintf(" AND t.date >= $%d", argIndex)
		args = append(args, filter.StartDate)
		argIndex++
	}
	if !filter.EndDate.IsZero() {
		query += fmt.Sprintf(" AND t.date <= $%d", argIndex)
		args = append(args, filter.EndDate)
		argIndex++
	}
	if filter.Category != "" {
		query += fmt.Sprintf(" AND t.category = $%d", argIndex)
		args = append(args, filter.Category)
		argIndex++
	}
	if filter.PaymentMethod != "" {
		query += fmt.Sprintf(" AND t.payment_method = $%d", argIndex)
		args = append(args, filter.PaymentMethod)
		argIndex++
	}
	if filter.MinAmount != 0 {
		query += fmt.Sprintf(" AND t.amount >= $%d", argIndex)
		args = append(args, filter.MinAmount)
		argIndex++
	}
	if filter.MaxAmount != 0 {
		query += fmt.Sprintf(" AND t.amount <= $%d", argIndex)
		args = append(args, filter.MaxAmount)
		argIndex++
	}

	// Add ordering
	query += " ORDER BY t.date DESC, t.created_at DESC"

	// Add limit/offset
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var transaction model.Transaction
		var recurringRuleJSON, tagsJSON, metadataJSON []byte

		err := rows.Scan(
			&transaction.ID,
			&transaction.CircleID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.Currency,
			&transaction.Category,
			&transaction.PaymentMethod,
			&transaction.Description,
			&transaction.Date,
			&transaction.IsRecurring,
			&recurringRuleJSON,
			&tagsJSON,
			&metadataJSON,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if len(recurringRuleJSON) > 0 {
			json.Unmarshal(recurringRuleJSON, &transaction.RecurringRule)
		}
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &transaction.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &transaction.Metadata)
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (r *PostgresTransactionRepository) Update(ctx context.Context, transaction *model.Transaction) error {
	query := `
		UPDATE transactions SET
			circle_id = $2,
			user_id = $3,
			amount = $4,
			currency = $5,
			category = $6,
			payment_method = $7,
			description = $8,
			date = $9,
			is_recurring = $10,
			recurring_rule = $11,
			tags = $12,
			metadata = $13,
			updated_at = $14
		WHERE id = $1
	`

	recurringRuleJSON, _ := json.Marshal(transaction.RecurringRule)
	tagsJSON, _ := json.Marshal(transaction.Tags)
	metadataJSON, _ := json.Marshal(transaction.Metadata)
	transaction.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		transaction.ID,
		transaction.CircleID,
		transaction.UserID,
		transaction.Amount,
		transaction.Currency,
		transaction.Category,
		transaction.PaymentMethod,
		transaction.Description,
		transaction.Date,
		transaction.IsRecurring,
		recurringRuleJSON,
		tagsJSON,
		metadataJSON,
		transaction.UpdatedAt,
	)

	return err
}

func (r *PostgresTransactionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM transactions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresTransactionRepository) GetSummary(ctx context.Context, circleID string, startDate, endDate time.Time) (*model.TransactionSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN amount >= 0 THEN amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as total_expenses,
			COALESCE(SUM(amount), 0) as net_balance
		FROM transactions 
		WHERE circle_id = $1 AND date BETWEEN $2 AND $3
	`

	var summary model.TransactionSummary
	err := r.db.QueryRowContext(ctx, query, circleID, startDate, endDate).Scan(
		&summary.TotalIncome,
		&summary.TotalExpenses,
		&summary.NetBalance,
	)
	if err != nil {
		return nil, err
	}

	// Get by category
	categoryQuery := `
		SELECT category, SUM(amount) as total
		FROM transactions 
		WHERE circle_id = $1 AND date BETWEEN $2 AND $3
		GROUP BY category
	`
	rows, err := r.db.QueryContext(ctx, categoryQuery, circleID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary.ByCategory = make(map[string]float64)
	for rows.Next() {
		var category string
		var total float64
		if err := rows.Scan(&category, &total); err == nil {
			summary.ByCategory[category] = total
		}
	}

	// Get by payment method
	paymentQuery := `
		SELECT payment_method, SUM(amount) as total
		FROM transactions 
		WHERE circle_id = $1 AND date BETWEEN $2 AND $3
		GROUP BY payment_method
	`
	rows2, err := r.db.QueryContext(ctx, paymentQuery, circleID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	summary.ByPaymentMethod = make(map[string]float64)
	for rows2.Next() {
		var method string
		var total float64
		if err := rows2.Scan(&method, &total); err == nil {
			summary.ByPaymentMethod[method] = total
		}
	}

	// Get monthly trend
	monthlyQuery := `
		SELECT 
			TO_CHAR(date, 'YYYY-MM') as month,
			COALESCE(SUM(CASE WHEN amount >= 0 THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as expenses,
			COALESCE(SUM(amount), 0) as balance
		FROM transactions 
		WHERE circle_id = $1 AND date BETWEEN $2 AND $3
		GROUP BY TO_CHAR(date, 'YYYY-MM')
		ORDER BY month
	`
	rows3, err := r.db.QueryContext(ctx, monthlyQuery, circleID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows3.Close()

	for rows3.Next() {
		var trend model.MonthlyTrend
		if err := rows3.Scan(&trend.Month, &trend.Income, &trend.Expenses, &trend.Balance); err == nil {
			summary.MonthlyTrend = append(summary.MonthlyTrend, trend)
		}
	}

	return &summary, nil
}

func (r *PostgresTransactionRepository) GetMonthlyTrend(ctx context.Context, circleID string, months int) ([]model.MonthlyTrend, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)

	query := `
		SELECT 
			TO_CHAR(date, 'YYYY-MM') as month,
			COALESCE(SUM(CASE WHEN amount >= 0 THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as expenses,
			COALESCE(SUM(amount), 0) as balance
		FROM transactions 
		WHERE circle_id = $1 AND date BETWEEN $2 AND $3
		GROUP BY TO_CHAR(date, 'YYYY-MM')
		ORDER BY month
	`

	rows, err := r.db.QueryContext(ctx, query, circleID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []model.MonthlyTrend
	for rows.Next() {
		var trend model.MonthlyTrend
		if err := rows.Scan(&trend.Month, &trend.Income, &trend.Expenses, &trend.Balance); err != nil {
			return nil, err
		}
		trends = append(trends, trend)
	}

	return trends, nil
}

func (r *PostgresTransactionRepository) BulkCreate(ctx context.Context, transactions []model.Transaction) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO transactions (
			id, circle_id, user_id, amount, currency, category, 
			payment_method, description, date, is_recurring, 
			recurring_rule, tags, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, transaction := range transactions {
		if transaction.ID == "" {
			transaction.ID = generateUUID()
		}
		if transaction.CreatedAt.IsZero() {
			transaction.CreatedAt = time.Now()
		}
		transaction.UpdatedAt = time.Now()

		recurringRuleJSON, _ := json.Marshal(transaction.RecurringRule)
		tagsJSON, _ := json.Marshal(transaction.Tags)
		metadataJSON, _ := json.Marshal(transaction.Metadata)

		_, err := stmt.ExecContext(ctx,
			transaction.ID,
			transaction.CircleID,
			transaction.UserID,
			transaction.Amount,
			transaction.Currency,
			transaction.Category,
			transaction.PaymentMethod,
			transaction.Description,
			transaction.Date,
			transaction.IsRecurring,
			recurringRuleJSON,
			tagsJSON,
			metadataJSON,
			transaction.CreatedAt,
			transaction.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresTransactionRepository) Close() error {
	return r.db.Close()
}

// Helper functions
func createTables(db *sql.DB) error {
	queries := []string{
		// Create users table
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			username VARCHAR(100) UNIQUE NOT NULL,
			full_name VARCHAR(255) NOT NULL,
			avatar_url TEXT,
			preferences JSONB DEFAULT '{}',
			is_active BOOLEAN DEFAULT true,
			last_login_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			metadata JSONB DEFAULT '{}'
		)`,

		// Create circles table
		`CREATE TABLE IF NOT EXISTS circles (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			currency VARCHAR(3) NOT NULL DEFAULT 'CLP',
			settings JSONB DEFAULT '{}',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			metadata JSONB DEFAULT '{}'
		)`,

		// Create circle_members table
		`CREATE TABLE IF NOT EXISTS circle_members (
			circle_id VARCHAR(36) NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL DEFAULT 'member',
			joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
			is_active BOOLEAN DEFAULT true,
			PRIMARY KEY (circle_id, user_id)
		)`,

		// Create transactions table
		`CREATE TABLE IF NOT EXISTS transactions (
			id VARCHAR(36) PRIMARY KEY,
			circle_id VARCHAR(36) NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			amount DECIMAL(15,2) NOT NULL,
			currency VARCHAR(3) NOT NULL DEFAULT 'CLP',
			category VARCHAR(100) NOT NULL,
			payment_method VARCHAR(50) NOT NULL,
			description TEXT,
			date TIMESTAMP NOT NULL,
			is_recurring BOOLEAN DEFAULT false,
			recurring_rule JSONB,
			tags JSONB DEFAULT '[]',
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			INDEX idx_transactions_circle_date (circle_id, date DESC),
			INDEX idx_transactions_user_date (user_id, date DESC),
			INDEX idx_transactions_category (category),
			INDEX idx_transactions_payment_method (payment_method)
		)`,

		// Create user_auth table
		`CREATE TABLE IF NOT EXISTS user_auth (
			user_id VARCHAR(36) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			password_hash VARCHAR(255) NOT NULL,
			salt VARCHAR(255) NOT NULL,
			mfa_enabled BOOLEAN DEFAULT false,
			mfa_secret VARCHAR(255),
			last_password_change TIMESTAMP NOT NULL DEFAULT NOW(),
			failed_attempts INTEGER DEFAULT 0,
			locked_until TIMESTAMP
		)`,

		// Create user_sessions table
		`CREATE TABLE IF NOT EXISTS user_sessions (
			session_id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			device_info TEXT,
			ip_address VARCHAR(45),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			last_used_at TIMESTAMP NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMP NOT NULL,
			INDEX idx_user_sessions_user_id (user_id),
			INDEX idx_user_sessions_expires_at (expires_at)
		)`,

		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_circles_name ON circles(name)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

func generateUUID() string {
	// Simple UUID generation for now
	// In production, use github.com/google/uuid
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Import json package
import "encoding/json"

// PostgresCircleRepository and PostgresUserRepository implementations would follow similar patterns
// For brevity, we'll create factory functions

func NewPostgresCircleRepository(db *sql.DB) repository.CircleRepository {
	return &PostgresCircleRepository{db: db}
}

func NewPostgresUserRepository(db *sql.DB) repository.UserRepository {
	return &PostgresUserRepository{db: db}
}

func NewPostgresAuthRepository(db *sql.DB) repository.AuthRepository {
	return &PostgresAuthRepository{db: db}
}


// ============================================================================
// FACTORY FUNCTIONS FOR NEW REPOSITORIES
// ============================================================================

// NewPostgresCreditCardRepository creates a new credit card repository
func NewPostgresCreditCardRepository(db *sql.DB) repository.CreditCardRepository {
	return &PostgresCreditCardRepository{db: db}
}

// NewPostgresExchangeRateRepository creates a new exchange rate repository
func NewPostgresExchangeRateRepository(db *sql.DB) repository.ExchangeRateRepository {
	return &PostgresExchangeRateRepository{db: db}
}

// NewPostgresMobileNotificationRepository creates a new mobile notification repository
func NewPostgresMobileNotificationRepository(db *sql.DB) repository.MobileNotificationRepository {
	return &PostgresMobileNotificationRepository{db: db}
}

// NewPostgresAnomalyDetectionRepository creates a new anomaly detection repository
func NewPostgresAnomalyDetectionRepository(db *sql.DB) repository.AnomalyDetectionRepository {
	return &PostgresAnomalyDetectionRepository{db: db}
}

// NewPostgresNotificationAppWhitelistRepository creates a new notification app whitelist repository
func NewPostgresNotificationAppWhitelistRepository(db *sql.DB) repository.NotificationAppWhitelistRepository {
	return &PostgresNotificationAppWhitelistRepository{db: db}
}


// Stub implementations for other repositories
type PostgresCircleRepository struct {
	db *sql.DB
}

func (r *PostgresCircleRepository) Create(ctx context.Context, circle *model.Circle) error {
	// Implementation would be similar to transaction repository
	return nil
}

func (r *PostgresCircleRepository) FindByID(ctx context.Context, id string) (*model.Circle, error) {
	return nil, nil
}

func (r *PostgresCircleRepository) FindByUser(ctx context.Context, userID string, filter model.CircleFilter) ([]model.Circle, error) {
	return nil, nil
}

func (r *PostgresCircleRepository) Update(ctx context.Context, circle *model.Circle) error {
	return nil
}

func (r *PostgresCircleRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (r *PostgresCircleRepository) AddMember(ctx context.Context, circleID string, member model.Member) error {
	return nil
}

func (r *PostgresCircleRepository) RemoveMember(ctx context.Context, circleID, userID string) error {
	return nil
}

func (r *PostgresCircleRepository) UpdateMember(ctx context.Context, circleID, userID, role string) error {
	return nil
}

func (r *PostgresCircleRepository) GetSummary(ctx context.Context, circleID string) (*model.CircleSummary, error) {
	return nil, nil
}

func (r *PostgresCircleRepository) Close() error {
	return nil
}

type PostgresUserRepository struct {
	db *sql.DB
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *model.User) error {
	return nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	return nil, nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, nil
}

func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	return nil, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *model.User) error {
	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (r *PostgresUserRepository) CreateSession(ctx context.Context, session *model.UserSession) error {
	return nil
}

func (r *PostgresUserRepository) FindSession(ctx context.Context, sessionID string) (*model.UserSession, error) {
	return nil, nil
}

func (r *PostgresUserRepository) UpdateSession(ctx context.Context, sessionID string) error {
	return nil
}

func (r *PostgresUserRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return nil
}

func (r *PostgresUserRepository) DeleteExpiredSessions(ctx context.Context) error {
	return nil
}

func (r *PostgresUserRepository) Close() error {
	return nil
}

type PostgresAuthRepository struct {
	db *sql.DB
}


// ============================================================================
// NEW REPOSITORY IMPLEMENTATIONS FOR REQUIREMENTS
// ============================================================================

// PostgresCreditCardRepository implements CreditCardRepository using PostgreSQL
type PostgresCreditCardRepository struct {
	db *sql.DB
}

// ============================================================================
// CREDIT CARD REPOSITORY IMPLEMENTATION
// ============================================================================

// validateCreditCard performs security validation on credit card data
func validateCreditCard(card *model.CreditCard) error {
	if card == nil {
		return fmt.Errorf("credit card cannot be nil")
	}

	// Validate last four digits (must be exactly 4 digits)
	if err := validateLastFour(card.LastFour); err != nil {
		return fmt.Errorf("invalid last four digits: %v", err)
	}

	// Validate bank name
	if err := validateBankName(card.BankName); err != nil {
		return fmt.Errorf("invalid bank name: %v", err)
	}

	// Validate card type
	if err := validateCardType(card.CardType); err != nil {
		return fmt.Errorf("invalid card type: %v", err)
	}

	// Validate credit limit (if provided)
	if card.CreditLimit < 0 {
		return fmt.Errorf("credit limit cannot be negative")
	}

	// Validate current balance (if provided)
	if card.CurrentBalance < 0 {
		return fmt.Errorf("current balance cannot be negative")
	}

	// Validate dates (payment due date must be after closing date if both provided)
	if !card.PaymentDueDate.IsZero() && !card.ClosingDate.IsZero() {
		if card.PaymentDueDate.Before(card.ClosingDate) {
			return fmt.Errorf("payment due date must be after closing date")
		}
	}

	return nil
}

// validateLastFour validates that last four digits are exactly 4 digits (0-9)
func validateLastFour(lastFour string) error {
	if len(lastFour) != 4 {
		return fmt.Errorf("last four must be exactly 4 characters, got %d", len(lastFour))
	}

	for _, ch := range lastFour {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("last four must contain only digits (0-9), got '%c'", ch)
		}
	}

	return nil
}

// validateBankName validates bank name is non-empty and reasonable length
func validateBankName(bankName string) error {
	if bankName == "" {
		return fmt.Errorf("bank name cannot be empty")
	}

	if len(bankName) > 100 {
		return fmt.Errorf("bank name cannot exceed 100 characters, got %d", len(bankName))
	}

	// Optional: Add more specific validation for bank names
	// This could include a list of known banks or regex patterns

	return nil
}

// validateCardType validates card type is one of the supported types
func validateCardType(cardType string) error {
	if cardType == "" {
		return fmt.Errorf("card type cannot be empty")
	}

	if len(cardType) > 20 {
		return fmt.Errorf("card type cannot exceed 20 characters, got %d", len(cardType))
	}

	// Supported card types
	supportedTypes := map[string]bool{
		"Visa":       true,
		"Mastercard": true,
		"American Express": true,
		"Amex":       true,
		"Discover":   true,
		"Diners Club": true,
		"JCB":        true,
		"UnionPay":   true,
		"Maestro":    true,
	}

	// Check if card type is supported
	if !supportedTypes[cardType] {
		// For flexibility, we'll allow other types but log a warning
		// In production, you might want to be more strict
		fmt.Printf("WARNING: Unsupported card type: %s\n", cardType)
	}

	return nil
}

// Create inserts a new credit card with secure validation
func (r *PostgresCreditCardRepository) Create(ctx context.Context, card *model.CreditCard) error {
	// Validate the credit card data
	if err := validateCreditCard(card); err != nil {
		return fmt.Errorf("credit card validation failed: %v", err)
	}

	// Ensure timestamps are set
	now := time.Now()
	if card.ID == uuid.Nil {
		card.ID = uuid.New()
	}
	card.CreatedAt = now
	card.UpdatedAt = now

	// Prepare SQL query
	query := \`
		INSERT INTO credit_cards (
			id, user_id, circle_id, bank_name, card_type, last_four,
			cardholder_name, credit_limit, current_balance,
			payment_due_date, closing_date, is_active, is_default,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	\`

	// Execute the query
	_, err := r.db.ExecContext(ctx, query,
		card.ID,
		card.UserID,
		card.CircleID,
		card.BankName,
		card.CardType,
		card.LastFour,
		card.CardholderName,
		card.CreditLimit,
		card.CurrentBalance,
		card.PaymentDueDate,
		card.ClosingDate,
		card.IsActive,
		card.IsDefault,
		card.CreatedAt,
		card.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create credit card: %v", err)
	}

	return nil
}

// FindByID retrieves a credit card by its ID
func (r *PostgresCreditCardRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.CreditCard, error) {
	query := \`
		SELECT 
			id, user_id, circle_id, bank_name, card_type, last_four,
			cardholder_name, credit_limit, current_balance,
			payment_due_date, closing_date, is_active, is_default,
			created_at, updated_at, deleted_at
		FROM credit_cards 
		WHERE id = $1 AND deleted_at IS NULL
	\`

	var card model.CreditCard
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&card.ID,
		&card.UserID,
		&card.CircleID,
		&card.BankName,
		&card.CardType,
		&card.LastFour,
		&card.CardholderName,
		&card.CreditLimit,
		&card.CurrentBalance,
		&card.PaymentDueDate,
		&card.ClosingDate,
		&card.IsActive,
		&card.IsDefault,
		&card.CreatedAt,
		&card.UpdatedAt,
		&card.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to find credit card by ID: %v", err)
	}

	return &card, nil
}

// FindByUser retrieves all credit cards for a user
func (r *PostgresCreditCardRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]model.CreditCard, error) {
	query := \`
		SELECT 
			id, user_id, circle_id, bank_name, card_type, last_four,
			cardholder_name, credit_limit, current_balance,
			payment_due_date, closing_date, is_active, is_default,
			created_at, updated_at, deleted_at
		FROM credit_cards 
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY is_default DESC, bank_name, card_type
	\`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query credit cards by user: %v", err)
	}
	defer rows.Close()

	var cards []model.CreditCard
	for rows.Next() {
		var card model.CreditCard
		err := rows.Scan(
			&card.ID,
			&card.UserID,
			&card.CircleID,
			&card.BankName,
			&card.CardType,
			&card.LastFour,
			&card.CardholderName,
			&card.CreditLimit,
			&card.CurrentBalance,
			&card.PaymentDueDate,
			&card.ClosingDate,
			&card.IsActive,
			&card.IsDefault,
			&card.CreatedAt,
			&card.UpdatedAt,
			&card.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credit card row: %v", err)
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credit card rows: %v", err)
	}

	return cards, nil
}

// FindByCircle retrieves all credit cards for a circle
func (r *PostgresCreditCardRepository) FindByCircle(ctx context.Context, circleID uuid.UUID) ([]model.CreditCard, error) {
	query := \`
		SELECT 
			id, user_id, circle_id, bank_name, card_type, last_four,
			cardholder_name, credit_limit, current_balance,
			payment_due_date, closing_date, is_active, is_default,
			created_at, updated_at, deleted_at
		FROM credit_cards 
		WHERE circle_id = $1 AND deleted_at IS NULL
		ORDER BY bank_name, card_type
	\`

	rows, err := r.db.QueryContext(ctx, query, circleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query credit cards by circle: %v", err)
	}
	defer rows.Close()

	var cards []model.CreditCard
	for rows.Next() {
		var card model.CreditCard
		err := rows.Scan(
			&card.ID,
			&card.UserID,
			&card.CircleID,
			&card.BankName,
			&card.CardType,
			&card.LastFour,
			&card.CardholderName,
			&card.CreditLimit,
			&card.CurrentBalance,
			&card.PaymentDueDate,
			&card.ClosingDate,
			&card.IsActive,
			&card.IsDefault,
			&card.CreatedAt,
			&card.UpdatedAt,
			&card.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credit card row: %v", err)
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credit card rows: %v", err)
	}

	return cards, nil
}


// PostgresExchangeRateRepository implements ExchangeRateRepository using PostgreSQL
type PostgresExchangeRateRepository struct {
	db *sql.DB
}


// Create inserts a new exchange rate into the database
func (r *PostgresExchangeRateRepository) Create(ctx context.Context, rate *model.ExchangeRate) error {
	// Validate currency codes
	if err := validateCurrencyCode(rate.BaseCurrency); err != nil {
		return fmt.Errorf("invalid base currency: %w", err)
	}
	if err := validateCurrencyCode(rate.TargetCurrency); err != nil {
		return fmt.Errorf("invalid target currency: %w", err)
	}
	if rate.BaseCurrency == rate.TargetCurrency {
		return fmt.Errorf("base and target currencies cannot be the same")
	}
	if rate.Rate <= 0 {
		return fmt.Errorf("exchange rate must be positive")
	}
	if rate.Date.IsZero() {
		rate.Date = time.Now().UTC()
	}
	if rate.Source == "" {
		rate.Source = "system"
	}

	// Set timestamps
	now := time.Now().UTC()
	rate.CreatedAt = now

	query := `
		INSERT INTO exchange_rates (
			id, base_currency, target_currency, rate, date, source, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		rate.ID,
		rate.BaseCurrency,
		rate.TargetCurrency,
		rate.Rate,
		rate.Date,
		rate.Source,
		rate.CreatedAt,
	)
	return err
}

// FindByID retrieves an exchange rate by its ID
func (r *PostgresExchangeRateRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ExchangeRate, error) {
	query := `
		SELECT id, base_currency, target_currency, rate, date, source, created_at
		FROM exchange_rates
		WHERE id = $1
	`

	var rate model.ExchangeRate
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rate.ID,
		&rate.BaseCurrency,
		&rate.TargetCurrency,
		&rate.Rate,
		&rate.Date,
		&rate.Source,
		&rate.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

// FindLatest retrieves the latest exchange rate for a currency pair
func (r *PostgresExchangeRateRepository) FindLatest(ctx context.Context, baseCurrency, targetCurrency string) (*model.ExchangeRate, error) {
	if err := validateCurrencyCode(baseCurrency); err != nil {
		return nil, fmt.Errorf("invalid base currency: %w", err)
	}
	if err := validateCurrencyCode(targetCurrency); err != nil {
		return nil, fmt.Errorf("invalid target currency: %w", err)
	}

	query := `
		SELECT id, base_currency, target_currency, rate, date, source, created_at
		FROM exchange_rates
		WHERE base_currency = $1 AND target_currency = $2
		ORDER BY date DESC, created_at DESC
		LIMIT 1
	`

	var rate model.ExchangeRate
	err := r.db.QueryRowContext(ctx, query, baseCurrency, targetCurrency).Scan(
		&rate.ID,
		&rate.BaseCurrency,
		&rate.TargetCurrency,
		&rate.Rate,
		&rate.Date,
		&rate.Source,
		&rate.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

// FindByDate retrieves an exchange rate for a specific date
func (r *PostgresExchangeRateRepository) FindByDate(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (*model.ExchangeRate, error) {
	if err := validateCurrencyCode(baseCurrency); err != nil {
		return nil, fmt.Errorf("invalid base currency: %w", err)
	}
	if err := validateCurrencyCode(targetCurrency); err != nil {
		return nil, fmt.Errorf("invalid target currency: %w", err)
	}
	if date.IsZero() {
		return nil, fmt.Errorf("date cannot be zero")
	}

	// Normalize date to midnight UTC for comparison
	dateMidnight := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	query := `
		SELECT id, base_currency, target_currency, rate, date, source, created_at
		FROM exchange_rates
		WHERE base_currency = $1 AND target_currency = $2 AND date = $3
		LIMIT 1
	`

	var rate model.ExchangeRate
	err := r.db.QueryRowContext(ctx, query, baseCurrency, targetCurrency, dateMidnight).Scan(
		&rate.ID,
		&rate.BaseCurrency,
		&rate.TargetCurrency,
		&rate.Rate,
		&rate.Date,
		&rate.Source,
		&rate.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

// FindByDateRange retrieves exchange rates for a currency pair within a date range
func (r *PostgresExchangeRateRepository) FindByDateRange(ctx context.Context, baseCurrency, targetCurrency string, startDate, endDate time.Time) ([]model.ExchangeRate, error) {
	if err := validateCurrencyCode(baseCurrency); err != nil {
		return nil, fmt.Errorf("invalid base currency: %w", err)
	}
	if err := validateCurrencyCode(targetCurrency); err != nil {
		return nil, fmt.Errorf("invalid target currency: %w", err)
	}
	if startDate.IsZero() || endDate.IsZero() {
		return nil, fmt.Errorf("start and end dates cannot be zero")
	}
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	// Normalize dates to midnight UTC
	startMidnight := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endMidnight := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

	query := `
		SELECT id, base_currency, target_currency, rate, date, source, created_at
		FROM exchange_rates
		WHERE base_currency = $1 AND target_currency = $2 
			AND date >= $3 AND date <= $4
		ORDER BY date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, baseCurrency, targetCurrency, startMidnight, endMidnight)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []model.ExchangeRate
	for rows.Next() {
		var rate model.ExchangeRate
		err := rows.Scan(
			&rate.ID,
			&rate.BaseCurrency,
			&rate.TargetCurrency,
			&rate.Rate,
			&rate.Date,
			&rate.Source,
			&rate.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rates, nil
}

// FindAllLatest retrieves the latest exchange rate for all currency pairs
func (r *PostgresExchangeRateRepository) FindAllLatest(ctx context.Context) ([]model.ExchangeRate, error) {
	query := `
		SELECT DISTINCT ON (base_currency, target_currency)
			id, base_currency, target_currency, rate, date, source, created_at
		FROM exchange_rates
		ORDER BY base_currency, target_currency, date DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []model.ExchangeRate
	for rows.Next() {
		var rate model.ExchangeRate
		err := rows.Scan(
			&rate.ID,
			&rate.BaseCurrency,
			&rate.TargetCurrency,
			&rate.Rate,
			&rate.Date,
			&rate.Source,
			&rate.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rates, nil
}

// Update updates an existing exchange rate
func (r *PostgresExchangeRateRepository) Update(ctx context.Context, rate *model.ExchangeRate) error {
	// Validate currency codes
	if err := validateCurrencyCode(rate.BaseCurrency); err != nil {
		return fmt.Errorf("invalid base currency: %w", err)
	}
	if err := validateCurrencyCode(rate.TargetCurrency); err != nil {
		return fmt.Errorf("invalid target currency: %w", err)
	}
	if rate.BaseCurrency == rate.TargetCurrency {
		return fmt.Errorf("base and target currencies cannot be the same")
	}
	if rate.Rate <= 0 {
		return fmt.Errorf("exchange rate must be positive")
	}
	if rate.Date.IsZero() {
		return fmt.Errorf("date cannot be zero")
	}

	query := `
		UPDATE exchange_rates
		SET base_currency = $2, target_currency = $3, rate = $4, date = $5, source = $6
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		rate.ID,
		rate.BaseCurrency,
		rate.TargetCurrency,
		rate.Rate,
		rate.Date,
		rate.Source,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete removes an exchange rate by ID
func (r *PostgresExchangeRateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM exchange_rates WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ConvertAmount converts an amount from one currency to another using the latest rate
func (r *PostgresExchangeRateRepository) ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string, date time.Time) (float64, error) {
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be positive")
	}
	if err := validateCurrencyCode(fromCurrency); err != nil {
		return 0, fmt.Errorf("invalid from currency: %w", err)
	}
	if err := validateCurrencyCode(toCurrency); err != nil {
		return 0, fmt.Errorf("invalid to currency: %w", err)
	}

	// If currencies are the same, no conversion needed
	if fromCurrency == toCurrency {
		return amount, nil
	}

	var rate *model.ExchangeRate
	var err error

	if date.IsZero() {
		// Use latest rate
		rate, err = r.FindLatest(ctx, fromCurrency, toCurrency)
	} else {
		// Use rate for specific date
		rate, err = r.FindByDate(ctx, fromCurrency, toCurrency, date)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			// Try inverse rate
			inverseRate, inverseErr := r.FindLatest(ctx, toCurrency, fromCurrency)
			if inverseErr == nil {
				// Convert using inverse rate: amount = amount * (1 / inverseRate.Rate)
				return amount * (1 / inverseRate.Rate), nil
			}
			return 0, fmt.Errorf("no exchange rate found for %s to %s", fromCurrency, toCurrency)
		}
		return 0, err
	}

	return amount * rate.Rate, nil
}

// GetSupportedCurrencies returns a list of all unique currencies in the system
func (r *PostgresExchangeRateRepository) GetSupportedCurrencies(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT base_currency FROM exchange_rates
		UNION
		SELECT DISTINCT target_currency FROM exchange_rates
		ORDER BY 1
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []string
	for rows.Next() {
		var currency string
		err := rows.Scan(&currency)
		if err != nil {
			return nil, err
		}
		currencies = append(currencies, currency)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return currencies, nil
}

// validateCurrencyCode validates a 3-letter currency code
func validateCurrencyCode(currency string) error {
	if len(currency) != 3 {
		return fmt.Errorf("currency code must be 3 characters")
	}
	// Check if all characters are uppercase letters
	for _, c := range currency {
		if c < 'A' || c > 'Z' {
			return fmt.Errorf("currency code must contain only uppercase letters")
		}
	}
	return nil
}


// PostgresMobileNotificationRepository implements MobileNotificationRepository using PostgreSQL
type PostgresMobileNotificationRepository struct {
	db *sql.DB
}

// PostgresAnomalyDetectionRepository implements AnomalyDetectionRepository using PostgreSQL
type PostgresAnomalyDetectionRepository struct {
	db *sql.DB
}


// CreateRule creates a new anomaly detection rule
func (r *PostgresAnomalyDetectionRepository) CreateRule(ctx context.Context, rule *model.AnomalyDetectionRule) error {
	// Validate rule
	if err := validateAnomalyRule(rule); err != nil {
		return fmt.Errorf("invalid anomaly rule: %w", err)
	}

	// Set timestamps
	now := time.Now().UTC()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	rule.UpdatedAt = now

	query := `
		INSERT INTO anomaly_detection_rules (
			id, circle_id, name, description, rule_type, condition, severity,
			amount_threshold, days_threshold, percentage_diff, is_active,
			last_triggered_at, trigger_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.CircleID,
		rule.Name,
		rule.Description,
		rule.RuleType,
		rule.Condition,
		rule.Severity,
		rule.AmountThreshold,
		rule.DaysThreshold,
		rule.PercentageDiff,
		rule.IsActive,
		rule.LastTriggeredAt,
		rule.TriggerCount,
		rule.CreatedAt,
		rule.UpdatedAt,
	)
	return err
}

// FindRuleByID retrieves an anomaly detection rule by ID
func (r *PostgresAnomalyDetectionRepository) FindRuleByID(ctx context.Context, id uuid.UUID) (*model.AnomalyDetectionRule, error) {
	query := `
		SELECT id, circle_id, name, description, rule_type, condition, severity,
			amount_threshold, days_threshold, percentage_diff, is_active,
			last_triggered_at, trigger_count, created_at, updated_at
		FROM anomaly_detection_rules
		WHERE id = $1
	`

	var rule model.AnomalyDetectionRule
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID,
		&rule.CircleID,
		&rule.Name,
		&rule.Description,
		&rule.RuleType,
		&rule.Condition,
		&rule.Severity,
		&rule.AmountThreshold,
		&rule.DaysThreshold,
		&rule.PercentageDiff,
		&rule.IsActive,
		&rule.LastTriggeredAt,
		&rule.TriggerCount,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// FindRulesByCircle retrieves all anomaly detection rules for a circle
func (r *PostgresAnomalyDetectionRepository) FindRulesByCircle(ctx context.Context, circleID uuid.UUID) ([]model.AnomalyDetectionRule, error) {
	query := `
		SELECT id, circle_id, name, description, rule_type, condition, severity,
			amount_threshold, days_threshold, percentage_diff, is_active,
			last_triggered_at, trigger_count, created_at, updated_at
		FROM anomaly_detection_rules
		WHERE circle_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, circleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []model.AnomalyDetectionRule
	for rows.Next() {
		var rule model.AnomalyDetectionRule
		err := rows.Scan(
			&rule.ID,
			&rule.CircleID,
			&rule.Name,
			&rule.Description,
			&rule.RuleType,
			&rule.Condition,
			&rule.Severity,
			&rule.AmountThreshold,
			&rule.DaysThreshold,
			&rule.PercentageDiff,
			&rule.IsActive,
			&rule.LastTriggeredAt,
			&rule.TriggerCount,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

// FindActiveRules retrieves all active anomaly detection rules for a circle
func (r *PostgresAnomalyDetectionRepository) FindActiveRules(ctx context.Context, circleID uuid.UUID) ([]model.AnomalyDetectionRule, error) {
	query := `
		SELECT id, circle_id, name, description, rule_type, condition, severity,
			amount_threshold, days_threshold, percentage_diff, is_active,
			last_triggered_at, trigger_count, created_at, updated_at
		FROM anomaly_detection_rules
		WHERE circle_id = $1 AND is_active = true
		ORDER BY severity DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, circleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []model.AnomalyDetectionRule
	for rows.Next() {
		var rule model.AnomalyDetectionRule
		err := rows.Scan(
			&rule.ID,
			&rule.CircleID,
			&rule.Name,
			&rule.Description,
			&rule.RuleType,
			&rule.Condition,
			&rule.Severity,
			&rule.AmountThreshold,
			&rule.DaysThreshold,
			&rule.PercentageDiff,
			&rule.IsActive,
			&rule.LastTriggeredAt,
			&rule.TriggerCount,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

// UpdateRule updates an existing anomaly detection rule
func (r *PostgresAnomalyDetectionRepository) UpdateRule(ctx context.Context, rule *model.AnomalyDetectionRule) error {
	// Validate rule
	if err := validateAnomalyRule(rule); err != nil {
		return fmt.Errorf("invalid anomaly rule: %w", err)
	}

	// Update timestamp
	rule.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE anomaly_detection_rules
		SET circle_id = $2, name = $3, description = $4, rule_type = $5, condition = $6,
			severity = $7, amount_threshold = $8, days_threshold = $9, percentage_diff = $10,
			is_active = $11, last_triggered_at = $12, trigger_count = $13, updated_at = $14
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.CircleID,
		rule.Name,
		rule.Description,
		rule.RuleType,
		rule.Condition,
		rule.Severity,
		rule.AmountThreshold,
		rule.DaysThreshold,
		rule.PercentageDiff,
		rule.IsActive,
		rule.LastTriggeredAt,
		rule.TriggerCount,
		rule.UpdatedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteRule removes an anomaly detection rule by ID
func (r *PostgresAnomalyDetectionRepository) DeleteRule(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM anomaly_detection_rules WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ActivateRule activates an anomaly detection rule
func (r *PostgresAnomalyDetectionRepository) ActivateRule(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE anomaly_detection_rules
		SET is_active = true, updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now().UTC())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeactivateRule deactivates an anomaly detection rule
func (r *PostgresAnomalyDetectionRepository) DeactivateRule(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE anomaly_detection_rules
		SET is_active = false, updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now().UTC())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// IncrementTriggerCount increments the trigger count and updates last triggered timestamp
func (r *PostgresAnomalyDetectionRepository) IncrementTriggerCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE anomaly_detection_rules
		SET trigger_count = trigger_count + 1, last_triggered_at = $2, updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now().UTC())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// validateAnomalyRule validates an anomaly detection rule
func validateAnomalyRule(rule *model.AnomalyDetectionRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}
	if rule.RuleType == "" {
		return fmt.Errorf("rule type cannot be empty")
	}
	// Validate rule type
	validRuleTypes := map[string]bool{
		"duplicate": true,
		"amount_mismatch": true,
		"orphan": true,
		"timeframe": true,
	}
	if !validRuleTypes[rule.RuleType] {
		return fmt.Errorf("invalid rule type: %s", rule.RuleType)
	}
	// Validate severity
	validSeverities := map[string]bool{
		"warning": true,
		"critical": true,
	}
	if rule.Severity != "" && !validSeverities[rule.Severity] {
		return fmt.Errorf("invalid severity: %s", rule.Severity)
	}
	// Validate thresholds
	if rule.AmountThreshold < 0 {
		return fmt.Errorf("amount threshold cannot be negative")
	}
	if rule.DaysThreshold < 0 {
		return fmt.Errorf("days threshold cannot be negative")
	}
	if rule.PercentageDiff < 0 || rule.PercentageDiff > 100 {
		return fmt.Errorf("percentage difference must be between 0 and 100")
	}
	return nil
}


// PostgresNotificationAppWhitelistRepository implements NotificationAppWhitelistRepository using PostgreSQL
type PostgresNotificationAppWhitelistRepository struct {
	db *sql.DB
}


// Create adds a new app to the whitelist
func (r *PostgresNotificationAppWhitelistRepository) Create(ctx context.Context, app *model.NotificationAppWhitelist) error {
	// Validate app
	if err := validateNotificationApp(app); err != nil {
		return fmt.Errorf("invalid notification app: %w", err)
	}

	// Set timestamps
	now := time.Now().UTC()
	if app.CreatedAt.IsZero() {
		app.CreatedAt = now
	}
	app.UpdatedAt = now

	query := `
		INSERT INTO notification_app_whitelist (
			id, user_id, app_package, app_name, is_enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		app.ID,
		app.UserID,
		app.AppPackage,
		app.AppName,
		app.IsEnabled,
		app.CreatedAt,
		app.UpdatedAt,
	)
	return err
}

// FindByID retrieves a whitelist entry by ID
func (r *PostgresNotificationAppWhitelistRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.NotificationAppWhitelist, error) {
	query := `
		SELECT id, user_id, app_package, app_name, is_enabled, created_at, updated_at
		FROM notification_app_whitelist
		WHERE id = $1
	`

	var app model.NotificationAppWhitelist
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&app.ID,
		&app.UserID,
		&app.AppPackage,
		&app.AppName,
		&app.IsEnabled,
		&app.CreatedAt,
		&app.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// FindByUser retrieves all whitelisted apps for a user
func (r *PostgresNotificationAppWhitelistRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]model.NotificationAppWhitelist, error) {
	query := `
		SELECT id, user_id, app_package, app_name, is_enabled, created_at, updated_at
		FROM notification_app_whitelist
		WHERE user_id = $1
		ORDER BY app_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.NotificationAppWhitelist
	for rows.Next() {
		var app model.NotificationAppWhitelist
		err := rows.Scan(
			&app.ID,
			&app.UserID,
			&app.AppPackage,
			&app.AppName,
			&app.IsEnabled,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return apps, nil
}

// FindByCircle retrieves all whitelisted apps for users in a circle
func (r *PostgresNotificationAppWhitelistRepository) FindByCircle(ctx context.Context, circleID uuid.UUID) ([]model.NotificationAppWhitelist, error) {
	query := `
		SELECT naw.id, naw.user_id, naw.app_package, naw.app_name, naw.is_enabled, naw.created_at, naw.updated_at
		FROM notification_app_whitelist naw
		INNER JOIN user_circles uc ON naw.user_id = uc.user_id
		WHERE uc.circle_id = $1
		ORDER BY naw.app_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, circleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.NotificationAppWhitelist
	for rows.Next() {
		var app model.NotificationAppWhitelist
		err := rows.Scan(
			&app.ID,
			&app.UserID,
			&app.AppPackage,
			&app.AppName,
			&app.IsEnabled,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return apps, nil
}

// FindByAppPackage retrieves all whitelist entries for a specific app package
func (r *PostgresNotificationAppWhitelistRepository) FindByAppPackage(ctx context.Context, appPackage string) ([]model.NotificationAppWhitelist, error) {
	query := `
		SELECT id, user_id, app_package, app_name, is_enabled, created_at, updated_at
		FROM notification_app_whitelist
		WHERE app_package = $1
		ORDER BY user_id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, appPackage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.NotificationAppWhitelist
	for rows.Next() {
		var app model.NotificationAppWhitelist
		err := rows.Scan(
			&app.ID,
			&app.UserID,
			&app.AppPackage,
			&app.AppName,
			&app.IsEnabled,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return apps, nil
}

// FindByUserAndApp retrieves a specific whitelist entry for a user and app package
func (r *PostgresNotificationAppWhitelistRepository) FindByUserAndApp(ctx context.Context, userID uuid.UUID, appPackage string) (*model.NotificationAppWhitelist, error) {
	query := `
		SELECT id, user_id, app_package, app_name, is_enabled, created_at, updated_at
		FROM notification_app_whitelist
		WHERE user_id = $1 AND app_package = $2
	`

	var app model.NotificationAppWhitelist
	err := r.db.QueryRowContext(ctx, query, userID, appPackage).Scan(
		&app.ID,
		&app.UserID,
		&app.AppPackage,
		&app.AppName,
		&app.IsEnabled,
		&app.CreatedAt,
		&app.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// Update modifies an existing whitelist entry
func (r *PostgresNotificationAppWhitelistRepository) Update(ctx context.Context, app *model.NotificationAppWhitelist) error {
	// Validate app
	if err := validateNotificationApp(app); err != nil {
		return fmt.Errorf("invalid notification app: %w", err)
	}

	// Update timestamp
	app.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE notification_app_whitelist
		SET app_name = $2, is_enabled = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		app.ID,
		app.AppName,
		app.IsEnabled,
		app.UpdatedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete removes a whitelist entry by ID
func (r *PostgresNotificationAppWhitelistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notification_app_whitelist WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// IsWhitelisted checks if an app is whitelisted for a user
func (r *PostgresNotificationAppWhitelistRepository) IsWhitelisted(ctx context.Context, userID uuid.UUID, appPackage string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM notification_app_whitelist
			WHERE user_id = $1 AND app_package = $2 AND is_enabled = true
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, appPackage).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// AutoWhitelistApp automatically adds an app to the whitelist if it doesn't exist
func (r *PostgresNotificationAppWhitelistRepository) AutoWhitelistApp(ctx context.Context, userID uuid.UUID, appPackage, appName string) error {
	// Check if already exists
	existing, err := r.FindByUserAndApp(ctx, userID, appPackage)
	if err == nil && existing != nil {
		// Already exists, ensure it's enabled
		if !existing.IsEnabled {
			existing.IsEnabled = true
			existing.UpdatedAt = time.Now().UTC()
			return r.Update(ctx, existing)
		}
		return nil // Already whitelisted and enabled
	}

	// Doesn't exist or error occurred (other than not found)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Create new entry
	app := &model.NotificationAppWhitelist{
		ID:         uuid.New(),
		UserID:     userID,
		AppPackage: appPackage,
		AppName:    appName,
		IsEnabled:  true,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	return r.Create(ctx, app)
}

// GetWhitelistStats retrieves statistics about a user's whitelist
func (r *PostgresNotificationAppWhitelistRepository) GetWhitelistStats(ctx context.Context, userID uuid.UUID) (*repository.WhitelistStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_apps,
			COUNT(CASE WHEN is_enabled = true THEN 1 END) as enabled_apps,
			COUNT(CASE WHEN is_enabled = false THEN 1 END) as disabled_apps
		FROM notification_app_whitelist
		WHERE user_id = $1
	`

	var stats repository.WhitelistStats
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalApps,
		&stats.EnabledApps,
		&stats.DisabledApps,
	)
	if err != nil {
		return nil, err
	}

	// Get most recent app
	recentQuery := `
		SELECT app_name, created_at
		FROM notification_app_whitelist
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var recentApp sql.NullString
	var recentTime sql.NullTime
	err = r.db.QueryRowContext(ctx, recentQuery, userID).Scan(&recentApp, &recentTime)
	if err == nil && recentApp.Valid {
		stats.MostRecentApp = recentApp.String
		stats.MostRecentAt = &recentTime.Time
	}

	return &stats, nil
}

// validateNotificationApp validates a notification app whitelist entry
func validateNotificationApp(app *model.NotificationAppWhitelist) error {
	if app.AppPackage == "" {
		return fmt.Errorf("app package cannot be empty")
	}
	// Validate Android package format (com.example.app)
	if !isValidAndroidPackage(app.AppPackage) {
		return fmt.Errorf("invalid Android package format: %s", app.AppPackage)
	}
	if app.AppName == "" {
		return fmt.Errorf("app name cannot be empty")
	}
	return nil
}

// isValidAndroidPackage checks if a string is a valid Android package name
func isValidAndroidPackage(pkg string) bool {
	// Android package names: com.example.app, org.example.app, etc.
	// Must contain at least one dot, start with letter, contain only [a-z0-9_.]
	if len(pkg) < 3 {
		return false
	}
	// Must start with letter
	if !('a' <= pkg[0] && pkg[0] <= 'z') && !('A' <= pkg[0] && pkg[0] <= 'Z') {
		return false
	}
	// Must contain at least one dot
	if !strings.Contains(pkg, ".") {
		return false
	}
	// Only allowed characters
	for _, ch := range pkg {
		if !(('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ('0' <= ch && ch <= '9') || ch == '.' || ch == '_') {
			return false
		}
	}
	return true
}



func (r *PostgresAuthRepository) CreateAuth(ctx context.Context, auth *model.UserAuth) error {
	return nil
}

func (r *PostgresAuthRepository) FindAuthByUserID(ctx context.Context, userID string) (*model.UserAuth, error) {
	return nil, nil
}

func (r *PostgresAuthRepository) UpdateAuth(ctx context.Context, auth *model.UserAuth) error {
	return nil
}

func (r *PostgresAuthRepository) IncrementFailedAttempts(ctx context.Context, userID string) error {
	return nil
}

func (r *PostgresAuthRepository) ResetFailedAttempts(ctx context.Context, userID string) error {
	return nil
}

func (r *PostgresAuthRepository) LockAccount(ctx context.Context, userID string, until time.Time) error {
	return nil
}

func (r *PostgresAuthRepository) UnlockAccount(ctx context.Context, userID string) error {
	return nil
}

func (r *PostgresAuthRepository) Close() error {
	return nil
}