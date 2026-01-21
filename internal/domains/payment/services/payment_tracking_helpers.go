package payment

import (
	"api/internal/domains/payment/tracking"
	creditPackageDTO "api/internal/domains/credit_package/dto"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
)

// trackCreditPackagePurchase tracks a credit package purchase in the centralized payment system
func (s *WebhookService) trackCreditPackagePurchase(session *stripe.CheckoutSession, customerID uuid.UUID, creditPackage interface{}, transactionDate time.Time, receiptURL string) {
	ctx := context.Background()

	// Get customer details
	user, err := s.UserRepo.GetUserInfo(ctx, "", customerID)
	if err != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to get user info for %s: %v", customerID, err)
		return
	}

	customerEmail := ""
	if user.Email != nil {
		customerEmail = *user.Email
	}
	customerName := user.FirstName + " " + user.LastName

	// Calculate amounts from session
	originalAmount := float64(session.AmountTotal) / 100.0
	customerPaid := originalAmount
	subsidyAmount := 0.0
	discountAmount := 0.0

	// Check for discounts
	if session.TotalDetails != nil && session.TotalDetails.AmountDiscount > 0 {
		discountAmount = float64(session.TotalDetails.AmountDiscount) / 100.0
	}

	// Extract credit package ID - creditPackage is of type *dto.CreditPackageResponse
	var packageID *uuid.UUID
	description := "Credit package purchase"

	// Type assert to the actual DTO type
	if pkg, ok := creditPackage.(*creditPackageDTO.CreditPackageResponse); ok {
		packageID = &pkg.ID
		description = fmt.Sprintf("Credit package purchase: %s", pkg.Name)
	} else {
		// Fallback: log the type and continue without package ID
		log.Printf("[PAYMENT-TRACKING] Warning: Could not extract credit package ID from type %T", creditPackage)
	}

	_, trackingErr := s.PaymentTracking.TrackPayment(ctx, tracking.TrackPaymentParams{
		CustomerID:              customerID,
		CustomerEmail:           customerEmail,
		CustomerName:            customerName,
		TransactionType:         "credit_package",
		TransactionDate:         transactionDate,
		OriginalAmount:          originalAmount,
		DiscountAmount:          discountAmount,
		SubsidyAmount:           subsidyAmount,
		CustomerPaid:            customerPaid,
		CreditPackageID:         packageID,
		StripeCustomerID:        session.Customer.ID,
		StripeCheckoutSessionID: session.ID,
		PaymentStatus:           "completed",
		Currency:                string(session.Currency),
		Description:             description,
		ReceiptURL:              receiptURL,
	})

	if trackingErr != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to track credit package purchase: %v", trackingErr)
	}
}

// trackProgramEnrollment tracks a program enrollment payment
func (s *WebhookService) trackProgramEnrollment(session *stripe.CheckoutSession, customerID, programID uuid.UUID, transactionDate time.Time, receiptURL string) {
	ctx := context.Background()

	// Get customer details
	user, err := s.UserRepo.GetUserInfo(ctx, "", customerID)
	if err != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to get user info for %s: %v", customerID, err)
		return
	}

	customerEmail := ""
	if user.Email != nil {
		customerEmail = *user.Email
	}
	customerName := user.FirstName + " " + user.LastName

	// Calculate amounts from session
	originalAmount := float64(session.AmountTotal) / 100.0
	customerPaid := originalAmount
	subsidyAmount := 0.0
	discountAmount := 0.0

	// Check for discounts
	if session.TotalDetails != nil && session.TotalDetails.AmountDiscount > 0 {
		discountAmount = float64(session.TotalDetails.AmountDiscount) / 100.0
	}

	_, trackingErr := s.PaymentTracking.TrackPayment(ctx, tracking.TrackPaymentParams{
		CustomerID:              customerID,
		CustomerEmail:           customerEmail,
		CustomerName:            customerName,
		TransactionType:         "program_enrollment",
		TransactionDate:         transactionDate,
		OriginalAmount:          originalAmount,
		DiscountAmount:          discountAmount,
		SubsidyAmount:           subsidyAmount,
		CustomerPaid:            customerPaid,
		ProgramID:               &programID,
		StripeCustomerID:        session.Customer.ID,
		StripeCheckoutSessionID: session.ID,
		PaymentStatus:           "completed",
		Currency:                string(session.Currency),
		Description:             "Program enrollment payment",
		ReceiptURL:              receiptURL,
	})

	if trackingErr != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to track program enrollment: %v", trackingErr)
	}
}

// trackEventRegistration tracks an event registration payment
func (s *WebhookService) trackEventRegistration(session *stripe.CheckoutSession, customerID, eventID uuid.UUID, transactionDate time.Time, receiptURL string) {
	ctx := context.Background()

	// Get customer details
	user, err := s.UserRepo.GetUserInfo(ctx, "", customerID)
	if err != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to get user info for %s: %v", customerID, err)
		return
	}

	customerEmail := ""
	if user.Email != nil {
		customerEmail = *user.Email
	}
	customerName := user.FirstName + " " + user.LastName

	// Calculate amounts from session
	originalAmount := float64(session.AmountTotal) / 100.0
	customerPaid := originalAmount
	subsidyAmount := 0.0
	discountAmount := 0.0

	// Check for discounts
	if session.TotalDetails != nil && session.TotalDetails.AmountDiscount > 0 {
		discountAmount = float64(session.TotalDetails.AmountDiscount) / 100.0
	}

	_, trackingErr := s.PaymentTracking.TrackPayment(ctx, tracking.TrackPaymentParams{
		CustomerID:              customerID,
		CustomerEmail:           customerEmail,
		CustomerName:            customerName,
		TransactionType:         "event_registration",
		TransactionDate:         transactionDate,
		OriginalAmount:          originalAmount,
		DiscountAmount:          discountAmount,
		SubsidyAmount:           subsidyAmount,
		CustomerPaid:            customerPaid,
		EventID:                 &eventID,
		StripeCustomerID:        session.Customer.ID,
		StripeCheckoutSessionID: session.ID,
		PaymentStatus:           "completed",
		Currency:                string(session.Currency),
		Description:             "Event registration payment",
		ReceiptURL:              receiptURL,
	})

	if trackingErr != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to track event registration: %v", trackingErr)
	}
}

// trackMembershipSubscription tracks a membership subscription payment
func (s *WebhookService) trackMembershipSubscription(session *stripe.CheckoutSession, customerID, planID uuid.UUID, transactionDate time.Time) {
	ctx := context.Background()

	// Get customer details
	user, err := s.UserRepo.GetUserInfo(ctx, "", customerID)
	if err != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to get user info for %s: %v", customerID, err)
		return
	}

	customerEmail := ""
	if user.Email != nil {
		customerEmail = *user.Email
	}
	customerName := user.FirstName + " " + user.LastName

	// Get plan details
	plan, planErr := s.PlansRepo.GetMembershipPlanById(ctx, planID)
	if planErr != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to get plan details: %v", planErr)
		return
	}

	// Calculate amounts from session
	originalAmount := float64(session.AmountTotal) / 100.0
	customerPaid := originalAmount
	subsidyAmount := 0.0
	discountAmount := 0.0

	// Check for subsidy
	var subsidyID *uuid.UUID
	if hasSubsidy, exists := session.Metadata["has_subsidy"]; exists && hasSubsidy == "true" {
		if subsidyAmountStr, exists := session.Metadata["subsidy_amount"]; exists {
			fmt.Sscanf(subsidyAmountStr, "%f", &subsidyAmount)
		}
		if subsidyIDStr, exists := session.Metadata["subsidy_id"]; exists {
			if parsedID, err := uuid.Parse(subsidyIDStr); err == nil {
				subsidyID = &parsedID
			} else {
				log.Printf("[PAYMENT-TRACKING] Failed to parse subsidy_id from metadata: %v", err)
			}
		}
	}

	// Check for discounts
	if session.TotalDetails != nil && session.TotalDetails.AmountDiscount > 0 {
		totalDiscount := float64(session.TotalDetails.AmountDiscount) / 100.0
		// Separate subsidy from regular discounts
		if subsidyAmount > 0 {
			discountAmount = totalDiscount - subsidyAmount
		} else {
			discountAmount = totalDiscount
		}
	}

	var subscriptionID string
	if session.Subscription != nil {
		subscriptionID = session.Subscription.ID
	}

	_, trackingErr := s.PaymentTracking.TrackPayment(ctx, tracking.TrackPaymentParams{
		CustomerID:              customerID,
		CustomerEmail:           customerEmail,
		CustomerName:            customerName,
		TransactionType:         "membership_subscription",
		TransactionDate:         transactionDate,
		OriginalAmount:          originalAmount + subsidyAmount + discountAmount,
		DiscountAmount:          discountAmount,
		SubsidyAmount:           subsidyAmount,
		CustomerPaid:            customerPaid,
		MembershipPlanID:        &planID,
		SubsidyID:               subsidyID,
		StripeCustomerID:        session.Customer.ID,
		StripeSubscriptionID:    subscriptionID,
		StripeCheckoutSessionID: session.ID,
		PaymentStatus:           "completed",
		Currency:                string(session.Currency),
		Description:             fmt.Sprintf("Membership subscription: %s", plan.Name),
	})

	if trackingErr != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to track membership subscription: %v", trackingErr)
	}
}

// trackFailedPayment tracks a failed payment attempt
func (s *WebhookService) trackFailedPayment(invoice *stripe.Invoice, customerID uuid.UUID, transactionDate time.Time) {
	ctx := context.Background()

	// Get customer details
	user, err := s.UserRepo.GetUserInfo(ctx, "", customerID)
	if err != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to get user info for %s: %v", customerID, err)
		return
	}

	customerEmail := ""
	if user.Email != nil {
		customerEmail = *user.Email
	}
	customerName := user.FirstName + " " + user.LastName

	// Get subscription to find membership plan
	var planID *uuid.UUID
	if invoice.Subscription != nil && invoice.Subscription.ID != "" {
		membership, mErr := s.container.Queries.UserDb.GetMembershipByStripeSubscriptionID(ctx, sql.NullString{
			String: invoice.Subscription.ID,
			Valid:  true,
		})
		if mErr == nil {
			planID = &membership.MembershipPlanID
		} else {
			log.Printf("[PAYMENT-TRACKING] Could not find membership for subscription %s: %v", invoice.Subscription.ID, mErr)
		}
	}

	// For failed payments, set all amounts to 0 to satisfy the constraint:
	// customer_paid = original_amount - discount_amount - subsidy_amount
	// Store the attempted amount in metadata for reference
	attemptedAmount := float64(invoice.Total) / 100.0

	var subscriptionID string
	if invoice.Subscription != nil {
		subscriptionID = invoice.Subscription.ID
	}

	_, trackingErr := s.PaymentTracking.TrackPayment(ctx, tracking.TrackPaymentParams{
		CustomerID:           customerID,
		CustomerEmail:        customerEmail,
		CustomerName:         customerName,
		TransactionType:      "membership_renewal",
		TransactionDate:      transactionDate,
		OriginalAmount:       0, // Set to 0 for failed payments (constraint requires customer_paid = original - discount - subsidy)
		DiscountAmount:       0,
		SubsidyAmount:        0,
		CustomerPaid:         0, // Payment failed, nothing was collected
		MembershipPlanID:     planID,
		StripeCustomerID:     invoice.Customer.ID,
		StripeSubscriptionID: subscriptionID,
		StripeInvoiceID:      invoice.ID,
		PaymentStatus:        "failed",
		Currency:             string(invoice.Currency),
		Description:          fmt.Sprintf("Failed payment attempt - $%.2f was attempted", attemptedAmount),
		Metadata: map[string]interface{}{
			"attempted_amount": attemptedAmount,
			"failure_reason":   "Payment failed",
		},
	})

	if trackingErr != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to track failed payment: %v", trackingErr)
	} else {
		log.Printf("[PAYMENT-TRACKING] Tracked failed payment for user %s, invoice %s (attempted: $%.2f)", customerID, invoice.ID, attemptedAmount)
	}
}

// trackMembershipRenewal tracks a membership renewal payment from invoice
func (s *WebhookService) trackMembershipRenewal(invoice *stripe.Invoice, customerID uuid.UUID, transactionDate time.Time) {
	ctx := context.Background()

	// Get customer details
	user, err := s.UserRepo.GetUserInfo(ctx, "", customerID)
	if err != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to get user info for %s: %v", customerID, err)
		return
	}

	customerEmail := ""
	if user.Email != nil {
		customerEmail = *user.Email
	}
	customerName := user.FirstName + " " + user.LastName

	// Get subscription to find membership plan
	var planID *uuid.UUID
	if invoice.Subscription != nil && invoice.Subscription.ID != "" {
		membership, mErr := s.container.Queries.UserDb.GetMembershipByStripeSubscriptionID(ctx, sql.NullString{
			String: invoice.Subscription.ID,
			Valid:  true,
		})
		if mErr == nil {
			planID = &membership.MembershipPlanID
		} else {
			log.Printf("[PAYMENT-TRACKING] Could not find membership for subscription %s: %v", invoice.Subscription.ID, mErr)
		}
	}

	// Calculate amounts from invoice
	originalAmount := float64(invoice.Total) / 100.0
	customerPaid := float64(invoice.AmountPaid) / 100.0
	subsidyAmount := 0.0
	discountAmount := 0.0

	// Check for discounts
	if len(invoice.TotalDiscountAmounts) > 0 {
		for _, discount := range invoice.TotalDiscountAmounts {
			discountAmount += float64(discount.Amount) / 100.0
		}
	}

	var subscriptionID string
	if invoice.Subscription != nil {
		subscriptionID = invoice.Subscription.ID
	}

	_, trackingErr := s.PaymentTracking.TrackPayment(ctx, tracking.TrackPaymentParams{
		CustomerID:           customerID,
		CustomerEmail:        customerEmail,
		CustomerName:         customerName,
		TransactionType:      "membership_renewal",
		TransactionDate:      transactionDate,
		OriginalAmount:       originalAmount,
		DiscountAmount:       discountAmount,
		SubsidyAmount:        subsidyAmount,
		CustomerPaid:         customerPaid,
		MembershipPlanID:     planID,
		StripeCustomerID:     invoice.Customer.ID,
		StripeSubscriptionID: subscriptionID,
		StripeInvoiceID:      invoice.ID,
		PaymentStatus:        "completed",
		Currency:             string(invoice.Currency),
		Description:          "Membership renewal payment",
		InvoiceURL:           invoice.HostedInvoiceURL,
		InvoicePDFURL:        invoice.InvoicePDF,
	})

	if trackingErr != nil {
		log.Printf("[PAYMENT-TRACKING] Failed to track membership renewal: %v", trackingErr)
	}
}
