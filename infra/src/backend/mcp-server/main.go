package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"
    
    "github.com/google/uuid"
)

type Transaction struct {
    ID          string    `json:"id"`
    Amount      float64   `json:"amount"`
    Currency    string    `json:"currency"`
    Category    string    `json:"category"`
    Description string    `json:"description"`
    Date        time.Time `json:"date"`
    Type        string    `json:"type"` // "income" or "expense"
    UserID      string    `json:"user_id"`
    CreatedAt   time.Time `json:"created_at"`
}

var (
    transactions []Transaction
    mu           sync.RWMutex
)

func main() {
    // Initialize with some sample data
    transactions = []Transaction{
        {
            ID:          "1",
            Amount:      100.00,
            Currency:    "USD",
            Category:    "Food",
            Description: "Groceries",
            Date:        time.Now().Add(-24 * time.Hour),
            Type:        "expense",
            UserID:      "user1",
            CreatedAt:   time.Now().Add(-24 * time.Hour),
        },
        {
            ID:          "2",
            Amount:      50.00,
            Currency:    "USD",
            Category:    "Transportation",
            Description: "Gas",
            Date:        time.Now().Add(-12 * time.Hour),
            Type:        "expense",
            UserID:      "user1",
            CreatedAt:   time.Now().Add(-12 * time.Hour),
        },
    }
    
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "ok",
            "service": "MCP Server",
            "version": "1.0.0",
            "message": "Enhanced MCP Server with expense tracking",
            "endpoints": "/transactions (GET, POST), /transactions/{id} (GET), /health",
        })
    })
    
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy",
        })
    })
    
    http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        switch r.Method {
        case "GET":
            mu.RLock()
            json.NewEncoder(w).Encode(map[string]interface{}{
                "status": "success",
                "data":   transactions,
                "count":  len(transactions),
            })
            mu.RUnlock()
            
        case "POST":
            var newTx Transaction
            if err := json.NewDecoder(r.Body).Decode(&newTx); err != nil {
                http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
                return
            }
            
            // Generate ID if not provided
            if newTx.ID == "" {
                newTx.ID = uuid.New().String()
            }
            if newTx.CreatedAt.IsZero() {
                newTx.CreatedAt = time.Now()
            }
            if newTx.Date.IsZero() {
                newTx.Date = time.Now()
            }
            if newTx.Currency == "" {
                newTx.Currency = "USD"
            }
            if newTx.UserID == "" {
                newTx.UserID = "user1"
            }
            
            mu.Lock()
            transactions = append(transactions, newTx)
            mu.Unlock()
            
            w.WriteHeader(http.StatusCreated)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "status":  "success",
                "message": "Transaction added successfully",
                "data":    newTx,
            })
            
        default:
            http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
        }
    })
    
    http.HandleFunc("/transactions/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        if r.Method != "GET" {
            http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
            return
        }
        
        // Simple ID extraction from path
        id := r.URL.Path[len("/transactions/"):]
        
        mu.RLock()
        for _, tx := range transactions {
            if tx.ID == id {
                json.NewEncoder(w).Encode(map[string]interface{}{
                    "status": "success",
                    "data":   tx,
                })
                mu.RUnlock()
                return
            }
        }
        mu.RUnlock()
        
        http.Error(w, `{"error":"Transaction not found"}`, http.StatusNotFound)
    })
    
    // Simple expense summary
    http.HandleFunc("/summary", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        mu.RLock()
        defer mu.RUnlock()
        
        var totalExpenses, totalIncome float64
        categories := make(map[string]float64)
        
        for _, tx := range transactions {
            if tx.Type == "expense" {
                totalExpenses += tx.Amount
                categories[tx.Category] += tx.Amount
            } else if tx.Type == "income" {
                totalIncome += tx.Amount
            }
        }
        
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status":         "success",
            "total_expenses": totalExpenses,
            "total_income":   totalIncome,
            "net_balance":    totalIncome - totalExpenses,
            "categories":     categories,
            "transaction_count": len(transactions),
        })
    })
    
    fmt.Println("Enhanced MCP Server starting on :8080")
    fmt.Println("Endpoints:")
    fmt.Println("  GET  /              - Server info")
    fmt.Println("  GET  /health        - Health check")
    fmt.Println("  GET  /transactions  - List all transactions")
    fmt.Println("  POST /transactions  - Add new transaction")
    fmt.Println("  GET  /transactions/{id} - Get specific transaction")
    fmt.Println("  GET  /summary       - Get expense summary")
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}