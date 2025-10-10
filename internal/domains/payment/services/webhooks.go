package payment

import (
	"api/internal/di"
	creditPackageRepo "api/internal/domains/credit_package/persistence/repository"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	enrollment "api/internal/domains/enrollment/service"
	enrollmentRepo "api/internal/domains/enrollment/persistence/repository"
	repository "api/internal/domains/payment/persistence/repositories"
	userServices "api/internal/domains/user/services"
	errLib "api/internal/libs/errors"
	"api/internal/libs/logger"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	identityRepo "api/internal/domains/identity/persistence/repository/user"
	membershipRepo "api/internal/domains/membership/persistence/repositories"
	"api/utils/email"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/subscription"
)

type WebhookService struct {
	PostCheckoutRepository *repository.PostCheckoutRepository
	EnrollmentService      *enrollment.CustomerEnrollmentService
	EnrollmentRepo         *enrollmentRepo.CustomerEnrollmentRepository
	UserRepo               *identityRepo.UsersRepository
	PlansRepo              *membershipRepo.PlansRepository
	CreditService          *userServices.CreditService
	CustomerCreditService  *userServices.CustomerCreditService
	CreditPackageRepo      *creditPackageRepo.CreditPackageRepository
	Idempotency            *WebhookIdempotency
	logger                 *logger.StructuredLogger
	db                     *sql.DB
}

func NewWebhookService(container *di.Container) *WebhookService {
	return &WebhookService{
		PostCheckoutRepository: repository.NewPostCheckoutRepository(container),
		EnrollmentService:      enrollment.NewCustomerEnrollmentService(container),
		EnrollmentRepo:         enrollmentRepo.NewEnrollmentRepository(container),
		UserRepo:               identityRepo.NewUserRepository(container),
		PlansRepo:              membershipRepo.NewMembershipPlansRepository(container),
		CreditService:          userServices.NewCreditService(container),
		CustomerCreditService:  userServices.NewCustomerCreditService(container),
		CreditPackageRepo:      creditPackageRepo.NewCreditPackageRepository(container),
		Idempotency:            NewWebhookIdempotency(24*time.Hour, 10000), // Store events for 24 hours, max 10k events
		logger:                 logger.WithComponent("stripe-webhooks"),
		db:                     container.DB,
	}
}

func (s *WebhookService) HandleCheckoutSessionCompleted(event stripe.Event) *errLib.CommonError {
	// Get event creation time for consistent timestamps
	eventCreatedAt := time.Unix(event.Created, 0)

	webhookLogger := s.logger.WithFields(map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.Type,
		"webhook":    "checkout_session_completed",
	})

	// Check idempotency first
	if s.Idempotency.IsProcessed(event.ID) {
		webhookLogger.Info("Event already processed, skipping due to idempotency")
		return nil
	}

	var checkSession stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &checkSession); err != nil {
		webhookLogger.Error("Failed to parse checkout session from webhook data", err)
		return errLib.New("Failed to parse session", http.StatusBadRequest)
	}

	webhookLogger = webhookLogger.WithFields(map[string]interface{}{
		"session_id":   checkSession.ID,
		"session_mode": string(checkSession.Mode),
	})

	webhookLogger.Info("Processing checkout session completed webhook")

	var err *errLib.CommonError
	switch checkSession.Mode {
	case stripe.CheckoutSessionModePayment:
		err = s.handleItemCheckoutComplete(checkSession, webhookLogger, eventCreatedAt)
	case stripe.CheckoutSessionModeSubscription:
		err = s.handleSubscriptionCheckoutComplete(checkSession, eventCreatedAt)
	default:
		webhookLogger.Warn("Unhandled session mode received")
		return nil
	}

	// Mark as processed only if successful
	if err == nil {
		s.Idempotency.MarkAsProcessed(event.ID)
		webhookLogger.Info("Webhook event processed successfully")
	} else {
		webhookLogger.Error("Webhook processing failed", err)
		
		// Send comprehensive Slack alert for debugging
		s.sendEnhancedWebhookAlert(event, checkSession, err)
	}

	return err
}

func (s *WebhookService) handleItemCheckoutComplete(checkoutSession stripe.CheckoutSession, parentLogger *logger.StructuredLogger, eventCreatedAt time.Time) *errLib.CommonError {
	itemLogger := parentLogger.WithFields(map[string]interface{}{
		"handler": "item_checkout_complete",
	})
	
	itemLogger.Info("Expanding checkout session for item purchase")

	fullSession, err := s.getExpandedSession(checkoutSession.ID)
	if err != nil {
		log.Printf("getExpandedSession failed: %v", err)
		return err
	}

	log.Println("Full session expanded")

	userIDStr := fullSession.Metadata["userID"]
	log.Println("🔍 Metadata userID:", userIDStr)

	if userIDStr == "" {
		log.Println("userID not found in metadata")
		return errLib.New("userID not found in metadata", http.StatusBadRequest)
	}

	customerID, uuidErr := uuid.Parse(userIDStr)
	if uuidErr != nil {
		log.Printf("Invalid UUID: %v", uuidErr)
		return errLib.New("Invalid user ID format", http.StatusBadRequest)
	}

	priceIDs, err := s.validateLineItems(fullSession.LineItems)
	if err != nil {
		log.Printf("validateLineItems failed: %v", err)
		return err
	}

	log.Println("Price IDs:", priceIDs)

	if len(priceIDs) == 0 {
		return errLib.New("No price IDs found in line items", http.StatusBadRequest)
	}

	priceID := priceIDs[0]
	log.Println("Checking mappings for priceID:", priceID)

	// First check metadata for explicit event/program IDs
	var programID, eventID uuid.UUID
	
	if fullSession.Metadata != nil {
		if eventIDStr, exists := fullSession.Metadata["eventID"]; exists && eventIDStr != "" {
			if parsedEventID, parseErr := uuid.Parse(eventIDStr); parseErr == nil {
				eventID = parsedEventID
				log.Printf("Found eventID in metadata: %s", eventID)
			}
		}
		// Note: We could also check for programID in metadata if needed
	}
	
	// If no eventID found in metadata, fall back to database lookup
	if eventID == uuid.Nil {
		dbProgramID, err := s.PostCheckoutRepository.GetProgramIdByStripePriceId(context.Background(), priceID)
		if err != nil {
			log.Printf("GetProgramIdByStripePriceId failed: %v", err)
			return err
		}
		dbEventID, err := s.PostCheckoutRepository.GetEventIdByStripePriceId(context.Background(), priceID)
		if err != nil {
			log.Printf("GetEventIdByStripePriceId failed: %v", err)
			return err
		}
		programID = dbProgramID
		eventID = dbEventID
	}

	log.Println("Final ProgramID:", programID)
	log.Println("Final EventID:", eventID)

	// Check if this is a credit package purchase
	creditPackage, creditErr := s.CreditPackageRepo.GetByStripePriceID(context.Background(), priceID)

	switch {
	case programID != uuid.Nil && eventID != uuid.Nil:
		return errLib.New("price ID maps to both program and event", http.StatusConflict)
	case creditPackage != nil && creditErr == nil:
		// This is a credit package purchase
		log.Printf("CREDIT PACKAGE PURCHASE DETECTED - Customer: %s, Package: %s (%s)", customerID, creditPackage.ID, creditPackage.Name)

		// Add credits to customer balance
		log.Printf("Adding %d credits to customer %s balance", creditPackage.CreditAllocation, customerID)
		if err := s.CustomerCreditService.AddCredits(context.Background(), customerID, creditPackage.CreditAllocation, "Credit package purchase"); err != nil {
			log.Printf("FAILED to add credits to customer %s: %v", customerID, err)
			return errLib.New(fmt.Sprintf("failed to add credits: %v", err), http.StatusInternalServerError)
		}

		// Set active credit package (overwrites previous package)
		log.Printf("Setting active credit package for customer %s: weekly limit=%d", customerID, creditPackage.WeeklyCreditLimit)
		if err := s.CreditPackageRepo.SetCustomerActivePackage(context.Background(), customerID, creditPackage.ID, creditPackage.WeeklyCreditLimit); err != nil {
			log.Printf("FAILED to set active credit package for customer %s: %v", customerID, err)
			return errLib.New(fmt.Sprintf("failed to set active package: %v", err), http.StatusInternalServerError)
		}

		log.Printf("CREDIT PACKAGE PURCHASE COMPLETE - Customer %s: +%d credits, %d/week limit", customerID, creditPackage.CreditAllocation, creditPackage.WeeklyCreditLimit)
		return nil
	case programID == uuid.Nil && eventID == uuid.Nil:
		return errLib.New("price ID doesn't map to any program, event, or credit package", http.StatusNotFound)
	case programID != uuid.Nil:
		log.Printf("Updating program reservation for user %s and program %s", customerID, programID)
		if err = s.EnrollmentService.UpdateReservationStatusInProgram(context.Background(), programID, customerID, dbEnrollment.PaymentStatusPaid); err != nil {
			log.Printf("Failed to update program reservation: %v", err)
			return errLib.New(fmt.Sprintf("failed to update program reservation (customer: %s, program: %s): %v", customerID, programID, err), http.StatusInternalServerError)
		}
	case eventID != uuid.Nil:
		log.Printf("Updating event reservation for user %s and event %s", customerID, eventID)
		if err = s.EnrollmentService.UpdateReservationStatusInEvent(context.Background(), eventID, customerID, dbEnrollment.PaymentStatusPaid); err != nil {
			log.Printf("Failed to update event reservation: %v", err)
			return errLib.New(fmt.Sprintf("failed to update event reservation (customer: %s, event: %s): %v", customerID, eventID, err), http.StatusInternalServerError)
		}
	}

	log.Println("handleItemCheckoutComplete completed")
	return nil
}

func (s *WebhookService) handleSubscriptionCheckoutComplete(checkoutSession stripe.CheckoutSession, eventCreatedAt time.Time) *errLib.CommonError {
	webhookLogger := s.logger.WithFields(map[string]interface{}{
		"handler":    "subscription_checkout_complete",
		"session_id": checkoutSession.ID,
	})
	
	webhookLogger.Info("Starting subscription checkout completion processing")
	
	// 1. Validate and expand session
	fullSession, err := s.getExpandedSession(checkoutSession.ID)
	if err != nil {
		webhookLogger.Error("Failed to expand checkout session", err)
		return err
	}

	// 2. Validate line items and get price ID
	priceIDs, err := s.validateLineItems(fullSession.LineItems)
	if err != nil {
		webhookLogger.Error("Failed to validate line items", err)
		return err
	}

	priceID := priceIDs[0]
	webhookLogger = webhookLogger.WithFields(map[string]interface{}{
		"stripe_price_id": priceID,
	})
	
	webhookLogger.Info("Processing subscription for price ID")

	// 3. Parse metadata
	userIdStr := fullSession.Metadata["userID"]

	if userIdStr == "" {
		webhookLogger.Error("userID not found in metadata", nil)
		return errLib.New("userID not found in metadata", http.StatusBadRequest)
	}

	userID, uuidErr := uuid.Parse(userIdStr)
	if uuidErr != nil {
		webhookLogger.Error("Invalid user ID format", uuidErr)
		return errLib.New("Invalid user ID format", http.StatusBadRequest)
	}
	
	webhookLogger = webhookLogger.WithFields(map[string]interface{}{
		"customer_id": userID,
	})

	planID, amtPeriods, err := s.PostCheckoutRepository.GetMembershipPlanByStripePriceID(context.Background(), priceID)
	if err != nil {
		webhookLogger.Error("Failed to get membership plan by Stripe price ID", err)
		return err
	}

	webhookLogger = webhookLogger.WithFields(map[string]interface{}{
		"membership_plan_id": planID,
		"amt_periods": amtPeriods,
	})
	
	webhookLogger.Info("Retrieved membership plan details")

	// Store Stripe customer ID in database for future reference
	if fullSession.Customer != nil && fullSession.Customer.ID != "" {
		if err := s.storeStripeCustomerID(userID, fullSession.Customer.ID); err != nil {
			log.Printf("WARNING: Failed to store Stripe customer ID: %v", err)
			// Don't fail the entire process for this
		} else {
			log.Printf("Successfully stored Stripe customer ID %s for user %s", fullSession.Customer.ID, userID)
		}
	}

	if amtPeriods != nil {
		if err := s.processSubscriptionWithEndDate(
			fullSession.Subscription.ID,
			*amtPeriods,
			userID,
			planID,
			eventCreatedAt,
		); err != nil {
			return err
		}
	} else {
		log.Printf("Checking existing enrollment for customer %s in plan %s", userID, planID)
		
		// Check if customer is already enrolled to handle webhook retries gracefully
		if isAlreadyEnrolled, checkErr := s.isCustomerAlreadyEnrolled(context.Background(), userID, planID); checkErr != nil {
			log.Printf("ERROR: Failed to check existing enrollment: %v", checkErr)
			return checkErr
		} else if isAlreadyEnrolled {
			log.Printf("🚫 SKIPPING ENROLLMENT - Customer %s is already enrolled in plan %s", userID, planID)
		} else {
			log.Printf("🚀 STARTING ENROLLMENT - Customer %s in membership plan %s with start time %s", userID, planID, eventCreatedAt.Format(time.RFC3339))
			if err := s.EnrollmentService.EnrollCustomerInMembershipPlan(context.Background(), userID, planID, time.Time{}, eventCreatedAt); err != nil {
				log.Printf("❌ ERROR: EnrollCustomerInMembershipPlan failed: %v", err)
				return err
			}
			log.Printf("✅ SUCCESS: Enrolled customer %s in membership plan %s", userID, planID)
		}
	}
	
	// NOTE: Credits are no longer allocated with memberships - they are only available via credit packages
	// Credit allocation has been moved to credit package purchases
	
	s.sendMembershipPurchaseEmail(userID, planID)
	return nil
}

func (s *WebhookService) getExpandedSession(sessionID string) (*stripe.CheckoutSession, *errLib.CommonError) {
	params := &stripe.CheckoutSessionParams{
		Expand: []*string{
			stripe.String("line_items"),
			stripe.String("line_items.data.price"),
			stripe.String("subscription"),
			stripe.String("customer"),
		},
	}

	checkoutSession, err := session.Get(sessionID, params)
	if err != nil {
		return nil, errLib.New("Failed to retrieve session details: "+err.Error(), http.StatusInternalServerError)
	}

	log.Println("Metadata in expanded session:", checkoutSession.Metadata)

	return checkoutSession, nil
}

func (s *WebhookService) validateLineItems(lineItems *stripe.LineItemList) ([]string, *errLib.CommonError) {
	if lineItems == nil || len(lineItems.Data) == 0 {
		return nil, errLib.New("No line items found in session", http.StatusBadRequest)
	}

	var priceIds []string

	for _, datum := range lineItems.Data {
		priceIds = append(priceIds, datum.Price.ID)
	}

	return priceIds, nil
}

func (s *WebhookService) processSubscriptionWithEndDate(subscriptionID string, totalBillingPeriods int32, userID, planID uuid.UUID, eventCreatedAt time.Time) *errLib.CommonError {
	log.Printf("Processing subscription with end date: %s, periods: %d, user: %s, plan: %s", subscriptionID, totalBillingPeriods, userID, planID)
	
	sub, err := s.getExpandedSubscription(subscriptionID)
	if err != nil {
		log.Printf("ERROR: getExpandedSubscription failed: %v", err)
		return err
	}

	cancelAtUnix, cancelAtDateTime, err := s.calculateCancelAt(sub, int(totalBillingPeriods), eventCreatedAt)
	if err != nil {
		log.Printf("ERROR: calculateCancelAt failed: %v", err)
		return err
	}

	log.Printf("Checking existing enrollment for customer %s in plan %s", userID, planID)
	
	// Check if customer is already enrolled to handle webhook retries gracefully
	if isAlreadyEnrolled, checkErr := s.isCustomerAlreadyEnrolled(context.Background(), userID, planID); checkErr != nil {
		log.Printf("ERROR: Failed to check existing enrollment: %v", checkErr)
		return checkErr
	} else if isAlreadyEnrolled {
		log.Printf("🚫 SKIPPING ENROLLMENT - Customer %s is already enrolled in plan %s", userID, planID)
	} else {
		log.Printf("🚀 STARTING ENROLLMENT - Customer %s in plan %s with end date: %s, start time: %s", userID, planID, cancelAtDateTime.Format(time.RFC3339), eventCreatedAt.Format(time.RFC3339))
		if err = s.EnrollmentService.EnrollCustomerInMembershipPlan(context.Background(), userID, planID, cancelAtDateTime, eventCreatedAt); err != nil {
			log.Printf("❌ ERROR: EnrollCustomerInMembershipPlan failed in processSubscriptionWithEndDate: %v", err)
			return err
		}
		log.Printf("✅ SUCCESS: Enrolled customer %s in plan %s", userID, planID)
	}

	// NOTE: Credits are no longer allocated with memberships - they are only available via credit packages
	// Credit allocation has been moved to credit package purchases

	s.sendMembershipPurchaseEmail(userID, planID)

	// Update Stripe subscription cancel date - don't fail webhook if this fails
	// The critical enrollment has already succeeded
	if stripeErr := s.updateSubscriptionCancelAt(sub.ID, cancelAtUnix); stripeErr != nil {
		log.Printf("WARNING: Failed to update subscription cancel date for %s: %v", sub.ID, stripeErr)
		log.Printf("WARNING: Enrollment succeeded but Stripe subscription cancel date could not be set")
		// Consider implementing a background retry job for this
	} else {
		log.Printf("Successfully set cancel date for subscription %s", sub.ID)
	}

	return nil
}

func (s *WebhookService) getExpandedSubscription(subscriptionID string) (*stripe.Subscription, *errLib.CommonError) {
	params := &stripe.SubscriptionParams{
		Expand: []*string{
			stripe.String("items.data.price"),
			stripe.String("items.data.price.product"),
		},
	}

	sub, err := subscription.Get(subscriptionID, params)
	if err != nil {
		return nil, errLib.New("Failed to get subscription: "+err.Error(), http.StatusInternalServerError)
	}

	if len(sub.Items.Data) == 0 {
		return nil, errLib.New("No items in subscription", http.StatusBadRequest)
	}

	item := sub.Items.Data[0]
	if item.Price == nil || item.Price.Product == nil {
		return nil, errLib.New("Price or product information missing", http.StatusBadRequest)
	}

	if item.Price.Recurring == nil {
		return nil, errLib.New("Price is not recurring", http.StatusBadRequest)
	}

	return sub, nil
}

func (s *WebhookService) calculateCancelAt(sub *stripe.Subscription, periods int, baseTime time.Time) (cancelAtUnix int64, cancelAtDate time.Time, error *errLib.CommonError) {
	item := sub.Items.Data[0]
	interval := item.Price.Recurring.Interval
	intervalCount := int(item.Price.Recurring.IntervalCount)

	log.Printf("calculateCancelAt: interval=%s, intervalCount=%d, periods=%d", interval, intervalCount, periods)

	var cancelTime time.Time

	switch interval {
	case stripe.PriceRecurringIntervalMonth:
		cancelTime = baseTime.AddDate(0, intervalCount*periods, 0)
	case stripe.PriceRecurringIntervalYear:
		cancelTime = baseTime.AddDate(intervalCount*periods, 0, 0)
	case stripe.PriceRecurringIntervalWeek:
		cancelTime = baseTime.AddDate(0, 0, 7*intervalCount*periods)
	case stripe.PriceRecurringIntervalDay:
		cancelTime = baseTime.AddDate(0, 0, intervalCount*periods)
	default:
		log.Printf("ERROR: Unsupported billing interval: %s", interval)
		return 0, time.Time{}, errLib.New("invalid billing interval: " + string(interval), http.StatusBadRequest)
	}

	return cancelTime.Unix(), cancelTime, nil

}

func (s *WebhookService) updateSubscriptionCancelAt(subscriptionID string, cancelAt int64) *errLib.CommonError {
	// Retry logic for Stripe API calls
	maxRetries := 3
	retryDelay := time.Second
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err := subscription.Update(
			subscriptionID,
			&stripe.SubscriptionParams{
				CancelAt: stripe.Int64(cancelAt),
			},
		)
		
		if err == nil {
			if attempt > 1 {
				log.Printf("Successfully updated subscription %s on attempt %d", subscriptionID, attempt)
			}
			return nil
		}
		
		lastErr = err
		log.Printf("Attempt %d/%d failed to update subscription %s: %v", attempt, maxRetries, subscriptionID, err)
		
		// Don't retry on client errors (4xx), only on server errors and network issues
		if stripeErr, ok := err.(*stripe.Error); ok {
			if stripeErr.HTTPStatusCode < 500 && stripeErr.HTTPStatusCode >= 400 {
				log.Printf("Client error, not retrying: HTTP %d", stripeErr.HTTPStatusCode)
				break
			}
		}
		
		if attempt < maxRetries {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}
	
	return errLib.New("Failed to set cancel date after retries: "+lastErr.Error(), http.StatusInternalServerError)
}

func (s *WebhookService) sendMembershipPurchaseEmail(userID, planID uuid.UUID) {
	log.Printf("[EMAIL] Attempting to send membership purchase email for user %s, plan %s", userID, planID)

	userInfo, err := s.UserRepo.GetUserInfo(context.Background(), "", userID)
	if err != nil {
		log.Printf("[EMAIL] Failed to get user info for %s: %v", userID, err)
		return
	}

	if userInfo.Email == nil {
		log.Printf("[EMAIL] User %s has no email address", userID)
		return
	}

	plan, pErr := s.PlansRepo.GetMembershipPlanById(context.Background(), planID)
	if pErr != nil {
		log.Printf("[EMAIL] Failed to get membership plan %s: %v", planID, pErr)
		return
	}

	log.Printf("[EMAIL] Sending membership purchase email to %s for plan %s", *userInfo.Email, plan.Name)
	email.SendMembershipPurchaseEmail(*userInfo.Email, userInfo.FirstName, plan.Name)
	log.Printf("[EMAIL] Membership purchase email sent successfully to %s", *userInfo.Email)
}

// HandleSubscriptionCreated processes subscription.created events
func (s *WebhookService) HandleSubscriptionCreated(event stripe.Event) *errLib.CommonError {
	// Check idempotency first
	if s.Idempotency.IsProcessed(event.ID) {
		log.Printf("Event %s already processed, skipping", event.ID)
		return nil
	}

	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("[WEBHOOK] Failed to parse subscription created event: %v", err)
		return errLib.New("Failed to parse subscription", http.StatusBadRequest)
	}

	log.Printf("[WEBHOOK] Subscription created: %s", sub.ID)

	// Extract user ID from customer metadata
	if sub.Customer == nil || sub.Customer.Metadata == nil {
		log.Printf("[WEBHOOK] No customer metadata found for subscription %s", sub.ID)
		return nil
	}

	userIDStr, exists := sub.Customer.Metadata["userID"]
	if !exists {
		log.Printf("[WEBHOOK] No userID in customer metadata for subscription %s", sub.ID)
		return nil
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("[WEBHOOK] Invalid userID format for subscription %s: %v", sub.ID, err)
		return errLib.New("Invalid user ID format", http.StatusBadRequest)
	}

	// Update database with subscription status
	// This would involve updating the membership plan status to active
	log.Printf("[WEBHOOK] Successfully processed subscription creation for user %s", userID)
	
	// Mark as processed
	s.Idempotency.MarkAsProcessed(event.ID)
	return nil
}

// HandleSubscriptionUpdated processes subscription.updated events
func (s *WebhookService) HandleSubscriptionUpdated(event stripe.Event) *errLib.CommonError {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("[WEBHOOK] Failed to parse subscription updated event: %v", err)
		return errLib.New("Failed to parse subscription", http.StatusBadRequest)
	}

	log.Printf("[WEBHOOK] Subscription updated: %s, Status: %s", sub.ID, sub.Status)

	// Extract user ID from customer metadata
	if sub.Customer == nil || sub.Customer.Metadata == nil {
		log.Printf("[WEBHOOK] No customer metadata found for subscription %s", sub.ID)
		return nil
	}

	userIDStr, exists := sub.Customer.Metadata["userID"]
	if !exists {
		log.Printf("[WEBHOOK] No userID in customer metadata for subscription %s", sub.ID)
		return nil
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("[WEBHOOK] Invalid userID format for subscription %s: %v", sub.ID, err)
		return errLib.New("Invalid user ID format", http.StatusBadRequest)
	}

	// Handle status changes and update database
	var dbStatus string

	// Check if subscription is scheduled for cancellation (cancel_at is set)
	// Stripe keeps status as "active" until the end of the billing period,
	// but we want to mark it as canceled immediately in our database
	if sub.CancelAt > 0 {
		log.Printf("[WEBHOOK] Subscription %s is scheduled for cancellation at %d", sub.ID, sub.CancelAt)
		dbStatus = "canceled"
	} else {
		switch sub.Status {
		case stripe.SubscriptionStatusActive:
			log.Printf("[WEBHOOK] Subscription %s is now active", sub.ID)
			dbStatus = "active"
		case stripe.SubscriptionStatusPastDue:
			log.Printf("[WEBHOOK] Subscription %s is past due", sub.ID)
			dbStatus = "inactive" // Map past due to inactive
		case stripe.SubscriptionStatusCanceled:
			log.Printf("[WEBHOOK] Subscription %s is canceled", sub.ID)
			dbStatus = "canceled"
		case stripe.SubscriptionStatusUnpaid:
			log.Printf("[WEBHOOK] Subscription %s is unpaid", sub.ID)
			dbStatus = "inactive" // Map unpaid to inactive
		default:
			log.Printf("[WEBHOOK] Unhandled subscription status: %s for subscription %s", sub.Status, sub.ID)
			return nil
		}
	}

	// Update membership status in database
	if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatus(context.Background(), userID, dbStatus); updateErr != nil {
		log.Printf("[WEBHOOK] Failed to update subscription status in database: %v", updateErr)
		return updateErr
	}

	log.Printf("[WEBHOOK] Successfully updated subscription %s status to %s for user %s", sub.ID, dbStatus, userID)

	// Mark as processed
	s.Idempotency.MarkAsProcessed(event.ID)
	return nil
}

// HandleSubscriptionDeleted processes subscription.deleted events
func (s *WebhookService) HandleSubscriptionDeleted(event stripe.Event) *errLib.CommonError {
	// Check idempotency first
	if s.Idempotency.IsProcessed(event.ID) {
		log.Printf("Event %s already processed, skipping", event.ID)
		return nil
	}

	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("[WEBHOOK] Failed to parse subscription deleted event: %v", err)
		return errLib.New("Failed to parse subscription", http.StatusBadRequest)
	}

	log.Printf("[WEBHOOK] Subscription deleted: %s", sub.ID)

	// Extract user ID and update membership status to expired
	if sub.Customer != nil && sub.Customer.Metadata != nil {
		if userIDStr, exists := sub.Customer.Metadata["userID"]; exists {
			userID, err := uuid.Parse(userIDStr)
			if err == nil {
				log.Printf("[WEBHOOK] Marking membership as expired for user %s", userID)
				
				// Update membership status to expired in database
				if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatus(context.Background(), userID, "expired"); updateErr != nil {
					log.Printf("[WEBHOOK] Failed to mark membership as expired: %v", updateErr)
					return updateErr
				}
				
				log.Printf("[WEBHOOK] Successfully marked subscription %s as expired for user %s", sub.ID, userID)
			} else {
				log.Printf("[WEBHOOK] Invalid userID format for deleted subscription %s: %v", sub.ID, err)
				return errLib.New("Invalid user ID format", http.StatusBadRequest)
			}
		} else {
			log.Printf("[WEBHOOK] No userID in customer metadata for deleted subscription %s", sub.ID)
		}
	} else {
		log.Printf("[WEBHOOK] No customer metadata found for deleted subscription %s", sub.ID)
	}

	// Mark as processed
	s.Idempotency.MarkAsProcessed(event.ID)
	return nil
}

// HandleInvoicePaymentSucceeded processes invoice.payment_succeeded events
func (s *WebhookService) HandleInvoicePaymentSucceeded(event stripe.Event) *errLib.CommonError {
	// Check idempotency first
	if s.Idempotency.IsProcessed(event.ID) {
		log.Printf("Event %s already processed, skipping", event.ID)
		return nil
	}

	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("[WEBHOOK] Failed to parse invoice payment succeeded event: %v", err)
		return errLib.New("Failed to parse invoice", http.StatusBadRequest)
	}

	subscriptionID := "none"
	if invoice.Subscription != nil {
		subscriptionID = invoice.Subscription.ID
	}
	log.Printf("[WEBHOOK] Invoice payment succeeded: %s for subscription: %s", invoice.ID, subscriptionID)

	// Update payment history and ensure subscription is active
	if invoice.Customer != nil && invoice.Customer.Metadata != nil {
		if userIDStr, exists := invoice.Customer.Metadata["userID"]; exists {
			userID, err := uuid.Parse(userIDStr)
			if err == nil {
				log.Printf("[WEBHOOK] Payment successful for user %s, ensuring membership is active", userID)
				
				// Ensure membership is active after successful payment
				if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatus(context.Background(), userID, "active"); updateErr != nil {
					log.Printf("[WEBHOOK] Failed to activate membership after successful payment: %v", updateErr)
					return updateErr
				}
				
				log.Printf("[WEBHOOK] Successfully activated membership for user %s after invoice payment %s", userID, invoice.ID)
			} else {
				log.Printf("[WEBHOOK] Invalid userID format for invoice %s: %v", invoice.ID, err)
				return errLib.New("Invalid user ID format", http.StatusBadRequest)
			}
		} else {
			log.Printf("[WEBHOOK] No userID in customer metadata for invoice %s", invoice.ID)
		}
	} else {
		log.Printf("[WEBHOOK] No customer metadata found for invoice %s", invoice.ID)
	}

	// Mark as processed
	s.Idempotency.MarkAsProcessed(event.ID)
	return nil
}

// HandleInvoicePaymentFailed processes invoice.payment_failed events
func (s *WebhookService) HandleInvoicePaymentFailed(event stripe.Event) *errLib.CommonError {
	// Check idempotency first
	if s.Idempotency.IsProcessed(event.ID) {
		log.Printf("Event %s already processed, skipping", event.ID)
		return nil
	}

	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("[WEBHOOK] Failed to parse invoice payment failed event: %v", err)
		return errLib.New("Failed to parse invoice", http.StatusBadRequest)
	}

	subscriptionID := "none"
	if invoice.Subscription != nil {
		subscriptionID = invoice.Subscription.ID
	}
	log.Printf("[WEBHOOK] Invoice payment failed: %s for subscription: %s", invoice.ID, subscriptionID)

	// Handle payment failure - mark subscription as inactive/past due
	if invoice.Customer != nil && invoice.Customer.Metadata != nil {
		if userIDStr, exists := invoice.Customer.Metadata["userID"]; exists {
			userID, err := uuid.Parse(userIDStr)
			if err == nil {
				log.Printf("[WEBHOOK] Payment failed for user %s, handling payment failure", userID)
				
				// Update membership status to inactive due to payment failure
				if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatus(context.Background(), userID, "inactive"); updateErr != nil {
					log.Printf("[WEBHOOK] Failed to update membership status after payment failure: %v", updateErr)
					return updateErr
				}
				
				log.Printf("[WEBHOOK] Successfully marked membership as inactive for user %s after payment failure %s", userID, invoice.ID)
				
				// TODO: Send email notification about payment failure
				// s.sendPaymentFailureEmail(userID, invoice.ID)
			} else {
				log.Printf("[WEBHOOK] Invalid userID format for failed invoice %s: %v", invoice.ID, err)
				return errLib.New("Invalid user ID format", http.StatusBadRequest)
			}
		} else {
			log.Printf("[WEBHOOK] No userID in customer metadata for failed invoice %s", invoice.ID)
		}
	} else {
		log.Printf("[WEBHOOK] No customer metadata found for failed invoice %s", invoice.ID)
	}

	// Mark as processed
	s.Idempotency.MarkAsProcessed(event.ID)
	return nil
}

// allocateCreditsForMembership allocates credits to a customer when they purchase a credit-based membership
func (s *WebhookService) allocateCreditsForMembership(ctx context.Context, customerID, membershipPlanID uuid.UUID) *errLib.CommonError {
	log.Printf("Allocating credits for customer %s with membership plan %s", customerID, membershipPlanID)
	
	// Use the credit service to handle allocation logic
	return s.CreditService.AllocateCreditsOnMembershipPurchase(ctx, customerID, membershipPlanID)
}

// isCustomerAlreadyEnrolled checks if a customer is already enrolled in a membership plan
func (s *WebhookService) isCustomerAlreadyEnrolled(ctx context.Context, customerID, planID uuid.UUID) (bool, *errLib.CommonError) {
	query := "SELECT id, created_at, start_date, status FROM users.customer_membership_plans WHERE customer_id = $1 AND membership_plan_id = $2 AND status = 'active'"

	rows, err := s.db.QueryContext(ctx, query, customerID, planID)
	if err != nil {
		log.Printf("Failed to check enrollment status for customer %s, plan %s: %v", customerID, planID, err)
		return false, errLib.New("Failed to check enrollment status", http.StatusInternalServerError)
	}
	defer rows.Close()

	var foundRecords []string
	for rows.Next() {
		var id string
		var createdAt, startDate time.Time
		var status string
		if err := rows.Scan(&id, &createdAt, &startDate, &status); err == nil {
			foundRecords = append(foundRecords, fmt.Sprintf("ID: %s, CreatedAt: %s, StartDate: %s, Status: %s", id, createdAt.Format(time.RFC3339), startDate.Format(time.RFC3339), status))
		}
	}

	exists := len(foundRecords) > 0
	if exists {
		log.Printf("🚫 FOUND EXISTING MEMBERSHIP - Customer %s already enrolled in plan %s. Found records: %v", customerID, planID, foundRecords)
	} else {
		log.Printf("✅ NO EXISTING MEMBERSHIP - Customer %s not enrolled in plan %s, proceeding with enrollment", customerID, planID)
	}

	return exists, nil
}

// storeStripeCustomerID stores the Stripe customer ID in the database for future reference
func (s *WebhookService) storeStripeCustomerID(userID uuid.UUID, stripeCustomerID string) *errLib.CommonError {
	query := "UPDATE users.users SET stripe_customer_id = $1 WHERE id = $2"
	result, err := s.db.Exec(query, stripeCustomerID, userID)
	if err != nil {
		log.Printf("Failed to store Stripe customer ID %s for user %s: %v", stripeCustomerID, userID, err)
		return errLib.New("Failed to store customer ID", http.StatusInternalServerError)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to check rows affected for user %s: %v", userID, err)
		return errLib.New("Failed to verify update", http.StatusInternalServerError)
	}

	if rowsAffected == 0 {
		log.Printf("No user found with ID %s to update Stripe customer ID", userID)
		return errLib.New("User not found", http.StatusNotFound)
	}

	return nil
}

// sendEnhancedWebhookAlert sends comprehensive debugging information to Slack
func (s *WebhookService) sendEnhancedWebhookAlert(event stripe.Event, checkSession stripe.CheckoutSession, webhookError *errLib.CommonError) {
	// Gather comprehensive context
	var customerID, userEmail, stripePriceID, planID, subscriptionID string
	var failureStep string
	
	// Try to extract session data for context
	if fullSession, err := s.getExpandedSession(checkSession.ID); err == nil {
		if fullSession.Metadata != nil {
			customerID = fullSession.Metadata["userID"]
		}
		
		if fullSession.LineItems != nil && len(fullSession.LineItems.Data) > 0 {
			stripePriceID = fullSession.LineItems.Data[0].Price.ID
		}
		
		if fullSession.Subscription != nil {
			subscriptionID = fullSession.Subscription.ID
		}
		
		// Try to get plan ID from price ID
		if stripePriceID != "" {
			if planUUID, _, planErr := s.PostCheckoutRepository.GetMembershipPlanByStripePriceID(context.Background(), stripePriceID); planErr == nil {
				planID = planUUID.String()
			}
		}
		
		// Get user email if we have customer ID
		if customerID != "" {
			userEmail = logger.GetUserEmailFromID(customerID)
		}
	}
	
	// Classify the error and get troubleshooting steps
	errorType, troubleshootingSteps := logger.ClassifyWebhookError(webhookError.Error())
	
	// Determine failure step based on error message
	errorMsg := webhookError.Error()
	if strings.Contains(strings.ToLower(errorMsg), "expand") || strings.Contains(strings.ToLower(errorMsg), "session") {
		failureStep = "Session Expansion"
	} else if strings.Contains(strings.ToLower(errorMsg), "line items") {
		failureStep = "Line Item Validation"
	} else if strings.Contains(strings.ToLower(errorMsg), "membership plan") {
		failureStep = "Plan Lookup"
	} else if strings.Contains(strings.ToLower(errorMsg), "enrollment") || strings.Contains(strings.ToLower(errorMsg), "already enrolled") {
		failureStep = "Customer Enrollment"
	} else if strings.Contains(strings.ToLower(errorMsg), "subscription") || strings.Contains(strings.ToLower(errorMsg), "cancel") {
		failureStep = "Subscription Update"
	} else if strings.Contains(strings.ToLower(errorMsg), "credit") {
		failureStep = "Credit Allocation"
	} else {
		failureStep = "Unknown Step"
	}
	
	// Build comprehensive alert details
	alertDetails := logger.WebhookAlertDetails{
		EventID:              event.ID,
		EventType:            string(event.Type),
		SessionID:            checkSession.ID,
		CustomerID:           customerID,
		StripePriceID:        stripePriceID,
		PlanID:              planID,
		UserEmail:           userEmail,
		ErrorType:           errorType,
		ErrorMessage:        webhookError.Error(),
		FailureStep:         failureStep,
		RetryAttempt:        1, // Could be enhanced to track actual retry count
		SessionStatus:       string(checkSession.Status),
		PaymentStatus:       string(checkSession.PaymentStatus),
		SubscriptionID:      subscriptionID,
		TroubleshootingSteps: troubleshootingSteps,
	}
	
	// Send the enhanced alert
	logger.SendWebhookFailureAlert(alertDetails)
	
	log.Printf("Enhanced webhook failure alert sent for event %s", event.ID)
}
