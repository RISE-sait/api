package stripe

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"api/internal/di"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"

	"github.com/stripe/stripe-go/v81"
	billingportal "github.com/stripe/stripe-go/v81/billingportal/session"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/coupon"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/webhook"
)

// init configures the Stripe client with proper timeouts when the package is imported
func init() {
	configureStripeTimeouts()
}

// configureStripeTimeouts sets up the Stripe client with proper timeouts
func configureStripeTimeouts() {
	httpClient := &http.Client{
		Timeout: CriticalStripeTimeout,
		Transport: &http.Transport{
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
		},
	}
	
	// Configure all Stripe backends to use our HTTP client with timeouts
	stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		HTTPClient: httpClient,
	})
	stripe.GetBackendWithConfig(stripe.UploadsBackend, &stripe.BackendConfig{
		HTTPClient: httpClient,
	})
}

// CreateOneTimePayment creates a Stripe Checkout Session for a one-time payment
func CreateOneTimePayment(
	ctx context.Context, // request-scoped context (for userID, cancellation, etc.)
	itemStripePriceID string, // Stripe Price ID of the item being purchased
	quantity int, // Number of items
	eventID *string, // Optional: Event ID for event enrollment payments
	stripeCouponID *string, // Optional: Stripe coupon ID for discounts
	successURL string, // Success redirect URL after payment
) (string, *errLib.CommonError) {
	// Create a timeout context for this operation
	timeoutCtx, cancel := withCriticalTimeout(ctx)
	defer cancel()

	// Check if the Stripe API key is set
	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	// Check if context is already cancelled or timed out
	select {
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return "", errLib.New("Request timeout while creating payment", http.StatusRequestTimeout)
		}
		return "", errLib.New("Request cancelled", http.StatusRequestTimeout)
	default:
		// Continue with the operation
	}

	// Validate input
	if itemStripePriceID == "" {
		return "", errLib.New("item stripe price ID cannot be empty", http.StatusBadRequest)
	}

	if quantity <= 0 {
		return "", errLib.New("quantity must be positive", http.StatusBadRequest)
	}

	if successURL == "" {
		return "", errLib.New("success URL cannot be empty", http.StatusBadRequest)
	}

	// Extract user ID from context (e.g. JWT or middleware-injected value)
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return "", err
	}

	// Metadata includes userID to track who made the payment and optionally eventID
	metadata := map[string]string{
		"userID": userID.String(),
	}
	
	// Add event ID to metadata if provided
	if eventID != nil && *eventID != "" {
		metadata["eventID"] = *eventID
	}
	
	params := &stripe.CheckoutSessionParams{
		Metadata: metadata,
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: metadata,
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(itemStripePriceID), // Price ID (pre-created in Stripe dashboard)
				Quantity: stripe.Int64(int64(quantity)),    // Number of items
			},
		},
		Mode:       stripe.String("payment"),                     // One-time payment mode
		SuccessURL: stripe.String(successURL), // Redirect URL after success
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
	}

	// Add discount coupon if provided
	if stripeCouponID != nil && *stripeCouponID != "" {
		params.Discounts = []*stripe.CheckoutSessionDiscountParams{
			{Coupon: stripe.String(*stripeCouponID)},
		}
		// Add coupon ID to metadata for tracking
		metadata["stripeCouponID"] = *stripeCouponID
		params.Metadata["stripeCouponID"] = *stripeCouponID
		params.PaymentIntentData.Metadata["stripeCouponID"] = *stripeCouponID
	}

	// Create Stripe session with timeout handling
	type sessionResult struct {
		session *stripe.CheckoutSession
		err     error
	}

	resultChan := make(chan sessionResult, 1)
	
	go func() {
		s, err := session.New(params)
		resultChan <- sessionResult{session: s, err: err}
	}()

	select {
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return "", errLib.New("Stripe API timeout while creating payment", http.StatusRequestTimeout)
		}
		return "", errLib.New("Request cancelled during payment creation", http.StatusRequestTimeout)
	case result := <-resultChan:
		if result.err != nil {
			return "", errLib.New("Payment session failed: "+result.err.Error(), http.StatusInternalServerError)
		}
		return result.session.URL, nil
	}
}

// CreateSubscriptionWithSetupFee creates a Stripe Checkout Session for a recurring subscription with optional setup fee
func CreateSubscriptionWithSetupFee(
	ctx context.Context,
	stripePlanPriceID string, // Stripe Price ID for the recurring plan
	setupFeeAmount int,       // Setup fee amount in cents (0 for no fee)
	successURL string, // Success redirect URL after payment
) (string, *errLib.CommonError) {
	// Create a timeout context for this operation
	timeoutCtx, cancel := withCriticalTimeout(ctx)
	defer cancel()

	// Extract user ID from context
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return "", err
	}

	// Check if Stripe is initialized
	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	// Check if context is already cancelled or timed out
	select {
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return "", errLib.New("Request timeout while creating subscription", http.StatusRequestTimeout)
		}
		return "", errLib.New("Request cancelled", http.StatusRequestTimeout)
	default:
		// Continue with the operation
	}

	// Validate input
	if stripePlanPriceID == "" {
		return "", errLib.New("item stripe price ID cannot be empty", http.StatusBadRequest)
	}

	if successURL == "" {
		return "", errLib.New("success URL cannot be empty", http.StatusBadRequest)
	}

	// Set up Checkout session with subscription mode
	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{
			"userID": userID.String(),
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"userID": userID.String(),
			},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(stripePlanPriceID), // Main subscription plan
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("subscription"), // Subscription mode
		SuccessURL: stripe.String(successURL),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
	}

	// Add setup fee if specified
	if setupFeeAmount > 0 {
		// Store setup fee amount in metadata for webhook processing
		params.SubscriptionData.Metadata["setup_fee_amount"] = fmt.Sprintf("%d", setupFeeAmount)
		params.Metadata["setup_fee_amount"] = fmt.Sprintf("%d", setupFeeAmount)

		// Add payment intent data to handle the setup fee on first payment
		params.PaymentIntentData = &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"setup_fee_amount": fmt.Sprintf("%d", setupFeeAmount),
			},
		}
	}

	// Ask Stripe to expand line item pricing and subscription in response
	params.AddExpand("line_items.data.price")
	params.AddExpand("subscription")

	// Create Stripe session with timeout handling
	type subscriptionResult struct {
		session *stripe.CheckoutSession
		err     error
	}

	resultChan := make(chan subscriptionResult, 1)
	
	go func() {
		s, err := session.New(params)
		resultChan <- subscriptionResult{session: s, err: err}
	}()

	select {
	case result := <-resultChan:
		if result.err != nil {
			return "", errLib.New("Failed to create Stripe session", http.StatusInternalServerError)
		}
		return result.session.URL, nil
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return "", errLib.New("Timeout while creating subscription session", http.StatusRequestTimeout)
		}
		return "", errLib.New("Request cancelled", http.StatusRequestTimeout)
	}
}

// CreateSubscription creates a Stripe Checkout Session for a recurring subscription
func CreateSubscription(
	ctx context.Context,
	stripePlanPriceID string, // Stripe Price ID for the recurring plan
	stripeJoiningFeesID string, // Optional one-time joining fee
	stripeCouponID *string, // Optional: Stripe coupon ID for discounts
	successURL string, // Success redirect URL after payment
) (string, *errLib.CommonError) {
	// Create a timeout context for this operation
	timeoutCtx, cancel := withCriticalTimeout(ctx)
	defer cancel()

	// Extract user ID from context
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return "", err
	}

	// Check if Stripe is initialized
	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	// Check if context is already cancelled or timed out
	select {
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return "", errLib.New("Request timeout while creating subscription", http.StatusRequestTimeout)
		}
		return "", errLib.New("Request cancelled", http.StatusRequestTimeout)
	default:
		// Continue with the operation
	}

	// Validate input
	if stripePlanPriceID == "" {
		return "", errLib.New("item stripe price ID cannot be empty", http.StatusBadRequest)
	}

	if successURL == "" {
		return "", errLib.New("success URL cannot be empty", http.StatusBadRequest)
	}

	// Set up Checkout session with subscription mode
	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{
			"userID": userID.String(),
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"userID": userID.String(),
			},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(stripePlanPriceID), // Main subscription plan
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("subscription"), // Subscription mode
		SuccessURL: stripe.String(successURL),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
	}

	// Ask Stripe to expand line item pricing and subscription in response
	params.AddExpand("line_items.data.price")
	params.AddExpand("subscription")

	// If there's a joining fee, add it as a second line item
	if stripeJoiningFeesID != "" {
		params.LineItems = append(params.LineItems, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(stripeJoiningFeesID),
			Quantity: stripe.Int64(1),
		})
	}

	// Add discount coupon if provided
	if stripeCouponID != nil && *stripeCouponID != "" {
		params.Discounts = []*stripe.CheckoutSessionDiscountParams{
			{Coupon: stripe.String(*stripeCouponID)},
		}
		// Add coupon ID to metadata for tracking
		params.Metadata["stripeCouponID"] = *stripeCouponID
		params.SubscriptionData.Metadata["stripeCouponID"] = *stripeCouponID
	}

	// Create Stripe session with timeout handling
	type subscriptionResult struct{
		session *stripe.CheckoutSession
		err     error
	}

	resultChan := make(chan subscriptionResult, 1)
	
	go func() {
		s, err := session.New(params)
		resultChan <- subscriptionResult{session: s, err: err}
	}()

	select {
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return "", errLib.New("Stripe API timeout while creating subscription", http.StatusRequestTimeout)
		}
		return "", errLib.New("Request cancelled during subscription creation", http.StatusRequestTimeout)
	case result := <-resultChan:
		if result.err != nil {
			return "", errLib.New("Subscription setup failed: "+result.err.Error(), http.StatusInternalServerError)
		}
		return result.session.URL, nil // Return URL to redirect client for payment
	}
}

// CreateSubscriptionWithDiscountPercent creates a subscription checkout session
// applying a percentage discount via a temporary coupon
func CreateSubscriptionWithDiscountPercent(
	ctx context.Context,
	stripePlanPriceID string,
	stripeJoiningFeesID string,
	discountPercent int,
	successURL string, // Success redirect URL after payment
) (string, *errLib.CommonError) {
	if discountPercent <= 0 || discountPercent > 100 {
		return "", errLib.New("invalid discount percent", http.StatusBadRequest)
	}

	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return "", err
	}

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	if stripePlanPriceID == "" {
		return "", errLib.New("item stripe price ID cannot be empty", http.StatusBadRequest)
	}

	if successURL == "" {
		return "", errLib.New("success URL cannot be empty", http.StatusBadRequest)
	}

	c, cuErr := coupon.New(&stripe.CouponParams{
		Duration:   stripe.String(string(stripe.CouponDurationOnce)),
		PercentOff: stripe.Float64(float64(discountPercent)),
	})
	if cuErr != nil {
		return "", errLib.New("failed to create coupon: "+cuErr.Error(), http.StatusInternalServerError)
	}

	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{
			"userID": userID.String(),
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"userID": userID.String(),
			},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(stripePlanPriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("subscription"),
		SuccessURL: stripe.String(successURL),
		Discounts: []*stripe.CheckoutSessionDiscountParams{
			{Coupon: stripe.String(c.ID)},
		},
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
	}

	params.AddExpand("line_items.data.price")
	params.AddExpand("subscription")

	if stripeJoiningFeesID != "" {
		params.LineItems = append(params.LineItems, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(stripeJoiningFeesID),
			Quantity: stripe.Int64(1),
		})
	}

	s, sessionErr := session.New(params)
	if sessionErr != nil {
		return "", errLib.New("Subscription setup failed: "+sessionErr.Error(), http.StatusInternalServerError)
	}

	return s.URL, nil
}

// SubscriptionService provides secure subscription management operations
type SubscriptionService struct {
	db *sql.DB
}

// NewSubscriptionService creates a new instance of SubscriptionService
func NewSubscriptionService(container *di.Container) *SubscriptionService {
	return &SubscriptionService{
		db: container.DB,
	}
}

// getDB returns the database connection
func (s *SubscriptionService) getDB() *sql.DB {
	return s.db
}

// GetSubscription retrieves a subscription by ID with security validation
func (s *SubscriptionService) GetSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, *errLib.CommonError) {
	if strings.TrimSpace(subscriptionID) == "" {
		return nil, errLib.New("subscription ID cannot be empty", http.StatusBadRequest)
	}

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return nil, errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Get subscription with expanded customer data
	params := &stripe.SubscriptionParams{
		Expand: []*string{
			stripe.String("customer"),
			stripe.String("items.data.price"),
			stripe.String("items.data.price.product"),
			stripe.String("latest_invoice"),
		},
	}

	sub, stripeErr := subscription.Get(subscriptionID, params)
	if stripeErr != nil {
		log.Printf("[STRIPE] Failed to get subscription %s: %v", subscriptionID, stripeErr)
		return nil, errLib.New("Failed to retrieve subscription: "+stripeErr.Error(), http.StatusInternalServerError)
	}

	// Verify ownership by checking if user has this Stripe customer ID in database
	var stripeCustomerID sql.NullString
	query := "SELECT stripe_customer_id FROM users.users WHERE id = $1"
	if dbErr := s.getDB().QueryRowContext(ctx, query, userID).Scan(&stripeCustomerID); dbErr != nil {
		if dbErr == sql.ErrNoRows {
			log.Printf("[STRIPE] Security violation: User %s not found in database", userID)
			return nil, errLib.New("Access denied", http.StatusForbidden)
		}
		log.Printf("[STRIPE] Database error during security check for user %s: %v", userID, dbErr)
		return nil, errLib.New("Security validation failed", http.StatusInternalServerError)
	}
	
	// Check if user has a Stripe customer ID and if it matches the subscription's customer
	if !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		log.Printf("[STRIPE] Security violation: User %s has no Stripe customer ID but tried to access subscription %s", userID, subscriptionID)
		return nil, errLib.New("Access denied", http.StatusForbidden)
	}
	
	// Verify the subscription belongs to this user's Stripe customer
	if sub.Customer == nil || sub.Customer.ID != stripeCustomerID.String {
		var actualCustomerID string
		if sub.Customer != nil {
			actualCustomerID = sub.Customer.ID
		}
		log.Printf("[STRIPE] Security violation: User %s (customer %s) attempted to access subscription %s owned by customer %s", userID, stripeCustomerID.String, subscriptionID, actualCustomerID)
		return nil, errLib.New("Access denied", http.StatusForbidden)
	}

	return sub, nil
}

// CancelSubscription cancels a subscription immediately or at period end
func (s *SubscriptionService) CancelSubscription(ctx context.Context, subscriptionID string, cancelImmediately bool) (*stripe.Subscription, *errLib.CommonError) {
	if strings.TrimSpace(subscriptionID) == "" {
		return nil, errLib.New("subscription ID cannot be empty", http.StatusBadRequest)
	}

	// First verify ownership
	sub, err := s.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	// Check if already cancelled
	if sub.Status == stripe.SubscriptionStatusCanceled {
		return sub, errLib.New("Subscription is already cancelled", http.StatusConflict)
	}

	// Cancel subscription
	var cancelledSub *stripe.Subscription
	var stripeErr error

	if cancelImmediately {
		// For immediate cancellation, use subscription.Cancel()
		log.Printf("[STRIPE] Attempting to cancel subscription %s immediately", subscriptionID)
		params := &stripe.SubscriptionCancelParams{}
		cancelledSub, stripeErr = subscription.Cancel(subscriptionID, params)
		if stripeErr == nil {
			log.Printf("[STRIPE] Stripe API returned cancelled subscription with status: %s", cancelledSub.Status)
		}
	} else {
		// For end-of-period cancellation, use subscription.Update()
		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
			Metadata: map[string]string{
				"cancelled_by": "user",
				"cancelled_at": time.Now().UTC().Format(time.RFC3339),
			},
		}
		cancelledSub, stripeErr = subscription.Update(subscriptionID, params)
	}
	if stripeErr != nil {
		log.Printf("[STRIPE] Failed to cancel subscription %s: %v", subscriptionID, stripeErr)
		return nil, errLib.New("Failed to cancel subscription: "+stripeErr.Error(), http.StatusInternalServerError)
	}

	log.Printf("[STRIPE] Successfully cancelled subscription %s (immediate: %v) - New status: %s", subscriptionID, cancelImmediately, cancelledSub.Status)
	return cancelledSub, nil
}

// PauseSubscription pauses a subscription for a specified duration
func (s *SubscriptionService) PauseSubscription(ctx context.Context, subscriptionID string, resumeAt *time.Time) (*stripe.Subscription, *errLib.CommonError) {
	if strings.TrimSpace(subscriptionID) == "" {
		return nil, errLib.New("subscription ID cannot be empty", http.StatusBadRequest)
	}

	// First verify ownership
	sub, err := s.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	// Check if subscription can be paused
	if sub.Status != stripe.SubscriptionStatusActive {
		return nil, errLib.New("Only active subscriptions can be paused", http.StatusConflict)
	}

	// Prepare pause collection behavior
	pauseBehavior := &stripe.SubscriptionPauseCollectionParams{
		Behavior: stripe.String(string(stripe.SubscriptionPauseCollectionBehaviorKeepAsDraft)),
	}

	if resumeAt != nil {
		pauseBehavior.ResumesAt = stripe.Int64(resumeAt.Unix())
	}

	params := &stripe.SubscriptionParams{
		PauseCollection: pauseBehavior,
		Metadata: map[string]string{
			"paused_by": "user",
			"paused_at": time.Now().UTC().Format(time.RFC3339),
		},
	}

	pausedSub, stripeErr := subscription.Update(subscriptionID, params)
	if stripeErr != nil {
		log.Printf("[STRIPE] Failed to pause subscription %s: %v", subscriptionID, stripeErr)
		return nil, errLib.New("Failed to pause subscription: "+stripeErr.Error(), http.StatusInternalServerError)
	}

	log.Printf("[STRIPE] Successfully paused subscription %s", subscriptionID)
	return pausedSub, nil
}

// ResumeSubscription resumes a paused subscription
func (s *SubscriptionService) ResumeSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, *errLib.CommonError) {
	if strings.TrimSpace(subscriptionID) == "" {
		return nil, errLib.New("subscription ID cannot be empty", http.StatusBadRequest)
	}

	// First verify ownership
	sub, err := s.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	// Check if subscription is paused
	if sub.PauseCollection == nil || sub.PauseCollection.Behavior != stripe.SubscriptionPauseCollectionBehaviorKeepAsDraft {
		return nil, errLib.New("Subscription is not paused", http.StatusConflict)
	}

	params := &stripe.SubscriptionParams{
		PauseCollection: &stripe.SubscriptionPauseCollectionParams{},
		Metadata: map[string]string{
			"resumed_by": "user",
			"resumed_at": time.Now().UTC().Format(time.RFC3339),
		},
	}

	resumedSub, stripeErr := subscription.Update(subscriptionID, params)
	if stripeErr != nil {
		log.Printf("[STRIPE] Failed to resume subscription %s: %v", subscriptionID, stripeErr)
		return nil, errLib.New("Failed to resume subscription: "+stripeErr.Error(), http.StatusInternalServerError)
	}

	log.Printf("[STRIPE] Successfully resumed subscription %s", subscriptionID)
	return resumedSub, nil
}

// GetCustomerSubscriptions retrieves all subscriptions for a customer with security validation
func (s *SubscriptionService) GetCustomerSubscriptions(ctx context.Context) ([]*stripe.Subscription, *errLib.CommonError) {
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return nil, errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	// Get Stripe customer ID from database
	var stripeCustomerID sql.NullString
	query := "SELECT stripe_customer_id FROM users.users WHERE id = $1"
	if dbErr := s.getDB().QueryRowContext(ctx, query, userID).Scan(&stripeCustomerID); dbErr != nil {
		if dbErr == sql.ErrNoRows {
			log.Printf("User %s not found in database", userID)
			return []*stripe.Subscription{}, nil
		}
		log.Printf("Database error getting user %s: %v", userID, dbErr)
		return nil, errLib.New("Failed to lookup user", http.StatusInternalServerError)
	}

	// If no Stripe customer ID is stored, return empty subscriptions
	if !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		log.Printf("No Stripe customer ID found for user %s", userID)
		return []*stripe.Subscription{}, nil
	}

	// Get subscriptions for this customer
	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(stripeCustomerID.String),
		Expand: []*string{
			stripe.String("data.items.data.price"),
			stripe.String("data.latest_invoice"),
		},
	}

	var subscriptions []*stripe.Subscription
	iter := subscription.List(params)
	for iter.Next() {
		subscriptions = append(subscriptions, iter.Subscription())
	}

	if iter.Err() != nil {
		log.Printf("[STRIPE] Failed to list subscriptions for customer %s: %v", stripeCustomerID.String, iter.Err())
		return nil, errLib.New("Failed to retrieve subscriptions: "+iter.Err().Error(), http.StatusInternalServerError)
	}

	return subscriptions, nil
}

// CreateCustomerPortalSession creates a secure customer portal session
func (s *SubscriptionService) CreateCustomerPortalSession(ctx context.Context, returnURL string) (string, *errLib.CommonError) {
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return "", err
	}

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	// Validate return URL
	if strings.TrimSpace(returnURL) == "" {
		return "", errLib.New("return URL cannot be empty", http.StatusBadRequest)
	}

	// Get Stripe customer ID from database
	var stripeCustomerID sql.NullString
	query := "SELECT stripe_customer_id FROM users.users WHERE id = $1"
	if dbErr := s.getDB().QueryRowContext(ctx, query, userID).Scan(&stripeCustomerID); dbErr != nil {
		if err == sql.ErrNoRows {
			log.Printf("User %s not found in database", userID)
			return "", errLib.New("User not found", http.StatusNotFound)
		}
		log.Printf("Database error getting user %s: %v", userID, err)
		return "", errLib.New("Failed to lookup user", http.StatusInternalServerError)
	}

	// If no Stripe customer ID is stored, return error
	if !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		log.Printf("No Stripe customer ID found for user %s", userID)
		return "", errLib.New("No subscription found - please contact support", http.StatusNotFound)
	}

	// Create billing portal session
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(stripeCustomerID.String),
		ReturnURL: stripe.String(returnURL),
	}

	session, stripeErr := billingportal.New(params)
	if stripeErr != nil {
		log.Printf("[STRIPE] Failed to create portal session for customer %s: %v", stripeCustomerID.String, stripeErr)
		return "", errLib.New("Failed to create portal session: "+stripeErr.Error(), http.StatusInternalServerError)
	}

	log.Printf("[STRIPE] Successfully created portal session for customer %s", stripeCustomerID.String)
	return session.URL, nil
}

// ValidateWebhookSignature validates Stripe webhook signatures for security
func ValidateWebhookSignature(payload []byte, signature, secret string) (*stripe.Event, *errLib.CommonError) {
	if len(payload) == 0 {
		return nil, errLib.New("Empty payload", http.StatusBadRequest)
	}

	if strings.TrimSpace(signature) == "" {
		return nil, errLib.New("Missing signature", http.StatusBadRequest)
	}

	if strings.TrimSpace(secret) == "" {
		return nil, errLib.New("Webhook secret not configured", http.StatusInternalServerError)
	}

	event, err := webhook.ConstructEventWithOptions(payload, signature, secret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("[STRIPE] Webhook signature verification failed: %v", err)
		return nil, errLib.New("Invalid signature", http.StatusBadRequest)
	}

	return &event, nil
}

// PriceService handles Stripe price operations
type PriceService struct{}

// NewPriceService creates a new price service instance
func NewPriceService() *PriceService {
	return &PriceService{}
}

// GetPrice retrieves a price from Stripe by price ID
func (s *PriceService) GetPrice(priceID string) (*stripe.Price, *errLib.CommonError) {
	if strings.TrimSpace(priceID) == "" {
		return nil, errLib.New("price ID cannot be empty", http.StatusBadRequest)
	}

	stripePrice, err := price.Get(priceID, nil)
	if err != nil {
		log.Printf("[STRIPE] Failed to get price %s: %v", priceID, err)
		return nil, errLib.New("Failed to retrieve price from Stripe: "+err.Error(), http.StatusInternalServerError)
	}

	return stripePrice, nil
}
