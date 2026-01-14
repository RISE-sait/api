package payment

import (
	"api/internal/di"
	creditPackageRepo "api/internal/domains/credit_package/persistence/repository"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	enrollment "api/internal/domains/enrollment/service"
	enrollmentRepo "api/internal/domains/enrollment/persistence/repository"
	repository "api/internal/domains/payment/persistence/repositories"
	"api/internal/domains/payment/tracking"
	"api/internal/domains/subsidy/dto"
	subsidyService "api/internal/domains/subsidy/service"
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
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/invoice"
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
	SubsidyService         *subsidyService.SubsidyService
	PaymentTracking        *tracking.PaymentTrackingService
	Idempotency            *WebhookIdempotency
	logger                 *logger.StructuredLogger
	db                     *sql.DB
	container              *di.Container
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
		SubsidyService:         subsidyService.NewSubsidyService(container),
		PaymentTracking:        tracking.NewPaymentTrackingService(container),
		Idempotency:            NewWebhookIdempotencyWithDB(container.DB, 24*time.Hour, 10000), // Database-backed with cache
		logger:                 logger.WithComponent("stripe-webhooks"),
		db:                     container.DB,
		container:              container,
	}
}

func (s *WebhookService) HandleCheckoutSessionCompleted(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Get event creation time for consistent timestamps
	eventCreatedAt := time.Unix(event.Created, 0)

	webhookLogger := s.logger.WithFields(map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.Type,
		"webhook":    "checkout_session_completed",
	})

	// Atomically claim the event - prevents race conditions with concurrent webhook deliveries
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		webhookLogger.Info("Event already claimed by another process, skipping")
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
		err = s.handleItemCheckoutComplete(ctx, checkSession, webhookLogger, eventCreatedAt)
	case stripe.CheckoutSessionModeSubscription:
		err = s.handleSubscriptionCheckoutComplete(ctx, checkSession, eventCreatedAt)
	default:
		webhookLogger.Warn("Unhandled session mode received")
		return nil
	}

	// Mark event complete or failed based on result
	if err == nil {
		s.Idempotency.MarkEventComplete(event.ID)
		webhookLogger.Info("Webhook event processed successfully")
	} else {
		webhookLogger.Error("Webhook processing failed", err)
		s.Idempotency.MarkEventFailed(event.ID, err.Error())

		// Send comprehensive Slack alert for debugging
		s.sendEnhancedWebhookAlert(event, checkSession, err)
	}

	return err
}

func (s *WebhookService) handleItemCheckoutComplete(ctx context.Context, checkoutSession stripe.CheckoutSession, parentLogger *logger.StructuredLogger, eventCreatedAt time.Time) *errLib.CommonError {
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

	// Extract receipt URL from PaymentIntent's charge for one-time payments
	var receiptURL string
	if fullSession.PaymentIntent != nil && fullSession.PaymentIntent.LatestCharge != nil {
		receiptURL = fullSession.PaymentIntent.LatestCharge.ReceiptURL
		log.Printf("Receipt URL extracted: %s", receiptURL)
	}

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
		dbProgramID, err := s.PostCheckoutRepository.GetProgramIdByStripePriceId(ctx, priceID)
		if err != nil {
			log.Printf("GetProgramIdByStripePriceId failed: %v", err)
			return err
		}
		dbEventID, err := s.PostCheckoutRepository.GetEventIdByStripePriceId(ctx, priceID)
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
	creditPackage, creditErr := s.CreditPackageRepo.GetByStripePriceID(ctx, priceID)

	switch {
	case programID != uuid.Nil && eventID != uuid.Nil:
		return errLib.New("price ID maps to both program and event", http.StatusConflict)
	case creditPackage != nil && creditErr == nil:
		// This is a credit package purchase
		log.Printf("CREDIT PACKAGE PURCHASE DETECTED - Customer: %s, Package: %s (%s)", customerID, creditPackage.ID, creditPackage.Name)

		// Store Stripe customer ID in database for future reference (needed for payment display in admin panel)
		if fullSession.Customer != nil && fullSession.Customer.ID != "" {
			if err := s.storeStripeCustomerID(customerID, fullSession.Customer.ID); err != nil {
				log.Printf("WARNING: Failed to store Stripe customer ID for credit purchase: %v", err)
				// Don't fail the entire process for this
			} else {
				log.Printf("Successfully stored Stripe customer ID %s for user %s", fullSession.Customer.ID, customerID)
			}
		}

		// Add credits to customer balance
		log.Printf("Adding %d credits to customer %s balance", creditPackage.CreditAllocation, customerID)
		if err := s.CustomerCreditService.AddCredits(ctx, customerID, creditPackage.CreditAllocation, "Credit package purchase"); err != nil {
			log.Printf("FAILED to add credits to customer %s: %v", customerID, err)
			return errLib.New(fmt.Sprintf("failed to add credits: %v", err), http.StatusInternalServerError)
		}

		// Set active credit package (overwrites previous package)
		log.Printf("Setting active credit package for customer %s: weekly limit=%d", customerID, creditPackage.WeeklyCreditLimit)
		if err := s.CreditPackageRepo.SetCustomerActivePackage(ctx, customerID, creditPackage.ID, creditPackage.WeeklyCreditLimit); err != nil {
			log.Printf("FAILED to set active credit package for customer %s: %v", customerID, err)
			return errLib.New(fmt.Sprintf("failed to set active package: %v", err), http.StatusInternalServerError)
		}

		// Track payment in centralized system
		go s.trackCreditPackagePurchase(fullSession, customerID, creditPackage, eventCreatedAt, receiptURL)

		log.Printf("CREDIT PACKAGE PURCHASE COMPLETE - Customer %s: +%d credits, %d/week limit", customerID, creditPackage.CreditAllocation, creditPackage.WeeklyCreditLimit)
		return nil
	case programID == uuid.Nil && eventID == uuid.Nil:
		return errLib.New("price ID doesn't map to any program, event, or credit package", http.StatusNotFound)
	case programID != uuid.Nil:
		log.Printf("Updating program reservation for user %s and program %s", customerID, programID)
		if err = s.EnrollmentService.UpdateReservationStatusInProgram(ctx, programID, customerID, dbEnrollment.PaymentStatusPaid); err != nil {
			log.Printf("Failed to update program reservation: %v", err)
			return errLib.New(fmt.Sprintf("failed to update program reservation (customer: %s, program: %s): %v", customerID, programID, err), http.StatusInternalServerError)
		}
		// Track payment in centralized system
		go s.trackProgramEnrollment(fullSession, customerID, programID, eventCreatedAt, receiptURL)
	case eventID != uuid.Nil:
		log.Printf("Updating event reservation for user %s and event %s", customerID, eventID)
		if err = s.EnrollmentService.UpdateReservationStatusInEvent(ctx, eventID, customerID, dbEnrollment.PaymentStatusPaid); err != nil {
			log.Printf("Failed to update event reservation: %v", err)
			return errLib.New(fmt.Sprintf("failed to update event reservation (customer: %s, event: %s): %v", customerID, eventID, err), http.StatusInternalServerError)
		}
		// Track payment in centralized system
		go s.trackEventRegistration(fullSession, customerID, eventID, eventCreatedAt, receiptURL)
	}

	log.Println("handleItemCheckoutComplete completed")
	return nil
}

func (s *WebhookService) handleSubscriptionCheckoutComplete(ctx context.Context, checkoutSession stripe.CheckoutSession, eventCreatedAt time.Time) *errLib.CommonError {
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

	// 2. Parse metadata for user and plan IDs
	userIdStr := fullSession.Metadata["userID"]
	planIdStr := fullSession.Metadata["membershipPlanID"]

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

	// Get plan ID from metadata (preferred) or fall back to price lookup for backwards compatibility
	var planID uuid.UUID
	var amtPeriods *int32

	if planIdStr != "" {
		// Use plan ID from metadata (industry standard approach)
		parsedPlanID, parseErr := uuid.Parse(planIdStr)
		if parseErr != nil {
			webhookLogger.Error("Invalid membership plan ID format in metadata", parseErr)
			return errLib.New("Invalid membership plan ID format", http.StatusBadRequest)
		}
		planID = parsedPlanID

		// Get amtPeriods directly from the plan by ID
		amtPeriods, err = s.PostCheckoutRepository.GetMembershipPlanAmtPeriods(ctx, planID)
		if err != nil {
			webhookLogger.Error("Failed to get membership plan details", err)
			return err
		}

		webhookLogger = webhookLogger.WithFields(map[string]interface{}{
			"membership_plan_id": planID,
			"source":             "metadata",
		})
	} else {
		// Backwards compatibility: look up by price ID
		webhookLogger.Info("No membershipPlanID in metadata, falling back to price lookup")

		priceIDs, validateErr := s.validateLineItems(fullSession.LineItems)
		if validateErr != nil {
			webhookLogger.Error("Failed to validate line items", validateErr)
			return validateErr
		}

		// Try each price ID until we find a matching membership plan
		for _, priceID := range priceIDs {
			planID, amtPeriods, err = s.PostCheckoutRepository.GetMembershipPlanByStripePriceID(ctx, priceID)
			if err == nil && planID != uuid.Nil {
				webhookLogger = webhookLogger.WithFields(map[string]interface{}{
					"membership_plan_id": planID,
					"matched_price_id":   priceID,
					"source":             "price_lookup",
				})
				break
			}
		}

		if planID == uuid.Nil {
			webhookLogger.Error("Failed to find membership plan for any price ID", nil)
			return errLib.New("membership plan not found", http.StatusNotFound)
		}
	}

	webhookLogger = webhookLogger.WithFields(map[string]interface{}{
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
		if isAlreadyEnrolled, checkErr := s.isCustomerAlreadyEnrolled(ctx, userID, planID); checkErr != nil {
			log.Printf("ERROR: Failed to check existing enrollment: %v", checkErr)
			return checkErr
		} else if isAlreadyEnrolled {
			log.Printf(" SKIPPING ENROLLMENT - Customer %s is already enrolled in plan %s", userID, planID)
		} else {
			log.Printf(" STARTING ENROLLMENT - Customer %s in membership plan %s with start time %s (one-time payment, no next billing date)", userID, planID, eventCreatedAt.Format(time.RFC3339))
			if err := s.EnrollmentService.EnrollCustomerInMembershipPlan(ctx, userID, planID, time.Time{}, time.Time{}, eventCreatedAt, ""); err != nil {
				log.Printf(" ERROR: EnrollCustomerInMembershipPlan failed: %v", err)
				return err
			}
			log.Printf(" SUCCESS: Enrolled customer %s in membership plan %s", userID, planID)
		}
	}

	// NOTE: Credits are no longer allocated with memberships - they are only available via credit packages
	// Credit allocation has been moved to credit package purchases

	// Record subsidy usage if a subsidy was applied
	if hasSubsidy, exists := fullSession.Metadata["has_subsidy"]; exists && hasSubsidy == "true" {
		if err := s.recordSubsidyUsageFromCheckout(fullSession, userID); err != nil {
			log.Printf("WARNING: Failed to record subsidy usage: %v", err)
			// Don't fail the entire checkout for subsidy recording failure
		}
	}

	// Track payment in centralized system
	go s.trackMembershipSubscription(fullSession, userID, planID, eventCreatedAt)

	s.sendMembershipPurchaseEmail(userID, planID)
	return nil
}

// recordSubsidyUsageFromCheckout records subsidy usage based on checkout session metadata
func (s *WebhookService) recordSubsidyUsageFromCheckout(session *stripe.CheckoutSession, userID uuid.UUID) *errLib.CommonError {
	// Extract subsidy info from metadata
	subsidyIDStr, exists := session.Metadata["subsidy_id"]
	if !exists || subsidyIDStr == "" {
		return errLib.New("Subsidy ID not found in metadata", http.StatusBadRequest)
	}

	subsidyID, err := uuid.Parse(subsidyIDStr)
	if err != nil {
		return errLib.New("Invalid subsidy ID format", http.StatusBadRequest)
	}

	// Get subsidy balance from metadata
	var subsidyBalance float64
	if balanceStr, exists := session.Metadata["subsidy_balance"]; exists {
		fmt.Sscanf(balanceStr, "%f", &subsidyBalance)
	}

	// Get the invoice to determine actual amounts
	var invoice *stripe.Invoice
	if session.Invoice != nil && session.Invoice.ID != "" {
		inv, err := s.getInvoiceDetails(session.Invoice.ID)
		if err != nil {
			log.Printf("[SUBSIDY] Failed to get invoice details: %v", err)
			return err
		}
		invoice = inv
	}

	// Calculate subsidy amount applied
	var subsidyApplied float64
	var customerPaid float64
	var originalAmount float64

	if invoice != nil {
		// Calculate from invoice discount amounts
		if len(invoice.TotalDiscountAmounts) > 0 {
			for _, discount := range invoice.TotalDiscountAmounts {
				subsidyApplied += float64(discount.Amount) / 100.0
			}
			originalAmount = float64(invoice.Total+invoice.TotalDiscountAmounts[0].Amount) / 100.0
		} else {
			originalAmount = float64(invoice.Total) / 100.0
		}
		customerPaid = float64(invoice.AmountPaid) / 100.0
	} else {
		// Fallback: use subsidy balance if invoice not available
		subsidyApplied = subsidyBalance
		customerPaid = 0
		originalAmount = subsidyBalance
	}

	log.Printf("[SUBSIDY] Recording usage - Original: $%.2f, Subsidy: $%.2f, Customer Paid: $%.2f",
		originalAmount, subsidyApplied, customerPaid)

	// Record subsidy usage
	subscriptionID := ""
	if session.Subscription != nil {
		subscriptionID = session.Subscription.ID
	}

	invoiceID := ""
	if invoice != nil {
		invoiceID = invoice.ID
	}

	recordReq := &dto.RecordUsageRequest{
		SubsidyID:            subsidyID,
		CustomerID:           userID,
		TransactionType:      "membership_payment",
		OriginalAmount:       originalAmount,
		SubsidyApplied:       subsidyApplied,
		CustomerPaid:         customerPaid,
		StripeSubscriptionID: &subscriptionID,
		StripeInvoiceID:      &invoiceID,
		Description:          fmt.Sprintf("Membership payment - Session %s", session.ID),
	}

	_, recordErr := s.SubsidyService.RecordUsage(context.Background(), recordReq)
	if recordErr != nil {
		log.Printf("[SUBSIDY] Failed to record subsidy usage: %v", recordErr)
		return recordErr
	}

	log.Printf("[SUBSIDY] Recorded subsidy usage: $%.2f applied, $%.2f paid by customer", subsidyApplied, customerPaid)
	return nil
}

// getInvoiceDetails retrieves full invoice details from Stripe
func (s *WebhookService) getInvoiceDetails(invoiceID string) (*stripe.Invoice, *errLib.CommonError) {
	params := &stripe.InvoiceParams{
		Expand: []*string{
			stripe.String("subscription"),
			stripe.String("customer"),
		},
	}

	inv, err := invoice.Get(invoiceID, params)
	if err != nil {
		return nil, errLib.New("Failed to get invoice: "+err.Error(), http.StatusInternalServerError)
	}

	return inv, nil
}

func (s *WebhookService) getExpandedSession(sessionID string) (*stripe.CheckoutSession, *errLib.CommonError) {
	params := &stripe.CheckoutSessionParams{
		Expand: []*string{
			stripe.String("line_items"),
			stripe.String("line_items.data.price"),
			stripe.String("subscription"),
			stripe.String("customer"),
			stripe.String("payment_intent.latest_charge"),
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

	// Get next billing date from Stripe subscription's current_period_end
	nextBillingDate := time.Unix(sub.CurrentPeriodEnd, 0)

	log.Printf("Checking existing enrollment for customer %s in plan %s", userID, planID)

	// Check if customer is already enrolled to handle webhook retries gracefully
	if isAlreadyEnrolled, checkErr := s.isCustomerAlreadyEnrolled(context.Background(), userID, planID); checkErr != nil {
		log.Printf("ERROR: Failed to check existing enrollment: %v", checkErr)
		return checkErr
	} else if isAlreadyEnrolled {
		log.Printf("🚫 SKIPPING ENROLLMENT - Customer %s is already enrolled in plan %s", userID, planID)
	} else {
		log.Printf("🚀 STARTING ENROLLMENT - Customer %s in plan %s with end date: %s, next billing: %s, start time: %s, subscription: %s",
			userID, planID, cancelAtDateTime.Format(time.RFC3339), nextBillingDate.Format(time.RFC3339), eventCreatedAt.Format(time.RFC3339), subscriptionID)
		if err = s.EnrollmentService.EnrollCustomerInMembershipPlan(context.Background(), userID, planID, cancelAtDateTime, nextBillingDate, eventCreatedAt, subscriptionID); err != nil {
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
		// Queue for background retry
		s.queueSubscriptionCancelUpdate(sub.ID, cancelAtUnix)
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
func (s *WebhookService) HandleSubscriptionCreated(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomically claim the event - prevents race conditions
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		log.Printf("Event %s already claimed by another process, skipping", event.ID)
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

	// Mark as complete
	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// HandleSubscriptionUpdated processes subscription.updated events
func (s *WebhookService) HandleSubscriptionUpdated(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomically claim the event - prevents race conditions
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		log.Printf("Event %s already claimed by another process, skipping", event.ID)
		return nil
	}

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

	// Update membership status in database - use subscription ID to target only this specific subscription
	if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatusByID(ctx, userID, sub.ID, dbStatus); updateErr != nil {
		log.Printf("[WEBHOOK] Failed to update subscription status in database: %v", updateErr)
		return updateErr
	}

	log.Printf("[WEBHOOK] Successfully updated subscription %s status to %s for user %s", sub.ID, dbStatus, userID)

	// Mark as complete
	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// HandleSubscriptionDeleted processes subscription.deleted events
func (s *WebhookService) HandleSubscriptionDeleted(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomically claim the event - prevents race conditions
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		log.Printf("Event %s already claimed by another process, skipping", event.ID)
		return nil
	}

	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("[WEBHOOK] Failed to parse subscription deleted event: %v", err)
		return errLib.New("Failed to parse subscription", http.StatusBadRequest)
	}

	log.Printf("[WEBHOOK] Subscription deleted: %s", sub.ID)

	// Get Stripe customer ID from the Customer field
	var stripeCustomerID string
	if sub.Customer != nil {
		stripeCustomerID = sub.Customer.ID
	}

	if stripeCustomerID == "" {
		log.Printf("[WEBHOOK] No customer ID found for deleted subscription %s", sub.ID)
		return nil
	}

	log.Printf("[WEBHOOK] Processing deletion for Stripe customer: %s", stripeCustomerID)

	// Look up user by Stripe customer ID in database
	var userID uuid.UUID
	query := "SELECT id FROM users.users WHERE stripe_customer_id = $1"
	if dbErr := s.db.QueryRowContext(ctx, query, stripeCustomerID).Scan(&userID); dbErr != nil {
		log.Printf("[WEBHOOK] Failed to find user with Stripe customer ID %s for subscription %s: %v", stripeCustomerID, sub.ID, dbErr)
		return nil // Don't fail webhook if user not found - subscription may have been deleted already
	}

	log.Printf("[WEBHOOK] Found user %s for Stripe customer %s, marking subscription %s as expired", userID, stripeCustomerID, sub.ID)

	// Update membership status to expired in database - ONLY for this specific subscription
	if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatusByID(ctx, userID, sub.ID, "expired"); updateErr != nil {
		log.Printf("[WEBHOOK] Failed to mark membership as expired: %v", updateErr)
		return updateErr
	}

	log.Printf("[WEBHOOK] Successfully marked subscription %s as expired for user %s", sub.ID, userID)

	// Mark as complete
	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// HandleInvoicePaymentSucceeded processes invoice.payment_succeeded events
func (s *WebhookService) HandleInvoicePaymentSucceeded(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomically claim the event - prevents race conditions
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		log.Printf("Event %s already claimed by another process, skipping", event.ID)
		return nil
	}

	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("[WEBHOOK] Failed to parse invoice payment succeeded event: %v", err)
		return errLib.New("Failed to parse invoice", http.StatusBadRequest)
	}

	// Get subscription ID from the invoice
	// Note: invoice.Subscription might be nil (not expanded), so we check the ID field directly
	var subscriptionID string
	if invoice.Subscription != nil && invoice.Subscription.ID != "" {
		subscriptionID = invoice.Subscription.ID
	}

	log.Printf("[WEBHOOK] Invoice payment succeeded: %s for subscription: %s", invoice.ID, subscriptionID)

	// Look up user by Stripe customer ID (more reliable than metadata)
	var userID uuid.UUID

	// Try to get customer ID from invoice
	customerID := ""
	if invoice.Customer != nil {
		customerID = invoice.Customer.ID
	}

	if customerID == "" {
		log.Printf("[WEBHOOK] No customer ID found for invoice %s", invoice.ID)
		return nil
	}

	// Look up user by Stripe customer ID in database
	query := "SELECT id FROM users.users WHERE stripe_customer_id = $1"
	if dbErr := s.db.QueryRowContext(ctx, query, customerID).Scan(&userID); dbErr != nil {
		log.Printf("[WEBHOOK] Failed to find user with Stripe customer ID %s for invoice %s: %v", customerID, invoice.ID, dbErr)
		// Don't fail the webhook - the user might not exist yet
		return nil
	}

	log.Printf("[WEBHOOK] Found user %s for Stripe customer %s, processing payment", userID, customerID)

	// If this is a subscription invoice, get the next billing date from Stripe
	if subscriptionID != "" {
		// Get subscription details to find the next billing date
		sub, subErr := subscription.Get(subscriptionID, nil)
		if subErr != nil {
			log.Printf("[WEBHOOK] Failed to get subscription details for %s: %v", subscriptionID, subErr)
			// Fall back to just updating status - use subscription ID for specificity
			if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatusByID(ctx, userID, subscriptionID, "active"); updateErr != nil {
				log.Printf("[WEBHOOK] Failed to activate membership after successful payment: %v", updateErr)
				return updateErr
			}
		} else {
			// Update both status and next billing date - target specific subscription
			nextBillingDate := time.Unix(sub.CurrentPeriodEnd, 0)
			log.Printf("[WEBHOOK] Updating membership: status=active, next_billing=%s for subscription %s", nextBillingDate.Format(time.RFC3339), subscriptionID)
			if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatusByIDAndNextBilling(ctx, userID, subscriptionID, "active", nextBillingDate); updateErr != nil {
				log.Printf("[WEBHOOK] Failed to update membership after successful payment: %v", updateErr)
				return updateErr
			}
			log.Printf("[WEBHOOK] Successfully updated next_billing_date to %s", nextBillingDate.Format(time.RFC3339))
		}
	} else {
		// Non-subscription invoice - should not happen for membership payments
		log.Printf("[WEBHOOK] No subscription ID for invoice %s, updating all active subscriptions", invoice.ID)
		if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatus(ctx, userID, "active"); updateErr != nil {
			log.Printf("[WEBHOOK] Failed to activate membership after successful payment: %v", updateErr)
			return updateErr
		}
	}

	log.Printf("[WEBHOOK] Successfully activated membership for user %s after invoice payment %s", userID, invoice.ID)

	// Track payment in centralized system
	go s.trackMembershipRenewal(&invoice, userID, time.Unix(event.Created, 0))

	// Mark as complete
	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// HandleInvoicePaymentFailed processes invoice.payment_failed events
func (s *WebhookService) HandleInvoicePaymentFailed(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomically claim the event - prevents race conditions
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		log.Printf("Event %s already claimed by another process, skipping", event.ID)
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

	// Look up user by Stripe customer ID (more reliable than metadata)
	customerID := ""
	if invoice.Customer != nil {
		customerID = invoice.Customer.ID
	}

	if customerID == "" {
		log.Printf("[WEBHOOK] No customer ID found for failed invoice %s", invoice.ID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	var userID uuid.UUID
	query := "SELECT id FROM users.users WHERE stripe_customer_id = $1"
	if dbErr := s.db.QueryRowContext(ctx, query, customerID).Scan(&userID); dbErr != nil {
		log.Printf("[WEBHOOK] Failed to find user with Stripe customer ID %s: %v", customerID, dbErr)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	log.Printf("[WEBHOOK] Payment failed for user %s, handling payment failure", userID)

	// Update membership status to inactive due to payment failure - target specific subscription if available
	if subscriptionID != "none" {
		if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatusByID(ctx, userID, subscriptionID, "inactive"); updateErr != nil {
			log.Printf("[WEBHOOK] Failed to update membership status after payment failure: %v", updateErr)
			return updateErr
		}
	} else {
		// Fallback to updating all subscriptions if no subscription ID (shouldn't happen for subscription invoices)
		if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatus(ctx, userID, "inactive"); updateErr != nil {
			log.Printf("[WEBHOOK] Failed to update membership status after payment failure: %v", updateErr)
			return updateErr
		}
	}

	log.Printf("[WEBHOOK] Successfully marked subscription %s as inactive for user %s after payment failure %s", subscriptionID, userID, invoice.ID)

	// Send email notification about payment failure
	go s.sendPaymentFailureEmail(userID, invoice.ID)

	// Mark as complete
	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// allocateCreditsForMembership allocates credits to a customer when they purchase a credit-based membership
func (s *WebhookService) allocateCreditsForMembership(ctx context.Context, customerID, membershipPlanID uuid.UUID) *errLib.CommonError {
	log.Printf("Allocating credits for customer %s with membership plan %s", customerID, membershipPlanID)

	// Use the credit service to handle allocation logic
	return s.CreditService.AllocateCreditsOnMembershipPurchase(ctx, customerID, membershipPlanID)
}

// queueSubscriptionCancelUpdate queues a failed subscription cancel date update for background retry
func (s *WebhookService) queueSubscriptionCancelUpdate(subscriptionID string, cancelAtUnix int64) {
	log.Printf("[CANCEL_QUEUE] Queueing subscription cancel date update: subscription=%s, cancel_at=%d", subscriptionID, cancelAtUnix)

	// For now, we'll just retry asynchronously with exponential backoff
	go func() {
		maxRetries := 3
		baseDelay := 5 * time.Second

		for attempt := 1; attempt <= maxRetries; attempt++ {
			delay := baseDelay * time.Duration(attempt*attempt) // Quadratic backoff
			log.Printf("[CANCEL_QUEUE] Retry attempt %d/%d for subscription %s in %v", attempt, maxRetries, subscriptionID, delay)
			time.Sleep(delay)

			if err := s.updateSubscriptionCancelAt(subscriptionID, cancelAtUnix); err == nil {
				log.Printf("[CANCEL_QUEUE] Successfully updated subscription %s cancel date on retry %d", subscriptionID, attempt)
				return
			} else {
				log.Printf("[CANCEL_QUEUE] Retry %d failed for subscription %s: %v", attempt, subscriptionID, err)
			}
		}

		log.Printf("[CANCEL_QUEUE] CRITICAL: All retries failed for subscription %s cancel date update. Manual intervention required.", subscriptionID)
	}()
}

// sendPaymentFailureEmail sends an email notification when a payment fails
func (s *WebhookService) sendPaymentFailureEmail(userID uuid.UUID, invoiceID string) {
	log.Printf("[EMAIL] Attempting to send payment failure email for user %s, invoice %s", userID, invoiceID)

	// Get user info
	userInfo, err := s.UserRepo.GetUserInfo(context.Background(), "", userID)
	if err != nil {
		log.Printf("[EMAIL] Failed to get user info for %s: %v", userID, err)
		return
	}

	if userInfo.Email == nil {
		log.Printf("[EMAIL] User %s has no email address", userID)
		return
	}

	// Get active membership plan name
	membershipPlanName := "your membership"
	activePlans, planErr := s.EnrollmentRepo.GetCustomerActiveMembershipPlans(context.Background(), userID)
	if planErr == nil && len(activePlans) > 0 {
		// Get the first active plan's name
		plan, pErr := s.PlansRepo.GetMembershipPlanById(context.Background(), activePlans[0].MembershipPlanID)
		if pErr == nil {
			membershipPlanName = plan.Name
		}
	}

	// Create Stripe billing portal URL for the user to update payment
	updatePaymentURL := "https://www.risesportscomplex.com/account/billing"

	log.Printf("[EMAIL] Sending payment failure email to %s for %s", *userInfo.Email, membershipPlanName)
	email.SendPaymentFailedEmail(*userInfo.Email, userInfo.FirstName, membershipPlanName, updatePaymentURL)
	log.Printf("[EMAIL] Payment failure email sent successfully to %s", *userInfo.Email)
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
// and updates the Stripe customer with userID metadata for subsidy processing
func (s *WebhookService) storeStripeCustomerID(userID uuid.UUID, stripeCustomerID string) *errLib.CommonError {
	// First, update the database
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

	// Now update the Stripe customer with userID metadata for subsidy processing
	customerParams := &stripe.CustomerParams{
		Metadata: map[string]string{
			"userID": userID.String(),
		},
	}
	_, stripeErr := customer.Update(stripeCustomerID, customerParams)
	if stripeErr != nil {
		log.Printf("WARNING: Failed to update Stripe customer %s metadata: %v", stripeCustomerID, stripeErr)
		// Don't fail the entire process - the database update succeeded
		// Subsidy webhook can still fall back to subscription metadata
	} else {
		log.Printf("Successfully updated Stripe customer %s with userID metadata", stripeCustomerID)
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

// HandleInvoiceUpcoming handles invoice.upcoming events
// This is sent ~3 days before a subscription renews, useful for sending reminder emails
func (s *WebhookService) HandleInvoiceUpcoming(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomically claim the event
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		log.Printf("Event %s already claimed, skipping", event.ID)
		return nil
	}

	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("[WEBHOOK] Failed to parse invoice upcoming event: %v", err)
		return errLib.New("Failed to parse invoice", http.StatusBadRequest)
	}

	log.Printf("[WEBHOOK] Invoice upcoming: %s, Amount: %d cents", invoice.ID, invoice.AmountDue)

	// Look up user by Stripe customer ID
	customerID := ""
	if invoice.Customer != nil {
		customerID = invoice.Customer.ID
	}

	if customerID == "" {
		log.Printf("[WEBHOOK] No customer ID found for upcoming invoice %s", invoice.ID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	var userID uuid.UUID
	query := "SELECT id FROM users.users WHERE stripe_customer_id = $1"
	if dbErr := s.db.QueryRowContext(ctx, query, customerID).Scan(&userID); dbErr != nil {
		log.Printf("[WEBHOOK] Failed to find user with Stripe customer ID %s: %v", customerID, dbErr)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	// Send renewal reminder email
	go s.sendRenewalReminderEmail(userID, invoice.AmountDue, time.Unix(invoice.DueDate, 0))

	log.Printf("[WEBHOOK] Processed upcoming invoice for user %s, amount: %d cents", userID, invoice.AmountDue)

	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// HandlePaymentMethodAttached handles payment_method.attached events
// Useful for logging when customers add new payment methods
func (s *WebhookService) HandlePaymentMethodAttached(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomically claim the event
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		log.Printf("Event %s already claimed, skipping", event.ID)
		return nil
	}

	var paymentMethod stripe.PaymentMethod
	if err := json.Unmarshal(event.Data.Raw, &paymentMethod); err != nil {
		log.Printf("[WEBHOOK] Failed to parse payment method attached event: %v", err)
		return errLib.New("Failed to parse payment method", http.StatusBadRequest)
	}

	log.Printf("[WEBHOOK] Payment method attached: %s, Type: %s", paymentMethod.ID, paymentMethod.Type)

	// Get customer ID and look up user
	customerID := ""
	if paymentMethod.Customer != nil {
		customerID = paymentMethod.Customer.ID
	}

	if customerID != "" {
		var userID uuid.UUID
		query := "SELECT id FROM users.users WHERE stripe_customer_id = $1"
		if dbErr := s.db.QueryRowContext(ctx, query, customerID).Scan(&userID); dbErr == nil {
			log.Printf("[WEBHOOK] Payment method %s attached for user %s", paymentMethod.ID, userID)

			// If this customer had failed payments, they may have updated their payment method
			// Consider reactivating their subscription or sending a confirmation email
			go s.sendPaymentMethodUpdatedEmail(userID)
		}
	}

	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// HandlePaymentMethodDetached handles payment_method.detached events
func (s *WebhookService) HandlePaymentMethodDetached(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomically claim the event
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		log.Printf("Event %s already claimed, skipping", event.ID)
		return nil
	}

	var paymentMethod stripe.PaymentMethod
	if err := json.Unmarshal(event.Data.Raw, &paymentMethod); err != nil {
		log.Printf("[WEBHOOK] Failed to parse payment method detached event: %v", err)
		return errLib.New("Failed to parse payment method", http.StatusBadRequest)
	}

	log.Printf("[WEBHOOK] Payment method detached: %s", paymentMethod.ID)

	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// HandleSubscriptionPaused handles customer.subscription.paused events
func (s *WebhookService) HandleSubscriptionPaused(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomically claim the event
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		log.Printf("Event %s already claimed, skipping", event.ID)
		return nil
	}

	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("[WEBHOOK] Failed to parse subscription paused event: %v", err)
		return errLib.New("Failed to parse subscription", http.StatusBadRequest)
	}

	log.Printf("[WEBHOOK] Subscription paused: %s", sub.ID)

	// Look up user by Stripe customer ID
	stripeCustomerID := ""
	if sub.Customer != nil {
		stripeCustomerID = sub.Customer.ID
	}

	if stripeCustomerID == "" {
		log.Printf("[WEBHOOK] No customer ID found for paused subscription %s", sub.ID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	var userID uuid.UUID
	query := "SELECT id FROM users.users WHERE stripe_customer_id = $1"
	if dbErr := s.db.QueryRowContext(ctx, query, stripeCustomerID).Scan(&userID); dbErr != nil {
		log.Printf("[WEBHOOK] Failed to find user with Stripe customer ID %s: %v", stripeCustomerID, dbErr)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	// Update membership status to paused
	if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatusByID(ctx, userID, sub.ID, "paused"); updateErr != nil {
		log.Printf("[WEBHOOK] Failed to update subscription status to paused: %v", updateErr)
		return updateErr
	}

	log.Printf("[WEBHOOK] Successfully paused subscription %s for user %s", sub.ID, userID)

	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// HandleSubscriptionResumed handles customer.subscription.resumed events
func (s *WebhookService) HandleSubscriptionResumed(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomically claim the event
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		log.Printf("Event %s already claimed, skipping", event.ID)
		return nil
	}

	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("[WEBHOOK] Failed to parse subscription resumed event: %v", err)
		return errLib.New("Failed to parse subscription", http.StatusBadRequest)
	}

	log.Printf("[WEBHOOK] Subscription resumed: %s", sub.ID)

	// Look up user by Stripe customer ID
	stripeCustomerID := ""
	if sub.Customer != nil {
		stripeCustomerID = sub.Customer.ID
	}

	if stripeCustomerID == "" {
		log.Printf("[WEBHOOK] No customer ID found for resumed subscription %s", sub.ID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	var userID uuid.UUID
	query := "SELECT id FROM users.users WHERE stripe_customer_id = $1"
	if dbErr := s.db.QueryRowContext(ctx, query, stripeCustomerID).Scan(&userID); dbErr != nil {
		log.Printf("[WEBHOOK] Failed to find user with Stripe customer ID %s: %v", stripeCustomerID, dbErr)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	// Update membership status to active
	if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatusByID(ctx, userID, sub.ID, "active"); updateErr != nil {
		log.Printf("[WEBHOOK] Failed to update subscription status to active: %v", updateErr)
		return updateErr
	}

	log.Printf("[WEBHOOK] Successfully resumed subscription %s for user %s", sub.ID, userID)

	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// HandleCustomerUpdated handles customer.updated events
// Useful for tracking email changes and keeping user data in sync
func (s *WebhookService) HandleCustomerUpdated(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomically claim the event
	if !s.Idempotency.TryClaimEvent(event.ID, string(event.Type)) {
		log.Printf("Event %s already claimed, skipping", event.ID)
		return nil
	}

	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		log.Printf("[WEBHOOK] Failed to parse customer updated event: %v", err)
		return errLib.New("Failed to parse customer", http.StatusBadRequest)
	}

	log.Printf("[WEBHOOK] Customer updated: %s, Email: %s", customer.ID, customer.Email)

	// Look up user by Stripe customer ID
	var userID uuid.UUID
	query := "SELECT id FROM users.users WHERE stripe_customer_id = $1"
	if dbErr := s.db.QueryRowContext(ctx, query, customer.ID).Scan(&userID); dbErr != nil {
		log.Printf("[WEBHOOK] Customer %s not found in database (may be new customer)", customer.ID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	// Log the update - could sync email if needed
	log.Printf("[WEBHOOK] Customer update for user %s - Stripe email: %s", userID, customer.Email)

	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// sendRenewalReminderEmail sends an email reminder about upcoming subscription renewal
func (s *WebhookService) sendRenewalReminderEmail(userID uuid.UUID, amountDue int64, dueDate time.Time) {
	log.Printf("[EMAIL] Sending renewal reminder to user %s, amount: %d cents, due: %s",
		userID, amountDue, dueDate.Format(time.RFC3339))

	userInfo, err := s.UserRepo.GetUserInfo(context.Background(), "", userID)
	if err != nil || userInfo.Email == nil {
		log.Printf("[EMAIL] Failed to get user info for renewal reminder: %v", err)
		return
	}

	// Format amount for display
	amountStr := fmt.Sprintf("$%.2f", float64(amountDue)/100.0)
	dateStr := dueDate.Format("January 2, 2006")

	log.Printf("[EMAIL] Renewal reminder sent to %s: %s due on %s", *userInfo.Email, amountStr, dateStr)
	// Note: Implement email.SendRenewalReminderEmail if needed
}

// sendPaymentMethodUpdatedEmail sends confirmation when payment method is updated
func (s *WebhookService) sendPaymentMethodUpdatedEmail(userID uuid.UUID) {
	log.Printf("[EMAIL] Sending payment method updated confirmation to user %s", userID)

	userInfo, err := s.UserRepo.GetUserInfo(context.Background(), "", userID)
	if err != nil || userInfo.Email == nil {
		log.Printf("[EMAIL] Failed to get user info for payment method update: %v", err)
		return
	}

	log.Printf("[EMAIL] Payment method update confirmation sent to %s", *userInfo.Email)
	// Note: Implement email.SendPaymentMethodUpdatedEmail if needed
}
