package repositories

import (
	"api/internal/di"
	dbUser "api/internal/domains/user/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

type CustomerCreditRepository struct {
	queries *dbUser.Queries
	db      *sql.DB
}

func NewCustomerCreditRepository(container *di.Container) *CustomerCreditRepository {
	return &CustomerCreditRepository{
		queries: dbUser.New(container.DB),
		db:      container.DB,
	}
}

// WithTx creates a new repository instance with transaction support
func (r *CustomerCreditRepository) WithTx(tx *sql.Tx) *CustomerCreditRepository {
	return &CustomerCreditRepository{
		queries: r.queries.WithTx(tx),
		db:      r.db,
	}
}

// GetCustomerCredits retrieves the current credit balance for a customer
func (r *CustomerCreditRepository) GetCustomerCredits(ctx context.Context, customerID uuid.UUID) (int32, *errLib.CommonError) {
	credits, err := r.queries.GetCustomerCredits(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Customer doesn't have credit record yet - return 0 without creating
			return 0, nil
		}
		log.Printf("Error getting customer credits: %v", err)
		return 0, errLib.New("Failed to retrieve customer credits", http.StatusInternalServerError)
	}
	return credits, nil
}

// EnsureCustomerCreditsExist ensures a credit record exists for the customer, creating one with 0 balance if needed
func (r *CustomerCreditRepository) EnsureCustomerCreditsExist(ctx context.Context, customerID uuid.UUID) *errLib.CommonError {
	_, err := r.queries.GetCustomerCredits(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create record with 0 balance
			if createErr := r.queries.CreateCustomerCredits(ctx, dbUser.CreateCustomerCreditsParams{
				CustomerID: customerID,
				Credits:    0,
			}); createErr != nil {
				log.Printf("Error creating customer credits record: %v", createErr)
				return errLib.New("Failed to initialize customer credits", http.StatusInternalServerError)
			}
			return nil
		}
		log.Printf("Error checking customer credits: %v", err)
		return errLib.New("Failed to check customer credits", http.StatusInternalServerError)
	}
	return nil
}

// HasSufficientCredits checks if customer has enough credits for a transaction
func (r *CustomerCreditRepository) HasSufficientCredits(ctx context.Context, customerID uuid.UUID, amount int32) (bool, *errLib.CommonError) {
	result, err := r.queries.CheckCustomerHasSufficientCredits(ctx, dbUser.CheckCustomerHasSufficientCreditsParams{
		CustomerID: customerID,
		Credits:    amount,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Customer doesn't have credits record, they have insufficient credits
			return false, nil
		}
		log.Printf("Error checking customer credits: %v", err)
		return false, errLib.New("Failed to check customer credits", http.StatusInternalServerError)
	}
	return result, nil
}

// DeductCredits removes credits from customer's account (only if they have sufficient balance)
func (r *CustomerCreditRepository) DeductCredits(ctx context.Context, customerID uuid.UUID, amount int32) *errLib.CommonError {
	rowsAffected, err := r.queries.DeductCredits(ctx, dbUser.DeductCreditsParams{
		CustomerID: customerID,
		Credits:    amount,
	})
	if err != nil {
		log.Printf("Error deducting credits: %v", err)
		return errLib.New("Failed to deduct credits", http.StatusInternalServerError)
	}
	if rowsAffected == 0 {
		return errLib.New("Insufficient credits", http.StatusBadRequest)
	}
	return nil
}

// RefundCredits adds credits back to customer's account
func (r *CustomerCreditRepository) RefundCredits(ctx context.Context, customerID uuid.UUID, amount int32) *errLib.CommonError {
	rowsAffected, err := r.queries.RefundCredits(ctx, dbUser.RefundCreditsParams{
		CustomerID: customerID,
		Credits:    amount,
	})
	if err != nil {
		log.Printf("Error refunding credits: %v", err)
		return errLib.New("Failed to refund credits", http.StatusInternalServerError)
	}
	if rowsAffected == 0 {
		return errLib.New("Customer credits record not found", http.StatusNotFound)
	}
	return nil
}

// LogCreditTransaction records a credit transaction for audit purposes
func (r *CustomerCreditRepository) LogCreditTransaction(ctx context.Context, customerID uuid.UUID, amount int32, transactionType dbUser.CreditTransactionType, eventID *uuid.UUID, description string) *errLib.CommonError {
	var eventIDParam uuid.NullUUID
	if eventID != nil {
		eventIDParam = uuid.NullUUID{UUID: *eventID, Valid: true}
	}

	params := dbUser.LogCreditTransactionParams{
		CustomerID:      customerID,
		Amount:          amount,
		TransactionType: transactionType,
		EventID:         eventIDParam,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
	}

	if err := r.queries.LogCreditTransaction(ctx, params); err != nil {
		log.Printf("Error logging credit transaction: %v", err)
		return errLib.New("Failed to log credit transaction", http.StatusInternalServerError)
	}
	return nil
}

// GetCustomerCreditTransactions retrieves credit transaction history for a customer with pagination
func (r *CustomerCreditRepository) GetCustomerCreditTransactions(ctx context.Context, customerID uuid.UUID, limit, offset int32) ([]dbUser.UsersCreditTransaction, *errLib.CommonError) {
	transactions, err := r.queries.GetCustomerCreditTransactions(ctx, dbUser.GetCustomerCreditTransactionsParams{
		CustomerID: customerID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		log.Printf("Error getting customer credit transactions: %v", err)
		return nil, errLib.New("Failed to retrieve credit transactions", http.StatusInternalServerError)
	}
	return transactions, nil
}

// GetEventCreditTransactions retrieves all credit transactions for a specific event (admin function)
func (r *CustomerCreditRepository) GetEventCreditTransactions(ctx context.Context, eventID uuid.UUID) ([]dbUser.GetEventCreditTransactionsRow, *errLib.CommonError) {
	eventIDParam := uuid.NullUUID{UUID: eventID, Valid: true}
	transactions, err := r.queries.GetEventCreditTransactions(ctx, eventIDParam)
	if err != nil {
		log.Printf("Error getting event credit transactions: %v", err)
		return nil, errLib.New("Failed to retrieve event credit transactions", http.StatusInternalServerError)
	}
	return transactions, nil
}

// GetEventCreditCost retrieves the credit cost for an event
func (r *CustomerCreditRepository) GetEventCreditCost(ctx context.Context, eventID uuid.UUID) (*int32, *errLib.CommonError) {
	creditCost, err := r.queries.GetEventCreditCost(ctx, eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Event not found", http.StatusNotFound)
		}
		log.Printf("Error getting event credit cost: %v", err)
		return nil, errLib.New("Failed to retrieve event credit cost", http.StatusInternalServerError)
	}
	
	if !creditCost.Valid {
		return nil, nil // Event doesn't have credit pricing
	}
	
	cost := creditCost.Int32
	return &cost, nil
}

// UpdateEventCreditCost updates the credit cost for an event (admin function)
func (r *CustomerCreditRepository) UpdateEventCreditCost(ctx context.Context, eventID uuid.UUID, creditCost *int32) *errLib.CommonError {
	var costParam sql.NullInt32
	if creditCost != nil {
		costParam = sql.NullInt32{Int32: *creditCost, Valid: true}
	}

	rowsAffected, err := r.queries.UpdateEventCreditCost(ctx, dbUser.UpdateEventCreditCostParams{
		ID:         eventID,
		CreditCost: costParam,
	})
	if err != nil {
		log.Printf("Error updating event credit cost: %v", err)
		return errLib.New("Failed to update event credit cost", http.StatusInternalServerError)
	}
	if rowsAffected == 0 {
		return errLib.New("Event not found", http.StatusNotFound)
	}
	return nil
}

// ExecuteInTransaction executes multiple credit operations in a transaction
func (r *CustomerCreditRepository) ExecuteInTransaction(ctx context.Context, fn func(*CustomerCreditRepository) *errLib.CommonError) *errLib.CommonError {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		log.Printf("Failed to begin credit transaction: %v", err)
		return errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("Credit transaction rollback error: %v", err)
		}
	}()

	if txErr := fn(r.WithTx(tx)); txErr != nil {
		return txErr
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit credit transaction: %v", err)
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}
	return nil
}

// UpdateWeeklyUsage updates the weekly credit usage tracking
func (r *CustomerCreditRepository) UpdateWeeklyUsage(ctx context.Context, customerID uuid.UUID, creditsUsed int32) *errLib.CommonError {
	// Calculate current week start (Monday of the current week)
	now := time.Now()
	weekday := now.Weekday()
	daysSinceMonday := int(weekday-time.Monday) % 7
	if daysSinceMonday < 0 {
		daysSinceMonday += 7
	}
	monday := now.AddDate(0, 0, -daysSinceMonday)
	weekStart := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())

	err := r.queries.UpdateWeeklyCreditsUsed(ctx, dbUser.UpdateWeeklyCreditsUsedParams{
		CustomerID:    customerID,
		WeekStartDate: weekStart,
		CreditsUsed:   creditsUsed,
	})

	if err != nil {
		return errLib.New("Failed to update weekly usage", http.StatusInternalServerError)
	}

	return nil
}

// CanUseCreditsWithinWeeklyLimit checks if customer can use credits without exceeding weekly limit
func (r *CustomerCreditRepository) CanUseCreditsWithinWeeklyLimit(ctx context.Context, customerID uuid.UUID, creditsToUse int32) (bool, *errLib.CommonError) {
	// Calculate current week start (Monday of the current week)
	now := time.Now()
	weekday := now.Weekday()
	daysSinceMonday := int(weekday-time.Monday) % 7
	if daysSinceMonday < 0 {
		daysSinceMonday += 7
	}
	monday := now.AddDate(0, 0, -daysSinceMonday)
	weekStart := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())

	result, err := r.queries.CheckWeeklyCreditLimit(ctx, dbUser.CheckWeeklyCreditLimitParams{
		CustomerID:    customerID,
		WeekStartDate: weekStart,
		CreditsUsed:   creditsToUse,
	})

	if err != nil {
		log.Printf("Error checking weekly credit limit for customer %s: %v", customerID, err)
		return false, errLib.New("Failed to check weekly credit limit", http.StatusInternalServerError)
	}

	return result.CanUseCredits, nil
}