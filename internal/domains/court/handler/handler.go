package court

import (
	"net/http"

	"api/internal/di"
	dto "api/internal/domains/court/dto"
	service "api/internal/domains/court/services"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"

	"github.com/go-chi/chi"
)

type Handler struct {
	Service *service.Service
}

func NewHandler(container *di.Container) *Handler {
	return &Handler{Service: service.NewService(container)}
}

// CreateCourt handles POST /courts
func (h *Handler) CreateCourt(w http.ResponseWriter, r *http.Request) {
	var req dto.RequestDto
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	details, err := req.ToCreateDetails()
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	court, err := h.Service.CreateCourt(r.Context(), details)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, dto.NewResponse(court), http.StatusCreated)
}

// GetCourt handles GET /courts/{id}
func (h *Handler) GetCourt(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	court, err := h.Service.GetCourt(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, dto.NewResponse(court), http.StatusOK)
}

// GetCourts handles GET /courts
func (h *Handler) GetCourts(w http.ResponseWriter, r *http.Request) {
	courts, err := h.Service.GetCourts(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	resp := make([]dto.ResponseDto, len(courts))
	for i, c := range courts {
		resp[i] = dto.NewResponse(c)
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}

// UpdateCourt handles PUT /courts/{id}
func (h *Handler) UpdateCourt(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var req dto.RequestDto
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	details, err := req.ToUpdateDetails(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	if err = h.Service.UpdateCourt(r.Context(), details); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteCourt handles DELETE /courts/{id}
func (h *Handler) DeleteCourt(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	if err = h.Service.DeleteCourt(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}