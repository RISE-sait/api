package enrollment

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	errLib "api/internal/libs/errors"
	"github.com/google/uuid"
)

// UpdateStripeSubscriptionStatus updates the status of all Stripe subscriptions for a customer
func (r *CustomerEnrollmentRepository) UpdateStripeSubscriptionStatus(ctx context.Context, customerID uuid.UUID, status string) *errLib.CommonError {
	query := `UPDATE users.customer_membership_plans
			  SET status = $1, updated_at = CURRENT_TIMESTAMP
			  WHERE customer_id = $2 AND subscription_source = 'stripe'`

	result, err := r.Db.ExecContext(ctx, query, status, customerID)
	if err != nil {
		log.Printf("Failed to update Stripe subscription status: %v", err)
		return errLib.New("Failed to update subscription status", http.StatusInternalServerError)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Updated %d subscription(s) to status '%s' for customer %s", rowsAffected, status, customerID)

	return nil
}

// UpdateStripeSubscriptionStatusByID updates the status of a specific Stripe subscription.
// If eventTime is non-zero, it guards against out-of-order webhook events by only updating
// when the event is newer than the last processed event.
func (r *CustomerEnrollmentRepository) UpdateStripeSubscriptionStatusByID(ctx context.Context, customerID uuid.UUID, stripeSubscriptionID string, status string, eventTime ...time.Time) *errLib.CommonError {
	var result sql.Result
	var err error

	if len(eventTime) > 0 && !eventTime[0].IsZero() {
		et := eventTime[0]
		query := `UPDATE users.customer_membership_plans
				  SET status = $1, updated_at = CURRENT_TIMESTAMP, last_stripe_event_at = $4
				  WHERE customer_id = $2 AND stripe_subscription_id = $3 AND subscription_source = 'stripe'
				  AND (last_stripe_event_at IS NULL OR last_stripe_event_at < $4)`
		result, err = r.Db.ExecContext(ctx, query, status, customerID, stripeSubscriptionID, et)
	} else {
		query := `UPDATE users.customer_membership_plans
				  SET status = $1, updated_at = CURRENT_TIMESTAMP
				  WHERE customer_id = $2 AND stripe_subscription_id = $3 AND subscription_source = 'stripe'`
		result, err = r.Db.ExecContext(ctx, query, status, customerID, stripeSubscriptionID)
	}

	if err != nil {
		log.Printf("Failed to update Stripe subscription status by ID: %v", err)
		return errLib.New("Failed to update subscription status", http.StatusInternalServerError)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 && len(eventTime) > 0 && !eventTime[0].IsZero() {
		log.Printf("[WEBHOOK] Stale event rejected: status '%s' for customer %s (subscription: %s) — a newer event was already processed", status, customerID, stripeSubscriptionID)
	} else {
		log.Printf("Updated %d subscription(s) to status '%s' for customer %s (subscription: %s)", rowsAffected, status, customerID, stripeSubscriptionID)
	}

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

// UpdateStripeSubscriptionStatusAndNextBilling updates both the status and next billing date for a customer's Stripe subscription
func (r *CustomerEnrollmentRepository) UpdateStripeSubscriptionStatusAndNextBilling(ctx context.Context, customerID uuid.UUID, status string, nextBillingDate time.Time) *errLib.CommonError {
	query := `UPDATE users.customer_membership_plans
			  SET status = $1, next_billing_date = $2, updated_at = CURRENT_TIMESTAMP
			  WHERE customer_id = $3 AND subscription_source = 'stripe'`

	result, err := r.Db.ExecContext(ctx, query, status, sql.NullTime{Time: nextBillingDate, Valid: !nextBillingDate.IsZero()}, customerID)
	if err != nil {
		log.Printf("Failed to update Stripe subscription status and next billing date: %v", err)
		return errLib.New("Failed to update subscription", http.StatusInternalServerError)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Updated %d subscription(s) to status '%s' with next billing %s for customer %s",
		rowsAffected, status, nextBillingDate.Format(time.RFC3339), customerID)

	return nil
}

// UpdateStripeSubscriptionStatusByIDAndNextBilling updates status and next billing date for a specific Stripe subscription.
// If eventTime is non-zero, it guards against out-of-order webhook events.
func (r *CustomerEnrollmentRepository) UpdateStripeSubscriptionStatusByIDAndNextBilling(ctx context.Context, customerID uuid.UUID, stripeSubscriptionID string, status string, nextBillingDate time.Time, eventTime ...time.Time) *errLib.CommonError {
	var result sql.Result
	var err error

	if len(eventTime) > 0 && !eventTime[0].IsZero() {
		et := eventTime[0]
		query := `UPDATE users.customer_membership_plans
				  SET status = $1, next_billing_date = $2, updated_at = CURRENT_TIMESTAMP, last_stripe_event_at = $5
				  WHERE customer_id = $3 AND stripe_subscription_id = $4 AND subscription_source = 'stripe'
				  AND (last_stripe_event_at IS NULL OR last_stripe_event_at < $5)`
		result, err = r.Db.ExecContext(ctx, query, status, sql.NullTime{Time: nextBillingDate, Valid: !nextBillingDate.IsZero()}, customerID, stripeSubscriptionID, et)
	} else {
		query := `UPDATE users.customer_membership_plans
				  SET status = $1, next_billing_date = $2, updated_at = CURRENT_TIMESTAMP
				  WHERE customer_id = $3 AND stripe_subscription_id = $4 AND subscription_source = 'stripe'`
		result, err = r.Db.ExecContext(ctx, query, status, sql.NullTime{Time: nextBillingDate, Valid: !nextBillingDate.IsZero()}, customerID, stripeSubscriptionID)
	}

	if err != nil {
		log.Printf("Failed to update Stripe subscription status and next billing by ID: %v", err)
		return errLib.New("Failed to update subscription", http.StatusInternalServerError)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 && len(eventTime) > 0 && !eventTime[0].IsZero() {
		log.Printf("[WEBHOOK] Stale event rejected: status '%s' with next billing %s for customer %s (subscription: %s) — a newer event was already processed",
			status, nextBillingDate.Format(time.RFC3339), customerID, stripeSubscriptionID)
	} else {
		log.Printf("Updated %d subscription(s) to status '%s' with next billing %s for customer %s (subscription: %s)",
			rowsAffected, status, nextBillingDate.Format(time.RFC3339), customerID, stripeSubscriptionID)
	}

	return nil
}