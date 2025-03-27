package payment

import (
	errLib "api/internal/libs/errors"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/product"
	"github.com/stripe/stripe-go/v81/subscription"
)

func HandleCheckoutSessionCompleted(event stripe.Event) *errLib.CommonError {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return errLib.New("Failed to parse session", http.StatusBadRequest)
	}

	totalBillingPeriodsStr := session.Metadata["totalBillingPeriods"]
	if totalBillingPeriodsStr == "" {
		return nil // Single payment, nothing to do
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
