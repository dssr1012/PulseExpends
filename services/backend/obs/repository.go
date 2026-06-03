package obs

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/dssr1012/pulse-expends/internal/config"
	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/google/uuid"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/rs/zerolog"
)

// OBSRepository implements the repository interfaces using Huawei OBS
type OBSRepository struct {
	client     *obs.ObsClient
	bucketName string
	encryptionKey string
	logger     zerolog.Logger
}

// NewRepository creates a new OBS repository instance
func NewRepository(cfg config.OBSConfig) (*OBSRepository, error) {
	// Initialize OBS client
	client, err := obs.New(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.endpoint)
	if err != nil {
		return nil, fmt.errorf("failed to create OBS client: %w", err)
	}

	// Create bucket if it doesn't exist
	_, err = client.HeadBucket(cfg.BucketName)
	if err != nil {
		// Bucket doesn't exist, create it
		input := &obs.CreateBucketInput{
			Bucket: cfg.BucketName,
			BucketLocation: obs.BucketLocation{
				Location: cfg.Region,
			},
		}
		_, err = client.CreateBucket(input)
		if err != nil {
			return nil, fmt.errorf("failed to create bucket: %w", err)
		}
	}

	return &OBSRepository{
		client:     client,
		bucketName: cfg.BucketName,
		encryptionKey: cfg.EncryptionKey,
		logger:     zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger(),
	}, nil
}

// Close closes the OBS client
func (r *OBSRepository) Close() error {
	if r.client != nil {
		r.client.Close()
	}
	return nil
}

// ========== Transaction Repository Implementation ==========

type OBSTransactionRepository struct {
	*OBSRepository
}

func NewTransactionRepository(obsRepo *OBSRepository) repository.TransactionRepository {
	return &OBSTransactionRepository{obsRepo}
}

func (r *OBSTransactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
	// Generate ID if not provided
	if transaction.ID == uuid.Nil {
		transaction.ID = uuid.New()
	}
	
	// Set timestamps
	now := time.Now()
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = now
	}
	transaction.UpdatedAt = now

	// Convert to JSON
	data, err := json.Marshal(transaction)
	if err != nil {
		return fmt.errorf("failed to marshal transaction: %w", err)
	}

	// Encrypt data if encryption key is provided
	if r.encryptionKey != "" {
		// TODO: Implement encryption
		// For now, we'll store as plain JSON
	}

	// Save to OBS
	key := r.transactionKey(transaction.FamilyID, transaction.ID)
	input := &obs.PutObjectInput{
		PutObjectBasicInput: obs.PutObjectBasicInput{
			Bucket: r.bucketName,
			Key:    key,
			Metadata: map[string]string{
				"family-id":    transaction.FamilyID.String(),
				"user-id":      transaction.UserID.String(),
				"created-at":   transaction.CreatedAt.Format(time.RFC3339),
				"category":     transaction.Category,
				"payment-method": string(transaction.PaymentMethod),
			},
		},
		Body: data,
	}

	_, err = r.client.PutObject(input)
	if err != nil {
		return fmt.errorf("failed to save transaction to OBS: %w", err)
	}

	// Update family transaction index
	err = r.updateFamilyTransactionIndex(ctx, transaction.FamilyID, transaction.ID, transaction.TransactionDate)
	if err != nil {
		r.logger.error().Err(err).Msg("Failed to update transaction index")
		// Don't fail the transaction creation if index update fails
	}

	// Update user transaction index
	err = r.updateUserTransactionIndex(ctx, transaction.UserID, transaction.ID, transaction.TransactionDate)
	if err != nil {
		r.logger.error().Err(err).Msg("Failed to update user transaction index")
	}

	return nil
}

func (r *OBSTransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	// We need to find the transaction by scanning all families
	// This is inefficient but acceptable for Phase 1 with small data
	// In Phase 2, we'll have proper database indexes
	
	// List all family folders
	families, err := r.listFamilies(ctx)
	if err != nil {
		return nil, err
	}

	for _, familyID := range families {
		key := r.transactionKey(familyID, id)
		output, err := r.client.GetObject(&obs.GetObjectInput{
			GetObjectMetadataInput: obs.GetObjectMetadataInput{
				Bucket: r.bucketName,
				Key:    key,
			},
		})
		if err != nil {
			// Transaction not found in this family, continue
			continue
		}
		defer output.Body.Close()

		var transaction model.Transaction
		if err := json.NewDecoder(output.Body).Decode(&transaction); err != nil {
			return nil, fmt.errorf("failed to decode transaction: %w", err)
		}

		return &transaction, nil
	}

	return nil, fmt.errorf("transaction not found: %s", id)
}

func (r *OBSTransactionRepository) FindByFamily(ctx context.Context, familyID uuid.UUID, filters repository.TransactionFilters) ([]model.Transaction, error) {
	// Get transaction IDs from index
	indexKey := r.familyTransactionIndexKey(familyID)
	index, err := r.getTransactionIndex(ctx, indexKey)
	if err != nil {
		return nil, err
	}

	var transactions []model.Transaction
	for _, txID := range index.TransactionIDs {
		key := r.transactionKey(familyID, txID)
		output, err := r.client.GetObject(&obs.GetObjectInput{
			GetObjectMetadataInput: obs.GetObjectMetadataInput{
				Bucket: r.bucketName,
				Key:    key,
			},
		})
		if err != nil {
			r.logger.error().Err(err).Str("transaction_id", txID.String()).Msg("Failed to fetch transaction")
			continue
		}
		defer output.Body.Close()

		var transaction model.Transaction
		if err := json.NewDecoder(output.Body).Decode(&transaction); err != nil {
			r.logger.error().Err(err).Str("transaction_id", txID.String()).Msg("Failed to decode transaction")
			continue
		}

		// Apply filters
		if r.filterTransaction(&transaction, filters) {
			transactions = append(transactions, transaction)
		}
	}

	// Apply sorting
	transactions = r.sortTransactions(transactions, filters)

	// Apply limit and offset
	if filters.Limit > 0 {
		start := filters.Offset
		if start >= len(transactions) {
			return []model.Transaction{}, nil
		}
		end := start + filters.Limit
		if end > len(transactions) {
			end = len(transactions)
		}
		transactions = transactions[start:end]
	}

	return transactions, nil
}

func (r *OBSTransactionRepository) Update(ctx context.Context, transaction *model.Transaction) error {
	// Verify transaction exists
	existing, err := r.FindByID(ctx, transaction.ID)
	if err != nil {
		return err
	}

	// Update timestamps
	transaction.UpdatedAt = time.Now()
	transaction.CreatedAt = existing.CreatedAt

	// Save updated transaction
	data, err := json.Marshal(transaction)
	if err != nil {
		return fmt.errorf("failed to marshal transaction: %w", err)
	}

	key := r.transactionKey(transaction.FamilyID, transaction.ID)
	input := &obs.PutObjectInput{
		PutObjectBasicInput: obs.PutObjectBasicInput{
			Bucket: r.bucketName,
			Key:    key,
			Metadata: map[string]string{
				"family-id":    transaction.FamilyID.String(),
				"user-id":      transaction.UserID.String(),
				"updated-at":   transaction.UpdatedAt.Format(time.RFC3339),
				"category":     transaction.Category,
				"payment-method": string(transaction.PaymentMethod),
			},
		},
		Body: data,
	}

	_, err = r.client.PutObject(input)
	if err != nil {
		return fmt.errorf("failed to update transaction in OBS: %w", err)
	}

	return nil
}

func (r *OBSTransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Find transaction to get family ID
	transaction, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from OBS
	key := r.transactionKey(transaction.FamilyID, id)
	_, err = r.client.DeleteObject(&obs.DeleteObjectInput{
		Bucket: r.bucketName,
		Key:    key,
	})
	if err != nil {
		return fmt.errorf("failed to delete transaction from OBS: %w", err)
	}

	// Remove from indexes
	err = r.removeFromFamilyTransactionIndex(ctx, transaction.FamilyID, id)
	if err != nil {
		r.logger.error().Err(err).Msg("Failed to remove from family transaction index")
	}

	err = r.removeFromUserTransactionIndex(ctx, transaction.UserID, id)
	if err != nil {
		r.logger.error().Err(err).Msg("Failed to remove from user transaction index")
	}

	return nil
}

func (r *OBSTransactionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	transaction, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	transaction.DeletedAt = time.Now()
	return r.Update(ctx, transaction)
}

// Helper methods
func (r *OBSTransactionRepository) transactionKey(familyID, transactionID uuid.UUID) string {
	return path.Join("families", familyID.String(), "transactions", transactionID.String()+".json")
}

func (r *OBSTransactionRepository) familyTransactionIndexKey(familyID uuid.UUID) string {
	return path.Join("families", familyID.String(), "index", "transactions.json")
}

func (r *OBSTransactionRepository) userTransactionIndexKey(userID uuid.UUID) string {
	return path.Join("users", userID.String(), "index", "transactions.json")
}

type transactionIndex struct {
	FamilyID       uuid.UUID   `json:"family_id"`
	TransactionIDs []uuid.UUID `json:"transaction_ids"`
	LastUpdated    time.Time   `json:"last_updated"`
}

func (r *OBSTransactionRepository) getTransactionIndex(ctx context.Context, key string) (*transactionIndex, error) {
	output, err := r.client.GetObject(&obs.GetObjectInput{
		GetObjectMetadataInput: obs.GetObjectMetadataInput{
			Bucket: r.bucketName,
			Key:    key,
		},
	})
	if err != nil {
		// Index doesn't exist yet, create empty one
		return &transactionIndex{
			TransactionIDs: []uuid.UUID{},
			LastUpdated:    time.Now(),
		}, nil
	}
	defer output.Body.Close()

	var index transactionIndex
	if err := json.NewDecoder(output.Body).Decode(&index); err != nil {
		return nil, fmt.errorf("failed to decode transaction index: %w", err)
	}

	return &index, nil
}

func (r *OBSTransactionRepository) updateFamilyTransactionIndex(ctx context.Context, familyID, transactionID uuid.UUID, date time.Time) error {
	key := r.familyTransactionIndexKey(familyID)
	index, err := r.getTransactionIndex(ctx, key)
	if err != nil {
		return err
	}

	// Add transaction ID if not already present
	found := false
	for _, id := range index.TransactionIDs {
		if id == transactionID {
			found = true
			break
		}
	}

	if !found {
		index.TransactionIDs = append(index.TransactionIDs, transactionID)
		index.LastUpdated = time.Now()

		data, err := json.Marshal(index)
		if err != nil {
			return fmt.errorf("failed to marshal transaction index: %w", err)
		}

		input := &obs.PutObjectInput{
			PutObjectBasicInput: obs.PutObjectBasicInput{
				Bucket: r.bucketName,
				Key:    key,
			},
			Body: data,
		}

		_, err = r.client.PutObject(input)
		if err != nil {
			return fmt.errorf("failed to update transaction index: %w", err)
		}
	}

	return nil
}

func (r *OBSTransactionRepository) updateUserTransactionIndex(ctx context.Context, userID, transactionID uuid.UUID, date time.Time) error {
	// Similar implementation to updateFamilyTransactionIndex
	// Omitted for brevity
	return nil
}

func (r *OBSTransactionRepository) removeFromFamilyTransactionIndex(ctx context.Context, familyID, transactionID uuid.UUID) error {
	key := r.familyTransactionIndexKey(familyID)
	index, err := r.getTransactionIndex(ctx, key)
	if err != nil {
		return err
	}

	// Remove transaction ID
	var newIDs []uuid.UUID
	for _, id := range index.TransactionIDs {
		if id != transactionID {
			newIDs = append(newIDs, id)
		}
	}

	index.TransactionIDs = newIDs
	index.LastUpdated = time.Now()

	data, err := json.Marshal(index)
	if err != nil {
		return fmt.errorf("failed to marshal transaction index: %w", err)
	}

	input := &obs.PutObjectInput{
		PutObjectBasicInput: obs.PutObjectBasicInput{
			Bucket: r.bucketName,
			Key:    key,
		},
		Body: data,
	}

	_, err = r.client.PutObject(input)
	if err != nil {
		return fmt.errorf("failed to update transaction index: %w", err)
	}

	return nil
}

func (r *OBSTransactionRepository) filterTransaction(transaction *model.Transaction, filters repository.TransactionFilters) bool {
	// Apply date filters
	if filters.StartDate != nil && transaction.TransactionDate.Before(*filters.StartDate) {
		return false
	}
	if filters.EndDate != nil && transaction.TransactionDate.After(*filters.EndDate) {
		return false
	}

	// Apply category filter
	if filters.Category != "" && transaction.Category != filters.Category {
		return false
	}

	// Apply payment method filter
	if filters.PaymentMethod != "" && transaction.PaymentMethod != filters.PaymentMethod {
		return false
	}

	// Apply amount filters
	if filters.MinAmount != nil && transaction.Amount < *filters.MinAmount {
		return false
	}
	if filters.MaxAmount != nil && transaction.Amount > *filters.MaxAmount {
		return false
	}

	// Apply status filter
	if filters.Status != "" && transaction.Status != filters.Status {
		return false
	}

	// Apply verification filter
	if filters.IsVerified != nil && transaction.IsVerified != *filters.IsVerified {
		return false
	}

	// Apply recurring filter
	if filters.IsRecurring != nil && transaction.IsRecurring != *filters.IsRecurring {
		return false
	}

	// Apply search filter
	if filters.Search != "" {
		searchLower := filters.Search
		descriptionMatch := containsIgnoreCase(transaction.Description, searchLower)
		merchantMatch := containsIgnoreCase(transaction.Merchant, searchLower)
		notesMatch := containsIgnoreCase(transaction.Notes, searchLower)
		
		if !descriptionMatch && !merchantMatch && !notesMatch {
			return false
		}
	}

	return true
}

func (r *OBSTransactionRepository) sortTransactions(transactions []model.Transaction, filters repository.TransactionFilters) []model.Transaction {
	// Default sort by date descending
	if filters.SortBy == "" {
		filters.SortBy = "date"
		filters.SortOrder = "desc"
	}

	// Implement sorting logic based on filters.SortBy and filters.SortOrder
	// Omitted for brevity - would implement proper sorting

	return transactions
}

func (r *OBSTransactionRepository) listFamilies(ctx context.Context) ([]uuid.UUID, error) {
	// List all family folders in OBS
	input := &obs.ListObjectsInput{
		Bucket: r.bucketName,
		Prefix: "families/",
		Delimiter: "/",
	}

	output, err := r.client.ListObjects(input)
	if err != nil {
		return nil, fmt.errorf("failed to list families: %w", err)
	}

	var families []uuid.UUID
	for _, prefix := range output.CommonPrefixes {
		// Extract family ID from prefix
		// Prefix format: "families/{family-id}/"
		// Omitted for brevity
	}

	return families, nil
}

// Helper function
func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains check
	// In production, use strings.Contains with strings.ToLower
	return true // Simplified for example
}

// Note: Other repository methods (FindByDateRange, GetMonthlySummary, etc.) would be implemented similarly
// For Phase 1, we implement basic CRUD operations
// For Phase 2, we'll add more complex queries and aggregations