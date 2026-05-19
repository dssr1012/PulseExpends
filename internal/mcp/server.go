package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// MCPServer represents the MCP server
type MCPServer struct {
	router     *chi.Mux
	toolMgr    *MCPToolManager
	repo       *repository.RepositoryManager
	httpServer *http.Server
}

// NewMCPServer creates a new MCP server
func NewMCPServer(repo *repository.RepositoryManager) *MCPServer {
	s := &MCPServer{
		router:  chi.NewRouter(),
		toolMgr: NewMCPToolManager(repo),
		repo:    repo,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures HTTP middleware
func (s *MCPServer) setupMiddleware() {
	// Basic middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, restrict this
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Content type middleware
	s.router.Use(middleware.SetHeader("Content-Type", "application/json"))
}

// setupRoutes configures HTTP routes
func (s *MCPServer) setupRoutes() {
	// Health check
	s.router.Get("/health", s.handleHealthCheck)

	// MCP protocol endpoints
	s.router.Route("/mcp", func(r chi.Router) {
		// Tool discovery
		r.Get("/tools", s.handleListTools)
		
		// Tool execution
		r.Post("/tools/{toolName}/execute", s.handleExecuteTool)
		
		// Batch tool execution
		r.Post("/tools/execute-batch", s.handleExecuteBatch)
		
		// Tool schemas
		r.Get("/tools/schemas", s.handleGetToolSchemas)
		
		// Server info
		r.Get("/info", s.handleServerInfo)
	})

	// API endpoints (for direct HTTP access)
	s.router.Route("/api/v1", func(r chi.Router) {
		// Transactions
		r.Route("/transactions", func(r chi.Router) {
			r.Post("/", s.handleAPICreateTransaction)
			r.Get("/", s.handleAPIGetTransactions)
			r.Get("/summary", s.handleAPIGetTransactionSummary)
			r.Get("/{transactionId}", s.handleAPIGetTransaction)
			r.Put("/{transactionId}", s.handleAPIUpdateTransaction)
			r.Delete("/{transactionId}", s.handleAPIDeleteTransaction)
		})

		// Circles
		r.Route("/circles", func(r chi.Router) {
			r.Post("/", s.handleAPICreateCircle)
			r.Get("/", s.handleAPIGetCircles)
			r.Get("/{circleId}", s.handleAPIGetCircle)
			r.Put("/{circleId}", s.handleAPIUpdateCircle)
			r.Delete("/{circleId}", s.handleAPIDeleteCircle)
			r.Post("/{circleId}/members", s.handleAPIAddCircleMember)
			r.Delete("/{circleId}/members/{userId}", s.handleAPIRemoveCircleMember)
		})

		// Document parsing
		r.Route("/documents", func(r chi.Router) {
			r.Post("/parse", s.handleAPIParseDocument)
			r.Get("/parse/{requestId}", s.handleAPIGetParseResult)
		})

		// Analytics
		r.Route("/analytics", func(r chi.Router) {
			r.Get("/monthly-summary/{circleId}", s.handleAPIGetMonthlySummary)
			r.Get("/spending-patterns/{circleId}", s.handleAPIDetectSpendingPatterns)
			r.Get("/budget-status/{circleId}", s.handleAPIGetBudgetStatus)
		})
	})

	// WebSocket for real-time updates
	s.router.HandleFunc("/ws", s.handleWebSocket)

	// Static files for documentation
	s.router.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))))
}

// Start starts the MCP server
func (s *MCPServer) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting MCP server on %s", addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *MCPServer) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// handleHealthCheck handles health check requests
func (s *MCPServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"services": map[string]string{
			"mcp_server":   "running",
			"repository":   "connected",
			"tool_manager": "ready",
		},
	}

	// Check repository connection
	if s.repo != nil {
		// Simple ping test for repository
		health["services"].(map[string]string)["repository"] = "connected"
	} else {
		health["services"].(map[string]string)["repository"] = "disconnected"
		health["status"] = "degraded"
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

// handleListTools lists all available MCP tools
func (s *MCPServer) handleListTools(w http.ResponseWriter, r *http.Request) {
	tools := s.toolMgr.GetTools()
	
	// Convert to MCP tool format
	mcpTools := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		mcpTools[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
	}

	response := map[string]interface{}{
		"tools": mcpTools,
		"count": len(tools),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleExecuteTool executes a single MCP tool
func (s *MCPServer) handleExecuteTool(w http.ResponseWriter, r *http.Request) {
	toolName := chi.URLParam(r, "toolName")
	
	// Parse request body
	var request struct {
		Params map[string]interface{} `json:"params"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Execute tool
	result, err := s.toolMgr.ExecuteTool(r.Context(), toolName, request.Params)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to execute tool %s", toolName), err)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"tool":    toolName,
		"result":  result,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleExecuteBatch executes multiple MCP tools in batch
func (s *MCPServer) handleExecuteBatch(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var requests []struct {
		Tool   string                 `json:"tool"`
		Params map[string]interface{} `json:"params"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Execute tools in parallel
	type toolResult struct {
		Tool   string      `json:"tool"`
		Result interface{} `json:"result"`
		Error  string      `json:"error,omitempty"`
	}
	
	results := make([]toolResult, len(requests))
	
	// Use a simple goroutine for each tool (in production, use worker pool)
	for i, req := range requests {
		result, err := s.toolMgr.ExecuteTool(r.Context(), req.Tool, req.Params)
		if err != nil {
			results[i] = toolResult{
				Tool:  req.Tool,
				Error: err.Error(),
			}
		} else {
			results[i] = toolResult{
				Tool:   req.Tool,
				Result: result,
			}
		}
	}

	response := map[string]interface{}{
		"success": true,
		"results": results,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleGetToolSchemas returns detailed schemas for all tools
func (s *MCPServer) handleGetToolSchemas(w http.ResponseWriter, r *http.Request) {
	tools := s.toolMgr.GetTools()
	
	schemas := make(map[string]interface{})
	for _, tool := range tools {
		schemas[tool.Name] = map[string]interface{}{
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
			"outputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"success": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the tool execution was successful",
					},
					"result": map[string]interface{}{
						"type":        "object",
						"description": "Tool-specific result data",
					},
					"error": map[string]interface{}{
						"type":        "string",
						"description": "Error message if execution failed",
					},
				},
				"required": []string{"success"},
			},
		}
	}

	response := map[string]interface{}{
		"schemas": schemas,
		"count":   len(schemas),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleServerInfo returns server information
func (s *MCPServer) handleServerInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"name":        "PulseExpends MCP Server",
		"version":     "1.0.0",
		"description": "Model Context Protocol server for PulseExpends family expense tracking",
		"capabilities": []string{
			"transaction_management",
			"document_parsing",
			"family_circle_management",
			"spending_analytics",
			"budget_tracking",
		},
		"repository": map[string]interface{}{
			"type": "hybrid",
			"phase": 1,
			"storage": "huawei-cloud-obs",
			"migration_path": "postgresql/mongodb",
		},
		"api": map[string]interface{}{
			"mcp_protocol": "1.0",
			"http_api":     "v1",
			"websocket":    true,
		},
		"limits": map[string]interface{}{
			"max_transactions_per_request": 1000,
			"max_document_size_mb":         50,
			"rate_limit_per_minute":        60,
			"concurrent_requests":          100,
		},
		"contact": map[string]interface{}{
			"email":    "support@pulseexpends.com",
			"docs":     "https://docs.pulseexpends.com",
			"source":   "https://github.com/dssr1012/PulseExpends",
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(info)
}

// API handlers
func (s *MCPServer) handleAPICreateTransaction(w http.ResponseWriter, r *http.Request) {
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := s.toolMgr.ExecuteTool(r.Context(), "save_transaction", params)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to create transaction", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (s *MCPServer) handleAPIGetTransactions(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})
	
	// Parse query parameters
	circleID := r.URL.Query().Get("circle_id")
	if circleID != "" {
		params["circle_id"] = circleID
	}
	
	startDate := r.URL.Query().Get("start_date")
	if startDate != "" {
		params["start_date"] = startDate
	}
	
	endDate := r.URL.Query().Get("end_date")
	if endDate != "" {
		params["end_date"] = endDate
	}
	
	category := r.URL.Query().Get("category")
	if category != "" {
		params["category"] = category
	}
	
	paymentMethod := r.URL.Query().Get("payment_method")
	if paymentMethod != "" {
		params["payment_method"] = paymentMethod
	}
	
	if limit := r.URL.Query().Get("limit"); limit != "" {
		params["limit"] = limit
	}
	
	if offset := r.URL.Query().Get("offset"); offset != "" {
		params["offset"] = offset
	}

	result, err := s.toolMgr.ExecuteTool(r.Context(), "get_transactions", params)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get transactions", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (s *MCPServer) handleAPIGetTransactionSummary(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})
	
	// Parse query parameters
	circleID := r.URL.Query().Get("circle_id")
	if circleID == "" {
		s.sendError(w, http.StatusBadRequest, "circle_id is required", nil)
		return
	}
	params["circle_id"] = circleID
	
	startDate := r.URL.Query().Get("start_date")
	if startDate != "" {
		params["start_date"] = startDate
	}
	
	endDate := r.URL.Query().Get("end_date")
	if endDate != "" {
		params["end_date"] = endDate
	}

	result, err := s.toolMgr.ExecuteTool(r.Context(), "get_transaction_summary", params)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get transaction summary", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (s *MCPServer) handleAPIGetTransaction(w http.ResponseWriter, r *http.Request) {
	transactionID := chi.URLParam(r, "transactionId")
	
	// This would use a direct repository call in a real implementation
	// For now, return not implemented
	s.sendError(w, http.StatusNotImplemented, "Direct transaction retrieval not yet implemented", nil)
}

func (s *MCPServer) handleAPIUpdateTransaction(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, http.StatusNotImplemented, "Transaction update not yet implemented", nil)
}

func (s *MCPServer) handleAPIDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, http.StatusNotImplemented, "Transaction deletion not yet implemented", nil)
}

func (s *MCPServer) handleAPICreateCircle(w http.ResponseWriter, r *http.Request) {
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := s.toolMgr.ExecuteTool(r.Context(), "create_family_circle", params)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to create family circle", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (s *MCPServer) handleAPIGetCircles(w http.ResponseWriter, r *http.Request) {
	// This would use a direct repository call in a real implementation
	// For now, return not implemented
	s.sendError(w, http.StatusNotImplemented, "Circle listing not yet implemented", nil)
}

func (s *MCPServer) handleAPIGetCircle(w http.ResponseWriter, r *http.Request) {
	circleID := chi.URLParam(r, "circleId")
	
	params := map[string]interface{}{
		"circle_id": circleID,
	}

	result, err := s.toolMgr.ExecuteTool(r.Context(), "get_family_circle", params)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get family circle", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (s *MCPServer) handleAPIUpdateCircle(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, http.StatusNotImplemented, "Circle update not yet implemented", nil)
}

func (s *MCPServer) handleAPIDeleteCircle(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, http.StatusNotImplemented, "Circle deletion not yet implemented", nil)
}

func (s *MCPServer) handleAPIAddCircleMember(w http.ResponseWriter, r *http.Request) {
	circleID := chi.URLParam(r, "circleId")
	
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	
	params["circle_id"] = circleID

	result, err := s.toolMgr.ExecuteTool(r.Context(), "add_circle_member", params)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to add circle member", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (s *MCPServer) handleAPIRemoveCircleMember(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, http.StatusNotImplemented, "Circle member removal not yet implemented", nil)
}

func (s *MCPServer) handleAPIParseDocument(w http.ResponseWriter, r *http.Request) {
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := s.toolMgr.ExecuteTool(r.Context(), "parse_credit_card_statement", params)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to parse document", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(result)
}

func (s *MCPServer) handleAPIGetParseResult(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "requestId")
	
	// This would query the Python microservice or check OBS for results
	// For now, return a mock response
	response := map[string]interface{}{
		"success":    true,
		"request_id": requestID,
		"status":     "completed",
		"message":    "Document parsing results retrieval not yet implemented",
		"note":       "In production, this would fetch results from the Python microservice",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *MCPServer) handleAPIGetMonthlySummary(w http.ResponseWriter, r *http.Request) {
	circleID := chi.URLParam(r, "circleId")
	
	params := map[string]interface{}{
		"circle_id": circleID,
	}
	
	months := r.URL.Query().Get("months")
	if months != "" {
		params["months"] = months
	}

	result, err := s.toolMgr.ExecuteTool(r.Context(), "get_monthly_summary", params)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get monthly summary", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (s *MCPServer) handleAPIDetectSpendingPatterns(w http.ResponseWriter, r *http.Request) {
	circleID := chi.URLParam(r, "circleId")
	
	params := map[string]interface{}{
		"circle_id": circleID,
	}
	
	startDate := r.URL.Query().Get("start_date")
	if startDate != "" {
		params["start_date"] = startDate
	}
	
	endDate := r.URL.Query().Get("end_date")
	if endDate != "" {
		params["end_date"] = endDate
	}

	result, err := s.toolMgr.ExecuteTool(r.Context(), "detect_spending_patterns", params)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to detect spending patterns", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (s *MCPServer) handleAPIGetBudgetStatus(w http.ResponseWriter, r *http.Request) {
	circleID := chi.URLParam(r, "circleId")
	
	params := map[string]interface{}{
		"circle_id": circleID,
	}

	result, err := s.toolMgr.ExecuteTool(r.Context(), "get_budget_status", params)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get budget status", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// handleWebSocket handles WebSocket connections for real-time updates
func (s *MCPServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// WebSocket implementation would go here
	// For now, return not implemented
	s.sendError(w, http.StatusNotImplemented, "WebSocket not yet implemented", nil)
}

// sendError sends an error response
func (s *MCPServer) sendError(w http.ResponseWriter, statusCode int, message string, err error) {
	errorResponse := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    statusCode,
			"message": message,
			"details": nil,
		},
		"timestamp": time.Now().UTC(),
	}

	if err != nil {
		errorResponse["error"].(map[string]interface{})["details"] = err.Error()
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

// MCPProtocolHandler handles MCP protocol-specific requests
func (s *MCPServer) MCPProtocolHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle MCP protocol version negotiation
		// This would implement the full MCP protocol
		// For now, route to appropriate handlers
		switch r.URL.Path {
		case "/mcp/tools":
			s.handleListTools(w, r)
		case "/mcp/execute":
			toolName := r.URL.Query().Get("tool")
			if toolName == "" {
				s.sendError(w, http.StatusBadRequest, "Tool name is required", nil)
				return
			}
			s.handleExecuteTool(w, r)
		default:
			s.sendError(w, http.StatusNotFound, "MCP endpoint not found", nil)
		}
	}
}