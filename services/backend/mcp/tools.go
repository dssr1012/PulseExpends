package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/dssr1012/pulse-expends/internal/repository"
)

// MCPTool represents an MCP tool that can be called by OpenClaw
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     ToolHandler            `json:"-"`
}

// ToolHandler is a function that handles tool execution
type ToolHandler func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// MCPToolManager manages all available MCP tools
type MCPToolManager struct {
	tools map[string]MCPTool
	repo  *repository.RepositoryManager
}

// NewMCPToolManager creates a new tool manager
func NewMCPToolManager(repo *repository.RepositoryManager) *MCPToolManager {
	mgr := &MCPToolManager{
		tools: make(map[string]MCPTool),
		repo:  repo,
	}
	mgr.registerTools()
	return mgr
}

// GetTools returns all available tools
func (m *MCPToolManager) GetTools() []MCPTool {
	tools := make([]MCPTool, 0, len(m.tools))
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ExecuteTool executes a tool by name
func (m *MCPToolManager) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	tool, exists := m.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	return tool.Handler(ctx, params)
}

// registerTools registers all available tools
func (m *MCPToolManager) registerTools() {
	m.tools["save_transaction"] = MCPTool{
		Name:        "save_transaction",
		Description: "Save a new transaction to the ledger",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"circle_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the family circle",
				},
				"user_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the user creating the transaction",
				},
				"amount": map[string]interface{}{
					"type":        "number",
					"description": "Transaction amount (negative for expenses, positive for income)",
				},
				"currency": map[string]interface{}{
					"type":        "string",
					"description": "Currency code (e.g., CLP, USD)",
					"default":     "CLP",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Transaction category",
				},
				"payment_method": map[string]interface{}{
					"type":        "string",
					"description": "Payment method (cash, debit_card, credit_card, transfer)",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Transaction description",
				},
				"date": map[string]interface{}{
					"type":        "string",
					"format":      "date-time",
					"description": "Transaction date (ISO 8601)",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Transaction tags",
				},
			},
			"required": []string{"circle_id", "user_id", "amount", "description"},
		},
		Handler: m.handleSaveTransaction,
	}

	m.tools["get_transactions"] = MCPTool{
		Name:        "get_transactions",
		Description: "Get transactions for a family circle with optional filters",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"circle_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the family circle",
				},
				"start_date": map[string]interface{}{
					"type":        "string",
					"format":      "date-time",
					"description": "Start date for filtering (ISO 8601)",
				},
				"end_date": map[string]interface{}{
					"type":        "string",
					"format":      "date-time",
					"description": "End date for filtering (ISO 8601)",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Filter by category",
				},
				"payment_method": map[string]interface{}{
					"type":        "string",
					"description": "Filter by payment method",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of transactions to return",
					"default":     50,
				},
				"offset": map[string]interface{}{
					"type":        "integer",
					"description": "Number of transactions to skip",
					"default":     0,
				},
			},
			"required": []string{"circle_id"},
		},
		Handler: m.handleGetTransactions,
	}

	m.tools["get_transaction_summary"] = MCPTool{
		Name:        "get_transaction_summary",
		Description: "Get summary statistics for transactions in a family circle",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"circle_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the family circle",
				},
				"start_date": map[string]interface{}{
					"type":        "string",
					"format":      "date-time",
					"description": "Start date for summary (ISO 8601)",
				},
				"end_date": map[string]interface{}{
					"type":        "string",
					"format":      "date-time",
					"description": "End date for summary (ISO 8601)",
				},
			},
			"required": []string{"circle_id"},
		},
		Handler: m.handleGetTransactionSummary,
	}

	m.tools["parse_credit_card_statement"] = MCPTool{
		Name:        "parse_credit_card_statement",
		Description: "Parse a credit card statement PDF/image and extract transactions",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"document_url": map[string]interface{}{
					"type":        "string",
					"description": "URL of the document to parse (OBS URL or external URL)",
				},
				"bank_name": map[string]interface{}{
					"type":        "string",
					"description": "Bank name for better parsing accuracy",
					"enum": []string{
						"Banco de Chile",
						"Banco Estado",
						"Santander",
						"BCI",
						"Scotiabank",
						"Itaú",
						"Falabella",
						"Ripley",
						"Paris",
						"Cencosud",
						"generic",
					},
				},
				"circle_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the family circle to save transactions to",
				},
				"user_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the user who uploaded the statement",
				},
				"auto_save": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to automatically save parsed transactions",
					"default":     false,
				},
				"extract_installments": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to extract installment information",
					"default":     true,
				},
			},
			"required": []string{"document_url", "circle_id", "user_id"},
		},
		Handler: m.handleParseCreditCardStatement,
	}

	m.tools["create_family_circle"] = MCPTool{
		Name:        "create_family_circle",
		Description: "Create a new family circle for expense tracking",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the family circle",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Description of the family circle",
				},
				"currency": map[string]interface{}{
					"type":        "string",
					"description": "Default currency for the circle",
					"default":     "CLP",
				},
				"member_ids": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "IDs of users to add as members",
				},
			},
			"required": []string{"name"},
		},
		Handler: m.handleCreateFamilyCircle,
	}

	m.tools["get_family_circle"] = MCPTool{
		Name:        "get_family_circle",
		Description: "Get details of a family circle",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"circle_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the family circle",
				},
			},
			"required": []string{"circle_id"},
		},
		Handler: m.handleGetFamilyCircle,
	}

	m.tools["add_circle_member"] = MCPTool{
		Name:        "add_circle_member",
		Description: "Add a member to a family circle",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"circle_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the family circle",
				},
				"user_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the user to add",
				},
				"role": map[string]interface{}{
					"type":        "string",
					"description": "Role of the member (admin, member, viewer)",
					"default":     "member",
				},
			},
			"required": []string{"circle_id", "user_id"},
	},
		Handler: m.handleAddCircleMember,
	}

	m.tools["get_monthly_summary"] = MCPTool{
		Name:        "get_monthly_summary",
		Description: "Get monthly summary of expenses and income",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"circle_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the family circle",
				},
				"months": map[string]interface{}{
					"type":        "integer",
					"description": "Number of months to include in summary",
					"default":     6,
				},
			},
			"required": []string{"circle_id"},
		},
		Handler: m.handleGetMonthlySummary,
	}

	m.tools["detect_spending_patterns"] = MCPTool{
		Name:        "detect_spending_patterns",
		Description: "Analyze transactions to detect spending patterns and anomalies",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"circle_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the family circle",
				},
				"start_date": map[string]interface{}{
					"type":        "string",
					"format":      "date-time",
					"description": "Start date for analysis (ISO 8601)",
				},
				"end_date": map[string]interface{}{
					"type":        "string",
					"format":      "date-time",
					"description": "End date for analysis (ISO 8601)",
				},
			},
			"required": []string{"circle_id"},
		},
		Handler: m.handleDetectSpendingPatterns,
	}

	m.tools["set_budget"] = MCPTool{
		Name:        "set_budget",
		Description: "Set or update budget for a family circle",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"circle_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the family circle",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Budget category",
				},
				"amount": map[string]interface{}{
					"type":        "number",
					"description": "Budget amount",
				},
				"period": map[string]interface{}{
					"type":        "string",
					"description": "Budget period (monthly, weekly, yearly)",
					"default":     "monthly",
				},
				"currency": map[string]interface{}{
					"type":        "string",
					"description": "Currency for budget",
					"default":     "CLP",
				},
			},
			"required": []string{"circle_id", "category", "amount"},
		},
		Handler: m.handleSetBudget,
	}

	m.tools["get_budget_status"] = MCPTool{
		Name:        "get_budget_status",
		Description: "Get current status of budgets for a family circle",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"circle_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the family circle",
				},
			},
			"required": []string{"circle_id"},
		},
		Handler: m.handleGetBudgetStatus,
	}
}

// Tool handlers implementation
func (m *MCPToolManager) handleSaveTransaction(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Parse parameters
	circleID, ok := params["circle_id"].(string)
	if !ok {
		return nil, fmt.Errorf("circle_id is required and must be a string")
	}

	userID, ok := params["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_id is required and must be a string")
	}

	amount, ok := params["amount"].(float64)
	if !ok {
		return nil, fmt.Errorf("amount is required and must be a number")
	}

	description, ok := params["description"].(string)
	if !ok {
		return nil, fmt.Errorf("description is required and must be a string")
	}

	// Create transaction
	transaction := &model.Transaction{
		ID:            generateUUID(),
		CircleID:      circleID,
		UserID:        userID,
		Amount:        amount,
		Currency:      getStringParam(params, "currency", "CLP"),
		Category:      getStringParam(params, "category", ""),
		PaymentMethod: getStringParam(params, "payment_method", "cash"),
		Description:   description,
		Date:          parseDateParam(params, "date", time.Now()),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Parse tags if provided
	if tags, ok := params["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				transaction.Tags = append(transaction.Tags, tagStr)
			}
		}
	}

	// Save transaction
	err := m.repo.Transaction.Create(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	return map[string]interface{}{
		"success":      true,
		"transaction":  transaction,
		"message":      "Transaction saved successfully",
		"transaction_id": transaction.ID,
	}, nil
}

func (m *MCPToolManager) handleGetTransactions(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	circleID, ok := params["circle_id"].(string)
	if !ok {
		return nil, fmt.Errorf("circle_id is required and must be a string")
	}

	// Build filter
	filter := model.TransactionFilter{
		CircleID: circleID,
		Limit:    getIntParam(params, "limit", 50),
		Offset:   getIntParam(params, "offset", 0),
	}

	// Parse optional filters
	if startDate, ok := params["start_date"].(string); ok {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter.StartDate = t
		}
	}

	if endDate, ok := params["end_date"].(string); ok {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filter.EndDate = t
		}
	}

	if category, ok := params["category"].(string); ok {
		filter.Category = category
	}

	if paymentMethod, ok := params["payment_method"].(string); ok {
		filter.PaymentMethod = paymentMethod
	}

	// Get transactions
	transactions, err := m.repo.Transaction.FindByCircle(ctx, circleID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	return map[string]interface{}{
		"success":      true,
		"transactions": transactions,
		"count":        len(transactions),
		"circle_id":    circleID,
	}, nil
}

func (m *MCPToolManager) handleGetTransactionSummary(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	circleID, ok := params["circle_id"].(string)
	if !ok {
		return nil, fmt.Errorf("circle_id is required and must be a string")
	}

	// Parse date range
	startDate := time.Now().AddDate(0, -1, 0) // Default: last month
	endDate := time.Now()

	if startDateStr, ok := params["start_date"].(string); ok {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = t
		}
	}

	if endDateStr, ok := params["end_date"].(string); ok {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = t
		}
	}

	// Get summary
	summary, err := m.repo.Transaction.GetSummary(ctx, circleID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction summary: %w", err)
	}

	return map[string]interface{}{
		"success":     true,
		"summary":     summary,
		"circle_id":   circleID,
		"start_date":  startDate,
		"end_date":    endDate,
		"period_days": int(endDate.Sub(startDate).Hours() / 24),
	}, nil
}

func (m *MCPToolManager) handleParseCreditCardStatement(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// This would call the Python microservice
	// For now, return a mock response
	documentURL, ok := params["document_url"].(string)
	if !ok {
		return nil, fmt.Errorf("document_url is required and must be a string")
	}

	circleID, ok := params["circle_id"].(string)
	if !ok {
		return nil, fmt.Errorf("circle_id is required and must be a string")
	}

	userID, ok := params["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_id is required and must be a string")
	}

	bankName := getStringParam(params, "bank_name", "generic")
	autoSave := getBoolParam(params, "auto_save", false)
	extractInstallments := getBoolParam(params, "extract_installments", true)

	// In a real implementation, this would call the Python microservice
	// For now, return a mock response
	return map[string]interface{}{
		"success":              true,
		"message":              "Document parsing initiated",
		"document_url":         documentURL,
		"bank_name":            bankName,
		"circle_id":            circleID,
		"user_id":              userID,
		"auto_save":            autoSave,
		"extract_installments": extractInstallments,
		"status":               "processing",
		"request_id":           generateUUID(),
		"estimated_completion": time.Now().Add(30 * time.Second).Format(time.RFC3339),
	}, nil
}

func (m *MCPToolManager) handleCreateFamilyCircle(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required and must be a string")
	}

	circle := &model.Circle{
		ID:          generateUUID(),
		Name:        name,
		Description: getStringParam(params, "description", ""),
		Currency:    getStringParam(params, "currency", "CLP"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Settings: model.CircleSettings{
			AllowMemberAddTransactions:    true,
			AllowMemberEditTransactions:   true,
			AllowMemberDeleteTransactions: false,
			DefaultCategories: []string{
				"supermercado", "restaurante", "transporte", "servicios",
				"salud", "educación", "entretenimiento", "ropa",
				"tecnología", "hogar", "otros",
			},
			BudgetAlertsEnabled: true,
			NotificationPreferences: model.NotificationPreferences{
				EmailNotifications:     true,
				PushNotifications:      true,
				WeeklySummary:          true,
				BudgetAlerts:           true,
				LargeTransactionAlerts: true,
			},
		},
	}

	// Parse members if provided
	if memberIDs, ok := params["member_ids"].([]interface{}); ok {
		for _, memberID := range memberIDs {
			if memberIDStr, ok := memberID.(string); ok {
				circle.Members = append(circle.Members, model.Member{
					UserID:   memberIDStr,
					Role:     "member",
					JoinedAt: time.Now(),
					IsActive: true,
				})
			}
		}
	}

	// Save circle
	err := m.repo.Circle.Create(ctx, circle)
	if err != nil {
		return nil, fmt.Errorf("failed to create family circle: %w", err)
	}

	return map[string]interface{}{
		"success":    true,
		"circle":     circle,
		"message":    "Family circle created successfully",
		"circle_id":  circle.ID,
	}, nil
}

func (m *MCPToolManager) handleGetFamilyCircle(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	circleID, ok := params["circle_id"].(string)
	if !ok {
		return nil, fmt.Errorf("circle_id is required and must be a string")
	}

	circle, err := m.repo.Circle.FindByID(ctx, circleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family circle: %w", err)
	}

	if circle == nil {
		return nil, fmt.Errorf("family circle not found: %s", circleID)
	}

	// Get summary
	summary, err := m.repo.Circle.GetSummary(ctx, circleID)
	if err != nil {
		// Don't fail if summary fails
		summary = &model.CircleSummary{CircleID: circleID}
	}

	return map[string]interface{}{
		"success": true,
		"circle":  circle,
		"summary": summary,
	}, nil
}

func (m *MCPToolManager) handleAddCircleMember(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	circleID, ok := params["circle_id"].(string)
	if !ok {
		return nil, fmt.Errorf("circle_id is required and must be a string")
	}

	userID, ok := params["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_id is required and must be a string")
	}

	role := getStringParam(params, "role", "member")

	member := model.Member{
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
		IsActive: true,
	}

	err := m.repo.Circle.AddMember(ctx, circleID, member)
	if err != nil {
		return nil, fmt.Errorf("failed to add circle member: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"message":   "Member added successfully",
		"circle_id": circleID,
		"user_id":   userID,
		"role":      role,
	}, nil
}

func (m *MCPToolManager) handleGetMonthlySummary(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	circleID, ok := params["circle_id"].(string)
	if !ok {
		return nil, fmt.Errorf("circle_id is required and must be a string")
	}

	months := getIntParam(params, "months", 6)

	// Get monthly trends
	trends, err := m.repo.Transaction.GetMonthlyTrend(ctx, circleID, months)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly summary: %w", err)
	}

	// Calculate totals
	var totalIncome, totalExpenses float64
	for _, trend := range trends {
		totalIncome += trend.Income
		totalExpenses += trend.Expenses
	}

	return map[string]interface{}{
		"success":        true,
		"circle_id":      circleID,
		"months":         months,
		"monthly_trends": trends,
		"total_income":   totalIncome,
		"total_expenses": totalExpenses,
		"net_balance":    totalIncome - totalExpenses,
		"average_income": totalIncome / float64(len(trends)),
		"average_expenses": totalExpenses / float64(len(trends)),
	}, nil
}

func (m *MCPToolManager) handleDetectSpendingPatterns(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	circleID, ok := params["circle_id"].(string)
	if !ok {
		return nil, fmt.Errorf("circle_id is required and must be a string")
	}

	// Parse date range
	startDate := time.Now().AddDate(0, -3, 0) // Default: last 3 months
	endDate := time.Now()

	if startDateStr, ok := params["start_date"].(string); ok {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = t
		}
	}

	if endDateStr, ok := params["end_date"].(string); ok {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = t
		}
	}

	// Get transactions for analysis
	filter := model.TransactionFilter{
		CircleID:  circleID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	transactions, err := m.repo.Transaction.FindByCircle(ctx, circleID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions for analysis: %w", err)
	}

	// Simple pattern detection (in real implementation, this would use ML)
	patterns := detectSpendingPatterns(transactions)

	return map[string]interface{}{
		"success":         true,
		"circle_id":       circleID,
		"period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
			"days":       int(endDate.Sub(startDate).Hours() / 24),
		},
		"transaction_count": len(transactions),
		"patterns":          patterns,
		"recommendations":   generateRecommendations(patterns),
	}, nil
}

func (m *MCPToolManager) handleSetBudget(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// This is a simplified implementation
	// In a real implementation, this would save to a budget repository
	return map[string]interface{}{
		"success": true,
		"message": "Budget feature coming soon",
		"note":    "Budget tracking will be implemented in Phase 2",
	}, nil
}

func (m *MCPToolManager) handleGetBudgetStatus(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// This is a simplified implementation
	// In a real implementation, this would retrieve from a budget repository
	return map[string]interface{}{
		"success": true,
		"message": "Budget feature coming soon",
		"note":    "Budget tracking will be implemented in Phase 2",
	}, nil
}

// Helper functions
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func getStringParam(params map[string]interface{}, key string, defaultValue string) string {
	if value, ok := params[key].(string); ok {
		return value
	}
	return defaultValue
}

func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if value, ok := params[key].(float64); ok {
		return int(value)
	}
	return defaultValue
}

func getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := params[key].(bool); ok {
		return value
	}
	return defaultValue
}

func parseDateParam(params map[string]interface{}, key string, defaultValue time.Time) time.Time {
	if value, ok := params[key].(string); ok {
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			return t
		}
	}
	return defaultValue
}

// Pattern detection helpers
func detectSpendingPatterns(transactions []model.Transaction) map[string]interface{} {
	patterns := map[string]interface{}{
		"recurring_expenses":   detectRecurringExpenses(transactions),
		"large_transactions":   detectLargeTransactions(transactions),
		"category_trends":      analyzeCategoryTrends(transactions),
		"payment_method_usage": analyzePaymentMethodUsage(transactions),
		"spending_velocity":    calculateSpendingVelocity(transactions),
	}
	return patterns
}

func detectRecurringExpenses(transactions []model.Transaction) []map[string]interface{} {
	// Simple recurring expense detection based on similar amounts and descriptions
	recurring := []map[string]interface{}{}
	
	// Group by description and amount (simplified)
	groups := make(map[string][]model.Transaction)
	for _, tx := range transactions {
		if tx.Amount < 0 { // Only expenses
			key := fmt.Sprintf("%s_%.2f", tx.Description, math.Abs(tx.Amount))
			groups[key] = append(groups[key], tx)
		}
	}
	
	// Find groups with multiple occurrences
	for key, txs := range groups {
		if len(txs) >= 2 {
			// Check if they're approximately monthly
			if isMonthlyRecurring(txs) {
				recurring = append(recurring, map[string]interface{}{
					"description": txs[0].Description,
					"amount":      math.Abs(txs[0].Amount),
					"count":       len(txs),
					"frequency":   "monthly",
					"category":    txs[0].Category,
				})
			}
		}
	}
	
	return recurring
}

func detectLargeTransactions(transactions []model.Transaction) []map[string]interface{} {
	// Detect transactions that are significantly larger than average
	if len(transactions) == 0 {
		return []map[string]interface{}{}
	}
	
	// Calculate average transaction amount (absolute value)
	var total float64
	for _, tx := range transactions {
		total += math.Abs(tx.Amount)
	}
	average := total / float64(len(transactions))
	
	// Find transactions more than 3x the average
	large := []map[string]interface{}{}
	for _, tx := range transactions {
		if math.Abs(tx.Amount) > average*3 {
			large = append(large, map[string]interface{}{
				"date":        tx.Date,
				"description": tx.Description,
				"amount":      tx.Amount,
				"category":    tx.Category,
				"deviation":   math.Abs(tx.Amount) / average,
			})
		}
	}
	
	return large
}

func analyzeCategoryTrends(transactions []model.Transaction) map[string]interface{} {
	// Analyze spending by category over time
	categoryMonthly := make(map[string]map[string]float64)
	
	for _, tx := range transactions {
		if tx.Amount >= 0 {
			continue // Skip income for now
		}
		
		month := tx.Date.Format("2006-01")
		category := tx.Category
		if category == "" {
			category = "uncategorized"
		}
		
		if categoryMonthly[category] == nil {
			categoryMonthly[category] = make(map[string]float64)
		}
		categoryMonthly[category][month] += math.Abs(tx.Amount)
	}
	
	// Calculate trends
	trends := make(map[string]interface{})
	for category, monthlyData := range categoryMonthly {
		months := make([]string, 0, len(monthlyData))
		amounts := make([]float64, 0, len(monthlyData))
		
		for month, amount := range monthlyData {
			months = append(months, month)
			amounts = append(amounts, amount)
		}
		
		// Sort by month
		sort.Strings(months)
		sortedAmounts := make([]float64, len(months))
		for i, month := range months {
			sortedAmounts[i] = monthlyData[month]
		}
		
		// Calculate trend (simple linear regression)
		trend := calculateTrend(sortedAmounts)
		
		trends[category] = map[string]interface{}{
			"monthly_data": monthlyData,
			"trend":        trend,
			"total":        sumAmounts(sortedAmounts),
			"average":      averageAmounts(sortedAmounts),
		}
	}
	
	return trends
}

func analyzePaymentMethodUsage(transactions []model.Transaction) map[string]interface{} {
	usage := make(map[string]interface{})
	methodCounts := make(map[string]int)
	methodAmounts := make(map[string]float64)
	
	for _, tx := range transactions {
		method := tx.PaymentMethod
		if method == "" {
			method = "unknown"
		}
		
		methodCounts[method]++
		methodAmounts[method] += math.Abs(tx.Amount)
	}
	
	for method, count := range methodCounts {
		usage[method] = map[string]interface{}{
			"count":  count,
			"amount": methodAmounts[method],
			"percentage": float64(count) / float64(len(transactions)) * 100,
		}
	}
	
	return usage
}

func calculateSpendingVelocity(transactions []model.Transaction) map[string]interface{} {
	if len(transactions) == 0 {
		return map[string]interface{}{
			"daily_average":   0,
			"weekly_average":  0,
			"monthly_average": 0,
			"trend":           "stable",
		}
	}
	
	// Sort by date
	sortedTxs := make([]model.Transaction, len(transactions))
	copy(sortedTxs, transactions)
	sort.Slice(sortedTxs, func(i, j int) bool {
		return sortedTxs[i].Date.Before(sortedTxs[j].Date)
	})
	
	// Calculate time span
	firstDate := sortedTxs[0].Date
	lastDate := sortedTxs[len(sortedTxs)-1].Date
	days := lastDate.Sub(firstDate).Hours() / 24
	if days < 1 {
		days = 1
	}
	
	// Calculate total expenses
	var totalExpenses float64
	for _, tx := range sortedTxs {
		if tx.Amount < 0 {
			totalExpenses += math.Abs(tx.Amount)
		}
	}
	
	// Calculate averages
	dailyAverage := totalExpenses / days
	weeklyAverage := dailyAverage * 7
	monthlyAverage := dailyAverage * 30
	
	// Calculate trend (last week vs previous week)
	lastWeekStart := lastDate.AddDate(0, 0, -7)
	prevWeekStart := lastWeekStart.AddDate(0, 0, -7)
	
	lastWeekTotal := 0.0
	prevWeekTotal := 0.0
	
	for _, tx := range sortedTxs {
		if tx.Amount >= 0 {
			continue
		}
		
		amount := math.Abs(tx.Amount)
		if tx.Date.After(lastWeekStart) && tx.Date.Before(lastDate) {
			lastWeekTotal += amount
		} else if tx.Date.After(prevWeekStart) && tx.Date.Before(lastWeekStart) {
			prevWeekTotal += amount
		}
	}
	
	var trend string
	if prevWeekTotal == 0 {
		trend = "new"
	} else if lastWeekTotal > prevWeekTotal * 1.2 {
		trend = "increasing"
	} else if lastWeekTotal < prevWeekTotal * 0.8 {
		trend = "decreasing"
	} else {
		trend = "stable"
	}
	
	return map[string]interface{}{
		"daily_average":   dailyAverage,
		"weekly_average":  weeklyAverage,
		"monthly_average": monthlyAverage,
		"trend":           trend,
		"last_week_total": lastWeekTotal,
		"prev_week_total": prevWeekTotal,
		"change_percentage": (lastWeekTotal - prevWeekTotal) / prevWeekTotal * 100,
	}
}

func generateRecommendations(patterns map[string]interface{}) []map[string]interface{} {
	recommendations := []map[string]interface{}{}
	
	// Analyze patterns and generate recommendations
	if recurring, ok := patterns["recurring_expenses"].([]map[string]interface{}); ok {
		if len(recurring) > 5 {
			recommendations = append(recommendations, map[string]interface{}{
				"type":        "recurring_expenses",
				"title":       "Many recurring expenses detected",
				"description": fmt.Sprintf("You have %d recurring expenses. Consider reviewing subscriptions and memberships.", len(recurring)),
				"priority":    "medium",
				"action":      "Review recurring expenses in settings",
			})
		}
	}
	
	if large, ok := patterns["large_transactions"].([]map[string]interface{}); ok {
		if len(large) > 0 {
			recommendations = append(recommendations, map[string]interface{}{
				"type":        "large_transactions",
				"title":       "Unusually large transactions detected",
				"description": fmt.Sprintf("Found %d transactions significantly larger than average. Verify these are legitimate.", len(large)),
				"priority":    "high",
				"action":      "Review large transactions",
			})
		}
	}
	
	if velocity, ok := patterns["spending_velocity"].(map[string]interface{}); ok {
		if trend, ok := velocity["trend"].(string); ok && trend == "increasing" {
			change, _ := velocity["change_percentage"].(float64)
			recommendations = append(recommendations, map[string]interface{}{
				"type":        "spending_trend",
				"title":       "Spending is increasing",
				"description": fmt.Sprintf("Your spending increased by %.1f%% compared to previous week.", change),
				"priority":    "medium",
				"action":      "Review recent expenses and set budget alerts",
			})
		}
	}
	
	// Add generic recommendations if none found
	if len(recommendations) == 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "general",
			"title":       "Spending patterns look normal",
			"description": "No unusual spending patterns detected. Continue monitoring your expenses.",
			"priority":    "low",
			"action":      "Keep tracking expenses regularly",
		})
	}
	
	return recommendations
}

func isMonthlyRecurring(transactions []model.Transaction) bool {
	if len(transactions) < 2 {
		return false
	}
	
	// Sort by date
	sorted := make([]model.Transaction, len(transactions))
	copy(sorted, transactions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.Before(sorted[j].Date)
	})
	
	// Check if transactions are approximately monthly
	for i := 1; i < len(sorted); i++ {
		daysBetween := sorted[i].Date.Sub(sorted[i-1].Date).Hours() / 24
		if daysBetween < 25 || daysBetween > 35 {
			return false
		}
	}
	
	return true
}

func calculateTrend(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	
	n := float64(len(values))
	var sumX, sumY, sumXY, sumX2 float64
	
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// Simple linear regression slope
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	return slope
}

func sumAmounts(values []float64) float64 {
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum
}

func averageAmounts(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return sumAmounts(values) / float64(len(values))
}

// Import math and sort packages
import (
	"math"
	"sort"
)