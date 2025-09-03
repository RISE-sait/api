package payment

import (
	"api/internal/di"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	enrollment "api/internal/domains/enrollment/service"
	enrollmentRepo "api/internal/domains/enrollment/persistence/repository"
	repository "api/internal/domains/payment/persistence/repositories"
	errLib "api/internal/libs/errors"
	"api/internal/libs/logger"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	Idempotency            *WebhookIdempotency
	logger                 *logger.StructuredLogger
}

func NewWebhookService(container *di.Container) *WebhookService {
	return &WebhookService{
		PostCheckoutRepository: repository.NewPostCheckoutRepository(container),
		EnrollmentService:      enrollment.NewCustomerEnrollmentService(container),
		EnrollmentRepo:         enrollmentRepo.NewEnrollmentRepository(container),
		UserRepo:               identityRepo.NewUserRepository(container),
		PlansRepo:              membershipRepo.NewMembershipPlansRepository(container),
		Idempotency:            NewWebhookIdempotency(24*time.Hour, 10000), // Store events for 24 hours, max 10k events
		logger:                 logger.WithComponent("stripe-webhooks"),
	}
}

func (s *WebhookService) HandleCheckoutSessionCompleted(event stripe.Event) *errLib.CommonError {
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
		err = s.handleItemCheckoutComplete(checkSession, webhookLogger)
	case stripe.CheckoutSessionModeSubscription:
		err = s.handleSubscriptionCheckoutComplete(checkSession)
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
	}

	return err
}

func (s *WebhookService) handleItemCheckoutComplete(checkoutSession stripe.CheckoutSession, parentLogger *logger.StructuredLogger) *errLib.CommonError {
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

	switch {
	case programID != uuid.Nil && eventID != uuid.Nil:
		return errLib.New("price ID maps to both program and event", http.StatusConflict)
	case programID == uuid.Nil && eventID == uuid.Nil:
		return errLib.New("price ID doesn't map to any program or event", http.StatusNotFound)
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

func (s *WebhookService) handleSubscriptionCheckoutComplete(checkoutSession stripe.CheckoutSession) *errLib.CommonError {
	// 1. Validate and expand session
	fullSession, err := s.getExpandedSession(checkoutSession.ID)
	if err != nil {
		return err
	}

	// 2. Validate line items and get price ID
	priceIDs, err := s.validateLineItems(fullSession.LineItems)
	if err != nil {
		return err
	}

	priceID := priceIDs[0]

	// 3. Parse metadata
	userIdStr := fullSession.Metadata["userID"]

	if userIdStr == "" {
		return errLib.New("userID not found in metadata", http.StatusBadRequest)
	}

	userID, uuidErr := uuid.Parse(userIdStr)

	if uuidErr != nil {
		return errLib.New("Invalid user ID format", http.StatusBadRequest)
	}

	planID, amtPeriods, err := s.PostCheckoutRepository.GetMembershipPlanByStripePriceID(context.Background(), priceID)

	if err != nil {
		log.Printf("ERROR: GetMembershipPlanByStripePriceID failed for priceID %s: %v", priceID, err)
		return err
	}

	log.Printf("Found planID: %s, amtPeriods: %v for priceID: %s", planID, amtPeriods, priceID)

	if amtPeriods != nil {
		if err := s.processSubscriptionWithEndDate(
			fullSession.Subscription.ID,
			*amtPeriods,
			userID,
			planID,
		); err != nil {
			return err
		}
	} else {
		log.Printf("Enrolling customer %s in membership plan %s", userID, planID)
		if err := s.EnrollmentService.EnrollCustomerInMembershipPlan(context.Background(), userID, planID, time.Time{}); err != nil {
			log.Printf("ERROR: EnrollCustomerInMembershipPlan failed: %v", err)
			return err
		}
		log.Printf("Successfully enrolled customer %s in membership plan %s", userID, planID)
	}
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

func (s *WebhookService) processSubscriptionWithEndDate(subscriptionID string, totalBillingPeriods int32, userID, planID uuid.UUID) *errLib.CommonError {
	log.Printf("Processing subscription with end date: %s, periods: %d, user: %s, plan: %s", subscriptionID, totalBillingPeriods, userID, planID)
	
	sub, err := s.getExpandedSubscription(subscriptionID)
	if err != nil {
		log.Printf("ERROR: getExpandedSubscription failed: %v", err)
		return err
	}

	cancelAtUnix, cancelAtDateTime, err := s.calculateCancelAt(sub, int(totalBillingPeriods))
	if err != nil {
		log.Printf("ERROR: calculateCancelAt failed: %v", err)
		return err
	}

	log.Printf("Enrolling customer %s in plan %s with end date: %s", userID, planID, cancelAtDateTime)
	if err = s.EnrollmentService.EnrollCustomerInMembershipPlan(context.Background(), userID, planID, cancelAtDateTime); err != nil {
		log.Printf("ERROR: EnrollCustomerInMembershipPlan failed in processSubscriptionWithEndDate: %v", err)
		return err
	}

	s.sendMembershipPurchaseEmail(userID, planID)

	return s.updateSubscriptionCancelAt(sub.ID, cancelAtUnix)
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

func (s *WebhookService) calculateCancelAt(sub *stripe.Subscription, periods int) (cancelAtUnix int64, cancelAtDate time.Time, error *errLib.CommonError) {
	item := sub.Items.Data[0]
	interval := item.Price.Recurring.Interval
	intervalCount := int(item.Price.Recurring.IntervalCount)

	log.Printf("calculateCancelAt: interval=%s, intervalCount=%d, periods=%d", interval, intervalCount, periods)

	now := time.Now()
	var cancelTime time.Time

	switch interval {
	case stripe.PriceRecurringIntervalMonth:
		cancelTime = now.AddDate(0, intervalCount*periods, 0)
	case stripe.PriceRecurringIntervalYear:
		cancelTime = now.AddDate(intervalCount*periods, 0, 0)
	case stripe.PriceRecurringIntervalWeek:
		cancelTime = now.AddDate(0, 0, 7*intervalCount*periods)
	case stripe.PriceRecurringIntervalDay:
		cancelTime = now.AddDate(0, 0, intervalCount*periods)
	default:
		log.Printf("ERROR: Unsupported billing interval: %s", interval)
		return 0, time.Time{}, errLib.New("invalid billing interval: " + string(interval), http.StatusBadRequest)
	}

	return cancelTime.Unix(), cancelTime, nil

}

func (s *WebhookService) updateSubscriptionCancelAt(subscriptionID string, cancelAt int64) *errLib.CommonError {
	_, err := subscription.Update(
		subscriptionID,
		&stripe.SubscriptionParams{
			CancelAt: stripe.Int64(cancelAt),
		},
	)
	if err != nil {
		log.Printf("Failed to update subscription: %v", err)
		return errLib.New("Failed to set cancel date: "+err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (s *WebhookService) sendMembershipPurchaseEmail(userID, planID uuid.UUID) {
	userInfo, err := s.UserRepo.GetUserInfo(context.Background(), "", userID)
	if err != nil || userInfo.Email == nil {
		return
	}

	plan, pErr := s.PlansRepo.GetMembershipPlanById(context.Background(), planID)
	if pErr != nil {
		return
	}

	email.SendMembershipPurchaseEmail(*userInfo.Email, userInfo.FirstName, plan.Name)
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
