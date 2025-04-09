package payment

import (
	"api/internal/di"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	enrollment "api/internal/domains/enrollment/service"
	repository "api/internal/domains/payment/persistence/repositories"
	errLib "api/internal/libs/errors"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"log"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscription"
)

type WebhookService struct {
	PostCheckoutRepository *repository.PostCheckoutRepository
	EnrollmentService      *enrollment.CustomerEnrollmentService
}

func NewWebhookService(container *di.Container) *WebhookService {
	return &WebhookService{
		PostCheckoutRepository: repository.NewPostCheckoutRepository(container),
		EnrollmentService:      enrollment.NewCustomerEnrollmentService(container),
	}
}

func (s *WebhookService) HandleCheckoutSessionCompleted(event stripe.Event) *errLib.CommonError {
	var checkSession stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &checkSession); err != nil {
		log.Printf("Failed to parse session: %v", err)
		return errLib.New("Failed to parse session", http.StatusBadRequest)
	}

	switch checkSession.Mode {
	case stripe.CheckoutSessionModePayment:
		return s.handleProgramCheckoutComplete(checkSession)
	case stripe.CheckoutSessionModeSubscription:
		return s.handleSubscriptionCheckoutComplete(checkSession)
	}

	return nil
}

func (s *WebhookService) handleProgramCheckoutComplete(checkoutSession stripe.CheckoutSession) *errLib.CommonError {

	fullSession, err := s.getExpandedSession(checkoutSession.ID)
	if err != nil {
		return err
	}

	// 2. Validate line items and get price ID
	priceIDs, err := s.validateLineItems(fullSession.LineItems)
	if err != nil {
		return err
	}

	priceId := priceIDs[0]

	programID, err := s.PostCheckoutRepository.GetProgramIdByStripePriceId(context.Background(), priceId)

	if err != nil {
		return errLib.New("Invalid program ID format", http.StatusBadRequest)
	}

	userIDStr := fullSession.Metadata["userID"]

	if userIDStr == "" {
		return errLib.New("userID not found in metadata", http.StatusBadRequest)
	}

	customerID, uuidErr := uuid.Parse(userIDStr)

	if uuidErr != nil {
		return errLib.New("Invalid user ID format", http.StatusBadRequest)
	}

	if repoErr := s.EnrollmentService.EnrollCustomerInProgram(context.Background(), customerID, programID); repoErr != nil {
		return repoErr
	}

	if err = s.EnrollmentService.UpdateReservationStatusInProgram(context.Background(), programID, customerID, dbEnrollment.PaymentStatusPaid); err != nil {
		log.Printf("Failed to update reserve status: %v", err)
		return errLib.New("Failed to update reserve status", http.StatusInternalServerError)
	}

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
		return err
	}

	if amtPeriods != nil {
		return s.processSubscriptionWithEndDate(
			fullSession.Subscription.ID,
			*amtPeriods,
			userID,
			planID,
		)
	}
	return s.EnrollmentService.EnrollCustomerInMembershipPlan(context.Background(), userID, planID, time.Time{})
}

func (s *WebhookService) getExpandedSession(sessionID string) (*stripe.CheckoutSession, *errLib.CommonError) {
	params := &stripe.CheckoutSessionParams{
		Expand: []*string{
			stripe.String("line_items.data.price"),
			stripe.String("subscription"),
		},
	}

	checkoutSession, err := session.Get(sessionID, params)
	if err != nil {
		return nil, errLib.New("Failed to retrieve session details: "+err.Error(), http.StatusInternalServerError)
	}
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
	sub, err := s.getExpandedSubscription(subscriptionID)
	if err != nil {
		return err
	}

	cancelAtUnix, cancelAtDateTime, err := s.calculateCancelAt(sub, int(totalBillingPeriods))
	if err != nil {
		return err
	}

	if err = s.EnrollmentService.EnrollCustomerInMembershipPlan(context.Background(), userID, planID, cancelAtDateTime); err != nil {
		return err
	}

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

	now := time.Now()
	var cancelTime time.Time

	switch interval {
	case stripe.PriceRecurringIntervalMonth:
		cancelTime = now.AddDate(0, intervalCount*periods, 0)
	case stripe.PriceRecurringIntervalYear:
		cancelTime = now.AddDate(intervalCount*periods, 0, 0)
	case stripe.PriceRecurringIntervalWeek:
		cancelTime = now.AddDate(0, 0, intervalCount*periods)
	default:
		return 0, time.Time{}, errLib.New("invalid billing interval", http.StatusBadRequest)
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
