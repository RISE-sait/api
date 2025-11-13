package jobs

import (
	"context"
	"database/sql"
	"log"
	"time"

	"api/internal/di"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
)

// MembershipReconciliationJob syncs membership status with Stripe
// This catches cases where webhooks failed or were missed
type MembershipReconciliationJob struct {
	db *sql.DB
}

// NewMembershipReconciliationJob creates a new reconciliation job
func NewMembershipReconciliationJob(container *di.Container) *MembershipReconciliationJob {
	return &MembershipReconciliationJob{
		db: container.DB,
	}
}

// Name returns the job name
func (j *MembershipReconciliationJob) Name() string {
	return "MembershipReconciliation"
}

// Interval returns how often this job runs (every 15 minutes)
func (j *MembershipReconciliationJob) Interval() time.Duration {
	return 15 * time.Minute
}

// Run executes the reconciliation logic
func (j *MembershipReconciliationJob) Run(ctx context.Context) error {
	log.Printf("[RECONCILIATION] Starting membership reconciliation")

	// Get all customers with active or inactive memberships that use Stripe
	rows, err := j.db.QueryContext(ctx, `
		SELECT
			u.id as customer_id,
			u.stripe_customer_id,
			cmp.status as db_status,
			cmp.renewal_date,
			cmp.id as membership_id
		FROM users.users u
		INNER JOIN users.customer_membership_plans cmp ON u.id = cmp.customer_id
		WHERE u.stripe_customer_id IS NOT NULL
		  AND cmp.subscription_source = 'stripe'
		  AND cmp.status IN ('active', 'inactive')
		ORDER BY cmp.updated_at ASC
		LIMIT 100 -- Process 100 at a time to avoid timeout
	`)
	if err != nil {
		log.Printf("[RECONCILIATION] Failed to query memberships: %v", err)
		return err
	}
	defer rows.Close()

	var (
		fixed         int
		checked       int
		errors        int
		drifts        []string
	)

	for rows.Next() {
		var (
			customerID       uuid.UUID
			stripeCustomerID string
			dbStatus         string
			renewalDate      sql.NullTime
			membershipID     uuid.UUID
		)

		if err := rows.Scan(&customerID, &stripeCustomerID, &dbStatus, &renewalDate, &membershipID); err != nil {
			log.Printf("[RECONCILIATION] Failed to scan row: %v", err)
			errors++
			continue
		}

		checked++

		// Check Stripe subscription status
		stripeStatus, err := j.getStripeSubscriptionStatus(stripeCustomerID)
		if err != nil {
			log.Printf("[RECONCILIATION] Failed to get Stripe status for customer %s: %v", customerID, err)
			errors++
			continue
		}

		// Check for drift
		expectedDBStatus := j.mapStripeStatusToDBStatus(stripeStatus)

		// Also check if membership is past renewal date
		if renewalDate.Valid && time.Now().After(renewalDate.Time) && dbStatus == "active" {
			log.Printf("[RECONCILIATION] ⚠️  DRIFT DETECTED: Membership %s is past renewal date (%s) but still active",
				membershipID, renewalDate.Time.Format(time.RFC3339))
			expectedDBStatus = "expired"
		}

		if dbStatus != expectedDBStatus {
			log.Printf("[RECONCILIATION] ⚠️  DRIFT DETECTED: Customer %s - DB status: %s, Expected: %s (Stripe: %s)",
				customerID, dbStatus, expectedDBStatus, stripeStatus)

			drifts = append(drifts, customerID.String())

			// Fix the drift
			if err := j.updateMembershipStatus(ctx, customerID, expectedDBStatus); err != nil {
				log.Printf("[RECONCILIATION] Failed to fix drift for customer %s: %v", customerID, err)
				errors++
				continue
			}

			log.Printf("[RECONCILIATION] ✅ Fixed drift for customer %s: %s → %s", customerID, dbStatus, expectedDBStatus)
			fixed++
		}
	}

	log.Printf("[RECONCILIATION] Summary: checked=%d, fixed=%d, errors=%d", checked, fixed, errors)

	if len(drifts) > 0 {
		log.Printf("[RECONCILIATION] ⚠️  Drifts detected for customers: %v", drifts)
	}

	// TODO: Add Prometheus metrics
	// metrics.ReconciliationChecked.Add(float64(checked))
	// metrics.ReconciliationFixed.Add(float64(fixed))
	// metrics.ReconciliationErrors.Add(float64(errors))

	return nil
}

// getStripeSubscriptionStatus fetches the subscription status from Stripe
func (j *MembershipReconciliationJob) getStripeSubscriptionStatus(stripeCustomerID string) (string, error) {
	// Get customer's subscriptions from Stripe
	params := &stripe.CustomerParams{}
	params.AddExpand("subscriptions")

	cust, err := customer.Get(stripeCustomerID, params)
	if err != nil {
		return "", err
	}

	// If no subscriptions, return canceled
	if cust.Subscriptions == nil || len(cust.Subscriptions.Data) == 0 {
		return "canceled", nil
	}

	// Get the first active subscription
	for _, sub := range cust.Subscriptions.Data {
		if sub.Status == stripe.SubscriptionStatusActive ||
		   sub.Status == stripe.SubscriptionStatusTrialing ||
		   sub.Status == stripe.SubscriptionStatusPastDue {
			return string(sub.Status), nil
		}
	}

	// If all subscriptions are canceled/incomplete
	return "canceled", nil
}

// mapStripeStatusToDBStatus maps Stripe subscription status to DB status
func (j *MembershipReconciliationJob) mapStripeStatusToDBStatus(stripeStatus string) string {
	switch stripeStatus {
	case "active", "trialing":
		return "active"
	case "past_due", "unpaid":
		return "inactive"
	case "canceled", "incomplete", "incomplete_expired":
		return "expired"
	default:
		log.Printf("[RECONCILIATION] Unknown Stripe status: %s, defaulting to inactive", stripeStatus)
		return "inactive"
	}
}

// updateMembershipStatus updates the membership status in the database
func (j *MembershipReconciliationJob) updateMembershipStatus(ctx context.Context, customerID uuid.UUID, status string) error {
	query := `
		UPDATE users.customer_membership_plans
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE customer_id = $2 AND subscription_source = 'stripe'
	`

	result, err := j.db.ExecContext(ctx, query, status, customerID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		log.Printf("[RECONCILIATION] Warning: No rows updated for customer %s", customerID)
	}

	return nil
}
