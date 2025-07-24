package payment

import (
	"api/internal/di"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	enrollment "api/internal/domains/enrollment/service"
	identityRepo "api/internal/domains/identity/persistence/repository/user"
	membershipRepo "api/internal/domains/membership/persistence/repositories"
	repository "api/internal/domains/payment/persistence/repositories"
	errLib "api/internal/libs/errors"
	"api/utils/email"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/subscription"
)

type WebhookService struct {
	PostCheckoutRepository *repository.PostCheckoutRepository
	EnrollmentService      *enrollment.CustomerEnrollmentService
	UserRepo               *identityRepo.UsersRepository
	PlansRepo              *membershipRepo.PlansRepository
	SquareServiceURL       string
}

func NewWebhookService(container *di.Container) *WebhookService {
	return &WebhookService{
		PostCheckoutRepository: repository.NewPostCheckoutRepository(container),
		EnrollmentService:      enrollment.NewCustomerEnrollmentService(container),
		UserRepo:               identityRepo.NewUserRepository(container),
		PlansRepo:              membershipRepo.NewMembershipPlansRepository(container),
		SquareServiceURL:       os.Getenv("SQUARE_SERVICE_URL"),
	}
}

func (s *WebhookService) HandleCheckoutSessionCompleted(event stripe.Event) *errLib.CommonError {
	var checkSession stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &checkSession); err != nil {
		log.Printf("Failed to parse session: %v", err)
		return errLib.New("Failed to parse session", http.StatusBadRequest)
	}

	log.Println("Parsed checkout session with ID:", checkSession.ID)
	log.Println("Session Mode:", checkSession.Mode)

	switch checkSession.Mode {
	case stripe.CheckoutSessionModePayment:
		return s.handleItemCheckoutComplete(checkSession)
	case stripe.CheckoutSessionModeSubscription:
		return s.handleSubscriptionCheckoutComplete(checkSession)
	default:
		log.Println("Unhandled session mode:", checkSession.Mode)
	}

	return nil
}

func (s *WebhookService) handleItemCheckoutComplete(checkoutSession stripe.CheckoutSession) *errLib.CommonError {
	log.Println("Expanding session:", checkoutSession.ID)

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

	programID, err := s.PostCheckoutRepository.GetProgramIdByStripePriceId(context.Background(), priceID)
	if err != nil {
		log.Printf("GetProgramIdByStripePriceId failed: %v", err)
		return err
	}
	eventID, err := s.PostCheckoutRepository.GetEventIdByStripePriceId(context.Background(), priceID)
	if err != nil {
		log.Printf("GetEventIdByStripePriceId failed: %v", err)
		return err
	}

	log.Println("ProgramID:", programID)
	log.Println("EventID:", eventID)

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
		return err
	}

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
		if err := s.EnrollmentService.EnrollCustomerInMembershipPlan(context.Background(), userID, planID, time.Time{}); err != nil {
			return err
		}
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

type squareWebhookEvent struct {
	Data struct {
		Object map[string]interface{} `json:"object"`
	} `json:"data"`
}

// HandleSquareWebhook processes Square webhook payloads sent from the Python service.
// It extracts the plan variation identifier for future processing.
func (s *WebhookService) HandleSquareWebhook(payload []byte) *errLib.CommonError {
	var event squareWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errLib.New("failed to parse Square webhook: "+err.Error(), http.StatusBadRequest)
	}

	obj := event.Data.Object

	var planVariationID string
	if v, ok := obj["plan_variation_id"].(string); ok && v != "" {
		planVariationID = v
	} else if v, ok := obj["plan_id"].(string); ok && v != "" {
		planVariationID = v
	}

	if planVariationID == "" {
		return errLib.New("plan_variation_id cannot be empty", http.StatusBadRequest)
	}

	customerID, _ := obj["customer_id"].(string)
	if customerID == "" {
		return errLib.New("customer_id cannot be empty", http.StatusBadRequest)
	}

	// Retrieve Square customer via the Python service
	ctx := context.Background()
	if s.SquareServiceURL == "" {
		return errLib.New("square service url not configured", http.StatusInternalServerError)
	}
	resp, err := http.Get(s.SquareServiceURL + "/customers/" + customerID)
	if err != nil {
		log.Printf("failed to fetch Square customer %s: %v", customerID, err)
		return errLib.New("failed to fetch Square customer", http.StatusInternalServerError)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return errLib.New(string(data), resp.StatusCode)
	}

	type customerResp struct {
		Customer struct {
			ReferenceID *string `json:"reference_id"`
		} `json:"customer"`
	}
	var custResp customerResp
	json.Unmarshal(data, &custResp)
	var referenceID string
	if custResp.Customer.ReferenceID != nil {
		referenceID = *custResp.Customer.ReferenceID
	}

	if referenceID == "" {
		return errLib.New("reference_id missing in customer", http.StatusBadRequest)
	}

	userUUID, uuidErr := uuid.Parse(referenceID)
	if uuidErr != nil {
		return errLib.New("invalid reference_id format", http.StatusBadRequest)
	}

	planID, _, err := s.PostCheckoutRepository.GetMembershipPlanByStripePriceID(ctx, planVariationID)
	if planID == uuid.Nil {
		if commonErr, ok := err.(*errLib.CommonError); ok {
			return commonErr
		}
		return errLib.New("membership plan not found for given plan_variation_id", http.StatusNotFound)
	}

	if err := s.EnrollmentService.EnrollCustomerInMembershipPlan(ctx, userUUID, planID, time.Time{}); err != nil {
		return err
	}

	s.sendMembershipPurchaseEmail(userUUID, planID)

	log.Println("Square webhook plan variation id:", planVariationID)
	return nil
}
