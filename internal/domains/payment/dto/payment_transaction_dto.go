package payment

import (
	"database/sql"
	"encoding/json"
	"time"

	db "api/internal/domains/payment/persistence/sqlc/generated"

	"github.com/google/uuid"
)

// PaymentTransactionResponse is the API response for payment transactions
// It handles nullable fields properly for JSON serialization
type PaymentTransactionResponse struct {
	ID              uuid.UUID  `json:"id"`
	CustomerID      uuid.UUID  `json:"customer_id"`
	CustomerEmail   string     `json:"customer_email"`
	CustomerName    string     `json:"customer_name"`
	TransactionType string     `json:"transaction_type"`
	TransactionDate time.Time  `json:"transaction_date"`
	OriginalAmount  float64    `json:"original_amount"`
	DiscountAmount  float64    `json:"discount_amount"`
	SubsidyAmount   float64    `json:"subsidy_amount"`
	CustomerPaid    float64    `json:"customer_paid"`

	MembershipPlanID        *uuid.UUID `json:"membership_plan_id,omitempty"`
	ProgramID               *uuid.UUID `json:"program_id,omitempty"`
	EventID                 *uuid.UUID `json:"event_id,omitempty"`
	CreditPackageID         *uuid.UUID `json:"credit_package_id,omitempty"`
	SubsidyID               *uuid.UUID `json:"subsidy_id,omitempty"`
	DiscountCodeID          *uuid.UUID `json:"discount_code_id,omitempty"`

	StripeCustomerID        *string `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID    *string `json:"stripe_subscription_id,omitempty"`
	StripeInvoiceID         *string `json:"stripe_invoice_id,omitempty"`
	StripePaymentIntentID   *string `json:"stripe_payment_intent_id,omitempty"`
	StripeCheckoutSessionID *string `json:"stripe_checkout_session_id,omitempty"`

	// Stripe URLs for receipts and invoices
	ReceiptURL    *string `json:"receipt_url,omitempty"`
	InvoiceURL    *string `json:"invoice_url,omitempty"`
	InvoicePDFURL *string `json:"invoice_pdf_url,omitempty"`

	PaymentStatus  string  `json:"payment_status"`
	PaymentMethod  *string `json:"payment_method,omitempty"`
	Currency       *string `json:"currency,omitempty"`
	Description    *string `json:"description,omitempty"`

	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	RefundedAmount float64                `json:"refunded_amount"`
	RefundReason   *string                `json:"refund_reason,omitempty"`
	RefundedAt     *time.Time             `json:"refunded_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// ToPaymentTransactionResponse converts a database model to a response DTO
func ToPaymentTransactionResponse(tx db.PaymentsPaymentTransaction) PaymentTransactionResponse {
	originalAmt, _ := tx.OriginalAmount.Float64()
	discountAmt, _ := tx.DiscountAmount.Float64()
	subsidyAmt, _ := tx.SubsidyAmount.Float64()
	customerPaid, _ := tx.CustomerPaid.Float64()
	refundedAmt, _ := tx.RefundedAmount.Float64()

	response := PaymentTransactionResponse{
		ID:              tx.ID,
		CustomerID:      tx.CustomerID,
		CustomerEmail:   tx.CustomerEmail,
		CustomerName:    tx.CustomerName,
		TransactionType: tx.TransactionType,
		TransactionDate: tx.TransactionDate,
		OriginalAmount:  originalAmt,
		DiscountAmount:  discountAmt,
		SubsidyAmount:   subsidyAmt,
		CustomerPaid:    customerPaid,
		PaymentStatus:   tx.PaymentStatus,
		RefundedAmount:  refundedAmt,
		CreatedAt:       tx.CreatedAt,
		UpdatedAt:       tx.UpdatedAt,
	}

	// Handle nullable UUIDs
	if tx.MembershipPlanID.Valid {
		response.MembershipPlanID = &tx.MembershipPlanID.UUID
	}
	if tx.ProgramID.Valid {
		response.ProgramID = &tx.ProgramID.UUID
	}
	if tx.EventID.Valid {
		response.EventID = &tx.EventID.UUID
	}
	if tx.CreditPackageID.Valid {
		response.CreditPackageID = &tx.CreditPackageID.UUID
	}
	if tx.SubsidyID.Valid {
		response.SubsidyID = &tx.SubsidyID.UUID
	}
	if tx.DiscountCodeID.Valid {
		response.DiscountCodeID = &tx.DiscountCodeID.UUID
	}

	// Handle nullable strings
	if tx.StripeCustomerID.Valid {
		response.StripeCustomerID = &tx.StripeCustomerID.String
	}
	if tx.StripeSubscriptionID.Valid {
		response.StripeSubscriptionID = &tx.StripeSubscriptionID.String
	}
	if tx.StripeInvoiceID.Valid {
		response.StripeInvoiceID = &tx.StripeInvoiceID.String
	}
	if tx.StripePaymentIntentID.Valid {
		response.StripePaymentIntentID = &tx.StripePaymentIntentID.String
	}
	if tx.StripeCheckoutSessionID.Valid {
		response.StripeCheckoutSessionID = &tx.StripeCheckoutSessionID.String
	}
	if tx.ReceiptUrl.Valid {
		response.ReceiptURL = &tx.ReceiptUrl.String
	}
	if tx.InvoiceUrl.Valid {
		response.InvoiceURL = &tx.InvoiceUrl.String
	}
	if tx.InvoicePdfUrl.Valid {
		response.InvoicePDFURL = &tx.InvoicePdfUrl.String
	}
	if tx.PaymentMethod.Valid {
		response.PaymentMethod = &tx.PaymentMethod.String
	}
	if tx.Currency.Valid {
		response.Currency = &tx.Currency.String
	}
	if tx.Description.Valid {
		response.Description = &tx.Description.String
	}
	if tx.RefundReason.Valid {
		response.RefundReason = &tx.RefundReason.String
	}

	// Handle nullable time - THIS IS THE KEY FIX
	if tx.RefundedAt.Valid {
		response.RefundedAt = &tx.RefundedAt.Time
	}

	// Handle metadata JSONB
	if tx.Metadata.Valid {
		var metadata map[string]interface{}
		if err := json.Unmarshal(tx.Metadata.RawMessage, &metadata); err == nil {
			response.Metadata = metadata
		}
	}

	return response
}

// ToPaymentTransactionResponses converts a slice of database models to response DTOs
func ToPaymentTransactionResponses(txs []db.PaymentsPaymentTransaction) []PaymentTransactionResponse {
	responses := make([]PaymentTransactionResponse, len(txs))
	for i, tx := range txs {
		responses[i] = ToPaymentTransactionResponse(tx)
	}
	return responses
}

// Helper functions for nullable types
func nullUUIDToPtr(nu uuid.NullUUID) *uuid.UUID {
	if nu.Valid {
		return &nu.UUID
	}
	return nil
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
