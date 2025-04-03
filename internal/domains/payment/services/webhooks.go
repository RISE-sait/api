package payment

import (
	"api/internal/di"
	repository "api/internal/domains/payment/persistence/repositories"
	types "api/internal/domains/payment/types"
	errLib "api/internal/libs/errors"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/product"
	"github.com/stripe/stripe-go/v81/subscription"
)

type WebhookService struct {
	PostCheckoutRepository *repository.PostCheckoutRepository
}

func NewWebhookService(container *di.Container) *WebhookService {
	return &WebhookService{
		PostCheckoutRepository: repository.NewPostCheckoutRepository(container),
	}
}

func (s *WebhookService) HandleCheckoutSessionCompleted(event stripe.Event) *errLib.CommonError {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return errLib.New("Failed to parse session", http.StatusBadRequest)
	}

	switch session.Mode {
	case stripe.CheckoutSessionModePayment:
		return s.handleItemCheckoutComplete(session)
	case stripe.CheckoutSessionModeSubscription:
		return s.handleSubscriptionWithEndDate(session)
	}

	return nil
}

func (s *WebhookService) handleItemCheckoutComplete(session stripe.CheckoutSession) *errLib.CommonError {

	userIDStr := session.Metadata["userID"]
	itemType := types.OneTimePaymentCheckoutItemType(session.Metadata["itemType"])

	if userIDStr == "" {
		return errLib.New("userID not found in metadata", http.StatusBadRequest)
	}

	if !types.IsOneTimePaymentCheckoutItemTypeValid(itemType) {
		return errLib.New("itemType not found in metadata", http.StatusBadRequest)
	}

	switch itemType {
	case types.Program:

		return s.handleProgramCheckoutComplete(session)
	}
	return nil
}

func (s *WebhookService) handleProgramCheckoutComplete(session stripe.CheckoutSession) *errLib.CommonError {

	userIDStr := session.Metadata["userID"]
	if userIDStr == "" {
		return errLib.New("userID not found in metadata", http.StatusBadRequest)
	}

	customerID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errLib.New("Failed to parse userID", http.StatusBadRequest)
	}

	var programID uuid.UUID

	if session.LineItems == nil || len(session.LineItems.Data) == 0 {
		log.Println(session.LineItems)
		return errLib.New("No line items found in session", http.StatusBadRequest)
	}

	// Get first line item (assuming single item checkout)
	lineItem := session.LineItems.Data[0]
	if lineItem.Price != nil && lineItem.Price.Product != nil {
		if idStr, exists := lineItem.Price.Product.Metadata["item_id"]; exists {
			programID, err = uuid.Parse(idStr)
			if err != nil {
				return errLib.New("failed to parse program ID from line items", http.StatusBadRequest)
			}
		}
	}

	if programID == uuid.Nil {
		return errLib.New("program ID not found in line items", http.StatusBadRequest)
	}

	if repoErr := s.PostCheckoutRepository.EnrollCustomerInProgramEvents(context.Background(), customerID, programID); repoErr != nil {
		return repoErr
	}

	return nil
}

func (s *WebhookService) handleSubscriptionWithEndDate(session stripe.CheckoutSession) *errLib.CommonError {

	totalBillingPeriodsStr := session.Metadata["totalBillingPeriods"]
	if totalBillingPeriodsStr == "" {
		return nil // Single payment, nothing to do
	}

	userIDStr := session.Metadata["userID"]

	if userIDStr == "" {
		return errLib.New("userID not found in metadata", http.StatusBadRequest)
	}

	totalBillingPeriods, err := strconv.Atoi(totalBillingPeriodsStr)
	if err != nil {
		return errLib.New("Failed to parse totalBillingPeriods", http.StatusBadRequest)
	}

	sub, err := subscription.Get(session.Subscription.ID, nil)
	if err != nil {
		return errLib.New("Failed to get subscription", http.StatusInternalServerError)
	}

	if len(sub.Items.Data) == 0 {
		return errLib.New("No items in subscription", http.StatusBadRequest)
	}

	subItem := sub.Items.Data[0]
	retrievedProduct, err := product.Get(subItem.Price.Product.ID, nil)
	if err != nil {
		log.Printf("Failed to get product: %v", err)
		return errLib.New("Failed to get product details", http.StatusInternalServerError)
	}

	retrievedPrice, err := price.Get(subItem.Price.ID, nil)
	if err != nil {
		return errLib.New("Failed to get price details", http.StatusInternalServerError)
	}

	_ = retrievedProduct.Name

	var cancelAt int64
	switch retrievedPrice.Recurring.Interval {
	case stripe.PriceRecurringIntervalMonth:
		cancelAt = time.Now().AddDate(0, int(retrievedPrice.Recurring.IntervalCount)*totalBillingPeriods, 0).Unix()
	case stripe.PriceRecurringIntervalYear:
		cancelAt = time.Now().AddDate(int(retrievedPrice.Recurring.IntervalCount)*totalBillingPeriods, 0, 0).Unix()
	case stripe.PriceRecurringIntervalWeek:
		cancelAt = time.Now().AddDate(0, 0, int(retrievedPrice.Recurring.IntervalCount)*totalBillingPeriods).Unix()
	default:
		return errLib.New("Invalid billing interval", http.StatusBadRequest)
	}

	if _, err = subscription.Update(
		sub.ID,
		&stripe.SubscriptionParams{
			CancelAt: stripe.Int64(cancelAt),
		},
	); err != nil {
		log.Printf("Failed to update subscription: %v", err)
		return errLib.New("Failed to set cancel date", http.StatusInternalServerError)
	}

	return nil
}
