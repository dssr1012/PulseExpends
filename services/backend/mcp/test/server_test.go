package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dssr1012/pulse-expends/internal/mcp"
	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryManager is a mock implementation of RepositoryManager
type MockRepositoryManager struct {
	mock.Mock
}

func (m *MockRepositoryManager) Transaction() repository.TransactionRepository {
	args := m.Called()
	return args.Get(0).(repository.TransactionRepository)
}

func (m *MockRepositoryManager) Circle() repository.CircleRepository {
	args := m.Called()
	return args.Get(0).(repository.CircleRepository)
}

func (m *MockRepositoryManager) User() repository.UserRepository {
	args := m.Called()
	return args.Get(0).(repository.UserRepository)
}

func (m *MockRepositoryManager) Budget() repository.BudgetRepository {
	args := m.Called()
	return args.Get(0).(repository.BudgetRepository)
}

func (m *MockRepositoryManager) Close() error {
	args := m.Called()
	return args.error(0)
}

// MockTransactionRepository is a mock implementation of TransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(ctx context.Context, transaction *repository.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.error(0)
}

func (m *MockTransactionRepository) FindByID(ctx context.Context, id string) (*repository.Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*repository.Transaction), args.error(1)
}

func (m *MockTransactionRepository) FindByCircle(ctx context.Context, circleID string, filter repository.TransactionFilter) ([]repository.Transaction, error) {
	args := m.Called(ctx, circleID, filter)
	return args.Get(0).([]repository.Transaction), args.error(1)
}

func (m *MockTransactionRepository) Update(ctx context.Context, transaction *repository.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.error(0)
}

func (m *MockTransactionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.error(0)
}

func (m *MockTransactionRepository) GetSummary(ctx context.Context, circleID string, startDate, endDate time.Time) (*repository.TransactionSummary, error) {
	args := m.Called(ctx, circleID, startDate, endDate)
	return args.Get(0).(*repository.TransactionSummary), args.error(1)
}

func (m *MockTransactionRepository) GetMonthlyTrend(ctx context.Context, circleID string, months int) ([]repository.MonthlyTrend, error) {
	args := m.Called(ctx, circleID, months)
	return args.Get(0).([]repository.MonthlyTrend), args.error(1)
}

// MockCircleRepository is a mock implementation of CircleRepository
type MockCircleRepository struct {
	mock.Mock
}

func (m *MockCircleRepository) Create(ctx context.Context, circle *repository.Circle) error {
	args := m.Called(ctx, circle)
	return args.error(0)
}

func (m *MockCircleRepository) FindByID(ctx context.Context, id string) (*repository.Circle, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*repository.Circle), args.error(1)
}

func (m *MockCircleRepository) FindByUser(ctx context.Context, userID string) ([]repository.Circle, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]repository.Circle), args.error(1)
}

func (m *MockCircleRepository) Update(ctx context.Context, circle *repository.Circle) error {
	args := m.Called(ctx, circle)
	return args.error(0)
}

func (m *MockCircleRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.error(0)
}

func (m *MockCircleRepository) AddMember(ctx context.Context, circleID string, member repository.Member) error {
	args := m.Called(ctx, circleID, member)
	return args.error(0)
}

func (m *MockCircleRepository) RemoveMember(ctx context.Context, circleID, userID string) error {
	args := m.Called(ctx, circleID, userID)
	return args.error(0)
}

func (m *MockCircleRepository) GetSummary(ctx context.Context, circleID string) (*repository.CircleSummary, error) {
	args := m.Called(ctx, circleID)
	return args.Get(0).(*repository.CircleSummary), args.error(1)
}

func TestMCPServer_HealthCheck(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.NotNil(t, response["timestamp"])
	assert.Equal(t, "1.0.0", response["version"])

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
}

func TestMCPServer_ListTools(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Create test request
	req := httptest.NewRequest("GET", "/mcp/tools", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "tools")
	assert.Contains(t, response, "count")

	tools := response["tools"].([]interface{})
	assert.Greater(t, len(tools), 0)

	// Check that we have the expected tools
	toolNames := make([]string, len(tools))
	for i, tool := range tools {
		toolMap := tool.(map[string]interface{})
		toolNames[i] = toolMap["name"].(string)
	}

	expectedTools := []string{
		"save_transaction",
		"get_transactions",
		"get_transaction_summary",
		"parse_credit_card_statement",
		"create_family_circle",
		"get_family_circle",
		"add_circle_member",
		"get_monthly_summary",
		"detect_spending_patterns",
		"set_budget",
		"get_budget_status",
	}

	for _, expectedTool := range expectedTools {
		assert.Contains(t, toolNames, expectedTool)
	}

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
}

func TestMCPServer_ExecuteTool_SaveTransaction(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Mock transaction creation
	mockTransactionRepo.On("Create", mock.Anything, mock.AnythingOfType("*repository.Transaction")).Return(nil)

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Create test request
	transactionData := map[string]interface{}{
		"circle_id":      "circle-123",
		"user_id":        "user-456",
		"amount":         -15000.50,
		"description":    "Supermercado",
		"currency":       "CLP",
		"category":       "supermercado",
		"payment_method": "debit_card",
		"date":           "2024-01-15T10:30:00Z",
		"tags":           []string{"comida", "necesario"},
	}

	body, _ := json.Marshal(map[string]interface{}{
		"params": transactionData,
	})

	req := httptest.NewRequest("POST", "/mcp/tools/save_transaction/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, true, response["success"])
	assert.Equal(t, "save_transaction", response["tool"])
	assert.Contains(t, response, "result")

	result := response["result"].(map[string]interface{})
	assert.Equal(t, true, result["success"])
	assert.Contains(t, result, "transaction_id")
	assert.Equal(t, "Transaction saved successfully", result["message"])

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
}

func TestMCPServer_ExecuteTool_GetTransactions(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Mock transaction retrieval
	expectedTransactions := []repository.Transaction{
		{
			ID:            "tx-1",
			CircleID:      "circle-123",
			UserID:        "user-456",
			Amount:        -15000.50,
			Description:   "Supermercado",
			Category:      "supermercado",
			PaymentMethod: "debit_card",
			Date:          time.Now().Add(-24 * time.Hour),
		},
		{
			ID:            "tx-2",
			CircleID:      "circle-123",
			UserID:        "user-456",
			Amount:        -5000.00,
			Description:   "Restaurante",
			Category:      "restaurante",
			PaymentMethod: "credit_card",
			Date:          time.Now().Add(-48 * time.Hour),
		},
	}

	filter := repository.TransactionFilter{
		CircleID: "circle-123",
		Limit:    50,
		Offset:   0,
	}

	mockTransactionRepo.On("FindByCircle", mock.Anything, "circle-123", filter).Return(expectedTransactions, nil)

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Create test request
	requestData := map[string]interface{}{
		"circle_id": "circle-123",
		"limit":     50,
		"offset":    0,
	}

	body, _ := json.Marshal(map[string]interface{}{
		"params": requestData,
	})

	req := httptest.NewRequest("POST", "/mcp/tools/get_transactions/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, true, response["success"])
	assert.Equal(t, "get_transactions", response["tool"])

	result := response["result"].(map[string]interface{})
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "circle-123", result["circle_id"])
	assert.Equal(t, float64(2), result["count"])

	transactions := result["transactions"].([]interface{})
	assert.Equal(t, 2, len(transactions))

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
}

func TestMCPServer_ExecuteTool_CreateFamilyCircle(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Mock circle creation
	mockCircleRepo.On("Create", mock.Anything, mock.AnythingOfType("*repository.Circle")).Return(nil)

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Create test request
	circleData := map[string]interface{}{
		"name":        "Familia Pérez",
		"description": "Gastos familiares",
		"currency":    "CLP",
		"member_ids":  []string{"user-1", "user-2"},
	}

	body, _ := json.Marshal(map[string]interface{}{
		"params": circleData,
	})

	req := httptest.NewRequest("POST", "/mcp/tools/create_family_circle/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, true, response["success"])
	assert.Equal(t, "create_family_circle", response["tool"])

	result := response["result"].(map[string]interface{})
	assert.Equal(t, true, result["success"])
	assert.Contains(t, result, "circle_id")
	assert.Equal(t, "Family circle created successfully", result["message"])

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
	mockCircleRepo.AssertExpectations(t)
}

func TestMCPServer_ExecuteTool_InvalidTool(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Create test request for non-existent tool
	requestData := map[string]interface{}{
		"circle_id": "circle-123",
	}

	body, _ := json.Marshal(map[string]interface{}{
		"params": requestData,
	})

	req := httptest.NewRequest("POST", "/mcp/tools/nonexistent_tool/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	errorData := response["error"].(map[string]interface{})
	assert.Contains(t, errorData["message"].(string), "Failed to execute tool")

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
}

func TestMCPServer_ExecuteTool_MissingRequiredParams(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Create test request with missing required params
	transactionData := map[string]interface{}{
		// Missing circle_id, user_id, amount, description
		"currency":       "CLP",
		"category":       "supermercado",
	}

	body, _ := json.Marshal(map[string]interface{}{
		"params": transactionData,
	})

	req := httptest.NewRequest("POST", "/mcp/tools/save_transaction/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	errorData := response["error"].(map[string]interface{})
	assert.Contains(t, errorData["message"].(string), "Failed to execute tool")

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
}

func TestMCPServer_GetToolSchemas(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Create test request
	req := httptest.NewRequest("GET", "/mcp/tools/schemas", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "schemas")
	assert.Contains(t, response, "count")

	schemas := response["schemas"].(map[string]interface{})
	assert.Greater(t, len(schemas), 0)

	// Check that we have schemas for all tools
	expectedTools := []string{
		"save_transaction",
		"get_transactions",
		"get_transaction_summary",
		"parse_credit_card_statement",
		"create_family_circle",
		"get_family_circle",
		"add_circle_member",
		"get_monthly_summary",
		"detect_spending_patterns",
		"set_budget",
		"get_budget_status",
	}

	for _, tool := range expectedTools {
		assert.Contains(t, schemas, tool)
		toolSchema := schemas[tool].(map[string]interface{})
		assert.Contains(t, toolSchema, "description")
		assert.Contains(t, toolSchema, "inputSchema")
		assert.Contains(t, toolSchema, "outputSchema")
	}

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
}

func TestMCPServer_ServerInfo(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Create test request
	req := httptest.NewRequest("GET", "/mcp/info", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "PulseExpends MCP Server", response["name"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "Model Context Protocol server for PulseExpends family expense tracking", response["description"])

	// Check capabilities
	capabilities := response["capabilities"].([]interface{})
	expectedCaps := []string{
		"transaction_management",
		"document_parsing",
		"family_circle_management",
		"spending_analytics",
		"budget_tracking",
	}

	for _, expectedCap := range expectedCaps {
		assert.Contains(t, capabilities, expectedCap)
	}

	// Check repository info
	repositoryInfo := response["repository"].(map[string]interface{})
	assert.Equal(t, "hybrid", repositoryInfo["type"])
	assert.Equal(t, float64(1), repositoryInfo["phase"])
	assert.Equal(t, "huawei-cloud-obs", repositoryInfo["storage"])

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
}

func TestMCPServer_ExecuteBatch(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Mock transaction creation
	mockTransactionRepo.On("Create", mock.Anything, mock.AnythingOfType("*repository.Transaction")).Return(nil).Times(2)

	// Mock circle creation
	mockCircleRepo.On("Create", mock.Anything, mock.AnythingOfType("*repository.Circle")).Return(nil)

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Create test request with batch of tools
	batchRequests := []map[string]interface{}{
		{
			"tool": "save_transaction",
			"params": map[string]interface{}{
				"circle_id":      "circle-123",
				"user_id":        "user-456",
				"amount":         -15000.50,
				"description":    "Supermercado",
				"currency":       "CLP",
				"category":       "supermercado",
				"payment_method": "debit_card",
			},
		},
		{
			"tool": "save_transaction",
			"params": map[string]interface{}{
				"circle_id":      "circle-123",
				"user_id":        "user-456",
				"amount":         -5000.00,
				"description":    "Restaurante",
				"currency":       "CLP",
				"category":       "restaurante",
				"payment_method": "credit_card",
			},
		},
		{
			"tool": "create_family_circle",
			"params": map[string]interface{}{
				"name":        "Familia Pérez",
				"description": "Gastos familiares",
				"currency":    "CLP",
			},
		},
	}

	body, _ := json.Marshal(batchRequests)

	req := httptest.NewRequest("POST", "/mcp/tools/execute-batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, true, response["success"])
	assert.Contains(t, response, "results")

	results := response["results"].([]interface{})
	assert.Equal(t, 3, len(results))

	// Check each result
	for i, result := range results {
		resultMap := result.(map[string]interface{})
		assert.Contains(t, resultMap, "tool")
		
		toolName := resultMap["tool"].(string)
		assert.Contains(t, []string{"save_transaction", "create_family_circle"}, toolName)
		
		// All should be successful in this test
		assert.NotContains(t, resultMap, "error")
		assert.Contains(t, resultMap, "result")
	}

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
	mockCircleRepo.AssertExpectations(t)
}

func TestMCPServer_APIEndpoints(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Mock transaction creation
	mockTransactionRepo.On("Create", mock.Anything, mock.AnythingOfType("*repository.Transaction")).Return(nil)

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Test api endpoint for creating transaction
	transactionData := map[string]interface{}{
		"circle_id":      "circle-123",
		"user_id":        "user-456",
		"amount":         -15000.50,
		"description":    "Supermercado",
		"currency":       "CLP",
		"category":       "supermercado",
		"payment_method": "debit_card",
	}

	body, _ := json.Marshal(transactionData)

	req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, true, response["success"])
	assert.Contains(t, response, "transaction_id")
	assert.Equal(t, "Transaction saved successfully", response["message"])

	// Test api endpoint for getting transaction summary
	req2 := httptest.NewRequest("GET", "/api/v1/transactions/summary?circle_id=circle-123", nil)
	w2 := httptest.NewRecorder()

	// Mock transaction summary
	summary := &repository.TransactionSummary{
		TotalTransactions: 10,
		TotalIncome:       500000,
		TotalExpenses:     450000,
		NetBalance:        50000,
		AverageTransaction: 95000,
		MostCommonCategory: "supermercado",
		MostCommonPaymentMethod: "debit_card",
	}

	mockTransactionRepo.On("GetSummary", mock.Anything, "circle-123", mock.Anything, mock.Anything).Return(summary, nil)

	server.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var response2 map[string]interface{}
	err2 := json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.NoError(t, err2)

	assert.Equal(t, true, response2["success"])
	assert.Contains(t, response2, "summary")
	assert.Equal(t, "circle-123", response2["circle_id"])

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
}

func TestMCPServer_ErrorHandling(t *testing.T) {
	// Create mocks
	mockRepoMgr := new(MockRepositoryManager)
	mockTransactionRepo := new(MockTransactionRepository)
	mockCircleRepo := new(MockCircleRepository)

	// Setup mock expectations
	mockRepoMgr.On("Transaction").Return(mockTransactionRepo)
	mockRepoMgr.On("Circle").Return(mockCircleRepo)
	mockRepoMgr.On("User").Return(nil)
	mockRepoMgr.On("Budget").Return(nil)

	// Mock transaction creation that fails
	mockTransactionRepo.On("Create", mock.Anything, mock.AnythingOfType("*repository.Transaction")).Return(fmt.errorf("database connection failed"))

	// Create server
	server := mcp.NewMCPServer(mockRepoMgr)

	// Test error handling
	transactionData := map[string]interface{}{
		"circle_id":      "circle-123",
		"user_id":        "user-456",
		"amount":         -15000.50,
		"description":    "Supermercado",
	}

	body, _ := json.Marshal(map[string]interface{}{
		"params": transactionData,
	})

	req := httptest.NewRequest("POST", "/mcp/tools/save_transaction/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	errorData := response["error"].(map[string]interface{})
	assert.Contains(t, errorData["message"].(string), "Failed to execute tool")
	assert.Contains(t, errorData["details"].(string), "database connection failed")

	// Test invalid JSON
	req2 := httptest.NewRequest("POST", "/mcp/tools/save_transaction/execute", bytes.NewReader([]byte("invalid json")))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	server.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)

	var response2 map[string]interface{}
	err2 := json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.NoError(t, err2)

	assert.Equal(t, false, response2["success"])
	errorData2 := response2["error"].(map[string]interface{})
	assert.Contains(t, errorData2["message"].(string), "Invalid request body")

	// Verify mock expectations
	mockRepoMgr.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
}