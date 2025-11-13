package payment

import (
	"context"
	"log"
	"net/http"

	errLib "api/internal/libs/errors"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
)

// getOrRecreateStripeCustomer retrieves existing Stripe customer ID or creates a new one if missing/deleted
// This handles the case where a Stripe customer was deleted but still exists in the database
func (s *Service) getOrRecreateStripeCustomer(ctx context.Context, userID uuid.UUID) (*string, *errLib.CommonError) {
	// First, try to get existing customer ID from database
	existingCustomerID := s.getExistingStripeCustomerID(ctx, userID)

	if existingCustomerID == nil {
		// No customer ID in database - will create new one via Stripe Checkout
		log.Printf("[CUSTOMER-RECOVERY] No existing Stripe customer for user %s", userID)
		return nil, nil
	}

	// Verify the customer exists in Stripe
	_, err := customer.Get(*existingCustomerID, nil)
	if err == nil {
		// Customer exists in Stripe - all good!
		log.Printf("[CUSTOMER-RECOVERY] Verified Stripe customer %s exists for user %s", *existingCustomerID, userID)
		return existingCustomerID, nil
	}

	// Check if it's a "customer not found" error
	if stripeErr, ok := err.(*stripe.Error); ok {
		if stripeErr.Code == stripe.ErrorCodeResourceMissing {
			log.Printf("[CUSTOMER-RECOVERY] ⚠️  Stripe customer %s not found for user %s - was deleted from Stripe", *existingCustomerID, userID)

			// Clear the invalid customer ID from database
			if clearErr := s.clearStripeCustomerID(ctx, userID); clearErr != nil {
				log.Printf("[CUSTOMER-RECOVERY] Failed to clear invalid customer ID: %v", clearErr)
				return nil, clearErr
			}

			log.Printf("[CUSTOMER-RECOVERY] ✅ Cleared invalid customer ID from database - new customer will be created")

			// Return nil so Stripe Checkout creates a fresh customer
			return nil, nil
		}
	}

	// Other Stripe errors (network, auth, etc.) - fail the request
	log.Printf("[CUSTOMER-RECOVERY] Error verifying Stripe customer %s: %v", *existingCustomerID, err)
	return nil, errLib.New("Failed to verify Stripe customer: "+err.Error(), http.StatusInternalServerError)
}

// clearStripeCustomerID removes an invalid Stripe customer ID from the database
func (s *Service) clearStripeCustomerID(ctx context.Context, userID uuid.UUID) *errLib.CommonError {
	query := "UPDATE users.users SET stripe_customer_id = NULL WHERE id = $1"
	_, err := s.DB.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("[CUSTOMER-RECOVERY] Failed to clear stripe_customer_id for user %s: %v", userID, err)
		return errLib.New("Failed to clear invalid customer ID", http.StatusInternalServerError)
	}

	log.Printf("[CUSTOMER-RECOVERY] Cleared stripe_customer_id for user %s", userID)
	return nil
}
