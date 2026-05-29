package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/google/uuid"
)

// ExchangeRateServiceImpl implements ExchangeRateService
type ExchangeRateServiceImpl struct {
	exchangeRateRepo repository.ExchangeRateRepository
	rateCache        sync.Map // Cache for exchange rates: currencyPair -> *model.ExchangeRate
	cacheMutex       sync.RWMutex
}

// NewExchangeRateService creates a new ExchangeRateService
func NewExchangeRateService(exchangeRateRepo repository.ExchangeRateRepository) ExchangeRateService {
	return &ExchangeRateServiceImpl{
		exchangeRateRepo: exchangeRateRepo,
		rateCache:        sync.Map{},
	}
}

// AddExchangeRate adds a new exchange rate
func (s *ExchangeRateServiceImpl) AddExchangeRate(ctx context.Context, rate *model.ExchangeRate) error {
	// Validate exchange rate
	if err := s.validateExchangeRate(rate); err != nil {
		return fmt.Errorf("exchange rate validation failed: %w", err)
	}

	// Set timestamps if not set
	now := time.Now().UTC()
	if rate.CreatedAt.IsZero() {
		rate.CreatedAt = now
	}
	rate.UpdatedAt = now

	// Generate ID if not set
	if rate.ID == uuid.Nil {
		rate.ID = uuid.New()
	}

	// Add to cache
	cacheKey := s.GetCurrencyPairKey(rate.BaseCurrency, rate.TargetCurrency)
	s.rateCache.Store(cacheKey, rate)

	return s.exchangeRateRepo.Create(ctx, rate)
}

// GetExchangeRate retrieves an exchange rate by ID
func (s *ExchangeRateServiceImpl) GetExchangeRate(ctx context.Context, id uuid.UUID) (*model.ExchangeRate, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid exchange rate ID")
	}

	return s.exchangeRateRepo.FindByID(ctx, id)
}

// UpdateExchangeRate updates an existing exchange rate
func (s *ExchangeRateServiceImpl) UpdateExchangeRate(ctx context.Context, rate *model.ExchangeRate) error {
	if rate.ID == uuid.Nil {
		return fmt.Errorf("exchange rate ID is required")
	}

	// Validate exchange rate
	if err := s.validateExchangeRate(rate); err != nil {
		return fmt.Errorf("exchange rate validation failed: %w", err)
	}

	// Update timestamp
	rate.UpdatedAt = time.Now().UTC()

	// Update cache
	cacheKey := s.GetCurrencyPairKey(rate.BaseCurrency, rate.TargetCurrency)
	s.rateCache.Store(cacheKey, rate)

	return s.exchangeRateRepo.Update(ctx, rate)
}

// DeleteExchangeRate deletes an exchange rate
func (s *ExchangeRateServiceImpl) DeleteExchangeRate(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid exchange rate ID")
	}

	// Get rate first to remove from cache
	rate, err := s.exchangeRateRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find exchange rate: %w", err)
	}

	// Remove from cache
	cacheKey := s.GetCurrencyPairKey(rate.BaseCurrency, rate.TargetCurrency)
	s.rateCache.Delete(cacheKey)

	return s.exchangeRateRepo.Delete(ctx, id)
}

// GetLatestRate retrieves the latest exchange rate for a currency pair
func (s *ExchangeRateServiceImpl) GetLatestRate(ctx context.Context, baseCurrency, targetCurrency string) (*model.ExchangeRate, error) {
	// Validate currencies
	if !s.ValidateCurrencyCode(baseCurrency) {
		return nil, fmt.Errorf("invalid base currency: %s", baseCurrency)
	}
	if !s.ValidateCurrencyCode(targetCurrency) {
		return nil, fmt.Errorf("invalid target currency: %s", targetCurrency)
	}

	// Check cache first
	cacheKey := s.GetCurrencyPairKey(baseCurrency, targetCurrency)
	if cached, ok := s.rateCache.Load(cacheKey); ok {
		if rate, ok := cached.(*model.ExchangeRate); ok {
			// Check if cache is still valid (less than 1 hour old)
			if time.Since(rate.UpdatedAt) < time.Hour {
				return rate, nil
			}
		}
	}

	// Get from repository
	rate, err := s.exchangeRateRepo.FindLatestRate(ctx, baseCurrency, targetCurrency)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if rate != nil {
		s.rateCache.Store(cacheKey, rate)
	}

	return rate, nil
}

// GetRatesByDate retrieves all exchange rates for a specific date
func (s *ExchangeRateServiceImpl) GetRatesByDate(ctx context.Context, date time.Time) ([]model.ExchangeRate, error) {
	if date.IsZero() {
		return nil, fmt.Errorf("date is required")
	}

	// Don't allow future dates
	if date.After(time.Now()) {
		return nil, fmt.Errorf("date cannot be in the future")
	}

	return s.exchangeRateRepo.FindRatesByDate(ctx, date)
}

// GetRatesByCurrency retrieves all exchange rates for a base currency
func (s *ExchangeRateServiceImpl) GetRatesByCurrency(ctx context.Context, baseCurrency string) ([]model.ExchangeRate, error) {
	if !s.ValidateCurrencyCode(baseCurrency) {
		return nil, fmt.Errorf("invalid base currency: %s", baseCurrency)
	}

	return s.exchangeRateRepo.FindRatesByCurrency(ctx, baseCurrency)
}

// GetHistoricalRates retrieves historical exchange rates for a currency pair
func (s *ExchangeRateServiceImpl) GetHistoricalRates(ctx context.Context, baseCurrency, targetCurrency string, startDate, endDate time.Time) ([]model.ExchangeRate, error) {
	// Validate currencies
	if !s.ValidateCurrencyCode(baseCurrency) {
		return nil, fmt.Errorf("invalid base currency: %s", baseCurrency)
	}
	if !s.ValidateCurrencyCode(targetCurrency) {
		return nil, fmt.Errorf("invalid target currency: %s", targetCurrency)
	}

	// Validate date range
	if startDate.IsZero() || endDate.IsZero() {
		return nil, fmt.Errorf("start date and end date are required")
	}
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	// Limit date range to 1 year max
	maxRange := 365 * 24 * time.Hour
	if endDate.Sub(startDate) > maxRange {
		return nil, fmt.Errorf("date range cannot exceed 1 year")
	}

	return s.exchangeRateRepo.FindHistoricalRates(ctx, baseCurrency, targetCurrency, startDate, endDate)
}

// ConvertAmount converts an amount from one currency to another
func (s *ExchangeRateServiceImpl) ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string, date time.Time) (float64, error) {
	// Validate amount
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be positive")
	}

	// Validate currencies
	if !s.ValidateCurrencyCode(fromCurrency) {
		return 0, fmt.Errorf("invalid from currency: %s", fromCurrency)
	}
	if !s.ValidateCurrencyCode(toCurrency) {
		return 0, fmt.Errorf("invalid to currency: %s", toCurrency)
	}

	// Same currency, no conversion needed
	if strings.EqualFold(fromCurrency, toCurrency) {
		return amount, nil
	}

	// Get exchange rate with fallback
	rate, err := s.GetExchangeRateWithFallback(ctx, fromCurrency, toCurrency, date)
	if err != nil {
		return 0, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	// Convert amount
	convertedAmount := amount * rate.Rate
	return convertedAmount, nil
}

// GetSupportedCurrencies returns list of supported currencies
func (s *ExchangeRateServiceImpl) GetSupportedCurrencies(ctx context.Context) ([]string, error) {
	return s.exchangeRateRepo.GetSupportedCurrencies(ctx)
}

// UpdateRatesFromAPI updates exchange rates from an external API
func (s *ExchangeRateServiceImpl) UpdateRatesFromAPI(ctx context.Context, rates map[string]float64, date time.Time, source string) error {
	if len(rates) == 0 {
		return fmt.Errorf("no rates provided")
	}

	if date.IsZero() {
		date = time.Now().UTC()
	}

	if source == "" {
		source = "api"
	}

	// Validate all rates
	for pair, rate := range rates {
		parts := strings.Split(pair, "_")
		if len(parts) != 2 {
			return fmt.Errorf("invalid currency pair format: %s. Expected format: USD_ARS", pair)
		}

		baseCurrency := strings.ToUpper(parts[0])
		targetCurrency := strings.ToUpper(parts[1])

		if !s.ValidateCurrencyCode(baseCurrency) {
			return fmt.Errorf("invalid base currency in pair %s: %s", pair, baseCurrency)
		}
		if !s.ValidateCurrencyCode(targetCurrency) {
			return fmt.Errorf("invalid target currency in pair %s: %s", pair, targetCurrency)
		}
		if rate <= 0 {
			return fmt.Errorf("invalid rate for pair %s: %f", pair, rate)
		}
	}

	return s.exchangeRateRepo.UpdateRatesFromAPI(ctx, rates, date, source)
}

// ValidateCurrencyCode validates a currency code
func (s *ExchangeRateServiceImpl) ValidateCurrencyCode(currency string) bool {
	if len(currency) != 3 {
		return false
	}

	// Check if all characters are uppercase letters
	for _, ch := range currency {
		if ch < 'A' || ch > 'Z' {
			return false
		}
	}

	// Supported currencies (ARS, USD, EUR for now)
	supportedCurrencies := map[string]bool{
		"ARS": true, // Argentine Peso
		"USD": true, // US Dollar
		"EUR": true, // Euro
	}

	return supportedCurrencies[currency]
}

// GetExchangeRateWithFallback gets exchange rate with fallback to inverse rate
func (s *ExchangeRateServiceImpl) GetExchangeRateWithFallback(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (*model.ExchangeRate, error) {
	// Try to get direct rate
	rate, err := s.GetLatestRate(ctx, baseCurrency, targetCurrency)
	if err == nil && rate != nil {
		return rate, nil
	}

	// If direct rate not found, try inverse rate
	inverseRate, err := s.GetLatestRate(ctx, targetCurrency, baseCurrency)
	if err == nil && inverseRate != nil {
		// Create inverse rate
		inverse := &model.ExchangeRate{
			ID:            uuid.New(),
			BaseCurrency:  baseCurrency,
			TargetCurrency: targetCurrency,
			Rate:          s.CalculateInverseRate(inverseRate.Rate),
			Date:          inverseRate.Date,
			Source:        inverseRate.Source,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		return inverse, nil
	}

	return nil, fmt.Errorf("no exchange rate found for %s to %s", baseCurrency, targetCurrency)
}

// CalculateInverseRate calculates the inverse of an exchange rate
func (s *ExchangeRateServiceImpl) CalculateInverseRate(rate float64) float64 {
	if rate == 0 {
		return 0
	}
	return 1 / rate
}

// GetCurrencyPairKey creates a cache key for a currency pair
func (s *ExchangeRateServiceImpl) GetCurrencyPairKey(baseCurrency, targetCurrency string) string {
	return strings.ToUpper(baseCurrency) + "_" + strings.ToUpper(targetCurrency)
}

// validateExchangeRate validates an exchange rate
func (s *ExchangeRateServiceImpl) validateExchangeRate(rate *model.ExchangeRate) error {
	if rate == nil {
		return fmt.Errorf("exchange rate cannot be nil")
	}

	// Validate currencies
	if !s.ValidateCurrencyCode(rate.BaseCurrency) {
		return fmt.Errorf("invalid base currency: %s", rate.BaseCurrency)
	}
	if !s.ValidateCurrencyCode(rate.TargetCurrency) {
		return fmt.Errorf("invalid target currency: %s", rate.TargetCurrency)
	}

	// Same currency not allowed
	if strings.EqualFold(rate.BaseCurrency, rate.TargetCurrency) {
		return fmt.Errorf("base and target currencies cannot be the same")
	}

	// Validate rate
	if rate.Rate <= 0 {
		return fmt.Errorf("exchange rate must be positive")
	}

	// Validate date
	if rate.Date.IsZero() {
		return fmt.Errorf("date is required")
	}
	if rate.Date.After(time.Now().Add(24 * time.Hour)) {
		return fmt.Errorf("date cannot be in the future")
	}

	// Validate source
	if rate.Source == "" {
		return fmt.Errorf("source is required")
	}
	if len(rate.Source) > 50 {
		return fmt.Errorf("source cannot exceed 50 characters")
	}

	return nil
}