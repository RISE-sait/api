package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"api/internal/di"
	creditPackageRepo "api/internal/domains/credit_package/persistence/repository"
	enrollment "api/internal/domains/enrollment/service"
	postCheckoutRepo "api/internal/domains/payment/persistence/repositories"
	stripeService "api/internal/domains/payment/services/stripe"
	userServices "api/internal/domains/user/services"
	"api/internal/libs/logger"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscription"
)

// CheckoutReconciliationJob catches paid checkout sessions that weren't processed by webhooks
// This is the safety net for when webhooks fail
type CheckoutReconciliationJob struct {
	db                     *sql.DB
	postCheckoutRepository *postCheckoutRepo.PostCheckoutRepository
	enrollmentService      *enrollment.CustomerEnrollmentService
	creditPackageRepo      *creditPackageRepo.CreditPackageRepository
	customerCreditService  *userServices.CustomerCreditService
	logger                 *logger.StructuredLogger
}

// NewCheckoutReconciliationJob creates a new checkout reconciliation job
func NewCheckoutReconciliationJob(container *di.Container) *CheckoutReconciliationJob {
	return &CheckoutReconciliationJob{
		db:                     container.DB,
		postCheckoutRepository: postCheckoutRepo.NewPostCheckoutRepository(container),
		enrollmentService:      enrollment.NewCustomerEnrollmentService(container),
		creditPackageRepo:      creditPackageRepo.NewCreditPackageRepository(container),
		customerCreditService:  userServices.NewCustomerCreditService(container),
		logger:                 logger.WithComponent("checkout-reconciliation"),
	}
}

// Name returns the job name
func (j *CheckoutReconciliationJob) Name() string {
	return "CheckoutReconciliation"
}

// Interval returns how often this job runs (every 30 minutes)
func (j *CheckoutReconciliationJob) Interval() time.Duration {
	return 30 * time.Minute
}

// Run executes the reconciliation logic
func (j *CheckoutReconciliationJob) Run(ctx context.Context) error {
	j.logger.Info("Starting checkout reconciliation job")

	// Look back 24 hours for missed checkouts
	sinceTime := time.Now().Add(-24 * time.Hour)

	// Get completed checkout sessions from Stripe
	sessions, err := stripeService.ListRecentCheckoutSessions(sinceTime, 100)
	if err != nil {
		j.logger.Error("Failed to list checkout sessions from Stripe", err)
		return err
	}

	var (
		checked       int
		reconciled    int
		skipped       int
		errorCount    int
		failedSessions []string // Track failed session IDs for summary alert
	)

	for _, session := range sessions {
		checked++

		// Only process completed sessions with paid status
		if session.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
			skipped++
			continue
		}

		// Get user ID from metadata
		userIDStr := session.Metadata["userID"]
		if userIDStr == "" {
			j.logger.WithFields(map[string]interface{}{
				"session_id": session.ID,
			}).Warn("Session has no userID metadata, skipping")
			skipped++
			continue
		}

		userID, parseErr := uuid.Parse(userIDStr)
		if parseErr != nil {
			j.logger.WithFields(map[string]interface{}{
				"session_id": session.ID,
				"user_id":    userIDStr,
			}).Warn("Invalid userID in session metadata")
			skipped++
			continue
		}

		// Check if this checkout was already processed
		processed, checkErr := j.wasCheckoutProcessed(ctx, session, userID)
		if checkErr != nil {
			j.logger.WithFields(map[string]interface{}{
				"session_id": session.ID,
				"user_id":    userID,
			}).Warn("Failed to check if checkout was processed")
			errorCount++
			continue
		}

		if processed {
			// Already processed, skip
			skipped++
			continue
		}

		// This checkout was missed - reconcile it now
		j.logger.WithFields(map[string]interface{}{
			"session_id": session.ID,
			"user_id":    userID,
			"mode":       session.Mode,
		}).Warn("MISSED CHECKOUT DETECTED - reconciling now")

		reconcileErr := j.reconcileCheckout(ctx, session, userID)
		if reconcileErr != nil {
			j.logger.WithFields(map[string]interface{}{
				"session_id": session.ID,
				"user_id":    userID,
				"error":      reconcileErr.Error(),
			}).Warn("Failed to reconcile checkout - will retry next run")
			errorCount++
			failedSessions = append(failedSessions, session.ID)
		} else {
			j.logger.WithFields(map[string]interface{}{
				"session_id": session.ID,
				"user_id":    userID,
				"product":    j.getProductType(session),
			}).Info("Successfully reconciled missed checkout")
			reconciled++
		}
	}

	j.logger.WithFields(map[string]interface{}{
		"checked":    checked,
		"reconciled": reconciled,
		"skipped":    skipped,
		"errors":     errorCount,
	}).Info("Checkout reconciliation job completed")

	// Send ONE summary Slack alert only if there were reconciled checkouts or persistent failures
	if reconciled > 0 || errorCount > 0 {
		j.sendSummaryAlert(reconciled, errorCount, failedSessions)
	}

	return nil
}

// wasCheckoutProcessed checks if a checkout session was already processed
func (j *CheckoutReconciliationJob) wasCheckoutProcessed(ctx context.Context, session *stripe.CheckoutSession, userID uuid.UUID) (bool, error) {
	// For subscriptions, check if user has a membership with this subscription ID
	if session.Mode == stripe.CheckoutSessionModeSubscription {
		// First, check by Stripe subscription ID (most reliable)
		if session.Subscription != nil && session.Subscription.ID != "" {
			exists, err := j.hasMembershipBySubscriptionID(ctx, session.Subscription.ID)
			if err == nil && exists {
				return true, nil
			}

			// Also check if the subscription was deleted/canceled in Stripe
			// If so, consider it "processed" (no need to reconcile a deleted subscription)
			sub, subErr := subscription.Get(session.Subscription.ID, nil)
			if subErr != nil {
				// Subscription no longer exists in Stripe - treat as processed (was deleted)
				j.logger.WithFields(map[string]interface{}{
					"session_id":      session.ID,
					"subscription_id": session.Subscription.ID,
				}).Info("Subscription no longer exists in Stripe, skipping")
				return true, nil
			}
			if sub.Status == stripe.SubscriptionStatusCanceled {
				// Subscription was canceled - treat as processed
				j.logger.WithFields(map[string]interface{}{
					"session_id":      session.ID,
					"subscription_id": session.Subscription.ID,
				}).Info("Subscription is canceled in Stripe, skipping")
				return true, nil
			}
		}

		// Fallback: check by membership plan ID from metadata
		membershipPlanIDStr := session.Metadata["membershipPlanID"]
		if membershipPlanIDStr == "" {
			// Can't determine, assume not processed
			return false, nil
		}

		planID, err := uuid.Parse(membershipPlanIDStr)
		if err != nil {
			return false, nil
		}

		return j.hasMembershipForPlan(ctx, userID, planID)
	}

	// For one-time payments, check based on the product type
	if session.Mode == stripe.CheckoutSessionModePayment {
		// Check line items to determine what was purchased
		if session.LineItems == nil || len(session.LineItems.Data) == 0 {
			return false, nil
		}

		for _, item := range session.LineItems.Data {
			if item.Price == nil {
				continue
			}

			// Check if this is a credit package
			pkg, _ := j.creditPackageRepo.GetByStripePriceID(ctx, item.Price.ID)
			if pkg != nil {
				return j.hasCreditPackageActive(ctx, userID, pkg.ID)
			}
		}
	}

	return false, nil
}

// hasMembershipBySubscriptionID checks if a membership exists with this Stripe subscription ID
func (j *CheckoutReconciliationJob) hasMembershipBySubscriptionID(ctx context.Context, subscriptionID string) (bool, error) {
	query := `SELECT EXISTS(
		SELECT 1 FROM users.customer_membership_plans
		WHERE stripe_subscription_id = $1
	)`

	var exists bool
	err := j.db.QueryRowContext(ctx, query, subscriptionID).Scan(&exists)
	return exists, err
}

// hasMembershipForPlan checks if user has an active membership for a plan
func (j *CheckoutReconciliationJob) hasMembershipForPlan(ctx context.Context, userID, planID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(
		SELECT 1 FROM users.customer_membership_plans
		WHERE customer_id = $1 AND membership_plan_id = $2 AND status = 'active'
	)`

	var exists bool
	err := j.db.QueryRowContext(ctx, query, userID, planID).Scan(&exists)
	return exists, err
}

// hasCreditPackageActive checks if user has an active credit package
func (j *CheckoutReconciliationJob) hasCreditPackageActive(ctx context.Context, userID, packageID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(
		SELECT 1 FROM users.customer_active_credit_package
		WHERE customer_id = $1 AND credit_package_id = $2
	)`

	var exists bool
	err := j.db.QueryRowContext(ctx, query, userID, packageID).Scan(&exists)
	return exists, err
}

// reconcileCheckout processes a missed checkout
func (j *CheckoutReconciliationJob) reconcileCheckout(ctx context.Context, session *stripe.CheckoutSession, userID uuid.UUID) error {
	eventCreatedAt := time.Unix(session.Created, 0)

	// Store Stripe customer ID if not already stored
	if session.Customer != nil && session.Customer.ID != "" {
		j.storeStripeCustomerID(userID, session.Customer.ID)
	}

	// Handle subscription checkout (membership)
	if session.Mode == stripe.CheckoutSessionModeSubscription {
		return j.reconcileMembershipCheckout(ctx, session, userID, eventCreatedAt)
	}

	// Handle one-time payment (credit package)
	if session.Mode == stripe.CheckoutSessionModePayment {
		return j.reconcileOneTimeCheckout(ctx, session, userID, eventCreatedAt)
	}

	return nil
}

// reconcileMembershipCheckout reconciles a missed membership subscription
func (j *CheckoutReconciliationJob) reconcileMembershipCheckout(ctx context.Context, session *stripe.CheckoutSession, userID uuid.UUID, eventCreatedAt time.Time) error {
	membershipPlanIDStr := session.Metadata["membershipPlanID"]
	if membershipPlanIDStr == "" {
		log.Printf("[CHECKOUT_RECONCILE] Missing membership plan ID in session metadata for session %s", session.ID)
		return nil // Can't reconcile without plan ID
	}

	planID, err := uuid.Parse(membershipPlanIDStr)
	if err != nil {
		return err
	}

	// Get amt_periods from membership plan to calculate renewal date
	amtPeriods, amtErr := j.postCheckoutRepository.GetMembershipPlanAmtPeriods(ctx, planID)
	if amtErr != nil {
		log.Printf("[CHECKOUT_RECONCILE] Warning: Could not get amt_periods for plan %s: %v", planID, amtErr)
	}

	subscriptionID := ""
	var cancelAtDateTime time.Time
	var nextBillingDate time.Time

	// Get subscription details from Stripe for proper dates
	if session.Subscription != nil {
		subscriptionID = session.Subscription.ID

		// Fetch full subscription details from Stripe
		sub, subErr := subscription.Get(subscriptionID, nil)
		if subErr != nil {
			log.Printf("[CHECKOUT_RECONCILE] Warning: Could not fetch subscription %s from Stripe: %v", subscriptionID, subErr)
		} else {
			// Get next billing date from current period end
			if sub.CurrentPeriodEnd > 0 {
				nextBillingDate = time.Unix(sub.CurrentPeriodEnd, 0)
			}

			// If plan has amt_periods, calculate the renewal/cancel date
			// This replicates the webhook's calculateCancelAt logic
			if amtPeriods != nil && *amtPeriods > 0 && len(sub.Items.Data) > 0 {
				item := sub.Items.Data[0]
				if item.Price != nil && item.Price.Recurring != nil {
					interval := item.Price.Recurring.Interval
					intervalCount := int(item.Price.Recurring.IntervalCount)
					periods := int(*amtPeriods)

					log.Printf("[CHECKOUT_RECONCILE] Calculating cancel date: interval=%s, intervalCount=%d, periods=%d", interval, intervalCount, periods)

					var cancelTime time.Time
					switch interval {
					case stripe.PriceRecurringIntervalMonth:
						cancelTime = eventCreatedAt.AddDate(0, intervalCount*periods, 0)
					case stripe.PriceRecurringIntervalYear:
						cancelTime = eventCreatedAt.AddDate(intervalCount*periods, 0, 0)
					case stripe.PriceRecurringIntervalWeek:
						cancelTime = eventCreatedAt.AddDate(0, 0, 7*intervalCount*periods)
					case stripe.PriceRecurringIntervalDay:
						cancelTime = eventCreatedAt.AddDate(0, 0, intervalCount*periods)
					default:
						log.Printf("[CHECKOUT_RECONCILE] Unsupported billing interval: %s", interval)
					}

					if !cancelTime.IsZero() {
						cancelAtDateTime = cancelTime
						log.Printf("[CHECKOUT_RECONCILE] Calculated cancel date: %v", cancelAtDateTime)

						// Update the Stripe subscription with the cancel date
						_, updateErr := subscription.Update(subscriptionID, &stripe.SubscriptionParams{
							CancelAt: stripe.Int64(cancelTime.Unix()),
						})
						if updateErr != nil {
							log.Printf("[CHECKOUT_RECONCILE] Warning: Failed to update subscription cancel date in Stripe: %v", updateErr)
						} else {
							log.Printf("[CHECKOUT_RECONCILE] Updated Stripe subscription %s with cancel_at: %v", subscriptionID, cancelTime)
						}
					}
				}
			} else if sub.CancelAt > 0 {
				// Fallback: use existing cancel date from Stripe if already set
				cancelAtDateTime = time.Unix(sub.CancelAt, 0)
			}

			log.Printf("[CHECKOUT_RECONCILE] Got subscription dates - Next billing: %v, Cancel at: %v", nextBillingDate, cancelAtDateTime)
		}
	}

	log.Printf("[CHECKOUT_RECONCILE] Enrolling customer %s in membership plan %s (subscription: %s)", userID, planID, subscriptionID)
	return j.enrollmentService.EnrollCustomerInMembershipPlan(ctx, userID, planID, cancelAtDateTime, nextBillingDate, eventCreatedAt, subscriptionID)
}

// reconcileOneTimeCheckout reconciles a missed one-time payment
func (j *CheckoutReconciliationJob) reconcileOneTimeCheckout(ctx context.Context, session *stripe.CheckoutSession, userID uuid.UUID, eventCreatedAt time.Time) error {
	if session.LineItems == nil || len(session.LineItems.Data) == 0 {
		return nil
	}

	for _, item := range session.LineItems.Data {
		if item.Price == nil {
			continue
		}

		priceID := item.Price.ID

		// Check if this is a credit package
		pkg, _ := j.creditPackageRepo.GetByStripePriceID(ctx, priceID)
		if pkg != nil {
			log.Printf("[CHECKOUT_RECONCILE] Adding %d credits to customer %s from package %s", pkg.CreditAllocation, userID, pkg.ID)

			if err := j.customerCreditService.AddCredits(ctx, userID, pkg.CreditAllocation, "Credit package purchase (reconciled)"); err != nil {
				return err
			}

			if err := j.creditPackageRepo.SetCustomerActivePackage(ctx, userID, pkg.ID, pkg.WeeklyCreditLimit); err != nil {
				return err
			}

			log.Printf("[CHECKOUT_RECONCILE] Successfully reconciled credit package %s for customer %s", pkg.ID, userID)
			return nil
		}
	}

	return nil
}

// getProductType determines the product type from the checkout session
func (j *CheckoutReconciliationJob) getProductType(session *stripe.CheckoutSession) string {
	if session.Mode == stripe.CheckoutSessionModeSubscription {
		return "membership"
	}

	// For one-time payments, check line items
	if session.LineItems != nil && len(session.LineItems.Data) > 0 {
		for _, item := range session.LineItems.Data {
			if item.Price != nil {
				// Check if this is a credit package
				pkg, _ := j.creditPackageRepo.GetByStripePriceID(context.Background(), item.Price.ID)
				if pkg != nil {
					return "credit_package"
				}
			}
		}
	}

	return "one_time_payment"
}

// storeStripeCustomerID stores the Stripe customer ID for the user
func (j *CheckoutReconciliationJob) storeStripeCustomerID(userID uuid.UUID, stripeCustomerID string) {
	query := `UPDATE users.users SET stripe_customer_id = $1 WHERE id = $2 AND (stripe_customer_id IS NULL OR stripe_customer_id = '')`
	_, err := j.db.Exec(query, stripeCustomerID, userID)
	if err != nil {
		log.Printf("[CHECKOUT_RECONCILE] Failed to store Stripe customer ID: %v", err)
	}
}

// sendSummaryAlert sends a single summary alert for the reconciliation job run
func (j *CheckoutReconciliationJob) sendSummaryAlert(reconciled int, errorCount int, failedSessions []string) {
	// Only alert if something noteworthy happened
	if reconciled == 0 && errorCount == 0 {
		return
	}

	var alertType string
	var message string

	if errorCount > 0 && reconciled == 0 {
		// Only failures
		alertType = "RECONCILIATION_FAILURE"
		message = fmt.Sprintf("Checkout reconciliation: %d failed (manual review needed)", errorCount)
	} else if errorCount > 0 {
		// Mixed results
		alertType = "RECONCILIATION_PARTIAL"
		message = fmt.Sprintf("Checkout reconciliation: %d recovered, %d failed", reconciled, errorCount)
	} else {
		// Only successes
		alertType = "RECONCILIATION_SUCCESS"
		message = fmt.Sprintf("Checkout reconciliation: %d missed checkouts recovered", reconciled)
	}

	// Build session list for failures (truncate if too many)
	sessionList := ""
	if len(failedSessions) > 0 {
		if len(failedSessions) <= 3 {
			sessionList = fmt.Sprintf(" Sessions: %v", failedSessions)
		} else {
			sessionList = fmt.Sprintf(" Sessions: %v... (+%d more)", failedSessions[:3], len(failedSessions)-3)
		}
	}

	logger.SendReconciliationAlert(logger.ReconciliationAlertDetails{
		AlertType:    alertType,
		ErrorMessage: message + sessionList,
		WasFixed:     reconciled > 0,
		Product:      "checkout_reconciliation",
	})
}
