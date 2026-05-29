package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dssr1012/pulse-expends/internal/service"
	"github.com/dssr1012/pulse-expends/services/backend/api/middleware"
	"github.com/gorilla/mux"
)

type ExchangeRateHandler struct {
	exchangeRateService service.ExchangeRateService
}

func NewExchangeRateHandler(exchangeRateService service.ExchangeRateService) *ExchangeRateHandler {
	return &ExchangeRateHandler{
		exchangeRateService: exchangeRateService,
	}
}

// GetLatestRate retrieves the latest exchange rate for a currency pair
func (h *ExchangeRateHandler) GetLatestRate(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	baseCurrency := vars["base"]
	targetCurrency := vars["target"]

	// Validate currency codes
	if !h.exchangeRateService.ValidateCurrencyCode(baseCurrency) {
		http.Error(w, fmt.Sprintf("Invalid base currency code: %s", baseCurrency), http.StatusBadRequest)
		return
	}
	if !h.exchangeRateService.ValidateCurrencyCode(targetCurrency) {
		http.Error(w, fmt.Sprintf("Invalid target currency code: %s", targetCurrency), http.StatusBadRequest)
		return
	}

	rate, err := h.exchangeRateService.GetLatestRate(r.Context(), baseCurrency, targetCurrency)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get exchange rate: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"rate":    rate,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConvertAmount converts an amount from one currency to another
func (h *ExchangeRateHandler) ConvertAmount(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		Amount         float64 `json:"amount"`
		FromCurrency   string  `json:"fromCurrency"`
		ToCurrency     string  `json:"toCurrency"`
		Date           string  `json:"date,omitempty"`
		UseLatest      bool    `json:"useLatest,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate currency codes
	if !h.exchangeRateService.ValidateCurrencyCode(request.FromCurrency) {
		http.Error(w, fmt.Sprintf("Invalid from currency code: %s", request.FromCurrency), http.StatusBadRequest)
		return
	}
	if !h.exchangeRateService.ValidateCurrencyCode(request.ToCurrency) {
		http.Error(w, fmt.Sprintf("Invalid to currency code: %s", request.ToCurrency), http.StatusBadRequest)
		return
	}

	// Validate amount
	if request.Amount <= 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	var convertedAmount float64
	var err error

	if request.Date != "" {
		// Convert with specific date
		date, err := time.Parse(time.RFC3339, request.Date)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid date format: %v", err), http.StatusBadRequest)
			return
		}
		convertedAmount, err = h.exchangeRateService.ConvertAmount(r.Context(), request.Amount, request.FromCurrency, request.ToCurrency, date)
	} else {
		// Convert with latest rate
		convertedAmount, err = h.exchangeRateService.ConvertAmount(r.Context(), request.Amount, request.FromCurrency, request.ToCurrency, time.Time{})
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert amount: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":         true,
		"originalAmount":  request.Amount,
		"originalCurrency": request.FromCurrency,
		"convertedAmount": convertedAmount,
		"targetCurrency":  request.ToCurrency,
		"date":            request.Date,
		"useLatest":       request.Date == "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetHistoricalRates retrieves historical exchange rates for a currency pair
func (h *ExchangeRateHandler) GetHistoricalRates(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	baseCurrency := query.Get("base")
	targetCurrency := query.Get("target")
	startDateStr := query.Get("startDate")
	endDateStr := query.Get("endDate")

	// Validate currency codes
	if baseCurrency == "" || targetCurrency == "" {
		http.Error(w, "Both base and target currency codes are required", http.StatusBadRequest)
		return
	}

	if !h.exchangeRateService.ValidateCurrencyCode(baseCurrency) {
		http.Error(w, fmt.Sprintf("Invalid base currency code: %s", baseCurrency), http.StatusBadRequest)
		return
	}
	if !h.exchangeRateService.ValidateCurrencyCode(targetCurrency) {
		http.Error(w, fmt.Sprintf("Invalid target currency code: %s", targetCurrency), http.StatusBadRequest)
		return
	}

	// Parse dates
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid start date format: %v", err), http.StatusBadRequest)
			return
		}
	} else {
		// Default to 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid end date format: %v", err), http.StatusBadRequest)
			return
		}
	} else {
		// Default to today
		endDate = time.Now()
	}

	// Validate date range
	if startDate.After(endDate) {
		http.Error(w, "Start date must be before end date", http.StatusBadRequest)
		return
	}

	// Limit to 365 days max
	if endDate.Sub(startDate).Hours() > 365*24 {
		http.Error(w, "Date range cannot exceed 365 days", http.StatusBadRequest)
		return
	}

	rates, err := h.exchangeRateService.GetHistoricalRates(r.Context(), baseCurrency, targetCurrency, startDate, endDate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get historical rates: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate statistics
	var minRate, maxRate, avgRate float64
	if len(rates) > 0 {
		minRate = rates[0].Rate
		maxRate = rates[0].Rate
		total := 0.0

		for _, rate := range rates {
			if rate.Rate < minRate {
				minRate = rate.Rate
			}
			if rate.Rate > maxRate {
				maxRate = rate.Rate
			}
			total += rate.Rate
		}
		avgRate = total / float64(len(rates))
	}

	response := map[string]interface{}{
		"success":         true,
		"baseCurrency":    baseCurrency,
		"targetCurrency":  targetCurrency,
		"startDate":       startDate.Format(time.RFC3339),
		"endDate":         endDate.Format(time.RFC3339),
		"rates":           rates,
		"count":           len(rates),
		"statistics": map[string]interface{}{
			"minRate": minRate,
			"maxRate": maxRate,
			"avgRate": avgRate,
			"range":   maxRate - minRate,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSupportedCurrencies retrieves all supported currency codes
func (h *ExchangeRateHandler) GetSupportedCurrencies(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	currencies, err := h.exchangeRateService.GetSupportedCurrencies(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get supported currencies: %v", err), http.StatusInternalServerError)
		return
	}

	// Group currencies by region/category
	currencyGroups := map[string][]string{
		"major": {"USD", "EUR", "GBP", "JPY", "CHF", "CAD", "AUD", "NZD"},
		"latin_america": {"ARS", "BRL", "MXN", "CLP", "COP", "PEN"},
		"asia": {"CNY", "INR", "KRW", "SGD", "HKD", "TWD"},
		"europe": {"CHF", "SEK", "NOK", "DKK", "PLN", "CZK", "HUF"},
		"middle_east": {"AED", "SAR", "QAR", "OMR", "KWD"},
		"africa": {"ZAR", "EGP", "NGN", "KES"},
	}

	// Filter to only include supported currencies
	filteredGroups := make(map[string][]string)
	for group, currenciesInGroup := range currencyGroups {
		var supported []string
		for _, currency := range currenciesInGroup {
			for _, supportedCurrency := range currencies {
				if currency == supportedCurrency {
					supported = append(supported, currency)
					break
				}
			}
		}
		if len(supported) > 0 {
			filteredGroups[group] = supported
		}
	}

	response := map[string]interface{}{
		"success":    true,
		"currencies": currencies,
		"groups":     filteredGroups,
		"count":      len(currencies),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateRates updates exchange rates from an external API
func (h *ExchangeRateHandler) UpdateRates(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		Rates map[string]float64 `json:"rates"`
		Date  string             `json:"date,omitempty"`
		Source string            `json:"source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if len(request.Rates) == 0 {
		http.Error(w, "No rates provided", http.StatusBadRequest)
		return
	}

	if request.Source == "" {
		request.Source = "manual"
	}

	// Parse date or use current date
	var date time.Time
	if request.Date != "" {
		var err error
		date, err = time.Parse(time.RFC3339, request.Date)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid date format: %v", err), http.StatusBadRequest)
			return
		}
	} else {
		date = time.Now()
	}

	// Validate currency codes in rates
	for currency := range request.Rates {
		if !h.exchangeRateService.ValidateCurrencyCode(currency) {
			http.Error(w, fmt.Sprintf("Invalid currency code in rates: %s", currency), http.StatusBadRequest)
			return
		}
	}

	// Update rates
	if err := h.exchangeRateService.UpdateRatesFromAPI(r.Context(), request.Rates, date, request.Source); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update rates: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Updated %d exchange rates", len(request.Rates)),
		"date":    date.Format(time.RFC3339),
		"source":  request.Source,
		"count":   len(request.Rates),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRatesByDate retrieves all exchange rates for a specific date
func (h *ExchangeRateHandler) GetRatesByDate(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	dateStr := query.Get("date")

	var date time.Time
	var err error

	if dateStr != "" {
		date, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid date format: %v", err), http.StatusBadRequest)
			return
		}
	} else {
		date = time.Now()
	}

	rates, err := h.exchangeRateService.GetRatesByDate(r.Context(), date)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get rates by date: %v", err), http.StatusInternalServerError)
		return
	}

	// Organize rates by base currency
	ratesByBase := make(map[string][]interface{})
	for _, rate := range rates {
		rateData := map[string]interface{}{
			"id":             rate.ID,
			"targetCurrency": rate.TargetCurrency,
			"rate":           rate.Rate,
			"source":         rate.Source,
			"createdAt":      rate.CreatedAt,
		}
		ratesByBase[rate.BaseCurrency] = append(ratesByBase[rate.BaseCurrency], rateData)
	}

	response := map[string]interface{}{
		"success": true,
		"date":    date.Format(time.RFC3339),
		"rates":   ratesByBase,
		"count":   len(rates),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRatesByCurrency retrieves all exchange rates for a specific base currency
func (h *ExchangeRateHandler) GetRatesByCurrency(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	baseCurrency := vars["currency"]

	if !h.exchangeRateService.ValidateCurrencyCode(baseCurrency) {
		http.Error(w, fmt.Sprintf("Invalid currency code: %s", baseCurrency), http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	limitStr := query.Get("limit")
	limit := 100 // Default limit

	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 1000 {
			limit = val
		}
	}

	rates, err := h.exchangeRateService.GetRatesByCurrency(r.Context(), baseCurrency)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get rates by currency: %v", err), http.StatusInternalServerError)
		return
	}

	// Apply limit
	if len(rates) > limit {
		rates = rates[:limit]
	}

	// Group by target currency with latest rate
	latestRates := make(map[string]interface{})
	for _, rate := range rates {
		rateData := map[string]interface{}{
			"rate":      rate.Rate,
			"date":      rate.Date,
			"source":    rate.Source,
			"createdAt": rate.CreatedAt,
		}
		latestRates[rate.TargetCurrency] = rateData
	}

	response := map[string]interface{}{
		"success":       true,
		"baseCurrency":  baseCurrency,
		"rates":         latestRates,
		"targetCount":   len(latestRates),
		"totalRates":    len(rates),
		"limitApplied":  len(rates) > limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}