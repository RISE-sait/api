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
// @Summary Create a court
// @Description Creates a new court for a location
// @Tags courts
// @Accept json
// @Produce json
// @Param court body dto.RequestDto true "Court details"
// @Security Bearer
// @Success 201 {object} dto.ResponseDto "Court created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courts [post]
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
// @Summary Get a court by ID
// @Description Retrieves a single court using its UUID
// @Tags courts
// @Produce json
// @Param id path string true "Court ID"
// @Success 200 {object} dto.ResponseDto "Court retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Court not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courts/{id} [get]
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
// @Summary List courts
// @Description Retrieves all courts
// @Tags courts
// @Produce json
// @Success 200 {array} dto.ResponseDto "List of courts retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courts [get]
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
// @Summary Update a court
// @Description Updates a court's details
// @Tags courts
// @Accept json
// @Produce json
// @Param id path string true "Court ID"
// @Param court body dto.RequestDto true "Updated court details"
// @Security Bearer
// @Success 204 "Court updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Court not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courts/{id} [put]
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

// @Summary Delete a court
// @Description Deletes a court by its ID
// @Tags courts
// @Produce json
// @Param id path string true "Court ID"
// @Security Bearer
// @Success 204 "Court deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Court not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courts/{id} [delete]
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
