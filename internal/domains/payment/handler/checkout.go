package payment

import (
	"api/internal/di"
	dto "api/internal/domains/payment/dto"
	service "api/internal/domains/payment/services"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
)

type CheckoutHandlers struct {
	Service *service.Service
}

func NewCheckoutHandlers(container *di.Container) *CheckoutHandlers {
	return &CheckoutHandlers{Service: service.NewPurchaseService(container)}
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
func (h *CheckoutHandlers) CheckoutMembership(w http.ResponseWriter, r *http.Request) {

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
func (h *CheckoutHandlers) CheckoutProgram(w http.ResponseWriter, r *http.Request) {

	var programID uuid.UUID

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

	var responseDto dto.CheckoutResponseDto

	if paymentLink, err := h.Service.CheckoutProgram(r.Context(), programID); err != nil {
		responseHandlers.RespondWithError(w, err)
	} else {
		responseDto.PaymentURL = paymentLink
		responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
	}
}
