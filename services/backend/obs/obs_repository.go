package obs

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

// OBSRepository implements repository interfaces using Huawei Cloud OBS
type OBSRepository struct {
	client     *obs.ObsClient
	bucketName string
	encryptionKey string
}

// NewOBSRepository creates a new OBS repository instance
func NewOBSRepository(endpoint, ak, sk, bucketName, encryptionKey string) (*OBSRepository, error) {
	client, err := obs.New(ak, sk, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create OBS client: %w", err)
	}

	// Check if bucket exists, create if not
	_, err = client.HeadBucket(bucketName)
	if err != nil {
		// Bucket doesn't exist, create it
		input := &obs.CreateBucketInput{
			Bucket: bucketName,
			BucketLocation: obs.BucketLocation{
				Location: "la-south-2", // Santiago region
			},
		}
		_, err = client.CreateBucket(input)
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
		}
	}

	return &OBSRepository{
		client:        client,
		bucketName:    bucketName,
		encryptionKey: encryptionKey,
	}, nil
}

// OBSTransactionRepository implements TransactionRepository using OBS
type OBSTransactionRepository struct {
	*OBSRepository
}

// NewOBSTransactionRepository creates a new transaction repository
func NewOBSTransactionRepository(repo *OBSRepository) repository.TransactionRepository {
	return &OBSTransactionRepository{OBSRepository: repo}
}

func (r *OBSTransactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
	if transaction.ID == "" {
		transaction.ID = generateUUID()
	}
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now()
	}
	transaction.UpdatedAt = time.Now()

	data, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	encryptedData, err := r.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt transaction data: %w", err)
	}

	key := r.transactionKey(transaction.CircleID, transaction.ID)
	input := &obs.PutObjectInput{
		Bucket: r.bucketName,
		Key:    key,
		Body:   strings.NewReader(string(encryptedData)),
	}

	_, err = r.client.PutObject(input)
	if err != nil {
		return fmt.Errorf("failed to store transaction in OBS: %w", err)
	}

	return nil
}

func (r *OBSTransactionRepository) FindByID(ctx context.Context, id string) (*model.Transaction, error) {
	// We need to know the circle ID to find the transaction
	// In OBS implementation, we would need to search across all circles
	// This is a limitation of OBS for Phase 1
	// For Phase 2, we'll implement proper indexing
	return nil, fmt.Errorf("FindByID not implemented for OBS - use FindByCircle with filter")
}

func (r *OBSTransactionRepository) FindByCircle(ctx context.Context, circleID string, filter model.TransactionFilter) ([]model.Transaction, error) {
	// List all transactions for the circle
	prefix := fmt.Sprintf("circles/%s/transactions/", circleID)
	input := &obs.ListObjectsInput{
		Bucket: r.bucketName,
		Prefix: prefix,
	}

	output, err := r.client.ListObjects(input)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions from OBS: %w", err)
	}

	var transactions []model.Transaction
	for _, obj := range output.Contents {
		// Get each transaction
		getInput := &obs.GetObjectInput{
			Bucket: r.bucketName,
			Key:    obj.Key,
		}

		getOutput, err := r.client.GetObject(getInput)
		if err != nil {
			// Skip corrupted objects
			continue
		}
		defer getOutput.Body.Close()

		// Read and decrypt data
		var encryptedData []byte
		_, err = getOutput.Body.Read(encryptedData)
		if err != nil {
			continue
		}

		data, err := r.decrypt(encryptedData)
		if err != nil {
			continue
		}

		var transaction model.Transaction
		if err := json.Unmarshal(data, &transaction); err != nil {
			continue
		}

		// Apply filters
		if r.matchesFilter(transaction, filter) {
			transactions = append(transactions, transaction)
		}
	}

	// Sort and limit results
	transactions = r.sortTransactions(transactions, filter)
	
	if filter.Limit > 0 && len(transactions) > filter.Limit {
		transactions = transactions[:filter.Limit]
	}

	return transactions, nil
}

func (r *OBSTransactionRepository) FindByUser(ctx context.Context, userID string, filter model.TransactionFilter) ([]model.Transaction, error) {
	// This is inefficient in OBS - we need to scan all circles
	// For Phase 1, we'll implement a simple scan
	// For Phase 2, we'll have proper indexing
	var allTransactions []model.Transaction
	
	// List all circles (simplified - in reality we'd need user-circle mapping)
	circlesPrefix := "circles/"
	listInput := &obs.ListObjectsInput{
		Bucket: r.bucketName,
		Prefix: circlesPrefix,
		Delimiter: "/",
	}

	output, err := r.client.ListObjects(listInput)
	if err != nil {
		return nil, fmt.Errorf("failed to list circles: %w", err)
	}

	for _, prefix := range output.CommonPrefixes {
		circleID := strings.TrimPrefix(strings.TrimSuffix(prefix.Prefix, "/"), "circles/")
		
		// Get transactions for this circle
		circleFilter := filter
		circleFilter.CircleID = circleID
		
		transactions, err := r.FindByCircle(ctx, circleID, circleFilter)
		if err != nil {
			continue
		}

		// Filter by user
		for _, transaction := range transactions {
			if transaction.UserID == userID {
				allTransactions = append(allTransactions, transaction)
			}
		}
	}

	return allTransactions, nil
}

func (r *OBSTransactionRepository) Update(ctx context.Context, transaction *model.Transaction) error {
	transaction.UpdatedAt = time.Now()
	
	// Delete old and create new (OBS doesn't support updates)
	oldKey := r.transactionKey(transaction.CircleID, transaction.ID)
	deleteInput := &obs.DeleteObjectInput{
		Bucket: r.bucketName,
		Key:    oldKey,
	}
	
	_, err := r.client.DeleteObject(deleteInput)
	if err != nil {
		return fmt.Errorf("failed to delete old transaction: %w", err)
	}

	return r.Create(ctx, transaction)
}

func (r *OBSTransactionRepository) Delete(ctx context.Context, id string) error {
	// We need circle ID - this is a limitation of OBS implementation
	// For now, we'll implement a scan-based delete
	return fmt.Errorf("Delete requires circle ID in OBS implementation")
}

func (r *OBSTransactionRepository) GetSummary(ctx context.Context, circleID string, startDate, endDate time.Time) (*model.TransactionSummary, error) {
	filter := model.TransactionFilter{
		CircleID:  circleID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	transactions, err := r.FindByCircle(ctx, circleID, filter)
	if err != nil {
		return nil, err
	}

	summary := &model.TransactionSummary{
		TotalIncome:    0,
		TotalExpenses:  0,
		NetBalance:     0,
		ByCategory:     make(map[string]float64),
		ByPaymentMethod: make(map[string]float64),
		MonthlyTrend:   []model.MonthlyTrend{},
	}

	for _, transaction := range transactions {
		if transaction.Amount >= 0 {
			summary.TotalIncome += transaction.Amount
		} else {
			summary.TotalExpenses += -transaction.Amount
		}

		summary.ByCategory[transaction.Category] += transaction.Amount
		summary.ByPaymentMethod[transaction.PaymentMethod] += transaction.Amount
	}

	summary.NetBalance = summary.TotalIncome - summary.TotalExpenses
	
	// Calculate monthly trend (simplified)
	summary.MonthlyTrend = r.calculateMonthlyTrend(transactions)

	return summary, nil
}

func (r *OBSTransactionRepository) GetMonthlyTrend(ctx context.Context, circleID string, months int) ([]model.MonthlyTrend, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)

	filter := model.TransactionFilter{
		CircleID:  circleID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	transactions, err := r.FindByCircle(ctx, circleID, filter)
	if err != nil {
		return nil, err
	}

	return r.calculateMonthlyTrend(transactions), nil
}

func (r *OBSTransactionRepository) BulkCreate(ctx context.Context, transactions []model.Transaction) error {
	for _, transaction := range transactions {
		if err := r.Create(ctx, &transaction); err != nil {
			return fmt.Errorf("failed to create transaction %s: %w", transaction.ID, err)
		}
	}
	return nil
}

func (r *OBSTransactionRepository) Close() error {
	// OBS client doesn't need explicit closing in this SDK
	return nil
}

// Helper methods
func (r *OBSTransactionRepository) transactionKey(circleID, transactionID string) string {
	return fmt.Sprintf("circles/%s/transactions/%s.json", circleID, transactionID)
}

func (r *OBSTransactionRepository) encrypt(data []byte) ([]byte, error) {
	// Implement encryption using encryptionKey
	// For Phase 1, we'll use simple XOR (not secure - replace with proper encryption)
	encrypted := make([]byte, len(data))
	key := []byte(r.encryptionKey)
	for i := range data {
		encrypted[i] = data[i] ^ key[i%len(key)]
	}
	return encrypted, nil
}

func (r *OBSTransactionRepository) decrypt(data []byte) ([]byte, error) {
	// Same as encrypt for XOR
	return r.encrypt(data)
}

func (r *OBSTransactionRepository) matchesFilter(transaction model.Transaction, filter model.TransactionFilter) bool {
	if filter.CircleID != "" && transaction.CircleID != filter.CircleID {
		return false
	}
	if filter.UserID != "" && transaction.UserID != filter.UserID {
		return false
	}
	if !filter.StartDate.IsZero() && transaction.Date.Before(filter.StartDate) {
		return false
	}
	if !filter.EndDate.IsZero() && transaction.Date.After(filter.EndDate) {
		return false
	}
	if filter.Category != "" && transaction.Category != filter.Category {
		return false
	}
	if filter.PaymentMethod != "" && transaction.PaymentMethod != filter.PaymentMethod {
		return false
	}
	if filter.MinAmount != 0 && transaction.Amount < filter.MinAmount {
		return false
	}
	if filter.MaxAmount != 0 && transaction.Amount > filter.MaxAmount {
		return false
	}
	if len(filter.Tags) > 0 {
		hasAllTags := true
		for _, tag := range filter.Tags {
			found := false
			for _, t := range transaction.Tags {
				if t == tag {
					found = true
					break
				}
			}
			if !found {
				hasAllTags = false
				break
			}
		}
		if !hasAllTags {
			return false
		}
	}
	return true
}

func (r *OBSTransactionRepository) sortTransactions(transactions []model.Transaction, filter model.TransactionFilter) []model.Transaction {
	// Simple date-based sorting for now
	// In Phase 2, we'll implement more sophisticated sorting
	for i := 0; i < len(transactions)-1; i++ {
		for j := i + 1; j < len(transactions); j++ {
			if transactions[j].Date.After(transactions[i].Date) {
				transactions[i], transactions[j] = transactions[j], transactions[i]
			}
		}
	}
	return transactions
}

func (r *OBSTransactionRepository) calculateMonthlyTrend(transactions []model.Transaction) []model.MonthlyTrend {
	monthlyData := make(map[string]*model.MonthlyTrend)
	
	for _, transaction := range transactions {
		monthKey := transaction.Date.Format("2006-01")
		
		if _, exists := monthlyData[monthKey]; !exists {
			monthlyData[monthKey] = &model.MonthlyTrend{
				Month:    monthKey,
				Income:   0,
				Expenses: 0,
				Balance:  0,
			}
		}
		
		if transaction.Amount >= 0 {
			monthlyData[monthKey].Income += transaction.Amount
		} else {
			monthlyData[monthKey].Expenses += -transaction.Amount
		}
		monthlyData[monthKey].Balance = monthlyData[monthKey].Income - monthlyData[monthKey].Expenses
	}
	
	var trends []model.MonthlyTrend
	for _, trend := range monthlyData {
		trends = append(trends, *trend)
	}
	
	// Sort by month
	for i := 0; i < len(trends)-1; i++ {
		for j := i + 1; j < len(trends); j++ {
			if trends[j].Month < trends[i].Month {
				trends[i], trends[j] = trends[j], trends[i]
			}
		}
	}
	
	return trends
}

func generateUUID() string {
	// Simple UUID generation for Phase 1
	// In production, use github.com/google/uuid
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// OBS Circle and User repository implementations would follow similar patterns
// For brevity, we'll create stubs for now

type OBSCircleRepository struct {
	*OBSRepository
}

func NewOBSCircleRepository(repo *OBSRepository) repository.CircleRepository {
	return &OBSCircleRepository{OBSRepository: repo}
}

func (r *OBSCircleRepository) Create(ctx context.Context, circle *model.Circle) error {
	// Implementation similar to transaction repository
	return nil
}

func (r *OBSCircleRepository) FindByID(ctx context.Context, id string) (*model.Circle, error) {
	return nil, nil
}

func (r *OBSCircleRepository) FindByUser(ctx context.Context, userID string, filter model.CircleFilter) ([]model.Circle, error) {
	return nil, nil
}

func (r *OBSCircleRepository) Update(ctx context.Context, circle *model.Circle) error {
	return nil
}

func (r *OBSCircleRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (r *OBSCircleRepository) AddMember(ctx context.Context, circleID string, member model.Member) error {
	return nil
}

func (r *OBSCircleRepository) RemoveMember(ctx context.Context, circleID, userID string) error {
	return nil
}

func (r *OBSCircleRepository) UpdateMember(ctx context.Context, circleID, userID, role string) error {
	return nil
}

func (r *OBSCircleRepository) GetSummary(ctx context.Context, circleID string) (*model.CircleSummary, error) {
	return nil, nil
}

func (r *OBSCircleRepository) Close() error {
	return nil
}