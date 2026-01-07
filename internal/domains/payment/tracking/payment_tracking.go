package tracking

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"api/internal/di"
	db "api/internal/domains/payment/persistence/sqlc/generated"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sqlc-dev/pqtype"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/invoice"
	"github.com/stripe/stripe-go/v81/paymentintent"
)

type PaymentTrackingService struct {
	queries *db.Queries
	db      *sql.DB
}

func NewPaymentTrackingService(container *di.Container) *PaymentTrackingService {
	return &PaymentTrackingService{
		queries: db.New(container.DB),
		db:      container.DB,
	}
}

// TrackPaymentParams contains all the details needed to track a payment
type TrackPaymentParams struct {
	CustomerID    uuid.UUID
	CustomerEmail string
	CustomerName  string

	TransactionType string // 'membership_subscription', 'membership_renewal', 'program_enrollment', 'event_registration', 'joining_fee', 'credit_package'
	TransactionDate time.Time

	OriginalAmount float64
	DiscountAmount float64
	SubsidyAmount  float64
	CustomerPaid   float64

	MembershipPlanID *uuid.UUID
	ProgramID        *uuid.UUID
	EventID          *uuid.UUID
	CreditPackageID  *uuid.UUID
	SubsidyID        *uuid.UUID
	DiscountCodeID   *uuid.UUID

	StripeCustomerID        string
	StripeSubscriptionID    string
	StripeInvoiceID         string
	StripePaymentIntentID   string
	StripeCheckoutSessionID string

	PaymentStatus string // 'pending', 'completed', 'failed', 'refunded', 'partially_refunded'
	PaymentMethod string
	Currency      string

	Description string
	Metadata    map[string]interface{}

	// Stripe URLs for receipts and invoices
	ReceiptURL    string // For one-time payments (events, programs, credit packages)
	InvoiceURL    string // For subscription payments
	InvoicePDFURL string // PDF download URL for subscription invoices
}

// TrackPayment creates a new payment transaction record
func (s *PaymentTrackingService) TrackPayment(ctx context.Context, params TrackPaymentParams) (*db.PaymentsPaymentTransaction, error) {
	// Validate payment calculation
	expectedCustomerPaid := params.OriginalAmount - params.DiscountAmount - params.SubsidyAmount
	if params.CustomerPaid != expectedCustomerPaid {
		return nil, fmt.Errorf("invalid payment calculation: customer_paid (%.2f) != original (%.2f) - discount (%.2f) - subsidy (%.2f)",
			params.CustomerPaid, params.OriginalAmount, params.DiscountAmount, params.SubsidyAmount)
	}

	// Convert metadata to JSONB
	var metadataJSON pqtype.NullRawMessage
	if params.Metadata != nil {
		jsonBytes, err := json.Marshal(params.Metadata)
		if err != nil {
			log.Printf("Failed to marshal payment metadata: %v", err)
		} else {
			metadataJSON = pqtype.NullRawMessage{
				RawMessage: jsonBytes,
				Valid:      true,
			}
		}
	}

	// Set default values
	if params.TransactionDate.IsZero() {
		params.TransactionDate = time.Now()
	}
	if params.PaymentStatus == "" {
		params.PaymentStatus = "pending"
	}
	if params.Currency == "" {
		params.Currency = "USD"
	}

	transaction, err := s.queries.CreatePaymentTransaction(ctx, db.CreatePaymentTransactionParams{
		CustomerID:              params.CustomerID,
		CustomerEmail:           params.CustomerEmail,
		CustomerName:            params.CustomerName,
		TransactionType:         params.TransactionType,
		TransactionDate:         params.TransactionDate,
		OriginalAmount:          decimal.NewFromFloat(params.OriginalAmount),
		DiscountAmount:          decimal.NewFromFloat(params.DiscountAmount),
		SubsidyAmount:           decimal.NewFromFloat(params.SubsidyAmount),
		CustomerPaid:            decimal.NewFromFloat(params.CustomerPaid),
		MembershipPlanID:        uuidToNullUUID(params.MembershipPlanID),
		ProgramID:               uuidToNullUUID(params.ProgramID),
		EventID:                 uuidToNullUUID(params.EventID),
		CreditPackageID:         uuidToNullUUID(params.CreditPackageID),
		SubsidyID:               uuidToNullUUID(params.SubsidyID),
		DiscountCodeID:          uuidToNullUUID(params.DiscountCodeID),
		StripeCustomerID:        stringToNullString(params.StripeCustomerID),
		StripeSubscriptionID:    stringToNullString(params.StripeSubscriptionID),
		StripeInvoiceID:         stringToNullString(params.StripeInvoiceID),
		StripePaymentIntentID:   stringToNullString(params.StripePaymentIntentID),
		StripeCheckoutSessionID: stringToNullString(params.StripeCheckoutSessionID),
		PaymentStatus:           params.PaymentStatus,
		PaymentMethod:           stringToNullString(params.PaymentMethod),
		Currency:                stringToNullString(params.Currency),
		Description:             stringToNullString(params.Description),
		Metadata:                metadataJSON,
		ReceiptUrl:              stringToNullString(params.ReceiptURL),
		InvoiceUrl:              stringToNullString(params.InvoiceURL),
		InvoicePdfUrl:           stringToNullString(params.InvoicePDFURL),
	})

	if err != nil {
		log.Printf("Failed to track payment: %v", err)
		return nil, err
	}

	log.Printf("✅ [PAYMENT-TRACKED] Type=%s, Customer=%s, Amount=$%.2f (Subsidy: $%.2f, Discount: $%.2f, Paid: $%.2f), Invoice=%s",
		params.TransactionType, params.CustomerEmail, params.OriginalAmount,
		params.SubsidyAmount, params.DiscountAmount, params.CustomerPaid, params.StripeInvoiceID)

	// Auto-backfill URLs (same logic as POST /admin/payments/backfill-urls)
	go s.BackfillTransactionURLs(transaction.ID)

	return &transaction, nil
}

// UpdatePaymentStatus updates the payment status
func (s *PaymentTrackingService) UpdatePaymentStatus(ctx context.Context, transactionID uuid.UUID, status string) error {
	_, err := s.queries.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		ID:            transactionID,
		PaymentStatus: status,
	})
	if err != nil {
		log.Printf("Failed to update payment status: %v", err)
		return err
	}

	log.Printf("✅ [PAYMENT-STATUS-UPDATED] Transaction=%s, Status=%s", transactionID, status)
	return nil
}

// RecordRefund records a payment refund
func (s *PaymentTrackingService) RecordRefund(ctx context.Context, transactionID uuid.UUID, refundedAmount float64, reason string) error {
	transaction, err := s.queries.GetPaymentTransaction(ctx, transactionID)
	if err != nil {
		return err
	}

	status := "refunded"
	currentPaid, _ := transaction.CustomerPaid.Float64()
	if refundedAmount < currentPaid {
		status = "partially_refunded"
	}

	_, err = s.queries.RecordRefund(ctx, db.RecordRefundParams{
		ID:             transactionID,
		PaymentStatus:  status,
		RefundedAmount: decimal.NewFromFloat(refundedAmount),
		RefundReason:   stringToNullString(reason),
	})

	if err != nil {
		log.Printf("Failed to record refund: %v", err)
		return err
	}

	log.Printf("✅ [REFUND-RECORDED] Transaction=%s, Amount=$%.2f, Reason=%s", transactionID, refundedAmount, reason)
	return nil
}

// GetPaymentByStripeInvoice retrieves a payment transaction by Stripe invoice ID
func (s *PaymentTrackingService) GetPaymentByStripeInvoice(ctx context.Context, invoiceID string) (*db.PaymentsPaymentTransaction, error) {
	transaction, err := s.queries.GetPaymentTransactionByStripeInvoice(ctx, sql.NullString{
		String: invoiceID,
		Valid:  true,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

// BackfillTransactionURLs fetches receipt/invoice URLs from Stripe for a single transaction.
// Uses the EXACT same logic as POST /admin/payments/backfill-urls endpoint.
func (s *PaymentTrackingService) BackfillTransactionURLs(transactionID uuid.UUID) {
	// Small delay to let Stripe finalize the charge/invoice
	time.Sleep(2 * time.Second)

	ctx := context.Background()

	tx, err := s.queries.GetPaymentTransaction(ctx, transactionID)
	if err != nil {
		log.Printf("[PAYMENT-BACKFILL] Error fetching transaction %s: %v", transactionID, err)
		return
	}

	var receiptURL, invoiceURL, invoicePDFURL sql.NullString

	// Try to get receipt URL from CheckoutSession first (most transactions have this)
	if tx.StripeCheckoutSessionID.Valid && tx.StripeCheckoutSessionID.String != "" {
		sess, sessErr := session.Get(tx.StripeCheckoutSessionID.String, &stripe.CheckoutSessionParams{
			Expand: []*string{
				stripe.String("payment_intent.latest_charge"),
				stripe.String("subscription.latest_invoice"),
			},
		})
		if sessErr != nil {
			// Handle deleted/expired checkout sessions gracefully (404 errors)
			if strings.Contains(sessErr.Error(), "resource_missing") || strings.Contains(sessErr.Error(), "No such checkout.session") {
				log.Printf("[PAYMENT-BACKFILL] Checkout session %s no longer exists (deleted/expired), skipping", tx.StripeCheckoutSessionID.String)
			} else {
				log.Printf("[PAYMENT-BACKFILL] Error fetching CheckoutSession %s: %v", tx.StripeCheckoutSessionID.String, sessErr)
			}
			return
		}
		// For one-time payments, get receipt from payment_intent
		if sess.PaymentIntent != nil && sess.PaymentIntent.LatestCharge != nil && sess.PaymentIntent.LatestCharge.ReceiptURL != "" {
			receiptURL = sql.NullString{String: sess.PaymentIntent.LatestCharge.ReceiptURL, Valid: true}
			log.Printf("[PAYMENT-BACKFILL] Found receipt URL for transaction %s via checkout session", tx.ID)
		}

		// For subscription payments, get invoice from the subscription's latest invoice
		if sess.Subscription != nil && sess.Subscription.LatestInvoice != nil {
			if sess.Subscription.LatestInvoice.HostedInvoiceURL != "" {
				invoiceURL = sql.NullString{String: sess.Subscription.LatestInvoice.HostedInvoiceURL, Valid: true}
			}
			if sess.Subscription.LatestInvoice.InvoicePDF != "" {
				invoicePDFURL = sql.NullString{String: sess.Subscription.LatestInvoice.InvoicePDF, Valid: true}
			}
			if invoiceURL.Valid {
				log.Printf("[PAYMENT-BACKFILL] Found invoice URLs for transaction %s via subscription", tx.ID)
			}
		}
	}

	// Fallback: Fetch receipt URL directly from PaymentIntent if we have it
	if !receiptURL.Valid && tx.StripePaymentIntentID.Valid && tx.StripePaymentIntentID.String != "" {
		pi, piErr := paymentintent.Get(tx.StripePaymentIntentID.String, &stripe.PaymentIntentParams{
			Expand: []*string{stripe.String("latest_charge")},
		})
		if piErr != nil {
			log.Printf("[PAYMENT-BACKFILL] Error fetching PaymentIntent %s: %v", tx.StripePaymentIntentID.String, piErr)
		} else if pi.LatestCharge != nil && pi.LatestCharge.ReceiptURL != "" {
			receiptURL = sql.NullString{String: pi.LatestCharge.ReceiptURL, Valid: true}
			log.Printf("[PAYMENT-BACKFILL] Found receipt URL for transaction %s via payment intent", tx.ID)
		}
	}

	// Fetch invoice URLs from Invoice (for subscription payments)
	if tx.StripeInvoiceID.Valid && tx.StripeInvoiceID.String != "" {
		inv, invErr := invoice.Get(tx.StripeInvoiceID.String, nil)
		if invErr != nil {
			log.Printf("[PAYMENT-BACKFILL] Error fetching Invoice %s: %v", tx.StripeInvoiceID.String, invErr)
			return
		}
		if inv.HostedInvoiceURL != "" {
			invoiceURL = sql.NullString{String: inv.HostedInvoiceURL, Valid: true}
		}
		if inv.InvoicePDF != "" {
			invoicePDFURL = sql.NullString{String: inv.InvoicePDF, Valid: true}
		}
		log.Printf("[PAYMENT-BACKFILL] Found invoice URLs for transaction %s", tx.ID)
	}

	// Update the transaction with the fetched URLs
	if receiptURL.Valid || invoiceURL.Valid || invoicePDFURL.Valid {
		updateErr := s.queries.UpdatePaymentUrls(ctx, db.UpdatePaymentUrlsParams{
			ID:            tx.ID,
			ReceiptUrl:    receiptURL,
			InvoiceUrl:    invoiceURL,
			InvoicePdfUrl: invoicePDFURL,
		})
		if updateErr != nil {
			log.Printf("[PAYMENT-BACKFILL] Error updating transaction %s: %v", tx.ID, updateErr)
			return
		}
		log.Printf("✅ [PAYMENT-BACKFILL] Transaction %s URLs updated", tx.ID)
	} else {
		log.Printf("[PAYMENT-BACKFILL] No URLs found for transaction %s", tx.ID)
	}
}

// Helper functions
func uuidToNullUUID(u *uuid.UUID) uuid.NullUUID {
	if u == nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: *u, Valid: true}
}

func stringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// ListPaymentTransactions retrieves payment transactions with filters and pagination
func (s *PaymentTrackingService) ListPaymentTransactions(ctx context.Context, filters PaymentFilters) ([]db.PaymentsPaymentTransaction, int64, error) {
	// Get total count
	count, err := s.queries.CountPaymentTransactions(ctx, db.CountPaymentTransactionsParams{
		CustomerID:      uuidToNullUUID(filters.CustomerID),
		TransactionType: stringToNullString(filters.TransactionType),
		PaymentStatus:   stringToNullString(filters.PaymentStatus),
		StartDate:       timeToNullTime(filters.StartDate),
		EndDate:         timeToNullTime(filters.EndDate),
		SubsidyID:       uuidToNullUUID(filters.SubsidyID),
	})
	if err != nil {
		return nil, 0, err
	}

	// Get transactions
	transactions, err := s.queries.ListPaymentTransactions(ctx, db.ListPaymentTransactionsParams{
		CustomerID:      uuidToNullUUID(filters.CustomerID),
		TransactionType: stringToNullString(filters.TransactionType),
		PaymentStatus:   stringToNullString(filters.PaymentStatus),
		StartDate:       timeToNullTime(filters.StartDate),
		EndDate:         timeToNullTime(filters.EndDate),
		SubsidyID:       uuidToNullUUID(filters.SubsidyID),
		Limit:           filters.Limit,
		Offset:          filters.Offset,
	})
	if err != nil {
		return nil, 0, err
	}

	return transactions, count, nil
}

// GetPaymentSummary retrieves aggregated payment statistics
func (s *PaymentTrackingService) GetPaymentSummary(ctx context.Context, filters PaymentSummaryFilters) (*db.GetPaymentSummaryRow, error) {
	result, err := s.queries.GetPaymentSummary(ctx, db.GetPaymentSummaryParams{
		StartDate:       timeToNullTime(filters.StartDate),
		EndDate:         timeToNullTime(filters.EndDate),
		TransactionType: stringToNullString(filters.TransactionType),
		PaymentStatus:   stringToNullString(filters.PaymentStatus),
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPaymentSummaryByType retrieves payment statistics grouped by transaction type
func (s *PaymentTrackingService) GetPaymentSummaryByType(ctx context.Context, startDate, endDate *time.Time) ([]db.GetPaymentSummaryByTypeRow, error) {
	return s.queries.GetPaymentSummaryByType(ctx, db.GetPaymentSummaryByTypeParams{
		StartDate: timeToNullTime(startDate),
		EndDate:   timeToNullTime(endDate),
	})
}

// GetSubsidyUsageSummary retrieves subsidy usage statistics
func (s *PaymentTrackingService) GetSubsidyUsageSummary(ctx context.Context, startDate, endDate *time.Time) (*db.GetSubsidyUsageSummaryRow, error) {
	result, err := s.queries.GetSubsidyUsageSummary(ctx, db.GetSubsidyUsageSummaryParams{
		StartDate: timeToNullTime(startDate),
		EndDate:   timeToNullTime(endDate),
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ExportPaymentTransactions retrieves transactions for export
func (s *PaymentTrackingService) ExportPaymentTransactions(ctx context.Context, filters ExportFilters) ([]db.ExportPaymentTransactionsRow, error) {
	return s.queries.ExportPaymentTransactions(ctx, db.ExportPaymentTransactionsParams{
		StartDate:       timeToNullTime(filters.StartDate),
		EndDate:         timeToNullTime(filters.EndDate),
		TransactionType: stringToNullString(filters.TransactionType),
		PaymentStatus:   stringToNullString(filters.PaymentStatus),
	})
}

// PaymentFilters holds filtering options for listing transactions
type PaymentFilters struct {
	CustomerID      *uuid.UUID
	TransactionType string
	PaymentStatus   string
	StartDate       *time.Time
	EndDate         *time.Time
	SubsidyID       *uuid.UUID
	Limit           int32
	Offset          int32
}

// PaymentSummaryFilters holds filtering options for payment summaries
type PaymentSummaryFilters struct {
	StartDate       *time.Time
	EndDate         *time.Time
	TransactionType string
	PaymentStatus   string
}

// ExportFilters holds filtering options for exports
type ExportFilters struct {
	StartDate       *time.Time
	EndDate         *time.Time
	TransactionType string
	PaymentStatus   string
}

// Helper function to convert time pointer to sql.NullTime
func timeToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
