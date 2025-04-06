package payment

import (
	"api/config"
	"api/internal/di"
	service "api/internal/domains/payment/services"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"io"
	"strings"

	"net/http"
)

type WebhookHandlers struct {
	Service *service.WebhookService
}

func NewWebhookHandlers(container *di.Container) *WebhookHandlers {
	return &WebhookHandlers{Service: service.NewWebhookService(container)}
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

	stripeWebhookSecret := config.Env.StripeWebhookSecret

	if stripeWebhookSecret == "" {
		responseHandlers.RespondWithError(w, errLib.New("Stripe webhook secret not configured", http.StatusInternalServerError))
		return
	}

	event, err := webhook.ConstructEvent(
		payload,
		r.Header.Get("Stripe-Signature"), stripeWebhookSecret,
	)

	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Signature verification failed", http.StatusBadRequest))
		return
	}

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		responseHandlers.RespondWithError(w, errLib.New("Stripe not configured with its API key", http.StatusInternalServerError))
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		if sessionErr := h.Service.HandleCheckoutSessionCompleted(event); sessionErr != nil {
			responseHandlers.RespondWithError(w, sessionErr)
			return
		}
	}
	w.WriteHeader(http.StatusOK)

}
