package payment

import (
	"api/config"
	"api/internal/di"
	dto "api/internal/domains/payment/dto"
	service "api/internal/domains/payment/services"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"io"
	"strings"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
)

type Handlers struct {
	Service *service.Service
}

func NewPaymentHandlers(container *di.Container) *Handlers {
	return &Handlers{Service: service.NewPurchaseService(container)}
}

// CheckoutMembership allows a customer to check out a membership plan.
// @Summary CheckoutMembershipPlan a membership plan
// @Description Generates a payment link for purchasing a membership plan.
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Membership plan ID"
// @Success 200 {object} dto.CheckoutResponseDto "Payment link generated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to process checkout"
// @Security Bearer
// @Router /checkout/membership_plans/{id} [post]
func (h *Handlers) CheckoutMembership(w http.ResponseWriter, r *http.Request) {

	var membershipPlanId uuid.UUID

	if idStr := chi.URLParam(r, "id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			membershipPlanId = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("membership planID must be provided", http.StatusBadRequest))
		return
	}

	var responseDto dto.CheckoutResponseDto

	if paymentLink, err := h.Service.CheckoutMembershipPlan(r.Context(), membershipPlanId); err != nil {
		responseDto.PaymentURL = paymentLink
		responseHandlers.RespondWithError(w, err)
	} else {
		responseDto.PaymentURL = paymentLink
		responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
	}
}

// CheckoutProgram allows a customer to check out a program.
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Program ID"
// @Success 200 {object} dto.CheckoutResponseDto "Payment link generated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to process checkout"
// @Security Bearer
// @Router /checkout/programs/{id} [post]
func (h *Handlers) CheckoutProgram(w http.ResponseWriter, r *http.Request) {

	var programID, userId uuid.UUID

	if idStr := chi.URLParam(r, "id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			programID = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("program id must be provided", http.StatusBadRequest))
		return
	}

	userId, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var responseDto dto.CheckoutResponseDto

	if paymentLink, err := h.Service.CheckoutProgram(r.Context(), userId, programID); err != nil {
		responseHandlers.RespondWithError(w, err)
	} else {
		responseDto.PaymentURL = paymentLink
		responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
	}
}

// HandleStripeWebhook processes incoming Stripe  webhook events.
// @Summary Receives and processes payment updates from Stripe .
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
func (h *Handlers) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Error reading request body", http.StatusBadRequest))
		return
	}

	// todo: use the actual webhook secret
	_ = config.Env.StripeSecretKey

	event, err := webhook.ConstructEvent(
		payload,
		r.Header.Get("Stripe-Signature"),
		"random-ass-string",
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
		if sessionErr := service.HandleCheckoutSessionCompleted(event); sessionErr != nil {
			responseHandlers.RespondWithError(w, sessionErr)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
