package services

import (
	"context"
	"database/sql"
	"log"
	"time"

	"api/internal/di"
	dbUser "api/internal/domains/user/persistence/sqlc/generated"
	dbMembership "api/internal/domains/membership/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"net/http"

	"github.com/google/uuid"
)

type CreditService struct {
	UserQueries       *dbUser.Queries
	MembershipQueries *dbMembership.Queries
}

func NewCreditService(container *di.Container) *CreditService {
	return &CreditService{
		UserQueries:       container.Queries.UserDb,
		MembershipQueries: container.Queries.MembershipDb,
	}
}

// GetCurrentWeekStart returns the Monday of the current week (ISO week)
func (s *CreditService) GetCurrentWeekStart() time.Time {
	now := time.Now()
	
	// Get the weekday (0 = Sunday, 1 = Monday, ..., 6 = Saturday)
	weekday := now.Weekday()
	
	// Calculate days since Monday (ISO week starts on Monday)
	daysSinceMonday := int(weekday-time.Monday) % 7
	if daysSinceMonday < 0 {
		daysSinceMonday += 7
	}
	
	// Get Monday of this week
	monday := now.AddDate(0, 0, -daysSinceMonday)
	
	// Return Monday at midnight
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// CanUseCredits checks if a customer can use the specified number of credits
// without exceeding their credit package's weekly limit
func (s *CreditService) CanUseCredits(ctx context.Context, customerID uuid.UUID, creditsToUse int32) (bool, *errLib.CommonError) {
	weekStart := s.GetCurrentWeekStart()

	result, err := s.UserQueries.CheckWeeklyCreditLimit(ctx, dbUser.CheckWeeklyCreditLimitParams{
		CustomerID:    customerID,
		WeekStartDate: weekStart,
		CreditsUsed:   creditsToUse,
	})

	if err != nil {
		// If no credit package found, assume they can't use credits
		return false, errLib.New("No active credit package found", http.StatusBadRequest)
	}

	return result.CanUseCredits, nil
}

// GetWeeklyUsage returns the customer's current weekly credit usage and their limit
func (s *CreditService) GetWeeklyUsage(ctx context.Context, customerID uuid.UUID) (int32, *int32, *errLib.CommonError) {
	weekStart := s.GetCurrentWeekStart()

	// Get current usage
	usage, err := s.UserQueries.GetWeeklyCreditsUsed(ctx, dbUser.GetWeeklyCreditsUsedParams{
		CustomerID:    customerID,
		WeekStartDate: weekStart,
	})

	if err != nil {
		usage = 0 // No usage record exists yet
	}

	// Get customer's active credit package
	activePackage, err := s.UserQueries.GetCustomerActiveCreditPackage(ctx, customerID)
	if err != nil {
		log.Printf("Failed to get active credit package for customer %s: %v", customerID, err)
		return usage, nil, errLib.New("No active credit package found", http.StatusBadRequest)
	}
	log.Printf("Found active credit package for customer %s: weekly_limit=%v", customerID, activePackage.WeeklyCreditLimit)

	// Return usage and weekly limit from credit package
	weeklyLimit := &activePackage.WeeklyCreditLimit

	return usage, weeklyLimit, nil
}

// UseCredits deducts credits and updates weekly usage tracking
func (s *CreditService) UseCredits(ctx context.Context, customerID uuid.UUID, creditsToUse int32, description string) *errLib.CommonError {
	// Check if they can use the credits
	canUse, err := s.CanUseCredits(ctx, customerID, creditsToUse)
	if err != nil {
		return err
	}
	
	if !canUse {
		return errLib.New("Weekly credit limit exceeded", http.StatusBadRequest)
	}
	
	// Deduct from customer's credit balance
	rowsAffected, dbErr := s.UserQueries.DeductCredits(ctx, dbUser.DeductCreditsParams{
		CustomerID: customerID,
		Credits:    creditsToUse,
	})
	
	if dbErr != nil {
		return errLib.New("Failed to deduct credits", http.StatusInternalServerError)
	}
	
	if rowsAffected == 0 {
		return errLib.New("Insufficient credits", http.StatusBadRequest)
	}
	
	// Update weekly usage tracking
	weekStart := s.GetCurrentWeekStart()
	dbErr = s.UserQueries.UpdateWeeklyCreditsUsed(ctx, dbUser.UpdateWeeklyCreditsUsedParams{
		CustomerID:    customerID,
		WeekStartDate: weekStart,
		CreditsUsed:   creditsToUse,
	})
	
	if dbErr != nil {
		// Log the error but don't fail the transaction - the credits were already deducted
		// In production, you might want to implement compensating actions
		return errLib.New("Failed to update weekly usage tracking", http.StatusInternalServerError)
	}
	
	return nil
}

// AllocateCreditsOnMembershipPurchase awards credits when a customer purchases a membership
func (s *CreditService) AllocateCreditsOnMembershipPurchase(ctx context.Context, customerID uuid.UUID, membershipPlanID uuid.UUID) *errLib.CommonError {
	// Get the membership plan to see how many credits to allocate
	plan, err := s.MembershipQueries.GetMembershipPlanById(ctx, membershipPlanID)
	if err != nil {
		return errLib.New("Membership plan not found", http.StatusNotFound)
	}
	
	// Only allocate credits if this is a credit-based membership
	if plan.CreditAllocation.Valid && plan.CreditAllocation.Int32 > 0 {
		// Add credits to customer's account
		_, err = s.UserQueries.RefundCredits(ctx, dbUser.RefundCreditsParams{
			CustomerID: customerID,
			Credits:    plan.CreditAllocation.Int32,
		})
		
		if err != nil {
			return errLib.New("Failed to allocate membership credits", http.StatusInternalServerError)
		}
		
		// Log the credit transaction
		description := "Credits allocated for membership purchase"
		
		err = s.UserQueries.LogCreditTransaction(ctx, dbUser.LogCreditTransactionParams{
			CustomerID:      customerID,
			Amount:          plan.CreditAllocation.Int32,
			TransactionType: dbUser.CreditTransactionTypeEnrollment,
			EventID:         uuid.NullUUID{Valid: false}, // No event associated
			Description:     sql.NullString{String: description, Valid: true},
		})
		
		if err != nil {
			// Log error but don't fail - credits were already allocated
			return errLib.New("Failed to log credit transaction", http.StatusInternalServerError)
		}
	}
	
	return nil
}

// getUserActiveMembershipPlanID gets the customer's active membership plan ID
func (s *CreditService) getUserActiveMembershipPlanID(ctx context.Context, customerID uuid.UUID) (uuid.UUID, error) {
	// This should be a simple query to users.customer_membership_plans table
	// We'll add this query to the user database queries
	rows, err := s.UserQueries.GetActiveCustomerMembershipPlanID(ctx, customerID)
	if err != nil {
		return uuid.Nil, err
	}
	return rows, nil
}