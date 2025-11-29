package payment

import (
	"api/config"
	"api/internal/di"
	service "api/internal/domains/payment/services"
	stripeService "api/internal/domains/payment/services/stripe"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/stripe/stripe-go/v81"
)

type WebhookHandlers struct {
	Service      *service.WebhookService
	RetryService *service.WebhookRetryService
}

func NewWebhookHandlers(container *di.Container) *WebhookHandlers {
	webhookService := service.NewWebhookService(container)
	retryService := service.NewWebhookRetryService(webhookService)
	

	
	return &WebhookHandlers{
		Service:      webhookService,
		RetryService: retryService,
	}
}

// HandleStripeWebhook processes incoming Stripe webhook payment events.
// @Description - checkout.session.completed: Logs completed checkout sessions
// @Tags payments
// @Accept json
// @Produce json
// @Param Stripe-Signature header string true "Stripe webhook signature"
// @Param request body string true "Raw webhook payload"
// @Success 200 {object} map[string]interface{} "Webhook processed successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to process webhook event"
// @Router /webhooks/stripe [post]
func (h *WebhookHandlers) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Error reading request body", http.StatusBadRequest))
		return
	}

	stripeWebhookSecret := strings.TrimSpace(config.Env.StripeWebhookSecret)

	if stripeWebhookSecret == "" {
		responseHandlers.RespondWithError(w, errLib.New("Stripe webhook secret not configured", http.StatusInternalServerError))
		return
	}

	log.Println(">>> Incoming Stripe webhook")
	log.Printf("[STRIPE] Event type: %s", string(payload)[:200]) // Log first 200 chars for debugging

	// Use the enhanced signature validation
	event, validationErr := stripeService.ValidateWebhookSignature(
		payload,
		r.Header.Get("Stripe-Signature"),
		stripeWebhookSecret,
	)

	if validationErr != nil {
		log.Printf("[STRIPE] Webhook signature validation failed: %v", validationErr)
		responseHandlers.RespondWithError(w, validationErr)
		return
	}

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		responseHandlers.RespondWithError(w, errLib.New("Stripe not configured with its API key", http.StatusInternalServerError))
		return
	}

	// Process the webhook event and handle retries on failure
	var webhookErr *errLib.CommonError
	ctx := r.Context()

	switch event.Type {
	case "checkout.session.completed":
		webhookErr = h.Service.HandleCheckoutSessionCompleted(ctx, *event)
	case "customer.subscription.created":
		webhookErr = h.Service.HandleSubscriptionCreated(ctx, *event)
	case "customer.subscription.updated":
		webhookErr = h.Service.HandleSubscriptionUpdated(ctx, *event)
	case "customer.subscription.deleted":
		webhookErr = h.Service.HandleSubscriptionDeleted(ctx, *event)
	case "invoice.created":
		// Apply subsidy credit BEFORE customer is charged
		webhookErr = h.Service.HandleInvoiceCreated(ctx, *event)
	case "invoice.finalized":
		// Apply subsidy credit at finalization (after subscription linked, before payment)
		webhookErr = h.Service.HandleInvoiceFinalized(ctx, *event)
	case "invoice.payment_succeeded":
		// Records subsidy usage after successful payment
		webhookErr = h.Service.HandleInvoicePaymentSucceededWithSubsidy(ctx, *event)
	case "invoice.payment_failed":
		webhookErr = h.Service.HandleInvoicePaymentFailed(ctx, *event)
	default:
		log.Printf("[STRIPE] Unhandled webhook event type: %s", event.Type)
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if webhookErr != nil {
		// Schedule for retry with exponential backoff
		h.RetryService.ScheduleRetry(*event, webhookErr)
		
		// Return 500 to tell Stripe to retry (though we handle our own retries)
		// Stripe will also retry on 5xx responses which provides additional redundancy
		responseHandlers.RespondWithError(w, errLib.New("Webhook processing failed, scheduled for retry", http.StatusInternalServerError))
		return
	}
	
	// Successfully processed - remove from retry queue if it was there
	h.RetryService.RemoveRetry(event.ID)

	w.WriteHeader(http.StatusOK)
}
