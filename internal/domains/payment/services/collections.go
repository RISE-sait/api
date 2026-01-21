package payment

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"api/internal/di"
	db "api/internal/domains/payment/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"api/utils/email"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/invoice"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/paymentlink"
	"github.com/stripe/stripe-go/v81/paymentmethod"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/product"
)

type CollectionsService struct {
	queries   *db.Queries
	db        *sql.DB
	container *di.Container
}

func NewCollectionsService(container *di.Container) *CollectionsService {
	return &CollectionsService{
		queries:   db.New(container.DB),
		db:        container.DB,
		container: container,
	}
}

// PaymentMethodInfo represents a saved payment method for a customer
type PaymentMethodInfo struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Brand       string `json:"brand,omitempty"`
	Last4       string `json:"last4,omitempty"`
	ExpMonth    int64  `json:"exp_month,omitempty"`
	ExpYear     int64  `json:"exp_year,omitempty"`
	IsDefault   bool   `json:"is_default"`
	DisplayName string `json:"display_name"`
}

// ChargeCardRequest represents a request to charge a saved card
type ChargeCardRequest struct {
	CustomerID      uuid.UUID `json:"customer_id"`
	PaymentMethodID string    `json:"payment_method_id"`
	Amount          float64   `json:"amount"`
	Notes           string    `json:"notes,omitempty"`
}

// SendPaymentLinkRequest represents a request to send a payment link
type SendPaymentLinkRequest struct {
	CustomerID    uuid.UUID `json:"customer_id"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description,omitempty"`
	SendEmail     bool      `json:"send_email"`
	EmailOverride string    `json:"email_override,omitempty"` // Override customer's email
	Notes         string    `json:"notes,omitempty"`
}

// RecordManualPaymentRequest represents a request to record a manual payment
type RecordManualPaymentRequest struct {
	CustomerID    uuid.UUID `json:"customer_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"` // cash, check, external_card, etc.
	Notes         string    `json:"notes,omitempty"`
}

// CollectionResult represents the result of a collection attempt
type CollectionResult struct {
	Success         bool      `json:"success"`
	CollectionID    uuid.UUID `json:"collection_id"`
	AmountCollected float64   `json:"amount_collected,omitempty"`
	ReceiptURL      string    `json:"receipt_url,omitempty"`
	PaymentLinkURL  string    `json:"payment_link_url,omitempty"`
	ErrorMessage    string    `json:"error_message,omitempty"`
}

// GetCustomerPaymentMethods fetches saved payment methods from Stripe
func (s *CollectionsService) GetCustomerPaymentMethods(ctx context.Context, customerID uuid.UUID) ([]PaymentMethodInfo, *errLib.CommonError) {
	// Get stripe_customer_id from database
	var stripeCustomerID sql.NullString
	query := "SELECT stripe_customer_id FROM users.users WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, customerID).Scan(&stripeCustomerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errLib.New("Customer not found", 404)
		}
		log.Printf("[COLLECTIONS] Error fetching customer: %v", err)
		return nil, errLib.New("Failed to fetch customer", 500)
	}

	if !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		return []PaymentMethodInfo{}, nil // No Stripe customer, return empty list
	}

	// Get default payment method from Stripe customer
	cust, custErr := customer.Get(stripeCustomerID.String, nil)
	if custErr != nil {
		log.Printf("[COLLECTIONS] Error fetching Stripe customer %s: %v", stripeCustomerID.String, custErr)
		return nil, errLib.New("Failed to fetch Stripe customer", 500)
	}

	var defaultPaymentMethodID string
	if cust.InvoiceSettings != nil && cust.InvoiceSettings.DefaultPaymentMethod != nil {
		defaultPaymentMethodID = cust.InvoiceSettings.DefaultPaymentMethod.ID
	}

	// List payment methods for customer
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(stripeCustomerID.String),
		Type:     stripe.String("card"),
	}

	var methods []PaymentMethodInfo
	iter := paymentmethod.List(params)
	for iter.Next() {
		pm := iter.PaymentMethod()

		info := PaymentMethodInfo{
			ID:        pm.ID,
			Type:      string(pm.Type),
			IsDefault: pm.ID == defaultPaymentMethodID,
		}

		if pm.Card != nil {
			info.Brand = string(pm.Card.Brand)
			info.Last4 = pm.Card.Last4
			info.ExpMonth = pm.Card.ExpMonth
			info.ExpYear = pm.Card.ExpYear
			info.DisplayName = fmt.Sprintf("%s ending in %s", capitalizeFirst(string(pm.Card.Brand)), pm.Card.Last4)
		}

		methods = append(methods, info)
	}

	if err := iter.Err(); err != nil {
		log.Printf("[COLLECTIONS] Error listing payment methods: %v", err)
		return nil, errLib.New("Failed to list payment methods", 500)
	}

	return methods, nil
}

// ChargeCard charges a saved payment method
func (s *CollectionsService) ChargeCard(ctx context.Context, adminID uuid.UUID, req ChargeCardRequest) (*CollectionResult, *errLib.CommonError) {
	// Get customer details
	var stripeCustomerID sql.NullString
	var customerEmail sql.NullString
	var firstName, lastName string
	query := "SELECT stripe_customer_id, email, first_name, last_name FROM users.users WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, req.CustomerID).Scan(&stripeCustomerID, &customerEmail, &firstName, &lastName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errLib.New("Customer not found", 404)
		}
		return nil, errLib.New("Failed to fetch customer", 500)
	}

	if !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		return nil, errLib.New("Customer has no saved payment methods", 400)
	}

	// Get previous balance (amount of failed/open invoices)
	previousBalance := s.getCustomerPastDueAmount(ctx, stripeCustomerID.String)

	// Get payment method details for display
	pm, pmErr := paymentmethod.Get(req.PaymentMethodID, nil)
	var paymentMethodDetails string
	if pmErr == nil && pm.Card != nil {
		paymentMethodDetails = fmt.Sprintf("%s ending in %s", capitalizeFirst(string(pm.Card.Brand)), pm.Card.Last4)
	}

	// Create collection attempt record
	attempt, createErr := s.queries.CreateCollectionAttempt(ctx, db.CreateCollectionAttemptParams{
		CustomerID:           req.CustomerID,
		AdminID:              adminID,
		AmountAttempted:      decimal.NewFromFloat(req.Amount),
		AmountCollected:      sql.NullString{String: "0", Valid: true},
		CollectionMethod:     "card_charge",
		PaymentMethodDetails: sql.NullString{String: paymentMethodDetails, Valid: paymentMethodDetails != ""},
		Status:               "pending",
		StripeCustomerID:     stripeCustomerID,
		Notes:                sql.NullString{String: req.Notes, Valid: req.Notes != ""},
		PreviousBalance:      sql.NullString{String: fmt.Sprintf("%.2f", previousBalance), Valid: true},
	})
	if createErr != nil {
		log.Printf("[COLLECTIONS] Error creating collection attempt: %v", createErr)
		return nil, errLib.New("Failed to create collection record", 500)
	}

	// Create PaymentIntent
	amountInCents := int64(req.Amount * 100)
	piParams := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amountInCents),
		Currency:      stripe.String("cad"),
		Customer:      stripe.String(stripeCustomerID.String),
		PaymentMethod: stripe.String(req.PaymentMethodID),
		Confirm:       stripe.Bool(true),
		OffSession:    stripe.Bool(true),
		Description:   stripe.String(fmt.Sprintf("Payment collection for %s %s", firstName, lastName)),
		Metadata: map[string]string{
			"collection_attempt_id": attempt.ID.String(),
			"admin_id":              adminID.String(),
			"customer_id":           req.CustomerID.String(),
		},
	}

	pi, piErr := paymentintent.New(piParams)
	if piErr != nil {
		// Update collection attempt as failed
		s.queries.UpdateCollectionAttemptStatus(ctx, db.UpdateCollectionAttemptStatusParams{
			ID:            attempt.ID,
			Status:        "failed",
			FailureReason: sql.NullString{String: piErr.Error(), Valid: true},
		})

		log.Printf("[COLLECTIONS] Card charge failed for customer %s: %v", req.CustomerID, piErr)
		return &CollectionResult{
			Success:      false,
			CollectionID: attempt.ID,
			ErrorMessage: piErr.Error(),
		}, nil
	}

	// Check payment status
	if pi.Status == stripe.PaymentIntentStatusSucceeded {
		// Calculate new balance
		newBalance := previousBalance - req.Amount
		if newBalance < 0 {
			newBalance = 0
		}

		// Update collection attempt as successful
		s.queries.UpdateCollectionAttemptStatus(ctx, db.UpdateCollectionAttemptStatusParams{
			ID:              attempt.ID,
			Status:          "success",
			AmountCollected: sql.NullString{String: fmt.Sprintf("%.2f", req.Amount), Valid: true},
			NewBalance:      sql.NullString{String: fmt.Sprintf("%.2f", newBalance), Valid: true},
		})

		// Get receipt URL
		var receiptURL string
		if pi.LatestCharge != nil && pi.LatestCharge.ReceiptURL != "" {
			receiptURL = pi.LatestCharge.ReceiptURL
		}

		log.Printf("[COLLECTIONS] Successfully charged $%.2f for customer %s", req.Amount, req.CustomerID)

		return &CollectionResult{
			Success:         true,
			CollectionID:    attempt.ID,
			AmountCollected: req.Amount,
			ReceiptURL:      receiptURL,
		}, nil
	}

	// Payment requires action or failed
	s.queries.UpdateCollectionAttemptStatus(ctx, db.UpdateCollectionAttemptStatusParams{
		ID:            attempt.ID,
		Status:        "failed",
		FailureReason: sql.NullString{String: fmt.Sprintf("Payment status: %s", pi.Status), Valid: true},
	})

	return &CollectionResult{
		Success:      false,
		CollectionID: attempt.ID,
		ErrorMessage: fmt.Sprintf("Payment requires additional action: %s", pi.Status),
	}, nil
}

// SendPaymentLink creates and sends a payment link to the customer
func (s *CollectionsService) SendPaymentLink(ctx context.Context, adminID uuid.UUID, req SendPaymentLinkRequest) (*CollectionResult, *errLib.CommonError) {
	// Get customer details
	var stripeCustomerID sql.NullString
	var customerEmail sql.NullString
	var firstName, lastName string
	query := "SELECT stripe_customer_id, email, first_name, last_name FROM users.users WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, req.CustomerID).Scan(&stripeCustomerID, &customerEmail, &firstName, &lastName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errLib.New("Customer not found", 404)
		}
		return nil, errLib.New("Failed to fetch customer", 500)
	}

	if req.SendEmail && (!customerEmail.Valid || customerEmail.String == "") {
		return nil, errLib.New("Customer has no email address", 400)
	}

	// Get previous balance
	previousBalance := 0.0
	if stripeCustomerID.Valid && stripeCustomerID.String != "" {
		previousBalance = s.getCustomerPastDueAmount(ctx, stripeCustomerID.String)
	}

	// Create collection attempt record
	attempt, createErr := s.queries.CreateCollectionAttempt(ctx, db.CreateCollectionAttemptParams{
		CustomerID:       req.CustomerID,
		AdminID:          adminID,
		AmountAttempted:  decimal.NewFromFloat(req.Amount),
		AmountCollected:  sql.NullString{String: "0", Valid: true},
		CollectionMethod: "payment_link",
		Status:           "pending",
		StripeCustomerID: stripeCustomerID,
		Notes:            sql.NullString{String: req.Notes, Valid: req.Notes != ""},
		PreviousBalance:  sql.NullString{String: fmt.Sprintf("%.2f", previousBalance), Valid: true},
	})
	if createErr != nil {
		log.Printf("[COLLECTIONS] Error creating collection attempt: %v", createErr)
		return nil, errLib.New("Failed to create collection record", 500)
	}

	// Create a one-time price
	description := req.Description
	if description == "" {
		description = fmt.Sprintf("Payment collection for %s %s", firstName, lastName)
	}

	// Create a product for this collection
	prod, prodErr := product.New(&stripe.ProductParams{
		Name: stripe.String(description),
	})
	if prodErr != nil {
		log.Printf("[COLLECTIONS] Error creating Stripe product: %v", prodErr)
		s.queries.UpdateCollectionAttemptStatus(ctx, db.UpdateCollectionAttemptStatusParams{
			ID:            attempt.ID,
			Status:        "failed",
			FailureReason: sql.NullString{String: prodErr.Error(), Valid: true},
		})
		return nil, errLib.New("Failed to create payment link", 500)
	}

	// Create a price
	amountInCents := int64(req.Amount * 100)
	priceObj, priceErr := price.New(&stripe.PriceParams{
		Product:    stripe.String(prod.ID),
		UnitAmount: stripe.Int64(amountInCents),
		Currency:   stripe.String("cad"),
	})
	if priceErr != nil {
		log.Printf("[COLLECTIONS] Error creating Stripe price: %v", priceErr)
		s.queries.UpdateCollectionAttemptStatus(ctx, db.UpdateCollectionAttemptStatusParams{
			ID:            attempt.ID,
			Status:        "failed",
			FailureReason: sql.NullString{String: priceErr.Error(), Valid: true},
		})
		return nil, errLib.New("Failed to create payment link", 500)
	}

	// Create the payment link
	plParams := &stripe.PaymentLinkParams{
		LineItems: []*stripe.PaymentLinkLineItemParams{
			{
				Price:    stripe.String(priceObj.ID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"collection_attempt_id": attempt.ID.String(),
			"admin_id":              adminID.String(),
			"customer_id":           req.CustomerID.String(),
		},
	}

	pl, plErr := paymentlink.New(plParams)
	if plErr != nil {
		log.Printf("[COLLECTIONS] Error creating payment link: %v", plErr)
		s.queries.UpdateCollectionAttemptStatus(ctx, db.UpdateCollectionAttemptStatusParams{
			ID:            attempt.ID,
			Status:        "failed",
			FailureReason: sql.NullString{String: plErr.Error(), Valid: true},
		})
		return nil, errLib.New("Failed to create payment link", 500)
	}

	// Update collection attempt with payment link ID
	s.db.ExecContext(ctx,
		"UPDATE payments.collection_attempts SET stripe_payment_link_id = $1 WHERE id = $2",
		pl.ID, attempt.ID)

	// Create payment link record
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	sentVia := []string{}
	if req.SendEmail {
		sentVia = append(sentVia, "email")
	}

	_, linkErr := s.queries.CreatePaymentLink(ctx, db.CreatePaymentLinkParams{
		CustomerID:           req.CustomerID,
		AdminID:              adminID,
		StripePaymentLinkID:  pl.ID,
		StripePaymentLinkUrl: pl.URL,
		Amount:               decimal.NewFromFloat(req.Amount),
		Description:          sql.NullString{String: description, Valid: true},
		CollectionAttemptID:  uuid.NullUUID{UUID: attempt.ID, Valid: true},
		Status:               "pending",
		SentVia:              sentVia,
		SentToEmail:          customerEmail,
		ExpiresAt:            sql.NullTime{Time: expiresAt, Valid: true},
	})
	if linkErr != nil {
		log.Printf("[COLLECTIONS] Error saving payment link record: %v", linkErr)
		// Don't fail - the payment link was created successfully
	}

	// Send email if requested
	if req.SendEmail {
		targetEmail := customerEmail.String
		if req.EmailOverride != "" {
			targetEmail = req.EmailOverride
		}
		if targetEmail != "" {
			go s.sendPaymentRequestEmail(targetEmail, firstName, req.Amount, pl.URL)
		}
	}

	log.Printf("[COLLECTIONS] Created payment link for customer %s: %s", req.CustomerID, pl.URL)

	return &CollectionResult{
		Success:        true,
		CollectionID:   attempt.ID,
		PaymentLinkURL: pl.URL,
	}, nil
}

// RecordManualPayment records a manual payment (cash, check, etc.)
func (s *CollectionsService) RecordManualPayment(ctx context.Context, adminID uuid.UUID, req RecordManualPaymentRequest) (*CollectionResult, *errLib.CommonError) {
	// Get customer details
	var stripeCustomerID sql.NullString
	query := "SELECT stripe_customer_id FROM users.users WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, req.CustomerID).Scan(&stripeCustomerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errLib.New("Customer not found", 404)
		}
		return nil, errLib.New("Failed to fetch customer", 500)
	}

	// Get previous balance
	previousBalance := 0.0
	if stripeCustomerID.Valid && stripeCustomerID.String != "" {
		previousBalance = s.getCustomerPastDueAmount(ctx, stripeCustomerID.String)
	}

	// Calculate new balance
	newBalance := previousBalance - req.Amount
	if newBalance < 0 {
		newBalance = 0
	}

	// Create collection attempt record (already successful since it's manual)
	attempt, createErr := s.queries.CreateCollectionAttempt(ctx, db.CreateCollectionAttemptParams{
		CustomerID:           req.CustomerID,
		AdminID:              adminID,
		AmountAttempted:      decimal.NewFromFloat(req.Amount),
		AmountCollected:      sql.NullString{String: fmt.Sprintf("%.2f", req.Amount), Valid: true},
		CollectionMethod:     "manual_entry",
		PaymentMethodDetails: sql.NullString{String: req.PaymentMethod, Valid: req.PaymentMethod != ""},
		Status:               "success",
		StripeCustomerID:     stripeCustomerID,
		Notes:                sql.NullString{String: req.Notes, Valid: req.Notes != ""},
		PreviousBalance:      sql.NullString{String: fmt.Sprintf("%.2f", previousBalance), Valid: true},
		NewBalance:           sql.NullString{String: fmt.Sprintf("%.2f", newBalance), Valid: true},
		CompletedAt:          sql.NullTime{Time: time.Now(), Valid: true},
	})
	if createErr != nil {
		log.Printf("[COLLECTIONS] Error creating manual payment record: %v", createErr)
		return nil, errLib.New("Failed to record payment", 500)
	}

	log.Printf("[COLLECTIONS] Recorded manual payment of $%.2f for customer %s (method: %s)",
		req.Amount, req.CustomerID, req.PaymentMethod)

	return &CollectionResult{
		Success:         true,
		CollectionID:    attempt.ID,
		AmountCollected: req.Amount,
	}, nil
}

// CustomerBalance represents a customer's outstanding balance
type CustomerBalance struct {
	CustomerID      uuid.UUID     `json:"customer_id"`
	PastDueAmount   float64       `json:"past_due_amount"`
	OpenInvoices    []OpenInvoice `json:"open_invoices"`
	HasPaymentMethod bool         `json:"has_payment_method"`
}

// OpenInvoice represents an unpaid invoice
type OpenInvoice struct {
	InvoiceID   string  `json:"invoice_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	DueDate     string  `json:"due_date,omitempty"`
	Status      string  `json:"status"`
}

// GetCustomerBalance returns the customer's past due balance and open invoices
func (s *CollectionsService) GetCustomerBalance(ctx context.Context, customerID uuid.UUID) (*CustomerBalance, *errLib.CommonError) {
	// Get customer's Stripe ID
	var stripeCustomerID sql.NullString
	query := "SELECT stripe_customer_id FROM users.users WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, customerID).Scan(&stripeCustomerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errLib.New("Customer not found", 404)
		}
		return nil, errLib.New("Failed to fetch customer", 500)
	}

	balance := &CustomerBalance{
		CustomerID:    customerID,
		OpenInvoices:  []OpenInvoice{},
	}

	if !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		return balance, nil // No Stripe customer, no balance
	}

	// Check if customer has payment methods
	pmParams := &stripe.PaymentMethodListParams{
		Customer: stripe.String(stripeCustomerID.String),
		Type:     stripe.String("card"),
	}
	pmParams.Limit = stripe.Int64(1)
	pmIter := paymentmethod.List(pmParams)
	balance.HasPaymentMethod = pmIter.Next()

	// Get open invoices from Stripe
	invoiceParams := &stripe.InvoiceListParams{
		Customer: stripe.String(stripeCustomerID.String),
	}
	invoiceParams.Limit = stripe.Int64(100)

	var totalDue float64
	i := invoice.List(invoiceParams)
	for i.Next() {
		inv := i.Invoice()
		// Include open and past_due invoices
		if inv.Status == stripe.InvoiceStatusOpen || inv.Status == stripe.InvoiceStatusUncollectible {
			amount := float64(inv.AmountDue) / 100.0
			totalDue += amount

			description := "Invoice"
			if len(inv.Lines.Data) > 0 && inv.Lines.Data[0].Description != "" {
				description = inv.Lines.Data[0].Description
			}

			var dueDate string
			if inv.DueDate > 0 {
				dueDate = time.Unix(inv.DueDate, 0).Format("2006-01-02")
			}

			balance.OpenInvoices = append(balance.OpenInvoices, OpenInvoice{
				InvoiceID:   inv.ID,
				Amount:      amount,
				Description: description,
				DueDate:     dueDate,
				Status:      string(inv.Status),
			})
		}
	}

	balance.PastDueAmount = totalDue
	return balance, nil
}

// GetCollectionAttempts returns collection attempts with optional filters
func (s *CollectionsService) GetCollectionAttempts(ctx context.Context, customerID, adminID *uuid.UUID, status, method string, startDate, endDate *time.Time, limit, offset int32) ([]db.PaymentsCollectionAttempt, int64, error) {
	attempts, err := s.queries.ListCollectionAttempts(ctx, db.ListCollectionAttemptsParams{
		CustomerID:       uuid.NullUUID{UUID: uuidOrNil(customerID), Valid: customerID != nil},
		AdminID:          uuid.NullUUID{UUID: uuidOrNil(adminID), Valid: adminID != nil},
		Status:           sql.NullString{String: status, Valid: status != ""},
		CollectionMethod: sql.NullString{String: method, Valid: method != ""},
		StartDate:        sql.NullTime{Time: timeOrZero(startDate), Valid: startDate != nil},
		EndDate:          sql.NullTime{Time: timeOrZero(endDate), Valid: endDate != nil},
		Limit:            limit,
		Offset:           offset,
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := s.queries.CountCollectionAttempts(ctx, db.CountCollectionAttemptsParams{
		CustomerID:       uuid.NullUUID{UUID: uuidOrNil(customerID), Valid: customerID != nil},
		AdminID:          uuid.NullUUID{UUID: uuidOrNil(adminID), Valid: adminID != nil},
		Status:           sql.NullString{String: status, Valid: status != ""},
		CollectionMethod: sql.NullString{String: method, Valid: method != ""},
		StartDate:        sql.NullTime{Time: timeOrZero(startDate), Valid: startDate != nil},
		EndDate:          sql.NullTime{Time: timeOrZero(endDate), Valid: endDate != nil},
	})
	if err != nil {
		return attempts, 0, err
	}

	return attempts, count, nil
}

// Helper functions

func (s *CollectionsService) getCustomerPastDueAmount(ctx context.Context, stripeCustomerID string) float64 {
	// Query Stripe for open invoices
	invoiceParams := &stripe.InvoiceListParams{
		Customer: stripe.String(stripeCustomerID),
		Status:   stripe.String("open"),
	}
	invoiceParams.Limit = stripe.Int64(100)

	var totalDue float64
	i := invoice.List(invoiceParams)
	for i.Next() {
		inv := i.Invoice()
		totalDue += float64(inv.AmountDue) / 100.0
	}

	return totalDue
}

func (s *CollectionsService) sendPaymentRequestEmail(toEmail, firstName string, amount float64, paymentURL string) {
	subject := "Payment Request from Rise"
	body := email.PaymentRequestBody(firstName, amount, paymentURL)

	if err := email.SendEmail(toEmail, subject, body); err != nil {
		log.Printf("[COLLECTIONS] Failed to send payment request email to %s: %v", toEmail, err)
	} else {
		log.Printf("[COLLECTIONS] Sent payment request email to %s", toEmail)
	}
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}

func uuidOrNil(u *uuid.UUID) uuid.UUID {
	if u == nil {
		return uuid.Nil
	}
	return *u
}

func timeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
