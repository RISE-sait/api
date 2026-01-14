package payment

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"api/internal/di"
	db "api/internal/domains/payment/persistence/sqlc/generated"
	"api/internal/domains/payment/tracking"
	errLib "api/internal/libs/errors"
	responses "api/internal/libs/responses"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/invoice"
	"github.com/stripe/stripe-go/v81/paymentintent"
)

type PaymentReportsHandler struct {
	trackingService *tracking.PaymentTrackingService
	queries         *db.Queries
	db              *sql.DB
}

func NewPaymentReportsHandler(container *di.Container) *PaymentReportsHandler {
	return &PaymentReportsHandler{
		trackingService: tracking.NewPaymentTrackingService(container),
		queries:         db.New(container.DB),
		db:              container.DB,
	}
}

// ListPaymentTransactions lists payment transactions with filters
// @Summary List payment transactions
// @Description Get paginated list of payment transactions with optional filters (admin only)
// @Tags Payments - Admin
// @Accept json
// @Produce json
// @Param customer_id query string false "Filter by customer ID"
// @Param transaction_type query string false "Filter by transaction type (membership_subscription, program_enrollment, event_registration, credit_package)"
// @Param payment_status query string false "Filter by payment status (pending, completed, failed, refunded, partially_refunded)"
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Param subsidy_id query string false "Filter by subsidy ID"
// @Param limit query int false "Page size" default(50)
// @Param offset query int false "Page offset" default(0)
// @Success 200 {object} map[string]interface{} "Payment transactions retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/payments/transactions [get]
func (h *PaymentReportsHandler) ListPaymentTransactions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	customerIDStr := r.URL.Query().Get("customer_id")
	transactionType := r.URL.Query().Get("transaction_type")
	paymentStatus := r.URL.Query().Get("payment_status")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	subsidyIDStr := r.URL.Query().Get("subsidy_id")

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Parse optional UUID fields
	var customerID, subsidyID *uuid.UUID
	if customerIDStr != "" {
		if parsed, err := uuid.Parse(customerIDStr); err == nil {
			customerID = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid customer_id format", http.StatusBadRequest))
			return
		}
	}

	if subsidyIDStr != "" {
		if parsed, err := uuid.Parse(subsidyIDStr); err == nil {
			subsidyID = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid subsidy_id format", http.StatusBadRequest))
			return
		}
	}

	// Parse optional date fields
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid start_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}

	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid end_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}

	// Fetch transactions from tracking service
	transactions, total, err := h.trackingService.ListPaymentTransactions(r.Context(), tracking.PaymentFilters{
		CustomerID:      customerID,
		TransactionType: transactionType,
		PaymentStatus:   paymentStatus,
		StartDate:       startDate,
		EndDate:         endDate,
		SubsidyID:       subsidyID,
		Limit:           int32(limit),
		Offset:          int32(offset),
	})

	if err != nil {
		log.Printf("[PAYMENT-REPORTS] Error fetching transactions: %v", err)
		responses.RespondWithError(w, errLib.New("Failed to fetch transactions", http.StatusInternalServerError))
		return
	}

	log.Printf("[PAYMENT-REPORTS] Listed %d transactions (total: %d)", len(transactions), total)

	responses.RespondWithSuccess(w, map[string]interface{}{
		"transactions": transactions,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	}, http.StatusOK)
}

// GetPaymentSummary gets payment summary statistics
// @Summary Get payment summary
// @Description Get aggregated payment statistics for a date range (admin only)
// @Tags Payments - Admin
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Param transaction_type query string false "Filter by transaction type"
// @Param payment_status query string false "Filter by payment status"
// @Success 200 {object} map[string]interface{} "Payment transactions retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/payments/summary [get]
func (h *PaymentReportsHandler) GetPaymentSummary(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	transactionType := r.URL.Query().Get("transaction_type")
	paymentStatus := r.URL.Query().Get("payment_status")

	// Parse optional date fields
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid start_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}

	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid end_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}

	// Fetch summary from tracking service
	summary, err := h.trackingService.GetPaymentSummary(r.Context(), tracking.PaymentSummaryFilters{
		StartDate:       startDate,
		EndDate:         endDate,
		TransactionType: transactionType,
		PaymentStatus:   paymentStatus,
	})

	if err != nil {
		log.Printf("[PAYMENT-REPORTS] Error fetching summary: %v", err)
		responses.RespondWithError(w, errLib.New("Failed to fetch payment summary", http.StatusInternalServerError))
		return
	}

	log.Printf("[PAYMENT-REPORTS] Payment summary: %d transactions, $%.2f total", summary.TotalTransactions, summary.TotalCustomerPaid)

	responses.RespondWithSuccess(w, summary, http.StatusOK)
}

// GetPaymentSummaryByType gets payment summary grouped by transaction type
// @Summary Get payment summary by type
// @Description Get payment statistics grouped by transaction type (admin only)
// @Tags Payments - Admin
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Success 200 {object} map[string]interface{} "Payment transactions retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/payments/summary/by-type [get]
func (h *PaymentReportsHandler) GetPaymentSummaryByType(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	// Parse optional date fields
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid start_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}

	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid end_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}

	// Fetch summary by type from tracking service
	summaryByType, err := h.trackingService.GetPaymentSummaryByType(r.Context(), startDate, endDate)
	if err != nil {
		log.Printf("[PAYMENT-REPORTS] Error fetching summary by type: %v", err)
		responses.RespondWithError(w, errLib.New("Failed to fetch payment summary by type", http.StatusInternalServerError))
		return
	}

	log.Printf("[PAYMENT-REPORTS] Summary by type: %d transaction types", len(summaryByType))

	responses.RespondWithSuccess(w, map[string]interface{}{
		"summary_by_type": summaryByType,
	}, http.StatusOK)
}

// GetSubsidyUsageSummary gets subsidy usage statistics
// @Summary Get subsidy usage summary
// @Description Get aggregated subsidy usage statistics (admin only)
// @Tags Payments - Admin
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Success 200 {object} map[string]interface{} "Payment transactions retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/payments/subsidy-usage [get]
func (h *PaymentReportsHandler) GetSubsidyUsageSummary(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	// Parse optional date fields
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid start_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}

	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid end_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}

	// Fetch subsidy usage summary from tracking service
	subsidySummary, err := h.trackingService.GetSubsidyUsageSummary(r.Context(), startDate, endDate)
	if err != nil {
		log.Printf("[PAYMENT-REPORTS] Error fetching subsidy usage summary: %v", err)
		responses.RespondWithError(w, errLib.New("Failed to fetch subsidy usage summary", http.StatusInternalServerError))
		return
	}

	log.Printf("[PAYMENT-REPORTS] Subsidy usage: %d transactions, $%.2f total", subsidySummary.TransactionsWithSubsidy, subsidySummary.TotalSubsidyUsed)

	responses.RespondWithSuccess(w, subsidySummary, http.StatusOK)
}

// ExportPaymentTransactions exports payment transactions to CSV
// @Summary Export payment transactions
// @Description Export payment transactions to CSV format (admin only)
// @Tags Payments - Admin
// @Accept json
// @Produce text/csv
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Param transaction_type query string false "Filter by transaction type"
// @Param payment_status query string false "Filter by payment status"
// @Success 200 {file} file "CSV file"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security Bearer
// @Router /admin/payments/export [get]
func (h *PaymentReportsHandler) ExportPaymentTransactions(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	transactionType := r.URL.Query().Get("transaction_type")
	paymentStatus := r.URL.Query().Get("payment_status")

	// Parse optional date fields
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid start_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}

	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid end_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}

	// Fetch transactions for export
	transactions, err := h.trackingService.ExportPaymentTransactions(r.Context(), tracking.ExportFilters{
		StartDate:       startDate,
		EndDate:         endDate,
		TransactionType: transactionType,
		PaymentStatus:   paymentStatus,
	})

	if err != nil {
		log.Printf("[PAYMENT-REPORTS] Error exporting transactions: %v", err)
		responses.RespondWithError(w, errLib.New("Failed to export transactions", http.StatusInternalServerError))
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=payment_transactions.csv")

	// Write CSV header
	w.Write([]byte("ID,Customer ID,Customer Email,Customer Name,Transaction Type,Transaction Date,Original Amount,Discount Amount,Subsidy Amount,Customer Paid,Stripe Invoice ID,Payment Status,Payment Method,Currency,Description,Created At\n"))

	// Write data rows
	for _, tx := range transactions {
		originalAmt, _ := tx.OriginalAmount.Float64()
		discountAmt, _ := tx.DiscountAmount.Float64()
		subsidyAmt, _ := tx.SubsidyAmount.Float64()
		customerPaid, _ := tx.CustomerPaid.Float64()

		line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%.2f,%.2f,%.2f,%.2f,%s,%s,%s,%s,%s,%s\n",
			tx.ID,
			tx.CustomerID,
			tx.CustomerEmail,
			tx.CustomerName,
			tx.TransactionType,
			tx.TransactionDate.Format(time.RFC3339),
			originalAmt,
			discountAmt,
			subsidyAmt,
			customerPaid,
			nullStringToString(tx.StripeInvoiceID),
			tx.PaymentStatus,
			nullStringToString(tx.PaymentMethod),
			nullStringToString(tx.Currency),
			nullStringToString(tx.Description),
			tx.CreatedAt.Format(time.RFC3339),
		)
		w.Write([]byte(line))
	}

	log.Printf("[PAYMENT-REPORTS] Exported %d transactions to CSV", len(transactions))
}

// Helper to convert nullable SQL types to JSON-friendly values
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullUUIDToString(nu uuid.NullUUID) string {
	if nu.Valid {
		return nu.UUID.String()
	}
	return ""
}

func nullTimeToString(nt sql.NullTime) string {
	if nt.Valid {
		return nt.Time.Format(time.RFC3339)
	}
	return ""
}

// Helper to marshal response with proper null handling
func marshalResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// BackfillPaymentURLs fetches and stores receipt/invoice URLs for existing transactions
// @Summary Backfill payment URLs
// @Description Fetches receipt and invoice URLs from Stripe for existing transactions that don't have them (admin only)
// @Tags Payments - Admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Backfill completed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/payments/backfill-urls [post]
func (h *PaymentReportsHandler) BackfillPaymentURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get all transactions that need backfilling
	transactions, err := h.queries.GetTransactionsForBackfill(ctx)
	if err != nil {
		log.Printf("[PAYMENT-BACKFILL] Error fetching transactions for backfill: %v", err)
		responses.RespondWithError(w, errLib.New("Failed to fetch transactions", http.StatusInternalServerError))
		return
	}

	log.Printf("[PAYMENT-BACKFILL] Found %d transactions to backfill", len(transactions))

	successCount := 0
	errorCount := 0

	for _, tx := range transactions {
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
					errorCount++
				}
				continue
			}
			// For one-time payments, get receipt from payment_intent
			if sess.PaymentIntent != nil && sess.PaymentIntent.LatestCharge != nil && sess.PaymentIntent.LatestCharge.ReceiptURL != "" {
				receiptURL = sql.NullString{String: sess.PaymentIntent.LatestCharge.ReceiptURL, Valid: true}
				log.Printf("[PAYMENT-BACKFILL] Found receipt URL for transaction %s via checkout session", tx.ID)
			}

			// For one-time payments (credit packages, etc.), get invoice from the charge's invoice
			if sess.PaymentIntent != nil && sess.PaymentIntent.LatestCharge != nil &&
				sess.PaymentIntent.LatestCharge.Invoice != nil && sess.PaymentIntent.LatestCharge.Invoice.ID != "" {
				inv, invErr := invoice.Get(sess.PaymentIntent.LatestCharge.Invoice.ID, nil)
				if invErr != nil {
					log.Printf("[PAYMENT-BACKFILL] Error fetching invoice %s for one-time payment: %v", sess.PaymentIntent.LatestCharge.Invoice.ID, invErr)
				} else {
					if inv.HostedInvoiceURL != "" {
						invoiceURL = sql.NullString{String: inv.HostedInvoiceURL, Valid: true}
					}
					if inv.InvoicePDF != "" {
						invoicePDFURL = sql.NullString{String: inv.InvoicePDF, Valid: true}
					}
					if invoiceURL.Valid {
						log.Printf("[PAYMENT-BACKFILL] Found invoice URLs for transaction %s via one-time payment charge", tx.ID)
					}
				}
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

		// Fallback: Fetch receipt URL and invoice URLs directly from PaymentIntent if we have it
		// This handles cases where checkout session has expired/been deleted
		if tx.StripePaymentIntentID.Valid && tx.StripePaymentIntentID.String != "" && (!receiptURL.Valid || !invoiceURL.Valid) {
			pi, piErr := paymentintent.Get(tx.StripePaymentIntentID.String, &stripe.PaymentIntentParams{
				Expand: []*string{stripe.String("latest_charge"), stripe.String("latest_charge.invoice")},
			})
			if piErr != nil {
				log.Printf("[PAYMENT-BACKFILL] Error fetching PaymentIntent %s: %v", tx.StripePaymentIntentID.String, piErr)
			} else if pi.LatestCharge != nil {
				// Get receipt URL
				if !receiptURL.Valid && pi.LatestCharge.ReceiptURL != "" {
					receiptURL = sql.NullString{String: pi.LatestCharge.ReceiptURL, Valid: true}
					log.Printf("[PAYMENT-BACKFILL] Found receipt URL for transaction %s via payment intent fallback", tx.ID)
				}
				// Get invoice URLs from the charge's invoice (for one-time payments)
				if !invoiceURL.Valid && pi.LatestCharge.Invoice != nil && pi.LatestCharge.Invoice.ID != "" {
					inv, invErr := invoice.Get(pi.LatestCharge.Invoice.ID, nil)
					if invErr != nil {
						log.Printf("[PAYMENT-BACKFILL] Error fetching invoice %s via payment intent fallback: %v", pi.LatestCharge.Invoice.ID, invErr)
					} else {
						if inv.HostedInvoiceURL != "" {
							invoiceURL = sql.NullString{String: inv.HostedInvoiceURL, Valid: true}
						}
						if inv.InvoicePDF != "" {
							invoicePDFURL = sql.NullString{String: inv.InvoicePDF, Valid: true}
						}
						if invoiceURL.Valid {
							log.Printf("[PAYMENT-BACKFILL] Found invoice URLs for transaction %s via payment intent fallback", tx.ID)
						}
					}
				}
			}
		}

		// Fetch invoice URLs from Invoice (for subscription payments)
		if tx.StripeInvoiceID.Valid && tx.StripeInvoiceID.String != "" {
			inv, invErr := invoice.Get(tx.StripeInvoiceID.String, nil)
			if invErr != nil {
				log.Printf("[PAYMENT-BACKFILL] Error fetching Invoice %s: %v", tx.StripeInvoiceID.String, invErr)
				errorCount++
				continue
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
			updateErr := h.queries.UpdatePaymentUrls(ctx, db.UpdatePaymentUrlsParams{
				ID:            tx.ID,
				ReceiptUrl:    receiptURL,
				InvoiceUrl:    invoiceURL,
				InvoicePdfUrl: invoicePDFURL,
			})
			if updateErr != nil {
				log.Printf("[PAYMENT-BACKFILL] Error updating transaction %s: %v", tx.ID, updateErr)
				errorCount++
				continue
			}
			successCount++
		} else {
			log.Printf("[PAYMENT-BACKFILL] No URLs found for transaction %s", tx.ID)
		}
	}

	log.Printf("[PAYMENT-BACKFILL] Backfill completed: %d succeeded, %d failed", successCount, errorCount)

	responses.RespondWithSuccess(w, map[string]interface{}{
		"message":       "Backfill completed",
		"total":         len(transactions),
		"success_count": successCount,
		"error_count":   errorCount,
	}, http.StatusOK)
}

// BackfillMissingTransactions creates payment_transactions records for historical Stripe payments
// that were made before the payment tracking system was implemented.
// @Summary Backfill missing payment transactions
// @Description Fetches checkout sessions from Stripe and creates missing payment_transactions records (admin only)
// @Tags Payments - Admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Backfill completed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/payments/backfill-transactions [post]
func (h *PaymentReportsHandler) BackfillMissingTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Printf("[TRANSACTION-BACKFILL] Starting backfill of missing payment transactions from Stripe")

	// Track results
	var created, skipped, errors int
	var results []map[string]interface{}

	// List all completed checkout sessions from Stripe
	params := &stripe.CheckoutSessionListParams{
		Status: stripe.String("complete"),
	}
	params.Limit = stripe.Int64(100)
	params.AddExpand("data.line_items")
	params.AddExpand("data.customer")
	params.AddExpand("data.payment_intent")

	iter := session.List(params)
	for iter.Next() {
		sess := iter.CheckoutSession()

		// Skip if no metadata with userID
		if sess.Metadata == nil || sess.Metadata["userID"] == "" {
			log.Printf("[TRANSACTION-BACKFILL] Skipping session %s: no userID in metadata", sess.ID)
			skipped++
			continue
		}

		userIDStr := sess.Metadata["userID"]
		userID, parseErr := uuid.Parse(userIDStr)
		if parseErr != nil {
			log.Printf("[TRANSACTION-BACKFILL] Skipping session %s: invalid userID %s", sess.ID, userIDStr)
			skipped++
			continue
		}

		// Check if transaction already exists for this checkout session
		existingTx, _ := h.queries.GetPaymentTransactionByStripeCheckoutSession(ctx, sql.NullString{
			String: sess.ID,
			Valid:  true,
		})
		if existingTx.ID != uuid.Nil {
			log.Printf("[TRANSACTION-BACKFILL] Skipping session %s: transaction already exists", sess.ID)
			skipped++
			continue
		}

		// Get user details
		var customerEmail, customerName string
		userQuery := "SELECT email, first_name, last_name FROM users.users WHERE id = $1"
		var email sql.NullString
		var firstName, lastName string
		if err := h.db.QueryRowContext(ctx, userQuery, userID).Scan(&email, &firstName, &lastName); err != nil {
			log.Printf("[TRANSACTION-BACKFILL] Warning: could not get user details for %s: %v", userID, err)
		} else {
			if email.Valid {
				customerEmail = email.String
			}
			customerName = firstName + " " + lastName
		}

		// Determine transaction type from metadata or line items
		transactionType := "unknown"
		var creditPackageID, membershipPlanID, programID, eventID *uuid.UUID

		if sess.Metadata["creditPackageID"] != "" {
			transactionType = "credit_package"
			if id, err := uuid.Parse(sess.Metadata["creditPackageID"]); err == nil {
				creditPackageID = &id
			}
		} else if sess.Metadata["membershipPlanID"] != "" {
			transactionType = "membership_subscription"
			if id, err := uuid.Parse(sess.Metadata["membershipPlanID"]); err == nil {
				membershipPlanID = &id
			}
		} else if sess.Metadata["programID"] != "" {
			transactionType = "program_enrollment"
			if id, err := uuid.Parse(sess.Metadata["programID"]); err == nil {
				programID = &id
			}
		} else if sess.Metadata["eventID"] != "" {
			transactionType = "event_registration"
			if id, err := uuid.Parse(sess.Metadata["eventID"]); err == nil {
				eventID = &id
			}
		} else if sess.Mode == stripe.CheckoutSessionModeSubscription {
			transactionType = "membership_subscription"
		} else if sess.Mode == stripe.CheckoutSessionModePayment {
			// Try to determine type from line items by looking up price IDs
			transactionType = "credit_package" // Default for one-time payments
		}

		// Calculate amounts
		originalAmount := float64(sess.AmountTotal) / 100.0
		discountAmount := 0.0
		if sess.TotalDetails != nil && sess.TotalDetails.AmountDiscount > 0 {
			discountAmount = float64(sess.TotalDetails.AmountDiscount) / 100.0
		}
		customerPaid := originalAmount - discountAmount

		// Get receipt URL
		var receiptURL string
		if sess.PaymentIntent != nil && sess.PaymentIntent.LatestCharge != nil {
			receiptURL = sess.PaymentIntent.LatestCharge.ReceiptURL
		}

		// Get Stripe customer ID
		var stripeCustomerID string
		if sess.Customer != nil {
			stripeCustomerID = sess.Customer.ID
		}

		// Get subscription ID if applicable
		var stripeSubscriptionID string
		if sess.Subscription != nil {
			stripeSubscriptionID = sess.Subscription.ID
		}

		// Create the transaction
		transactionDate := time.Unix(sess.Created, 0)
		_, createErr := h.trackingService.TrackPayment(ctx, tracking.TrackPaymentParams{
			CustomerID:              userID,
			CustomerEmail:           customerEmail,
			CustomerName:            customerName,
			TransactionType:         transactionType,
			TransactionDate:         transactionDate,
			OriginalAmount:          originalAmount,
			DiscountAmount:          discountAmount,
			SubsidyAmount:           0,
			CustomerPaid:            customerPaid,
			CreditPackageID:         creditPackageID,
			MembershipPlanID:        membershipPlanID,
			ProgramID:               programID,
			EventID:                 eventID,
			StripeCustomerID:        stripeCustomerID,
			StripeSubscriptionID:    stripeSubscriptionID,
			StripeCheckoutSessionID: sess.ID,
			PaymentStatus:           "completed",
			Currency:                string(sess.Currency),
			ReceiptURL:              receiptURL,
			Description:             fmt.Sprintf("Backfilled from Stripe session %s", sess.ID),
		})

		if createErr != nil {
			log.Printf("[TRANSACTION-BACKFILL] Error creating transaction for session %s: %v", sess.ID, createErr)
			errors++
			results = append(results, map[string]interface{}{
				"session_id": sess.ID,
				"user_id":    userID,
				"status":     "error",
				"error":      createErr.Error(),
			})
		} else {
			log.Printf("[TRANSACTION-BACKFILL] Created transaction for session %s, user %s, type %s, amount $%.2f",
				sess.ID, userID, transactionType, customerPaid)
			created++
			results = append(results, map[string]interface{}{
				"session_id":       sess.ID,
				"user_id":          userID,
				"transaction_type": transactionType,
				"amount":           customerPaid,
				"status":           "created",
			})
		}
	}

	if err := iter.Err(); err != nil {
		log.Printf("[TRANSACTION-BACKFILL] Error iterating Stripe sessions: %v", err)
		responses.RespondWithError(w, errLib.New("Failed to list Stripe sessions: "+err.Error(), http.StatusInternalServerError))
		return
	}

	log.Printf("[TRANSACTION-BACKFILL] Backfill completed: %d created, %d skipped, %d errors", created, skipped, errors)

	responses.RespondWithSuccess(w, map[string]interface{}{
		"message":       "Transaction backfill completed",
		"created":       created,
		"skipped":       skipped,
		"error_count":   errors,
		"details":       results,
	}, http.StatusOK)
}
