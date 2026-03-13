package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"api/internal/domains/subsidy/dto"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	invoicepkg "github.com/stripe/stripe-go/v81/invoice"
	"github.com/stripe/stripe-go/v81/invoiceitem"
	"github.com/stripe/stripe-go/v81/subscription"
)

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
func (s *WebhookService) HandleInvoiceCreated(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomic idempotency claim
	claimed, claimErr := s.Idempotency.TryClaimEvent(event.ID, string(event.Type))
	if claimErr != nil {
		log.Printf("[IDEMPOTENCY] DB error claiming event %s, failing closed: %v", event.ID, claimErr)
		return errLib.New("Idempotency check unavailable, will retry", http.StatusInternalServerError)
	}
	if !claimed {
		log.Printf("[SUBSIDY] Event %s already claimed, skipping", event.ID)
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
		var rawInvoiceData map[string]any
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
		s.Idempotency.MarkEventComplete(event.ID)
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

	// Fallback to database lookup by Stripe customer ID if still no userID
	if userID == uuid.Nil && invoice.Customer != nil && invoice.Customer.ID != "" {
		log.Printf("[SUBSIDY] Looking up userID by Stripe customer ID: %s", invoice.Customer.ID)
		uid, err := s.getUserIDByStripeCustomerID(ctx, invoice.Customer.ID)
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
		subsidy, err := s.SubsidyService.GetActiveSubsidy(ctx, userID)
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
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	// Get full subsidy details
	subsidy, err := s.SubsidyService.GetSubsidy(ctx, *subsidyID)
	if err != nil {
		log.Printf("[SUBSIDY] Error getting subsidy: %v", err)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil // Don't fail invoice creation
	}

	if subsidy.RemainingBalance <= 0 {
		log.Printf("[SUBSIDY] Subsidy %s depleted, skipping", subsidyID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	// Calculate how much subsidy to apply
	invoiceAmount := float64(invoice.AmountDue) / 100.0 // Convert cents to dollars
	subsidyToApply := math.Min(invoiceAmount, subsidy.RemainingBalance)
	subsidyInCents := int64(subsidyToApply * 100)

	log.Printf("[SUBSIDY] Applying $%.2f subsidy to invoice %s (balance: $%.2f)",
		subsidyToApply, invoice.ID, subsidy.RemainingBalance)

	// Add negative line item to invoice (credit)
	itemParams := &stripe.InvoiceItemParams{
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
	}
	itemParams.IdempotencyKey = stripe.String("subsidy-credit:" + event.ID + ":" + invoice.ID)
	_, itemErr := invoiceitem.New(itemParams)

	if itemErr != nil {
		log.Printf("[SUBSIDY] Failed to add subsidy credit to invoice: %v", itemErr)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil // Don't fail invoice - proceed with normal amount
	}

	log.Printf("[SUBSIDY] Successfully applied $%.2f subsidy credit to invoice %s",
		subsidyToApply, invoice.ID)

	// Update invoice metadata with subsidy info so invoice.payment_succeeded can record usage
	invoiceUpdateParams := &stripe.InvoiceParams{
		Metadata: map[string]string{
			"has_subsidy":     "true",
			"subsidy_id":      subsidyID.String(),
			"subsidy_balance": fmt.Sprintf("%.2f", subsidy.RemainingBalance),
			"subsidy_applied": fmt.Sprintf("%.2f", subsidyToApply),
			"userID":          userID.String(),
		},
	}
	invoiceUpdateParams.IdempotencyKey = stripe.String("subsidy-meta:" + event.ID + ":" + invoice.ID)
	_, updateErr := invoicepkg.Update(invoice.ID, invoiceUpdateParams)

	if updateErr != nil {
		log.Printf("[SUBSIDY] Warning: Failed to update invoice metadata: %v", updateErr)
		// Don't fail - the credit was already applied
	} else {
		log.Printf("[SUBSIDY] Updated invoice metadata with subsidy info for usage recording")
	}

	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// HandleInvoiceFinalized applies subsidy credit at invoice finalization
// This fires AFTER subscription is linked and AFTER customer metadata is set, but BEFORE payment
func (s *WebhookService) HandleInvoiceFinalized(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomic idempotency claim
	claimed, claimErr := s.Idempotency.TryClaimEvent(event.ID, string(event.Type))
	if claimErr != nil {
		log.Printf("[IDEMPOTENCY] DB error claiming event %s, failing closed: %v", event.ID, claimErr)
		return errLib.New("Idempotency check unavailable, will retry", http.StatusInternalServerError)
	}
	if !claimed {
		log.Printf("[SUBSIDY] Event %s already claimed, skipping", event.ID)
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
		s.Idempotency.MarkEventComplete(event.ID)
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

	// Fallback to database lookup by Stripe customer ID if still no userID
	if userID == uuid.Nil && invoice.Customer != nil && invoice.Customer.ID != "" {
		log.Printf("[SUBSIDY] Looking up userID by Stripe customer ID: %s", invoice.Customer.ID)
		uid, err := s.getUserIDByStripeCustomerID(ctx, invoice.Customer.ID)
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
		subsidy, err := s.SubsidyService.GetActiveSubsidy(ctx, userID)
		if err == nil && subsidy != nil && subsidy.RemainingBalance > 0 {
			subsidyID = &subsidy.ID
			log.Printf("[SUBSIDY] Found active subsidy for user %s: $%.2f", userID, subsidy.RemainingBalance)
		}
	}

	if subsidyID == nil {
		log.Printf("[SUBSIDY] No subsidy for invoice %s", invoice.ID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	// Get full subsidy details
	subsidy, err := s.SubsidyService.GetSubsidy(ctx, *subsidyID)
	if err != nil {
		log.Printf("[SUBSIDY] Error getting subsidy: %v", err)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil // Don't fail invoice
	}

	if subsidy.RemainingBalance <= 0 {
		log.Printf("[SUBSIDY] Subsidy %s depleted, skipping", subsidyID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	// Calculate how much subsidy to apply
	invoiceAmount := float64(invoice.AmountDue) / 100.0
	subsidyToApply := math.Min(invoiceAmount, subsidy.RemainingBalance)
	subsidyInCents := int64(subsidyToApply * 100)

	log.Printf("[SUBSIDY] Applying $%.2f subsidy to invoice %s (balance: $%.2f)",
		subsidyToApply, invoice.ID, subsidy.RemainingBalance)

	// Add negative line item to invoice (credit)
	finalizedItemParams := &stripe.InvoiceItemParams{
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
	}
	finalizedItemParams.IdempotencyKey = stripe.String("subsidy-finalized:" + event.ID + ":" + invoice.ID)
	_, itemErr := invoiceitem.New(finalizedItemParams)

	if itemErr != nil {
		log.Printf("[SUBSIDY] Failed to add subsidy credit to invoice: %v", itemErr)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil // Don't fail invoice
	}

	log.Printf("[SUBSIDY] Successfully applied $%.2f subsidy credit to invoice %s",
		subsidyToApply, invoice.ID)

	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}

// HandleInvoicePaymentSucceededWithSubsidy handles invoice.payment_succeeded for ALL customers.
// It ALWAYS runs core payment logic (status update, next_billing_date, tracking),
// then additionally records subsidy usage if the customer has a subsidy.
func (s *WebhookService) HandleInvoicePaymentSucceededWithSubsidy(ctx context.Context, event stripe.Event) *errLib.CommonError {
	// Atomic idempotency claim
	claimed, claimErr := s.Idempotency.TryClaimEvent(event.ID, string(event.Type))
	if claimErr != nil {
		log.Printf("[IDEMPOTENCY] DB error claiming event %s, failing closed: %v", event.ID, claimErr)
		return errLib.New("Idempotency check unavailable, will retry", http.StatusInternalServerError)
	}
	if !claimed {
		log.Printf("[WEBHOOK] Event %s already claimed, skipping", event.ID)
		return nil
	}

	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("[WEBHOOK] Failed to parse invoice.payment_succeeded: %v", err)
		return errLib.New("Failed to parse invoice", http.StatusBadRequest)
	}

	log.Printf("[WEBHOOK] Invoice payment succeeded: %s", invoice.ID)

	// Get subscription ID
	var subscriptionID string
	if invoice.Subscription != nil && invoice.Subscription.ID != "" {
		subscriptionID = invoice.Subscription.ID
		log.Printf("[WEBHOOK] Found subscription ID in invoice: %s", subscriptionID)
	} else {
		// Try to extract from raw data
		var rawInvoiceData map[string]any
		if err := json.Unmarshal(event.Data.Raw, &rawInvoiceData); err == nil {
			if subID, ok := rawInvoiceData["subscription"]; ok && subID != nil {
				if subIDStr, ok := subID.(string); ok && subIDStr != "" {
					subscriptionID = subIDStr
					log.Printf("[WEBHOOK] Found subscription ID in raw data: %s", subscriptionID)
				}
			}
		}
	}

	// --- CORE LOGIC: runs for ALL customers ---

	// Look up user by Stripe customer ID
	customerID := ""
	if invoice.Customer != nil {
		customerID = invoice.Customer.ID
	}

	if customerID == "" {
		log.Printf("[WEBHOOK] No customer ID found for invoice %s", invoice.ID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	var userID uuid.UUID

	// Try database lookup first (most reliable)
	uid, err := s.getUserIDByStripeCustomerID(ctx, customerID)
	if err == nil {
		userID = uid
		log.Printf("[WEBHOOK] Found user %s for Stripe customer %s", userID, customerID)
	}

	// Fallback: try customer metadata
	if userID == uuid.Nil && invoice.Customer != nil {
		if userIDStr, exists := invoice.Customer.Metadata["userID"]; exists {
			if uid, err := uuid.Parse(userIDStr); err == nil {
				userID = uid
				log.Printf("[WEBHOOK] Found userID in customer metadata: %s", userID)
			}
		}
	}

	if userID == uuid.Nil {
		log.Printf("[WEBHOOK] No userID for invoice %s, skipping", invoice.ID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	// Update membership status and next_billing_date
	eventTime := time.Unix(event.Created, 0)
	if subscriptionID != "" {
		sub, subErr := subscription.Get(subscriptionID, nil)
		if subErr != nil {
			log.Printf("[WEBHOOK] Failed to get subscription details for %s: %v", subscriptionID, subErr)
			// Fall back to just updating status by subscription ID
			if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatusByID(ctx, userID, subscriptionID, "active", eventTime); updateErr != nil {
				log.Printf("[WEBHOOK] Failed to activate membership: %v", updateErr)
				return updateErr
			}
		} else {
			nextBillingDate := time.Unix(sub.CurrentPeriodEnd, 0)
			log.Printf("[WEBHOOK] Updating membership: status=active, next_billing=%s for subscription %s", nextBillingDate.Format(time.RFC3339), subscriptionID)
			if updateErr := s.EnrollmentRepo.UpdateStripeSubscriptionStatusByIDAndNextBilling(ctx, userID, subscriptionID, "active", nextBillingDate, eventTime); updateErr != nil {
				log.Printf("[WEBHOOK] Failed to update membership: %v", updateErr)
				return updateErr
			}
			log.Printf("[WEBHOOK] Successfully updated next_billing_date to %s", nextBillingDate.Format(time.RFC3339))
		}
	} else {
		// No subscription ID — log warning and skip rather than updating all subscriptions
		log.Printf("[WEBHOOK] WARNING: No subscription ID for invoice %s, skipping membership status update to avoid affecting other subscriptions", invoice.ID)
	}

	log.Printf("[WEBHOOK] Successfully activated membership for user %s after invoice payment %s", userID, invoice.ID)

	// Track payment — skip initial subscription invoice (already tracked by checkout)
	if invoice.BillingReason != stripe.InvoiceBillingReasonSubscriptionCreate {
		safeGo("trackMembershipRenewal", func() { s.trackMembershipRenewal(&invoice, userID, eventTime) })
	} else {
		log.Printf("[WEBHOOK] Skipping payment tracking for initial subscription invoice %s (already tracked by checkout)", invoice.ID)
	}

	// --- SUBSIDY LOGIC: only runs if customer has a subsidy ---

	var subsidyID *uuid.UUID
	var subsidyBalance float64
	var subsidyAppliedFromMetadata float64

	// Check invoice metadata (set by invoice.created handler)
	if invoice.Metadata != nil {
		if hasSubsidy, exists := invoice.Metadata["has_subsidy"]; exists && hasSubsidy == "true" {
			log.Printf("[SUBSIDY] Invoice metadata: %v", invoice.Metadata)
			if sidStr, exists := invoice.Metadata["subsidy_id"]; exists {
				sid, parseErr := uuid.Parse(sidStr)
				if parseErr == nil {
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
		}
	}

	// Fallback to subscription metadata
	if subscriptionID != "" && subsidyID == nil {
		log.Printf("[SUBSIDY] Fetching subscription metadata for: %s", subscriptionID)
		sub, subErr := s.getExpandedSubscription(subscriptionID)
		if subErr == nil && sub != nil {
			if hasSubsidy, exists := sub.Metadata["has_subsidy"]; exists && hasSubsidy == "true" {
				if sidStr, exists := sub.Metadata["subsidy_id"]; exists {
					sid, parseErr := uuid.Parse(sidStr)
					if parseErr == nil {
						subsidyID = &sid
					}
				}
				if balanceStr, exists := sub.Metadata["subsidy_balance"]; exists {
					fmt.Sscanf(balanceStr, "%f", &subsidyBalance)
				}
			}
		}
	}

	// No subsidy — we're done (core logic already ran above)
	if subsidyID == nil {
		log.Printf("[WEBHOOK] No subsidy for invoice %s, core logic complete", invoice.ID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	// Calculate subsidy amount
	var subsidyApplied float64
	if subsidyAppliedFromMetadata > 0 {
		subsidyApplied = subsidyAppliedFromMetadata
		log.Printf("[SUBSIDY] Using subsidy amount from invoice metadata: $%.2f", subsidyApplied)
	} else if len(invoice.TotalDiscountAmounts) > 0 {
		for _, discount := range invoice.TotalDiscountAmounts {
			subsidyApplied += float64(discount.Amount) / 100.0
		}
		log.Printf("[SUBSIDY] Total discount on invoice: $%.2f", subsidyApplied)
	} else if subsidyBalance > 0 {
		invoiceTotal := float64(invoice.Total) / 100.0
		subsidyApplied = math.Min(invoiceTotal, subsidyBalance)
		log.Printf("[SUBSIDY] No discount found, calculated subsidy from balance: $%.2f", subsidyApplied)
	}

	if subsidyApplied == 0 {
		log.Printf("[SUBSIDY] No subsidy discount applied to invoice %s", invoice.ID)
		s.Idempotency.MarkEventComplete(event.ID)
		return nil
	}

	// Record subsidy usage
	originalAmount := (float64(invoice.AmountDue) + subsidyApplied*100) / 100.0
	customerPaid := float64(invoice.AmountPaid) / 100.0

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

	_, recordErr := s.SubsidyService.RecordUsage(ctx, recordReq)
	if recordErr != nil {
		log.Printf("[SUBSIDY] Failed to record subsidy usage: %v", recordErr)
		// Don't fail the webhook - payment was successful and core logic already ran
	} else {
		log.Printf("[SUBSIDY] Recorded subsidy usage: $%.2f applied, $%.2f paid by customer",
			subsidyApplied, customerPaid)
	}

	s.Idempotency.MarkEventComplete(event.ID)
	return nil
}
