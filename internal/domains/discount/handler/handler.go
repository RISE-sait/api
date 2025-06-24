package discount

import (
	"net/http"

	"api/internal/di"
	dto "api/internal/domains/discount/dto"
	service "api/internal/domains/discount/service"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	Service *service.Service
}

func NewHandler(container *di.Container) *Handler {
	return &Handler{Service: service.NewService(container)}
}

// CreateDiscount creates a new discount
func (h *Handler) CreateDiscount(w http.ResponseWriter, r *http.Request) {
	var req dto.RequestDto
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	details, err := req.ToCreateValues()
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	created, err := h.Service.CreateDiscount(r.Context(), details)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	resp := dto.ResponseDto{
		ID:              created.ID,
		Name:            created.Name,
		Description:     created.Description,
		DiscountPercent: created.DiscountPercent,
		IsUseUnlimited:  created.IsUseUnlimited,
		UsePerClient:    created.UsePerClient,
		IsActive:        created.IsActive,
		ValidFrom:       created.ValidFrom,
	}
	if !created.ValidTo.IsZero() {
		resp.ValidTo = &created.ValidTo
	}
	resp.CreatedAt = created.CreatedAt
	resp.UpdatedAt = created.UpdatedAt

	responseHandlers.RespondWithSuccess(w, resp, http.StatusCreated)
}

// GetDiscount retrieves a discount by ID
func (h *Handler) GetDiscount(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	discount, err := h.Service.GetDiscount(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	resp := dto.ResponseDto{
		ID:              discount.ID,
		Name:            discount.Name,
		Description:     discount.Description,
		DiscountPercent: discount.DiscountPercent,
		IsUseUnlimited:  discount.IsUseUnlimited,
		UsePerClient:    discount.UsePerClient,
		IsActive:        discount.IsActive,
		ValidFrom:       discount.ValidFrom,
	}
	if !discount.ValidTo.IsZero() {
		resp.ValidTo = &discount.ValidTo
	}
	resp.CreatedAt = discount.CreatedAt
	resp.UpdatedAt = discount.UpdatedAt

	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}

// GetDiscounts retrieves all discounts
func (h *Handler) GetDiscounts(w http.ResponseWriter, r *http.Request) {
	discounts, err := h.Service.GetDiscounts(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	resp := make([]dto.ResponseDto, len(discounts))
	for i, d := range discounts {
		resp[i] = dto.ResponseDto{
			ID:              d.ID,
			Name:            d.Name,
			Description:     d.Description,
			DiscountPercent: d.DiscountPercent,
			IsUseUnlimited:  d.IsUseUnlimited,
			UsePerClient:    d.UsePerClient,
			IsActive:        d.IsActive,
			ValidFrom:       d.ValidFrom,
			CreatedAt:       d.CreatedAt,
			UpdatedAt:       d.UpdatedAt,
		}
		if !d.ValidTo.IsZero() {
			resp[i].ValidTo = &d.ValidTo
		}
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}

// UpdateDiscount updates a discount by ID
func (h *Handler) UpdateDiscount(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var req dto.RequestDto
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	details, err := req.ToUpdateValues(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	updated, err := h.Service.UpdateDiscount(r.Context(), details)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	resp := dto.ResponseDto{
		ID:              updated.ID,
		Name:            updated.Name,
		Description:     updated.Description,
		DiscountPercent: updated.DiscountPercent,
		IsUseUnlimited:  updated.IsUseUnlimited,
		UsePerClient:    updated.UsePerClient,
		IsActive:        updated.IsActive,
		ValidFrom:       updated.ValidFrom,
		CreatedAt:       updated.CreatedAt,
		UpdatedAt:       updated.UpdatedAt,
	}
	if !updated.ValidTo.IsZero() {
		resp.ValidTo = &updated.ValidTo
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}

// DeleteDiscount deletes a discount by ID
func (h *Handler) DeleteDiscount(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	if err := h.Service.DeleteDiscount(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// ApplyDiscount allows a logged in customer to apply a discount code by name
func (h *Handler) ApplyDiscount(w http.ResponseWriter, r *http.Request) {
	var req dto.ApplyRequestDto
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	if err := req.Validate(); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	var planID *uuid.UUID
	if req.MembershipPlanID != nil {
		id, err := validators.ParseUUID(*req.MembershipPlanID)
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
		planID = &id
	}

	applied, err := h.Service.ApplyDiscount(r.Context(), req.Name, planID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	resp := dto.ResponseDto{
		ID:              applied.ID,
		Name:            applied.Name,
		Description:     applied.Description,
		DiscountPercent: applied.DiscountPercent,
		IsUseUnlimited:  applied.IsUseUnlimited,
		UsePerClient:    applied.UsePerClient,
		IsActive:        applied.IsActive,
		ValidFrom:       applied.ValidFrom,
		CreatedAt:       applied.CreatedAt,
		UpdatedAt:       applied.UpdatedAt,
	}
	if !applied.ValidTo.IsZero() {
		resp.ValidTo = &applied.ValidTo
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}
