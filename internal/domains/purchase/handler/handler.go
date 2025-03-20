package purchase

import (
	"api/internal/di"
	dto "api/internal/domains/purchase/dto"
	service "api/internal/domains/purchase/services"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"api/internal/middlewares"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
)

type Handlers struct {
	Service *service.Service
}

func NewPurchaseHandlers(container *di.Container) *Handlers {
	return &Handlers{Service: service.NewPurchaseService(container)}
}

// CheckoutMembership allows a customer to check out a membership plan.
// @Summary Checkout a membership plan
// @Description Generates a payment link for purchasing a membership plan.
// @Tags purchases
// @Accept json
// @Produce json
// @Param id path string true "Membership plan ID"
// @Success 200 {object} dto.CheckoutResponseDto "Payment link generated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to process checkout"
// @Security Bearer
// @Router /checkout/membership_plans/{id} [post]
func (h *Handlers) CheckoutMembership(w http.ResponseWriter, r *http.Request) {

	var membershipPlanId, userId uuid.UUID

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

	if ctxUserId := r.Context().Value(middlewares.UserIDKey); ctxUserId == nil {
		responseHandlers.RespondWithError(w, errLib.New("User ID not found", http.StatusUnauthorized))
		return
	} else {
		userId = ctxUserId.(uuid.UUID)
	}

	var responseDto dto.CheckoutResponseDto

	if paymentLink, err := h.Service.Checkout(r.Context(), membershipPlanId, userId); err != nil {
		responseHandlers.RespondWithError(w, err)
	} else {
		responseDto.PaymentURL = paymentLink
		responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
	}
}

// HandleSquareWebhook processes incoming Square webhook events.
// @Summary Handle Square Webhook
// @Description Receives and processes payment updates from Square.
// @Tags purchases
// @Accept json
// @Produce json
// @Param request body dto.SquareWebhookEventDto true "Square Webhook Event"
// @Success 200 {object} map[string]interface{} "Webhook processed successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to process webhook event"
// @Router /purchases/square/webhook [post]
func (h *Handlers) HandleSquareWebhook(w http.ResponseWriter, r *http.Request) {
	var event dto.SquareWebhookEventDto

	if err := validators.ParseJSON(r.Body, &event); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.ProcessSquareWebhook(r.Context(), event); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusOK)
}
