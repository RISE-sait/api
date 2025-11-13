package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"

	"api/internal/domains/subsidy/dto"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	invoicepkg "github.com/stripe/stripe-go/v81/invoice"
	"github.com/stripe/stripe-go/v81/invoiceitem"
)

// getMapKeys returns the keys of a map for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getUserIDByStripeCustomerID retrieves userID from database using Stripe customer ID
func (s *WebhookService) getUserIDByStripeCustomerID(ctx context.Context, stripeCustomerID string) (uuid.UUID, error) {
	var userID uuid.UUID
	query := `SELECT id FROM users.users WHERE stripe_customer_id = $1`
	err := s.EnrollmentRepo.Db.QueryRowContext(ctx, query, stripeCustomerID).Scan(&userID)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

// HandleInvoiceCreated applies subsidy credit to invoices BEFORE customer is charged
func (s *WebhookService) HandleInvoiceCreated(event stripe.Event) *errLib.CommonError {
	// Idempotency check
	if s.Idempotency.IsProcessed(event.ID) {
		log.Printf("[SUBSIDY] Event %s already processed, skipping", event.ID)
		return nil
	}

	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("[SUBSIDY] Failed to parse invoice.created: %v", err)
		return errLib.New("Failed to parse invoice", http.StatusBadRequest)
	}

	log.Printf("[SUBSIDY] Invoice created: %s for customer: %s", invoice.ID, invoice.Customer.ID)

	// Check if invoice has a subscription ID directly in the parsed struct
	var subscriptionID string
	if invoice.Subscription != nil && invoice.Subscription.ID != "" {
		subscriptionID = invoice.Subscription.ID
		log.Printf("[SUBSIDY] Found subscription ID in parsed invoice: %s", subscriptionID)
	} else {
		log.Printf("[SUBSIDY] Invoice.Subscription is nil or empty, checking raw data...")
		// Try raw data as fallback
		var rawInvoiceData map[string]interface{}
		if err := json.Unmarshal(event.Data.Raw, &rawInvoiceData); err == nil {
			if subID, ok := rawInvoiceData["subscription"]; ok && subID != nil {
				if subIDStr, ok := subID.(string); ok && subIDStr != "" {
					subscriptionID = subIDStr
					log.Printf("[SUBSIDY] Found subscription ID in raw data: %s", subscriptionID)
				}
			} else {
				log.Printf("[SUBSIDY] No subscription field in invoice (might be a one-time invoice)")
			}
		}
	}

	// Extract customer metadata - in Stripe v81, customer is expanded
	if invoice.Customer == nil {
		log.Printf("[SUBSIDY] No customer info for invoice %s", invoice.ID)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil
	}

	// Check for subsidy in subscription metadata
	var subsidyID *uuid.UUID
	var userID uuid.UUID

	// Try to get subsidy info from subscription metadata
	if subscriptionID != "" {
		// Fetch the full subscription object to get metadata
		log.Printf("[SUBSIDY] Fetching subscription metadata for: %s", subscriptionID)
		sub, subErr := s.getExpandedSubscription(subscriptionID)
		if subErr == nil && sub != nil {
			subscriptionMetadata := sub.Metadata
			log.Printf("[SUBSIDY] Fetched subscription metadata: %v", subscriptionMetadata)

			// Check for subsidy in subscription metadata
			if subscriptionMetadata != nil {
				if hasSubsidy, exists := subscriptionMetadata["has_subsidy"]; exists && hasSubsidy == "true" {
					if sidStr, exists := subscriptionMetadata["subsidy_id"]; exists {
						sid, err := uuid.Parse(sidStr)
						if err == nil {
							subsidyID = &sid
							log.Printf("[SUBSIDY] Found subsidy ID in subscription metadata: %s", subsidyID)
						}
					}
					// Also get userID from subscription metadata
					if userIDStr, exists := subscriptionMetadata["userID"]; exists {
						uid, err := uuid.Parse(userIDStr)
						if err == nil {
							userID = uid
						}
					}
				}
			}
		} else {
			log.Printf("[SUBSIDY] Failed to fetch subscription: %v", subErr)
		}
	}

	// Fallback: Try customer metadata for userID
	log.Printf("[SUBSIDY] Customer metadata: %v", invoice.Customer.Metadata)
	if userID == uuid.Nil {
		if userIDStr, exists := invoice.Customer.Metadata["userID"]; exists {
			uid, err := uuid.Parse(userIDStr)
			if err == nil {
				userID = uid
				log.Printf("[SUBSIDY] Found userID in customer metadata: %s", userID)
			}
		} else {
			log.Printf("[SUBSIDY] No userID in customer metadata")
		}
	}

	// NEW: Fallback to database lookup by Stripe customer ID if still no userID
	if userID == uuid.Nil && invoice.Customer != nil && invoice.Customer.ID != "" {
		log.Printf("[SUBSIDY] Looking up userID by Stripe customer ID: %s", invoice.Customer.ID)
		uid, err := s.getUserIDByStripeCustomerID(context.Background(), invoice.Customer.ID)
		if err == nil {
			userID = uid
			log.Printf("[SUBSIDY] Found userID from database: %s", userID)
		} else {
			log.Printf("[SUBSIDY] Failed to find userID by Stripe customer ID: %v", err)
		}
	}

	// Fallback: Check database for active subsidy
	if subsidyID == nil && userID != uuid.Nil {
		log.Printf("[SUBSIDY] Checking database for active subsidy for user: %s", userID)
		subsidy, err := s.SubsidyService.GetActiveSubsidy(context.Background(), userID)
		if err == nil && subsidy != nil && subsidy.RemainingBalance > 0 {
			subsidyID = &subsidy.ID
			log.Printf("[SUBSIDY] Found active subsidy for user %s: $%.2f", userID, subsidy.RemainingBalance)
		} else {
			log.Printf("[SUBSIDY] No active subsidy found for user %s (err: %v, subsidy: %v)", userID, err, subsidy)
		}
	} else {
		log.Printf("[SUBSIDY] Skipping database lookup - subsidyID: %v, userID: %v", subsidyID, userID)
	}

	if subsidyID == nil {
		log.Printf("[SUBSIDY] No subsidy for invoice %s", invoice.ID)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil
	}

	// Get full subsidy details
	subsidy, err := s.SubsidyService.GetSubsidy(context.Background(), *subsidyID)
	if err != nil {
		log.Printf("[SUBSIDY] Error getting subsidy: %v", err)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil // Don't fail invoice creation
	}

	if subsidy.RemainingBalance <= 0 {
		log.Printf("[SUBSIDY] Subsidy %s depleted, skipping", subsidyID)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil
	}

	// Calculate how much subsidy to apply
	invoiceAmount := float64(invoice.AmountDue) / 100.0 // Convert cents to dollars
	subsidyToApply := math.Min(invoiceAmount, subsidy.RemainingBalance)
	subsidyInCents := int64(subsidyToApply * 100)

	log.Printf("[SUBSIDY] Applying $%.2f subsidy to invoice %s (balance: $%.2f)",
		subsidyToApply, invoice.ID, subsidy.RemainingBalance)

	// Add negative line item to invoice (credit)
	_, itemErr := invoiceitem.New(&stripe.InvoiceItemParams{
		Customer: stripe.String(invoice.Customer.ID),
		Invoice:  stripe.String(invoice.ID),
		Amount:   stripe.Int64(-subsidyInCents), // NEGATIVE = credit
		Currency: stripe.String("cad"),
		Description: stripe.String(fmt.Sprintf("Subsidy Credit - %s (Balance: $%.2f)",
			subsidy.Provider.Name, subsidy.RemainingBalance)),
		Metadata: map[string]string{
			"subsidy_id":      subsidyID.String(),
			"subsidy_applied": fmt.Sprintf("%.2f", subsidyToApply),
		},
	})

	if itemErr != nil {
		log.Printf("[SUBSIDY] Failed to add subsidy credit to invoice: %v", itemErr)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil // Don't fail invoice - proceed with normal amount
	}

	log.Printf("[SUBSIDY] Successfully applied $%.2f subsidy credit to invoice %s",
		subsidyToApply, invoice.ID)

	// Update invoice metadata with subsidy info so invoice.payment_succeeded can record usage
	_, updateErr := invoicepkg.Update(invoice.ID, &stripe.InvoiceParams{
		Metadata: map[string]string{
			"has_subsidy":     "true",
			"subsidy_id":      subsidyID.String(),
			"subsidy_balance": fmt.Sprintf("%.2f", subsidy.RemainingBalance),
			"subsidy_applied": fmt.Sprintf("%.2f", subsidyToApply),
			"userID":          userID.String(),
		},
	})

	if updateErr != nil {
		log.Printf("[SUBSIDY] Warning: Failed to update invoice metadata: %v", updateErr)
		// Don't fail - the credit was already applied
	} else {
		log.Printf("[SUBSIDY] Updated invoice metadata with subsidy info for usage recording")
	}

	s.Idempotency.MarkAsProcessed(event.ID)
	return nil
}

// HandleInvoiceFinalized applies subsidy credit at invoice finalization
// This fires AFTER subscription is linked and AFTER customer metadata is set, but BEFORE payment
func (s *WebhookService) HandleInvoiceFinalized(event stripe.Event) *errLib.CommonError {
	// Idempotency check
	if s.Idempotency.IsProcessed(event.ID) {
		log.Printf("[SUBSIDY] Event %s already processed, skipping", event.ID)
		return nil
	}

	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("[SUBSIDY] Failed to parse invoice.finalized: %v", err)
		return errLib.New("Failed to parse invoice", http.StatusBadRequest)
	}

	log.Printf("[SUBSIDY] Invoice finalized: %s for customer: %s", invoice.ID, invoice.Customer.ID)

	// Check if invoice has a subscription ID directly in the parsed struct
	var subscriptionID string
	if invoice.Subscription != nil && invoice.Subscription.ID != "" {
		subscriptionID = invoice.Subscription.ID
		log.Printf("[SUBSIDY] Found subscription ID in finalized invoice: %s", subscriptionID)
	}

	// Extract customer metadata
	if invoice.Customer == nil {
		log.Printf("[SUBSIDY] No customer info for invoice %s", invoice.ID)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil
	}

	// Check for subsidy in subscription metadata
	var subsidyID *uuid.UUID
	var userID uuid.UUID

	// Try to get subsidy info from subscription metadata
	if subscriptionID != "" {
		log.Printf("[SUBSIDY] Fetching subscription metadata for: %s", subscriptionID)
		sub, subErr := s.getExpandedSubscription(subscriptionID)
		if subErr == nil && sub != nil {
			subscriptionMetadata := sub.Metadata
			log.Printf("[SUBSIDY] Fetched subscription metadata: %v", subscriptionMetadata)

			// Check for subsidy in subscription metadata
			if subscriptionMetadata != nil {
				if hasSubsidy, exists := subscriptionMetadata["has_subsidy"]; exists && hasSubsidy == "true" {
					if sidStr, exists := subscriptionMetadata["subsidy_id"]; exists {
						sid, err := uuid.Parse(sidStr)
						if err == nil {
							subsidyID = &sid
							log.Printf("[SUBSIDY] Found subsidy ID in subscription metadata: %s", subsidyID)
						}
					}
					// Also get userID from subscription metadata
					if userIDStr, exists := subscriptionMetadata["userID"]; exists {
						uid, err := uuid.Parse(userIDStr)
						if err == nil {
							userID = uid
						}
					}
				}
			}
		} else {
			log.Printf("[SUBSIDY] Failed to fetch subscription: %v", subErr)
		}
	}

	// Fallback: Try customer metadata for userID
	log.Printf("[SUBSIDY] Customer metadata: %v", invoice.Customer.Metadata)
	if userID == uuid.Nil {
		if userIDStr, exists := invoice.Customer.Metadata["userID"]; exists {
			uid, err := uuid.Parse(userIDStr)
			if err == nil {
				userID = uid
				log.Printf("[SUBSIDY] Found userID in customer metadata: %s", userID)
			}
		}
	}

	// NEW: Fallback to database lookup by Stripe customer ID if still no userID
	if userID == uuid.Nil && invoice.Customer != nil && invoice.Customer.ID != "" {
		log.Printf("[SUBSIDY] Looking up userID by Stripe customer ID: %s", invoice.Customer.ID)
		uid, err := s.getUserIDByStripeCustomerID(context.Background(), invoice.Customer.ID)
		if err == nil {
			userID = uid
			log.Printf("[SUBSIDY] Found userID from database: %s", userID)
		} else {
			log.Printf("[SUBSIDY] Failed to find userID by Stripe customer ID: %v", err)
		}
	}

	// Fallback: Check database for active subsidy
	if subsidyID == nil && userID != uuid.Nil {
		log.Printf("[SUBSIDY] Checking database for active subsidy for user: %s", userID)
		subsidy, err := s.SubsidyService.GetActiveSubsidy(context.Background(), userID)
		if err == nil && subsidy != nil && subsidy.RemainingBalance > 0 {
			subsidyID = &subsidy.ID
			log.Printf("[SUBSIDY] Found active subsidy for user %s: $%.2f", userID, subsidy.RemainingBalance)
		}
	}

	if subsidyID == nil {
		log.Printf("[SUBSIDY] No subsidy for invoice %s", invoice.ID)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil
	}

	// Get full subsidy details
	subsidy, err := s.SubsidyService.GetSubsidy(context.Background(), *subsidyID)
	if err != nil {
		log.Printf("[SUBSIDY] Error getting subsidy: %v", err)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil // Don't fail invoice
	}

	if subsidy.RemainingBalance <= 0 {
		log.Printf("[SUBSIDY] Subsidy %s depleted, skipping", subsidyID)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil
	}

	// Calculate how much subsidy to apply
	invoiceAmount := float64(invoice.AmountDue) / 100.0
	subsidyToApply := math.Min(invoiceAmount, subsidy.RemainingBalance)
	subsidyInCents := int64(subsidyToApply * 100)

	log.Printf("[SUBSIDY] Applying $%.2f subsidy to invoice %s (balance: $%.2f)",
		subsidyToApply, invoice.ID, subsidy.RemainingBalance)

	// Add negative line item to invoice (credit)
	_, itemErr := invoiceitem.New(&stripe.InvoiceItemParams{
		Customer: stripe.String(invoice.Customer.ID),
		Invoice:  stripe.String(invoice.ID),
		Amount:   stripe.Int64(-subsidyInCents), // NEGATIVE = credit
		Currency: stripe.String("cad"),
		Description: stripe.String(fmt.Sprintf("Subsidy Credit - %s (Balance: $%.2f)",
			subsidy.Provider.Name, subsidy.RemainingBalance)),
		Metadata: map[string]string{
			"subsidy_id":      subsidyID.String(),
			"subsidy_applied": fmt.Sprintf("%.2f", subsidyToApply),
		},
	})

	if itemErr != nil {
		log.Printf("[SUBSIDY] Failed to add subsidy credit to invoice: %v", itemErr)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil // Don't fail invoice
	}

	log.Printf("[SUBSIDY] Successfully applied $%.2f subsidy credit to invoice %s",
		subsidyToApply, invoice.ID)

	s.Idempotency.MarkAsProcessed(event.ID)
	return nil
}

// HandleInvoicePaymentSucceededWithSubsidy records subsidy usage after successful payment
func (s *WebhookService) HandleInvoicePaymentSucceededWithSubsidy(event stripe.Event) *errLib.CommonError {
	// Idempotency check
	if s.Idempotency.IsProcessed(event.ID) {
		log.Printf("[SUBSIDY] Event %s already processed, skipping", event.ID)
		return nil
	}

	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("[SUBSIDY] Failed to parse invoice.payment_succeeded: %v", err)
		return errLib.New("Failed to parse invoice", http.StatusBadRequest)
	}

	log.Printf("[SUBSIDY] Invoice payment succeeded: %s", invoice.ID)
	log.Printf("[SUBSIDY] Invoice.Subscription: %v", invoice.Subscription)

	// Get subscription ID to fetch metadata
	var subscriptionID string
	if invoice.Subscription != nil && invoice.Subscription.ID != "" {
		subscriptionID = invoice.Subscription.ID
		log.Printf("[SUBSIDY] Found subscription ID in invoice: %s", subscriptionID)
	} else {
		log.Printf("[SUBSIDY] Invoice.Subscription is nil, checking raw data...")
		// Try to extract from raw data
		var rawInvoiceData map[string]interface{}
		if err := json.Unmarshal(event.Data.Raw, &rawInvoiceData); err == nil {
			if subID, ok := rawInvoiceData["subscription"]; ok && subID != nil {
				if subIDStr, ok := subID.(string); ok && subIDStr != "" {
					subscriptionID = subIDStr
					log.Printf("[SUBSIDY] Found subscription ID in raw data: %s", subscriptionID)
				}
			}
		}
	}

	// Extract user ID and subsidy info from subscription metadata
	var userID uuid.UUID
	var subsidyID *uuid.UUID
	var subsidyBalance float64
	var subsidyAppliedFromMetadata float64

	// FIRST: Check invoice metadata (set by invoice.created for recurring invoices)
	if invoice.Metadata != nil {
		log.Printf("[SUBSIDY] Invoice metadata: %v", invoice.Metadata)
		if hasSubsidy, exists := invoice.Metadata["has_subsidy"]; exists && hasSubsidy == "true" {
			if sidStr, exists := invoice.Metadata["subsidy_id"]; exists {
				sid, err := uuid.Parse(sidStr)
				if err == nil {
					subsidyID = &sid
					log.Printf("[SUBSIDY] Found subsidy ID in invoice metadata: %s", subsidyID)
				}
			}
			if balanceStr, exists := invoice.Metadata["subsidy_balance"]; exists {
				fmt.Sscanf(balanceStr, "%f", &subsidyBalance)
			}
			if appliedStr, exists := invoice.Metadata["subsidy_applied"]; exists {
				fmt.Sscanf(appliedStr, "%f", &subsidyAppliedFromMetadata)
			}
			if userIDStr, exists := invoice.Metadata["userID"]; exists {
				uid, err := uuid.Parse(userIDStr)
				if err == nil {
					userID = uid
					log.Printf("[SUBSIDY] Found userID in invoice metadata: %s", userID)
				}
			}
		}
	}

	// SECOND: Fallback to subscription metadata if invoice metadata doesn't have subsidy info
	if subscriptionID != "" && subsidyID == nil {
		log.Printf("[SUBSIDY] Fetching subscription metadata for: %s", subscriptionID)
		sub, subErr := s.getExpandedSubscription(subscriptionID)
		if subErr == nil && sub != nil {
			log.Printf("[SUBSIDY] Subscription metadata: %v", sub.Metadata)

			// Get userID from subscription metadata
			if userIDStr, exists := sub.Metadata["userID"]; exists {
				uid, err := uuid.Parse(userIDStr)
				if err == nil {
					userID = uid
				}
			}

			// Get subsidy info from subscription metadata
			if hasSubsidy, exists := sub.Metadata["has_subsidy"]; exists && hasSubsidy == "true" {
				if sidStr, exists := sub.Metadata["subsidy_id"]; exists {
					sid, err := uuid.Parse(sidStr)
					if err == nil {
						subsidyID = &sid
					}
				}
				if balanceStr, exists := sub.Metadata["subsidy_balance"]; exists {
					fmt.Sscanf(balanceStr, "%f", &subsidyBalance)
				}
			}
		}
	}

	// Fallback: try customer metadata for userID
	if userID == uuid.Nil && invoice.Customer != nil {
		if userIDStr, exists := invoice.Customer.Metadata["userID"]; exists {
			uid, err := uuid.Parse(userIDStr)
			if err == nil {
				userID = uid
			}
		}
	}

	// NEW: Fallback to database lookup by Stripe customer ID if still no userID
	if userID == uuid.Nil && invoice.Customer != nil && invoice.Customer.ID != "" {
		log.Printf("[SUBSIDY] Looking up userID by Stripe customer ID: %s", invoice.Customer.ID)
		uid, err := s.getUserIDByStripeCustomerID(context.Background(), invoice.Customer.ID)
		if err == nil {
			userID = uid
			log.Printf("[SUBSIDY] Found userID from database: %s", userID)
		} else {
			log.Printf("[SUBSIDY] Failed to find userID by Stripe customer ID: %v", err)
		}
	}

	if userID == uuid.Nil {
		log.Printf("[SUBSIDY] No userID for invoice %s", invoice.ID)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil
	}

	if subsidyID == nil {
		log.Printf("[SUBSIDY] No subsidy for invoice %s", invoice.ID)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil
	}

	// Calculate subsidy amount from invoice metadata (most accurate) or discount/total
	var subsidyApplied float64

	// FIRST: Use metadata if available (set by invoice.created)
	if subsidyAppliedFromMetadata > 0 {
		subsidyApplied = subsidyAppliedFromMetadata
		log.Printf("[SUBSIDY] Using subsidy amount from invoice metadata: $%.2f", subsidyApplied)
	} else if invoice.TotalDiscountAmounts != nil && len(invoice.TotalDiscountAmounts) > 0 {
		// SECOND: Calculate from invoice discounts
		for _, discount := range invoice.TotalDiscountAmounts {
			subsidyApplied += float64(discount.Amount) / 100.0
		}
		log.Printf("[SUBSIDY] Total discount on invoice: $%.2f", subsidyApplied)
	} else if subsidyBalance > 0 {
		// THIRD: Fallback to balance calculation
		invoiceTotal := float64(invoice.Total) / 100.0
		subsidyApplied = math.Min(invoiceTotal, subsidyBalance)
		log.Printf("[SUBSIDY] No discount found, calculated subsidy from balance: $%.2f", subsidyApplied)
	}

	if subsidyApplied == 0 {
		log.Printf("[SUBSIDY] No subsidy discount applied to invoice %s", invoice.ID)
		s.Idempotency.MarkAsProcessed(event.ID)
		return nil
	}

	// Calculate amounts
	originalAmount := (float64(invoice.AmountDue) + subsidyApplied*100) / 100.0
	customerPaid := float64(invoice.AmountPaid) / 100.0

	// Record subsidy usage - reuse subscriptionID from earlier
	if subscriptionID == "" && invoice.Subscription != nil {
		subscriptionID = invoice.Subscription.ID
	}

	recordReq := &dto.RecordUsageRequest{
		SubsidyID:            *subsidyID,
		CustomerID:           userID,
		TransactionType:      "membership_payment",
		OriginalAmount:       originalAmount,
		SubsidyApplied:       subsidyApplied,
		CustomerPaid:         customerPaid,
		StripeSubscriptionID: &subscriptionID,
		StripeInvoiceID:      &invoice.ID,
		Description:          fmt.Sprintf("Membership payment - Invoice %s", invoice.Number),
	}

	_, err := s.SubsidyService.RecordUsage(context.Background(), recordReq)
	if err != nil {
		log.Printf("[SUBSIDY] Failed to record subsidy usage: %v", err)
		// Don't fail the webhook - payment was successful
	} else {
		log.Printf("[SUBSIDY] Recorded subsidy usage: $%.2f applied, $%.2f paid by customer",
			subsidyApplied, customerPaid)
	}

	// Update membership status (existing logic)
	updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatus(context.Background(), userID, "active")
	if updateErr != nil {
		log.Printf("[SUBSIDY] Failed to activate membership: %v", updateErr)
		return updateErr
	}

	s.Idempotency.MarkAsProcessed(event.ID)
	return nil
}
