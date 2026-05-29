package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/google/uuid"
)

// TransactionServiceImpl implements TransactionService
type TransactionServiceImpl struct {
	transactionRepo repository.TransactionRepository
}

// NewTransactionService creates a new TransactionService
func NewTransactionService(transactionRepo repository.TransactionRepository) TransactionService {
	return &TransactionServiceImpl{
		transactionRepo: transactionRepo,
	}
}

// CreateTransaction creates a new transaction with validation
func (s *TransactionServiceImpl) CreateTransaction(ctx context.Context, transaction *model.Transaction) error {
	// Validate transaction
	if err := s.ValidateTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}

	// Set timestamps if not set
	now := time.Now().UTC()
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = now
	}
	transaction.UpdatedAt = now

	// Generate ID if not set
	if transaction.ID == uuid.Nil {
		transaction.ID = uuid.New()
	}

	// Check for duplicates before creating
	duplicates, err := s.DetectDuplicateTransactions(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("potential duplicate transaction found: %d similar transactions", len(duplicates))
	}

	// Create transaction
	return s.transactionRepo.Create(ctx, transaction)
}

// GetTransaction retrieves a transaction by ID
func (s *TransactionServiceImpl) GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid transaction ID")
	}

	return s.transactionRepo.FindByID(ctx, id)
}

// UpdateTransaction updates an existing transaction
func (s *TransactionServiceImpl) UpdateTransaction(ctx context.Context, transaction *model.Transaction) error {
	if transaction.ID == uuid.Nil {
		return fmt.Errorf("transaction ID is required")
	}

	// Validate transaction
	if err := s.ValidateTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}

	// Update timestamp
	transaction.UpdatedAt = time.Now().UTC()

	return s.transactionRepo.Update(ctx, transaction)
}

// DeleteTransaction permanently deletes a transaction
func (s *TransactionServiceImpl) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid transaction ID")
	}

	return s.transactionRepo.Delete(ctx, id)
}

// SoftDeleteTransaction marks a transaction as deleted
func (s *TransactionServiceImpl) SoftDeleteTransaction(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid transaction ID")
	}

	return s.transactionRepo.SoftDelete(ctx, id)
}

// GetTransactionsByFamily retrieves transactions for a family with filters
func (s *TransactionServiceImpl) GetTransactionsByFamily(ctx context.Context, familyID uuid.UUID, filters repository.TransactionFilters) ([]model.Transaction, error) {
	if familyID == uuid.Nil {
		return nil, fmt.Errorf("invalid family ID")
	}

	// Apply default limit if not set
	if filters.Limit <= 0 {
		filters.Limit = 50
	}

	// Validate date range if provided
	if filters.StartDate != nil && filters.EndDate != nil && filters.EndDate.Before(*filters.StartDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	return s.transactionRepo.FindByFamily(ctx, familyID, filters)
}

// GetTransactionsByUser retrieves transactions for a user with filters
func (s *TransactionServiceImpl) GetTransactionsByUser(ctx context.Context, userID uuid.UUID, filters repository.TransactionFilters) ([]model.Transaction, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Apply default limit if not set
	if filters.Limit <= 0 {
		filters.Limit = 50
	}

	// Validate date range if provided
	if filters.StartDate != nil && filters.EndDate != nil && filters.EndDate.Before(*filters.StartDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	return s.transactionRepo.FindByUser(ctx, userID, filters)
}

// GetTransactionsByDateRange retrieves transactions within a date range
func (s *TransactionServiceImpl) GetTransactionsByDateRange(ctx context.Context, familyID uuid.UUID, start, end time.Time) ([]model.Transaction, error) {
	if familyID == uuid.Nil {
		return nil, fmt.Errorf("invalid family ID")
	}

	if end.Before(start) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	// Limit date range to 1 year max
	maxRange := 365 * 24 * time.Hour
	if end.Sub(start) > maxRange {
		return nil, fmt.Errorf("date range cannot exceed 1 year")
	}

	return s.transactionRepo.FindByDateRange(ctx, familyID, start, end)
}

// GetTransactionsByCategory retrieves transactions by category within date range
func (s *TransactionServiceImpl) GetTransactionsByCategory(ctx context.Context, familyID uuid.UUID, category string, start, end time.Time) ([]model.Transaction, error) {
	if familyID == uuid.Nil {
		return nil, fmt.Errorf("invalid family ID")
	}

	if category == "" {
		return nil, fmt.Errorf("category is required")
	}

	if end.Before(start) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	return s.transactionRepo.FindByCategory(ctx, familyID, category, start, end)
}

// GetTransactionsByPaymentMethod retrieves transactions by payment method within date range
func (s *TransactionServiceImpl) GetTransactionsByPaymentMethod(ctx context.Context, familyID uuid.UUID, method model.PaymentMethod, start, end time.Time) ([]model.Transaction, error) {
	if familyID == uuid.Nil {
		return nil, fmt.Errorf("invalid family ID")
	}

	if method == "" {
		return nil, fmt.Errorf("payment method is required")
	}

	if end.Before(start) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	return s.transactionRepo.FindByPaymentMethod(ctx, familyID, method, start, end)
}

// GetUncategorizedTransactions retrieves uncategorized transactions
func (s *TransactionServiceImpl) GetUncategorizedTransactions(ctx context.Context, familyID uuid.UUID, limit int) ([]model.Transaction, error) {
	if familyID == uuid.Nil {
		return nil, fmt.Errorf("invalid family ID")
	}

	if limit <= 0 {
		limit = 100
	}

	return s.transactionRepo.FindUncategorized(ctx, familyID, limit)
}

// GetUnverifiedTransactions retrieves unverified transactions
func (s *TransactionServiceImpl) GetUnverifiedTransactions(ctx context.Context, familyID uuid.UUID, limit int) ([]model.Transaction, error) {
	if familyID == uuid.Nil {
		return nil, fmt.Errorf("invalid family ID")
	}

	if limit <= 0 {
		limit = 100
	}

	return s.transactionRepo.FindUnverified(ctx, familyID, limit)
}

// GetMonthlySummary retrieves monthly summary for a family
func (s *TransactionServiceImpl) GetMonthlySummary(ctx context.Context, familyID uuid.UUID, year int, month time.Month) (*model.ReportSummary, error) {
	if familyID == uuid.Nil {
		return nil, fmt.Errorf("invalid family ID")
	}

	if year < 2000 || year > 2100 {
		return nil, fmt.Errorf("invalid year")
	}

	return s.transactionRepo.GetMonthlySummary(ctx, familyID, year, month)
}

// GetCategorySummary retrieves category summary for a date range
func (s *TransactionServiceImpl) GetCategorySummary(ctx context.Context, familyID uuid.UUID, start, end time.Time) ([]repository.CategorySummary, error) {
	if familyID == uuid.Nil {
		return nil, fmt.Errorf("invalid family ID")
	}

	if end.Before(start) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	// Limit date range to 1 year max
	maxRange := 365 * 24 * time.Hour
	if end.Sub(start) > maxRange {
		return nil, fmt.Errorf("date range cannot exceed 1 year")
	}

	return s.transactionRepo.GetCategorySummary(ctx, familyID, start, end)
}

// GetSpendingTrends retrieves spending trends for a family
func (s *TransactionServiceImpl) GetSpendingTrends(ctx context.Context, familyID uuid.UUID, period string, months int) ([]repository.TrendData, error) {
	if familyID == uuid.Nil {
		return nil, fmt.Errorf("invalid family ID")
	}

	// Validate period
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
		"yearly":  true,
	}
	if !validPeriods[period] {
		return nil, fmt.Errorf("invalid period. Must be: daily, weekly, monthly, yearly")
	}

	// Validate months
	if months <= 0 || months > 36 {
		return nil, fmt.Errorf("months must be between 1 and 36")
	}

	return s.transactionRepo.GetSpendingTrends(ctx, familyID, period, months)
}

// GetBudgetStatus retrieves budget status
func (s *TransactionServiceImpl) GetBudgetStatus(ctx context.Context, familyID uuid.UUID, budgetID uuid.UUID) (*repository.BudgetStatus, error) {
	if familyID == uuid.Nil {
		return nil, fmt.Errorf("invalid family ID")
	}

	if budgetID == uuid.Nil {
		return nil, fmt.Errorf("invalid budget ID")
	}

	return s.transactionRepo.GetBudgetStatus(ctx, familyID, budgetID)
}

// FindSimilarTransactions finds transactions similar to a statement transaction
func (s *TransactionServiceImpl) FindSimilarTransactions(ctx context.Context, statementTx model.StatementTransaction) ([]model.Transaction, error) {
	if statementTx.Amount <= 0 {
		return nil, fmt.Errorf("invalid statement transaction amount")
	}

	if statementTx.Date.IsZero() {
		return nil, fmt.Errorf("statement transaction date is required")
	}

	return s.transactionRepo.FindSimilarTransactions(ctx, statementTx)
}

// MatchStatementTransaction matches a statement transaction to an existing transaction
func (s *TransactionServiceImpl) MatchStatementTransaction(ctx context.Context, statementTxID, transactionID uuid.UUID, confidence float64) error {
	if statementTxID == uuid.Nil {
		return fmt.Errorf("invalid statement transaction ID")
	}

	if transactionID == uuid.Nil {
		return fmt.Errorf("invalid transaction ID")
	}

	if confidence < 0 || confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}

	return s.transactionRepo.MatchStatementTransaction(ctx, statementTxID, transactionID, confidence)
}

// CreateTransactionsBatch creates multiple transactions in a batch
func (s *TransactionServiceImpl) CreateTransactionsBatch(ctx context.Context, transactions []model.Transaction) error {
	if len(transactions) == 0 {
		return fmt.Errorf("no transactions to create")
	}

	if len(transactions) > 1000 {
		return fmt.Errorf("batch size cannot exceed 1000 transactions")
	}

	// Validate each transaction
	for i := range transactions {
		if err := s.ValidateTransaction(ctx, &transactions[i]); err != nil {
			return fmt.Errorf("transaction validation failed at index %d: %w", i, err)
		}

		// Set timestamps
		now := time.Now().UTC()
		if transactions[i].CreatedAt.IsZero() {
			transactions[i].CreatedAt = now
		}
		transactions[i].UpdatedAt = now

		// Generate ID if not set
		if transactions[i].ID == uuid.Nil {
			transactions[i].ID = uuid.New()
		}
	}

	return s.transactionRepo.CreateBatch(ctx, transactions)
}

// UpdateTransactionsBatch updates multiple transactions in a batch
func (s *TransactionServiceImpl) UpdateTransactionsBatch(ctx context.Context, transactions []model.Transaction) error {
	if len(transactions) == 0 {
		return fmt.Errorf("no transactions to update")
	}

	if len(transactions) > 1000 {
		return fmt.Errorf("batch size cannot exceed 1000 transactions")
	}

	// Validate each transaction
	for i := range transactions {
		if transactions[i].ID == uuid.Nil {
			return fmt.Errorf("transaction ID is required at index %d", i)
		}

		if err := s.ValidateTransaction(ctx, &transactions[i]); err != nil {
			return fmt.Errorf("transaction validation failed at index %d: %w", i, err)
		}

		// Update timestamp
		transactions[i].UpdatedAt = time.Now().UTC()
	}

	return s.transactionRepo.UpdateBatch(ctx, transactions)
}

// ValidateTransaction validates a transaction
func (s *TransactionServiceImpl) ValidateTransaction(ctx context.Context, transaction *model.Transaction) error {
	if transaction == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	// Validate amount
	if transaction.Amount <= 0 {
		return fmt.Errorf("transaction amount must be positive")
	}

	// Validate currency
	if transaction.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if len(transaction.Currency) != 3 {
		return fmt.Errorf("currency must be 3 characters (ISO 4217)")
	}

	// Validate date
	if transaction.Date.IsZero() {
		return fmt.Errorf("transaction date is required")
	}
	if transaction.Date.After(time.Now().Add(24 * time.Hour)) {
		return fmt.Errorf("transaction date cannot be in the future")
	}

	// Validate user and family IDs
	if transaction.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if transaction.CircleID == uuid.Nil {
		return fmt.Errorf("family/circle ID is required")
	}

	// Validate payment method
	if transaction.PaymentMethod == "" {
		return fmt.Errorf("payment method is required")
	}

	// Validate category (if provided)
	if transaction.Category != "" && len(transaction.Category) > 100 {
		return fmt.Errorf("category cannot exceed 100 characters")
	}

	// Validate description (if provided)
	if transaction.Description != "" && len(transaction.Description) > 500 {
		return fmt.Errorf("description cannot exceed 500 characters")
	}

	return nil
}

// CalculateTransactionTotals calculates totals for a family in a date range
func (s *TransactionServiceImpl) CalculateTransactionTotals(ctx context.Context, familyID uuid.UUID, start, end time.Time) (*TransactionTotals, error) {
	if familyID == uuid.Nil {
		return nil, fmt.Errorf("invalid family ID")
	}

	if end.Before(start) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	// Get transactions in date range
	transactions, err := s.transactionRepo.FindByDateRange(ctx, familyID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Calculate totals
	var totals TransactionTotals
	totals.TransactionCount = len(transactions)

	for _, tx := range transactions {
		if tx.Amount > 0 {
			totals.TotalIncome += tx.Amount
		} else {
			totals.TotalExpenses += -tx.Amount // Amount is negative for expenses
		}
	}

	totals.NetBalance = totals.TotalIncome - totals.TotalExpenses

	return &totals, nil
}

// DetectDuplicateTransactions finds potential duplicate transactions
func (s *TransactionServiceImpl) DetectDuplicateTransactions(ctx context.Context, transaction *model.Transaction) ([]model.Transaction, error) {
	if transaction == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}

	// Create a statement transaction for similarity matching
	statementTx := model.StatementTransaction{
		Amount:      transaction.Amount,
		Currency:    transaction.Currency,
		Description: transaction.Description,
		Date:        transaction.Date,
		Merchant:    "", // We don't have merchant in transaction model
	}

	return s.transactionRepo.FindSimilarTransactions(ctx, statementTx)
}