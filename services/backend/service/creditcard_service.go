package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dssr1012/pulse-expends/internal/model"
	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/google/uuid"
)

// CreditCardServiceImpl implements CreditCardService
type CreditCardServiceImpl struct {
	creditCardRepo repository.CreditCardRepository
}

// NewCreditCardService creates a new CreditCardService
func NewCreditCardService(creditCardRepo repository.CreditCardRepository) CreditCardService {
	return &CreditCardServiceImpl{
		creditCardRepo: creditCardRepo,
	}
}

// AddCreditCard adds a new credit card with validation
func (s *CreditCardServiceImpl) AddCreditCard(ctx context.Context, card *model.CreditCard) error {
	// Validate credit card
	if err := s.ValidateCreditCard(ctx, card); err != nil {
		return fmt.Errorf("credit card validation failed: %w", err)
	}

	// Set timestamps if not set
	now := time.Now().UTC()
	if card.CreatedAt.IsZero() {
		card.CreatedAt = now
	}
	card.UpdatedAt = now

	// Generate ID if not set
	if card.ID == uuid.Nil {
		card.ID = uuid.New()
	}

	// Mask card number for storage (NO PAN storage)
	card.LastFour = s.extractLastFour(card.CardNumber)
	card.CardNumber = "" // Clear full card number - NEVER store PAN

	// Set default values
	if card.Status == "" {
		card.Status = model.CardStatusActive
	}
	if card.CreditLimit <= 0 {
		card.CreditLimit = 10000.00 // Default limit
	}

	return s.creditCardRepo.Create(ctx, card)
}

// GetCreditCard retrieves a credit card by ID
func (s *CreditCardServiceImpl) GetCreditCard(ctx context.Context, id uuid.UUID) (*model.CreditCard, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid credit card ID")
	}

	card, err := s.creditCardRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Ensure card number is never returned
	card.CardNumber = ""
	return card, nil
}

// UpdateCreditCard updates an existing credit card
func (s *CreditCardServiceImpl) UpdateCreditCard(ctx context.Context, card *model.CreditCard) error {
	if card.ID == uuid.Nil {
		return fmt.Errorf("credit card ID is required")
	}

	// Get existing card to preserve last four
	existing, err := s.creditCardRepo.FindByID(ctx, card.ID)
	if err != nil {
		return fmt.Errorf("failed to find existing card: %w", err)
	}

	// Preserve last four if card number is being cleared
	if card.CardNumber == "" {
		card.LastFour = existing.LastFour
	} else {
		card.LastFour = s.extractLastFour(card.CardNumber)
	}

	// Clear card number for storage
	card.CardNumber = ""

	// Validate card (without full card number)
	if err := s.ValidateCreditCard(ctx, card); err != nil {
		return fmt.Errorf("credit card validation failed: %w", err)
	}

	// Update timestamp
	card.UpdatedAt = time.Now().UTC()

	return s.creditCardRepo.Update(ctx, card)
}

// DeleteCreditCard permanently deletes a credit card
func (s *CreditCardServiceImpl) DeleteCreditCard(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid credit card ID")
	}

	return s.creditCardRepo.Delete(ctx, id)
}

// SoftDeleteCreditCard marks a credit card as deleted
func (s *CreditCardServiceImpl) SoftDeleteCreditCard(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid credit card ID")
	}

	// Get existing card
	card, err := s.creditCardRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find card: %w", err)
	}

	// Update status to deleted
	card.Status = model.CardStatusDeleted
	card.UpdatedAt = time.Now().UTC()

	return s.creditCardRepo.Update(ctx, card)
}

// GetUserCreditCards retrieves all credit cards for a user
func (s *CreditCardServiceImpl) GetUserCreditCards(ctx context.Context, userID uuid.UUID) ([]model.CreditCard, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	cards, err := s.creditCardRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Clear card numbers from all cards
	for i := range cards {
		cards[i].CardNumber = ""
	}

	return cards, nil
}

// GetFamilyCreditCards retrieves all credit cards for a family/circle
func (s *CreditCardServiceImpl) GetFamilyCreditCards(ctx context.Context, circleID uuid.UUID) ([]model.CreditCard, error) {
	if circleID == uuid.Nil {
		return nil, fmt.Errorf("invalid circle ID")
	}

	cards, err := s.creditCardRepo.FindByCircle(ctx, circleID)
	if err != nil {
		return nil, err
	}

	// Clear card numbers from all cards
	for i := range cards {
		cards[i].CardNumber = ""
	}

	return cards, nil
}

// GetCreditCardsByBank retrieves credit cards by bank name
func (s *CreditCardServiceImpl) GetCreditCardsByBank(ctx context.Context, bankName string) ([]model.CreditCard, error) {
	if bankName == "" {
		return nil, fmt.Errorf("bank name is required")
	}

	cards, err := s.creditCardRepo.FindByBank(ctx, bankName)
	if err != nil {
		return nil, err
	}

	// Clear card numbers from all cards
	for i := range cards {
		cards[i].CardNumber = ""
	}

	return cards, nil
}

// GetActiveCreditCards retrieves active credit cards for a user
func (s *CreditCardServiceImpl) GetActiveCreditCards(ctx context.Context, userID uuid.UUID) ([]model.CreditCard, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Get all user cards
	cards, err := s.creditCardRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filter active cards
	var activeCards []model.CreditCard
	for _, card := range cards {
		if card.Status == model.CardStatusActive {
			// Clear card number
			card.CardNumber = ""
			activeCards = append(activeCards, card)
		}
	}

	return activeCards, nil
}

// GetDefaultCreditCard retrieves the default credit card for a user
func (s *CreditCardServiceImpl) GetDefaultCreditCard(ctx context.Context, userID uuid.UUID) (*model.CreditCard, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	card, err := s.creditCardRepo.FindDefaultCard(ctx, userID)
	if err != nil {
		return nil, err
	}

	if card != nil {
		// Clear card number
		card.CardNumber = ""
	}

	return card, nil
}

// UpdateCardBalance updates a credit card's balance
func (s *CreditCardServiceImpl) UpdateCardBalance(ctx context.Context, cardID uuid.UUID, newBalance float64) error {
	if cardID == uuid.Nil {
		return fmt.Errorf("invalid credit card ID")
	}

	if newBalance < 0 {
		return fmt.Errorf("balance cannot be negative")
	}

	// Get existing card
	card, err := s.creditCardRepo.FindByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("failed to find card: %w", err)
	}

	// Validate balance against credit limit
	if newBalance > card.CreditLimit {
		return fmt.Errorf("balance cannot exceed credit limit of %.2f", card.CreditLimit)
	}

	// Update balance
	card.CurrentBalance = newBalance
	card.UpdatedAt = time.Now().UTC()

	return s.creditCardRepo.UpdateBalance(ctx, cardID, newBalance)
}

// GetUpcomingPayments retrieves credit cards with upcoming payments
func (s *CreditCardServiceImpl) GetUpcomingPayments(ctx context.Context, userID uuid.UUID, days int) ([]model.CreditCard, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	if days <= 0 || days > 60 {
		return nil, fmt.Errorf("days must be between 1 and 60")
	}

	cards, err := s.creditCardRepo.GetUpcomingPayments(ctx, userID, days)
	if err != nil {
		return nil, err
	}

	// Clear card numbers from all cards
	for i := range cards {
		cards[i].CardNumber = ""
	}

	return cards, nil
}

// GetStatementPeriod retrieves the statement period for a card
func (s *CreditCardServiceImpl) GetStatementPeriod(ctx context.Context, cardID uuid.UUID) (*time.Time, *time.Time, error) {
	if cardID == uuid.Nil {
		return nil, nil, fmt.Errorf("invalid credit card ID")
	}

	return s.creditCardRepo.GetStatementPeriod(ctx, cardID)
}

// ValidateCreditCard validates a credit card
func (s *CreditCardServiceImpl) ValidateCreditCard(ctx context.Context, card *model.CreditCard) error {
	if card == nil {
		return fmt.Errorf("credit card cannot be nil")
	}

	// Validate user ID
	if card.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	// Validate circle ID
	if card.CircleID == uuid.Nil {
		return fmt.Errorf("circle ID is required")
	}

	// Validate bank name
	if card.BankName == "" {
		return fmt.Errorf("bank name is required")
	}
	if len(card.BankName) > 100 {
		return fmt.Errorf("bank name cannot exceed 100 characters")
	}

	// Validate card type
	if card.CardType == "" {
		return fmt.Errorf("card type is required")
	}
	validCardTypes := map[string]bool{
		"visa":       true,
		"mastercard": true,
		"amex":       true,
		"discover":   true,
		"other":      true,
	}
	if !validCardTypes[strings.ToLower(card.CardType)] {
		return fmt.Errorf("invalid card type. Must be: visa, mastercard, amex, discover, or other")
	}

	// Validate last four digits (if provided)
	if card.LastFour != "" {
		if len(card.LastFour) != 4 {
			return fmt.Errorf("last four digits must be exactly 4 characters")
		}
		if !regexp.MustCompile(`^\d{4}$`).MatchString(card.LastFour) {
			return fmt.Errorf("last four digits must be numeric")
		}
	}

	// Validate card number (if provided - for new cards)
	if card.CardNumber != "" {
		// Basic Luhn algorithm check
		if !s.isValidCardNumber(card.CardNumber) {
			return fmt.Errorf("invalid card number")
		}

		// Extract and set last four
		card.LastFour = s.extractLastFour(card.CardNumber)
	}

	// Validate credit limit
	if card.CreditLimit <= 0 {
		return fmt.Errorf("credit limit must be positive")
	}

	// Validate current balance
	if card.CurrentBalance < 0 {
		return fmt.Errorf("current balance cannot be negative")
	}
	if card.CurrentBalance > card.CreditLimit {
		return fmt.Errorf("current balance cannot exceed credit limit")
	}

	// Validate status
	if card.Status != "" {
		validStatuses := map[string]bool{
			model.CardStatusActive:   true,
			model.CardStatusInactive: true,
			model.CardStatusBlocked:  true,
			model.CardStatusExpired:  true,
			model.CardStatusDeleted:  true,
		}
		if !validStatuses[card.Status] {
			return fmt.Errorf("invalid card status")
		}
	}

	// Validate expiration (if provided)
	if !card.ExpirationDate.IsZero() && card.ExpirationDate.Before(time.Now()) {
		return fmt.Errorf("card is expired")
	}

	return nil
}

// MaskCardNumber masks a card number showing only last four digits
func (s *CreditCardServiceImpl) MaskCardNumber(cardNumber string) string {
	if cardNumber == "" {
		return ""
	}

	// Remove any spaces or dashes
	cleaned := strings.ReplaceAll(strings.ReplaceAll(cardNumber, " ", ""), "-", "")

	if len(cleaned) <= 4 {
		return cleaned
	}

	// Show only last 4 digits
	lastFour := cleaned[len(cleaned)-4:]
	return "•••• •••• •••• " + lastFour
}

// CalculateMinimumPayment calculates the minimum payment for a credit card
func (s *CreditCardServiceImpl) CalculateMinimumPayment(balance float64, interestRate float64) float64 {
	if balance <= 0 {
		return 0
	}

	// Minimum payment is the greater of:
	// 1. $25 (or local currency equivalent)
	// 2. 1% of balance + interest
	// 3. All past due amounts + late fees (simplified)

	minPayment := balance * 0.01 // 1% of balance

	// Add interest if rate is provided
	if interestRate > 0 {
		monthlyInterest := balance * (interestRate / 100 / 12)
		minPayment += monthlyInterest
	}

	// Ensure minimum of $25
	if minPayment < 25 {
		minPayment = 25
	}

	// Cap at balance
	if minPayment > balance {
		minPayment = balance
	}

	return minPayment
}

// CheckPaymentDueDate checks if a payment is due for a card
func (s *CreditCardServiceImpl) CheckPaymentDueDate(cardID uuid.UUID) (bool, time.Time, error) {
	// This would typically check the card's statement period
	// For now, return a simplified implementation
	dueDate := time.Now().AddDate(0, 0, 15) // 15 days from now
	isDue := time.Now().After(dueDate.AddDate(0, 0, -3)) // Due if within 3 days of due date

	return isDue, dueDate, nil
}

// Helper function to extract last four digits
func (s *CreditCardServiceImpl) extractLastFour(cardNumber string) string {
	if cardNumber == "" {
		return ""
	}

	// Remove any spaces or dashes
	cleaned := strings.ReplaceAll(strings.ReplaceAll(cardNumber, " ", ""), "-", "")

	if len(cleaned) < 4 {
		return cleaned
	}

	return cleaned[len(cleaned)-4:]
}

// Helper function to validate card number using Luhn algorithm
func (s *CreditCardServiceImpl) isValidCardNumber(cardNumber string) bool {
	// Remove any spaces or dashes
	cleaned := strings.ReplaceAll(strings.ReplaceAll(cardNumber, " ", ""), "-", "")

	// Check if all characters are digits
	if !regexp.MustCompile(`^\d+$`).MatchString(cleaned) {
		return false
	}

	// Check length (typical credit card lengths)
	if len(cleaned) < 13 || len(cleaned) > 19 {
		return false
	}

	// Luhn algorithm
	sum := 0
	alternate := false

	for i := len(cleaned) - 1; i >= 0; i-- {
		digit := int(cleaned[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}