package payment

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"api/internal/di"
	creditPackageRepo "api/internal/domains/credit_package/persistence/repository"
	enrollment "api/internal/domains/enrollment/service"
	repository "api/internal/domains/payment/persistence/repositories"
	"api/internal/domains/payment/services/stripe"
	userServices "api/internal/domains/user/services"
	errLib "api/internal/libs/errors"
	"api/internal/libs/logger"

	"github.com/google/uuid"
	stripeLib "github.com/stripe/stripe-go/v81"
)

// CheckoutVerificationService handles verification of checkout sessions
// This provides a safety net when webhooks fail or are delayed
type CheckoutVerificationService struct {
	PostCheckoutRepository *repository.PostCheckoutRepository
	EnrollmentService      *enrollment.CustomerEnrollmentService
	CreditPackageRepo      *creditPackageRepo.CreditPackageRepository
	CustomerCreditService  *userServices.CustomerCreditService
	db                     *sql.DB
	logger                 *logger.StructuredLogger
}

// NewCheckoutVerificationService creates a new checkout verification service
func NewCheckoutVerificationService(container *di.Container) *CheckoutVerificationService {
	return &CheckoutVerificationService{
		PostCheckoutRepository: repository.NewPostCheckoutRepository(container),
		EnrollmentService:      enrollment.NewCustomerEnrollmentService(container),
		CreditPackageRepo:      creditPackageRepo.NewCreditPackageRepository(container),
		CustomerCreditService:  userServices.NewCustomerCreditService(container),
		db:                     container.DB,
		logger:                 logger.WithComponent("checkout-verification"),
	}
}

// VerificationResult contains the result of a checkout verification
type VerificationResult struct {
	SessionID        string `json:"session_id"`
	PaymentStatus    string `json:"payment_status"`
	EnrollmentStatus string `json:"enrollment_status"`
	WasReconciled    bool   `json:"was_reconciled"`
	Message          string `json:"message"`
}

// VerifyCheckoutSession verifies a checkout session and ensures the customer is enrolled
// This is called by the frontend after redirect from Stripe checkout
func (s *CheckoutVerificationService) VerifyCheckoutSession(ctx context.Context, sessionID string, userID uuid.UUID) (*VerificationResult, *errLib.CommonError) {
	s.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"user_id":    userID,
	}).Info("Verifying checkout session")

	// 1. Retrieve the checkout session from Stripe
	checkoutSession, err := stripe.GetCheckoutSession(sessionID)
	if err != nil {
		s.logger.Error("Failed to retrieve checkout session from Stripe", err)
		return nil, err
	}

	// 2. Validate that this session belongs to the requesting user
	sessionUserIDStr := checkoutSession.Metadata["userID"]
	if sessionUserIDStr == "" {
		s.logger.Error("Session has no userID in metadata", nil)
		return nil, errLib.New("Invalid checkout session", http.StatusBadRequest)
	}

	sessionUserID, parseErr := uuid.Parse(sessionUserIDStr)
	if parseErr != nil || sessionUserID != userID {
		s.logger.WithFields(map[string]interface{}{
			"session_user_id":  sessionUserIDStr,
			"request_user_id":  userID,
		}).Error("User ID mismatch - potential security issue", nil)
		return nil, errLib.New("Access denied", http.StatusForbidden)
	}

	// 3. Check payment status
	result := &VerificationResult{
		SessionID:     sessionID,
		PaymentStatus: string(checkoutSession.PaymentStatus),
	}

	if checkoutSession.PaymentStatus != stripeLib.CheckoutSessionPaymentStatusPaid {
		result.EnrollmentStatus = "payment_incomplete"
		result.Message = "Payment has not been completed"
		return result, nil
	}

	// 4. Check if customer is already enrolled (webhook already processed)
	isEnrolled, enrollErr := s.checkIfAlreadyEnrolled(ctx, checkoutSession, userID)
	if enrollErr != nil {
		s.logger.Error("Failed to check enrollment status", enrollErr)
		return nil, enrollErr
	}

	if isEnrolled {
		result.EnrollmentStatus = "enrolled"
		result.WasReconciled = false
		result.Message = "Customer is already enrolled"
		return result, nil
	}

	// 5. Payment is complete but customer not enrolled - reconcile now
	s.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"user_id":    userID,
	}).Warn("Payment complete but enrollment missing - reconciling now")

	reconcileErr := s.reconcileCheckout(ctx, checkoutSession, userID)
	if reconcileErr != nil {
		s.logger.Error("Failed to reconcile checkout", reconcileErr)
		result.EnrollmentStatus = "reconciliation_failed"
		result.Message = "Payment received but enrollment failed - please contact support"
		// Alert for manual intervention
		s.alertReconciliationFailure(sessionID, userID, reconcileErr)
		return result, nil
	}

	result.EnrollmentStatus = "enrolled"
	result.WasReconciled = true
	result.Message = "Payment verified and enrollment completed"

	// Log successful reconciliation for audit
	s.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"user_id":    userID,
	}).Info("Successfully reconciled missed webhook - customer enrolled")

	return result, nil
}

// checkIfAlreadyEnrolled checks if the customer is already enrolled based on checkout type
func (s *CheckoutVerificationService) checkIfAlreadyEnrolled(ctx context.Context, session *stripeLib.CheckoutSession, userID uuid.UUID) (bool, *errLib.CommonError) {
	// Check membership plan enrollment
	if membershipPlanIDStr := session.Metadata["membershipPlanID"]; membershipPlanIDStr != "" {
		planID, err := uuid.Parse(membershipPlanIDStr)
		if err != nil {
			return false, errLib.New("Invalid membership plan ID in session", http.StatusBadRequest)
		}
		return s.checkMembershipEnrollment(ctx, userID, planID)
	}

	// Check credit package purchase by looking at price IDs
	if session.LineItems != nil && len(session.LineItems.Data) > 0 {
		for _, item := range session.LineItems.Data {
			if item.Price != nil {
				// Check if this is a credit package
				pkg, _ := s.CreditPackageRepo.GetByStripePriceID(ctx, item.Price.ID)
				if pkg != nil {
					// For credit packages, check if customer already has this package active
					return s.checkCreditPackageActive(ctx, userID, pkg.ID)
				}
			}
		}
	}

	// For programs/events, check by session metadata
	// Programs and events use reservation status, not a separate enrollment table
	return false, nil
}

// checkMembershipEnrollment checks if customer has an active membership for the plan
func (s *CheckoutVerificationService) checkMembershipEnrollment(ctx context.Context, customerID, planID uuid.UUID) (bool, *errLib.CommonError) {
	query := `SELECT EXISTS(
		SELECT 1 FROM users.customer_membership_plans
		WHERE customer_id = $1 AND membership_plan_id = $2 AND status = 'active'
	)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, customerID, planID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking membership enrollment: %v", err)
		return false, errLib.New("Failed to check enrollment status", http.StatusInternalServerError)
	}

	return exists, nil
}

// checkCreditPackageActive checks if customer has an active credit package
func (s *CheckoutVerificationService) checkCreditPackageActive(ctx context.Context, customerID, packageID uuid.UUID) (bool, *errLib.CommonError) {
	query := `SELECT EXISTS(
		SELECT 1 FROM users.customer_credit_packages
		WHERE customer_id = $1 AND credit_package_id = $2
	)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, customerID, packageID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking credit package: %v", err)
		return false, errLib.New("Failed to check credit package status", http.StatusInternalServerError)
	}

	return exists, nil
}

// reconcileCheckout processes a checkout that was missed by webhooks
func (s *CheckoutVerificationService) reconcileCheckout(ctx context.Context, session *stripeLib.CheckoutSession, userID uuid.UUID) *errLib.CommonError {
	eventCreatedAt := time.Unix(session.Created, 0)

	// Handle subscription checkout (membership)
	if session.Mode == stripeLib.CheckoutSessionModeSubscription {
		return s.reconcileMembershipCheckout(ctx, session, userID, eventCreatedAt)
	}

	// Handle one-time payment (credit package, program, event)
	if session.Mode == stripeLib.CheckoutSessionModePayment {
		return s.reconcileOneTimeCheckout(ctx, session, userID, eventCreatedAt)
	}

	return errLib.New("Unknown checkout mode", http.StatusBadRequest)
}

// reconcileMembershipCheckout reconciles a missed membership subscription checkout
func (s *CheckoutVerificationService) reconcileMembershipCheckout(ctx context.Context, session *stripeLib.CheckoutSession, userID uuid.UUID, eventCreatedAt time.Time) *errLib.CommonError {
	membershipPlanIDStr := session.Metadata["membershipPlanID"]
	if membershipPlanIDStr == "" {
		return errLib.New("Missing membership plan ID in session metadata", http.StatusBadRequest)
	}

	planID, err := uuid.Parse(membershipPlanIDStr)
	if err != nil {
		return errLib.New("Invalid membership plan ID", http.StatusBadRequest)
	}

	// Store Stripe customer ID if not already stored
	if session.Customer != nil && session.Customer.ID != "" {
		s.storeStripeCustomerID(userID, session.Customer.ID)
	}

	// Get amt_periods from membership plan to calculate renewal date
	amtPeriods, amtErr := s.PostCheckoutRepository.GetMembershipPlanAmtPeriods(ctx, planID)
	if amtErr != nil {
		log.Printf("[RECONCILE] Warning: Could not get amt_periods for plan %s: %v", planID, amtErr)
	}

	subscriptionID := ""
	var cancelAtDateTime time.Time
	var nextBillingDate time.Time

	// Get subscription details from Stripe for proper dates
	if session.Subscription != nil {
		subscriptionID = session.Subscription.ID

		// Fetch full subscription details from Stripe
		sub, subErr := stripe.GetSubscriptionDetails(subscriptionID)
		if subErr != nil {
			log.Printf("[RECONCILE] Warning: Could not fetch subscription %s from Stripe: %v", subscriptionID, subErr)
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

					log.Printf("[RECONCILE] Calculating cancel date: interval=%s, intervalCount=%d, periods=%d", interval, intervalCount, periods)

					var cancelTime time.Time
					switch interval {
					case stripeLib.PriceRecurringIntervalMonth:
						cancelTime = eventCreatedAt.AddDate(0, intervalCount*periods, 0)
					case stripeLib.PriceRecurringIntervalYear:
						cancelTime = eventCreatedAt.AddDate(intervalCount*periods, 0, 0)
					case stripeLib.PriceRecurringIntervalWeek:
						cancelTime = eventCreatedAt.AddDate(0, 0, 7*intervalCount*periods)
					case stripeLib.PriceRecurringIntervalDay:
						cancelTime = eventCreatedAt.AddDate(0, 0, intervalCount*periods)
					default:
						log.Printf("[RECONCILE] Unsupported billing interval: %s", interval)
					}

					if !cancelTime.IsZero() {
						cancelAtDateTime = cancelTime
						log.Printf("[RECONCILE] Calculated cancel date: %v", cancelAtDateTime)

						// Update the Stripe subscription with the cancel date
						if updateErr := stripe.UpdateSubscriptionCancelAt(subscriptionID, cancelTime.Unix()); updateErr != nil {
							log.Printf("[RECONCILE] Warning: Failed to update subscription cancel date in Stripe: %v", updateErr)
						} else {
							log.Printf("[RECONCILE] Updated Stripe subscription %s with cancel_at: %v", subscriptionID, cancelTime)
						}
					}
				}
			} else if sub.CancelAt > 0 {
				// Fallback: use existing cancel date from Stripe if already set
				cancelAtDateTime = time.Unix(sub.CancelAt, 0)
			}

			log.Printf("[RECONCILE] Got subscription dates - Next billing: %v, Cancel at: %v", nextBillingDate, cancelAtDateTime)
		}
	}

	// Enroll customer in membership plan
	log.Printf("[RECONCILE] Enrolling customer %s in membership plan %s (subscription: %s)", userID, planID, subscriptionID)
	if enrollErr := s.EnrollmentService.EnrollCustomerInMembershipPlan(ctx, userID, planID, cancelAtDateTime, nextBillingDate, eventCreatedAt, subscriptionID); enrollErr != nil {
		return enrollErr
	}

	log.Printf("[RECONCILE] Successfully enrolled customer %s in membership plan %s", userID, planID)
	return nil
}

// reconcileOneTimeCheckout reconciles a missed one-time payment checkout
func (s *CheckoutVerificationService) reconcileOneTimeCheckout(ctx context.Context, session *stripeLib.CheckoutSession, userID uuid.UUID, eventCreatedAt time.Time) *errLib.CommonError {
	if session.LineItems == nil || len(session.LineItems.Data) == 0 {
		return errLib.New("No line items in checkout session", http.StatusBadRequest)
	}

	for _, item := range session.LineItems.Data {
		if item.Price == nil {
			continue
		}

		priceID := item.Price.ID

		// Check if this is a credit package
		pkg, _ := s.CreditPackageRepo.GetByStripePriceID(ctx, priceID)
		if pkg != nil {
			// Process credit package purchase
			log.Printf("[RECONCILE] Adding %d credits to customer %s from package %s", pkg.CreditAllocation, userID, pkg.ID)
			if err := s.CustomerCreditService.AddCredits(ctx, userID, pkg.CreditAllocation, "Credit package purchase (reconciled)"); err != nil {
				return errLib.New("Failed to add credits: "+err.Error(), http.StatusInternalServerError)
			}

			// Set active credit package
			if err := s.CreditPackageRepo.SetCustomerActivePackage(ctx, userID, pkg.ID, pkg.WeeklyCreditLimit); err != nil {
				return errLib.New("Failed to set active package: "+err.Error(), http.StatusInternalServerError)
			}

			log.Printf("[RECONCILE] Successfully processed credit package %s for customer %s", pkg.ID, userID)
			return nil
		}

		// Check if this is a program enrollment
		programID, progErr := s.PostCheckoutRepository.GetProgramIdByStripePriceId(ctx, priceID)
		if progErr == nil && programID != uuid.Nil {
			log.Printf("[RECONCILE] Updating program reservation for customer %s, program %s", userID, programID)
			if err := s.EnrollmentService.UpdateReservationStatusInProgram(ctx, programID, userID, "paid"); err != nil {
				return err
			}
			return nil
		}

		// Check if this is an event enrollment
		eventID, evtErr := s.PostCheckoutRepository.GetEventIdByStripePriceId(ctx, priceID)
		if evtErr == nil && eventID != uuid.Nil {
			log.Printf("[RECONCILE] Updating event reservation for customer %s, event %s", userID, eventID)
			if err := s.EnrollmentService.UpdateReservationStatusInEvent(ctx, eventID, userID, "paid"); err != nil {
				return err
			}
			return nil
		}
	}

	return errLib.New("Could not identify checkout product type", http.StatusBadRequest)
}

// storeStripeCustomerID stores the Stripe customer ID for the user
func (s *CheckoutVerificationService) storeStripeCustomerID(userID uuid.UUID, stripeCustomerID string) {
	query := `UPDATE users.users SET stripe_customer_id = $1 WHERE id = $2 AND (stripe_customer_id IS NULL OR stripe_customer_id = '')`
	_, err := s.db.Exec(query, stripeCustomerID, userID)
	if err != nil {
		log.Printf("[RECONCILE] Failed to store Stripe customer ID: %v", err)
	}
}

// alertReconciliationFailure sends an alert when reconciliation fails
func (s *CheckoutVerificationService) alertReconciliationFailure(sessionID string, userID uuid.UUID, err *errLib.CommonError) {
	// Log the error
	s.logger.WithFields(map[string]interface{}{
		"session_id":  sessionID,
		"user_id":     userID,
		"error":       err.Error(),
		"alert_type":  "reconciliation_failure",
		"severity":    "critical",
	}).Error("CRITICAL: Checkout reconciliation failed - manual intervention required", err)

	// Send Slack alert
	logger.SendReconciliationAlert(logger.ReconciliationAlertDetails{
		SessionID:    sessionID,
		CustomerID:   userID.String(),
		AlertType:    "RECONCILIATION_FAILURE",
		ErrorMessage: err.Error(),
		WasFixed:     false,
		Product:      "unknown",
	})
}
