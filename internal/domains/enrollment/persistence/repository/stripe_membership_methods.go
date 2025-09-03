package enrollment

import (
	"context"
	"log"
	"net/http"

	errLib "api/internal/libs/errors"
	"github.com/google/uuid"
)

// UpdateStripeSubscriptionStatus updates the status of all Stripe subscriptions for a customer
func (r *CustomerEnrollmentRepository) UpdateStripeSubscriptionStatus(ctx context.Context, customerID uuid.UUID, status string) *errLib.CommonError {
	query := `UPDATE users.customer_membership_plans 
			  SET status = $1, updated_at = CURRENT_TIMESTAMP
			  WHERE customer_id = $2 AND subscription_source = 'subscription'`
	
	result, err := r.Db.ExecContext(ctx, query, status, customerID)
	if err != nil {
		log.Printf("Failed to update Stripe subscription status: %v", err)
		return errLib.New("Failed to update subscription status", http.StatusInternalServerError)
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("Updated %d subscription(s) to status '%s' for customer %s", rowsAffected, status, customerID)
	
	return nil
}

// GetStripeSubscriptionByCustomerID gets the most recent Stripe subscription for a customer
func (r *CustomerEnrollmentRepository) GetStripeSubscriptionByCustomerID(ctx context.Context, customerID uuid.UUID) (*uuid.UUID, *errLib.CommonError) {
	query := `SELECT membership_plan_id 
			  FROM users.customer_membership_plans
			  WHERE customer_id = $1 AND subscription_source = 'subscription'
			  ORDER BY created_at DESC
			  LIMIT 1`
	
	var planID uuid.UUID
	err := r.Db.QueryRowContext(ctx, query, customerID).Scan(&planID)
	
	if err != nil {
		log.Printf("No Stripe subscription found for customer %s: %v", customerID, err)
		return nil, errLib.New("No Stripe subscription found for customer", http.StatusNotFound)
	}
	
	return &planID, nil
}