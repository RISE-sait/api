package purchase

import (
	"api/internal/di"
	dto "api/internal/domains/purchase/dto"
	service "api/internal/domains/purchase/services"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"net/http"
)

type Handlers struct {
	Service *service.Service
}

func NewPurchaseHandlers(container *di.Container) *Handlers {
	return &Handlers{Service: service.NewPurchaseService(container)}
}

// PurchaseMembership allows a customer to purchase a membership plan.
// @Summary Purchase a membership plan
// @Description Allows a customer to purchase a membership plan by providing the plan details.
// @Tags purchases
// @Accept json
// @Produce json
// @Param request body dto.MembershipPlanRequestDto true "Membership purchase details"
// @Success 200 {object} map[string]interface{} "Membership purchased successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to process membership purchase"
// @Security Bearer
// @Router /purchases/memberships [post]
func (h *Handlers) PurchaseMembership(w http.ResponseWriter, r *http.Request) {

	var requestDto dto.MembershipPlanRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	userId := r.Context().Value("userId").(uuid.UUID)

	purchaseDetails, err := requestDto.ToPurchaseRequestInfo(userId)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.Purchase(r.Context(), purchaseDetails); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusOK)
}
