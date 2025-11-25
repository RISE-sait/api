package services

import (
	"api/internal/di"
	dbUser "api/internal/domains/user/persistence/sqlc/generated"
	"api/internal/domains/user/persistence/repositories"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type CustomerCreditService struct {
	repo *repositories.CustomerCreditRepository
}

func NewCustomerCreditService(container *di.Container) *CustomerCreditService {
	return &CustomerCreditService{
		repo: repositories.NewCustomerCreditRepository(container),
	}
}

// GetCustomerCredits retrieves the current credit balance for a customer
func (s *CustomerCreditService) GetCustomerCredits(ctx context.Context, customerID uuid.UUID) (int32, *errLib.CommonError) {
	return s.repo.GetCustomerCredits(ctx, customerID)
}

// EnrollWithCredits attempts to enroll a customer in an event using credits
func (s *CustomerCreditService) EnrollWithCredits(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	return s.repo.ExecuteInTransaction(ctx, func(txRepo *repositories.CustomerCreditRepository) *errLib.CommonError {
		// Check if customer is already enrolled in this event (prevent duplicate payments)
		isEnrolled, err := txRepo.IsCustomerEnrolledInEvent(ctx, eventID, customerID)
		if err != nil {
			return err
		}
		if isEnrolled {
			return errLib.New("Already enrolled in this event", http.StatusConflict)
		}

		// Get event credit cost
		creditCost, err := txRepo.GetEventCreditCost(ctx, eventID)
		if err != nil {
			return err
		}
		if creditCost == nil {
			return errLib.New("Event does not accept credit payments", http.StatusBadRequest)
		}

		// Check if customer has sufficient credits
		hasSufficient, err := txRepo.HasSufficientCredits(ctx, customerID, *creditCost)
		if err != nil {
			return err
		}
		if !hasSufficient {
			return errLib.New("Insufficient credits", http.StatusBadRequest)
		}

		// Check weekly credit limit
		canUseCredits, err := txRepo.CanUseCreditsWithinWeeklyLimit(ctx, customerID, *creditCost)
		if err != nil {
			return err
		}
		if !canUseCredits {
			return errLib.New("Weekly credit limit exceeded", http.StatusBadRequest)
		}

		// Deduct credits
		if err := txRepo.DeductCredits(ctx, customerID, *creditCost); err != nil {
			return err
		}

		// Log the transaction
		description := "Event enrollment payment"
		if err := txRepo.LogCreditTransaction(
			ctx,
			customerID,
			-*creditCost, // negative amount for deduction
			dbUser.CreditTransactionTypeEnrollment,
			&eventID,
			description,
		); err != nil {
			log.Printf("Failed to log credit transaction: %v", err)
			// Continue even if logging fails - the main operation succeeded
		}

		// Update weekly usage tracking
		log.Printf("Updating weekly usage for customer %s, credits: %d", customerID, *creditCost)
		if err := txRepo.UpdateWeeklyUsage(ctx, customerID, *creditCost); err != nil {
			log.Printf("Failed to update weekly usage tracking: %v", err)
			// Continue even if weekly tracking fails - the main operation succeeded
		} else {
			log.Printf("Successfully updated weekly usage")
		}

		return nil
	})
}

// RefundCreditsForCancellation refunds credits when a customer cancels their enrollment
func (s *CustomerCreditService) RefundCreditsForCancellation(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	return s.repo.ExecuteInTransaction(ctx, func(txRepo *repositories.CustomerCreditRepository) *errLib.CommonError {
		// Get event credit cost to know how much to refund
		creditCost, err := txRepo.GetEventCreditCost(ctx, eventID)
		if err != nil {
			return err
		}
		if creditCost == nil {
			// Event doesn't have credit pricing, no refund needed
			return nil
		}

		// Refund credits
		if err := txRepo.RefundCredits(ctx, customerID, *creditCost); err != nil {
			return err
		}

		// Log the refund transaction
		description := "Refund for cancelled event enrollment"
		if err := txRepo.LogCreditTransaction(
			ctx,
			customerID,
			*creditCost, // positive amount for refund
			dbUser.CreditTransactionTypeRefund,
			&eventID,
			description,
		); err != nil {
			log.Printf("Failed to log credit refund transaction: %v", err)
			// Continue even if logging fails - the main operation succeeded
		}

		return nil
	})
}

// AddCredits adds credits to a customer's account (admin function)
func (s *CustomerCreditService) AddCredits(ctx context.Context, customerID uuid.UUID, amount int32, description string) *errLib.CommonError {
	return s.repo.ExecuteInTransaction(ctx, func(txRepo *repositories.CustomerCreditRepository) *errLib.CommonError {
		// Ensure customer credit record exists (will create with 0 balance if not exists)
		if err := txRepo.EnsureCustomerCreditsExist(ctx, customerID); err != nil {
			return err
		}

		// Add credits (using refund method for positive addition)
		if err := txRepo.RefundCredits(ctx, customerID, amount); err != nil {
			return err
		}

		// Log the transaction
		if err := txRepo.LogCreditTransaction(
			ctx,
			customerID,
			amount, // positive amount for addition
			dbUser.CreditTransactionTypeAdminAdjustment,
			nil, // no event associated
			description,
		); err != nil {
			return err
		}

		return nil
	})
}

// DeductCredits removes credits from a customer's account (admin function)
func (s *CustomerCreditService) DeductCredits(ctx context.Context, customerID uuid.UUID, amount int32, description string) *errLib.CommonError {
	return s.repo.ExecuteInTransaction(ctx, func(txRepo *repositories.CustomerCreditRepository) *errLib.CommonError {
		// Check if customer has sufficient credits
		hasSufficient, err := txRepo.HasSufficientCredits(ctx, customerID, amount)
		if err != nil {
			return err
		}
		if !hasSufficient {
			return errLib.New("Insufficient credits for deduction", http.StatusBadRequest)
		}

		// Deduct credits
		if err := txRepo.DeductCredits(ctx, customerID, amount); err != nil {
			return err
		}

		// Log the transaction
		if err := txRepo.LogCreditTransaction(
			ctx,
			customerID,
			-amount, // negative amount for deduction
			dbUser.CreditTransactionTypeAdminAdjustment,
			nil, // no event associated
			description,
		); err != nil {
			return err
		}

		return nil
	})
}

// GetCustomerCreditTransactions retrieves credit transaction history with pagination
func (s *CustomerCreditService) GetCustomerCreditTransactions(ctx context.Context, customerID uuid.UUID, limit, offset int32) ([]dbUser.UsersCreditTransaction, *errLib.CommonError) {
	return s.repo.GetCustomerCreditTransactions(ctx, customerID, limit, offset)
}

// GetEventCreditTransactions retrieves all credit transactions for a specific event (admin function)
func (s *CustomerCreditService) GetEventCreditTransactions(ctx context.Context, eventID uuid.UUID) ([]dbUser.GetEventCreditTransactionsRow, *errLib.CommonError) {
	return s.repo.GetEventCreditTransactions(ctx, eventID)
}

// GetEventCreditCost retrieves the credit cost for an event
func (s *CustomerCreditService) GetEventCreditCost(ctx context.Context, eventID uuid.UUID) (*int32, *errLib.CommonError) {
	return s.repo.GetEventCreditCost(ctx, eventID)
}

// UpdateEventCreditCost updates the credit cost for an event (admin function)
func (s *CustomerCreditService) UpdateEventCreditCost(ctx context.Context, eventID uuid.UUID, creditCost *int32) *errLib.CommonError {
	return s.repo.UpdateEventCreditCost(ctx, eventID, creditCost)
}

// ValidateEventCreditPayment checks if a customer can pay for an event with credits
func (s *CustomerCreditService) ValidateEventCreditPayment(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	// Check if event accepts credit payments
	creditCost, err := s.repo.GetEventCreditCost(ctx, eventID)
	if err != nil {
		return err
	}
	if creditCost == nil {
		return errLib.New("Event does not accept credit payments", http.StatusBadRequest)
	}

	// Check if customer has sufficient credits
	hasSufficient, err := s.repo.HasSufficientCredits(ctx, customerID, *creditCost)
	if err != nil {
		return err
	}
	if !hasSufficient {
		return errLib.New("Insufficient credits", http.StatusBadRequest)
	}

	return nil
}