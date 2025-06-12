package playground

import (
	"api/internal/di"
	systemDto "api/internal/domains/playground/dto/system"
	service "api/internal/domains/playground/services"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type SystemsHandlers struct {
	Service *service.Service
}

func NewSystemsHandlers(container *di.Container) *SystemsHandlers {
	return &SystemsHandlers{Service: service.NewService(container)}
}

// CreateSystem creates a new playground system.
// @Tags playground-systems
// @Accept json
// @Produce json
// @Param system body systemDto.RequestDto true "System details"
// @Security Bearer
// @Success 201 {object} systemDto.ResponseDto
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /playground/systems [post]
func (h *SystemsHandlers) CreateSystem(w http.ResponseWriter, r *http.Request) {
	var dto systemDto.RequestDto
	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	value, err := dto.ToCreateValue()
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	system, err := h.Service.CreateSystem(r.Context(), value)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, systemDto.NewResponse(system), http.StatusCreated)
}

// GetSystems lists all playground systems.
// @Tags playground-systems
// @Produce json
// @Success 200 {array} systemDto.ResponseDto
// @Router /playground/systems [get]
func (h *SystemsHandlers) GetSystems(w http.ResponseWriter, r *http.Request) {
	systems, err := h.Service.GetSystems(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	resp := make([]systemDto.ResponseDto, len(systems))
	for i, s := range systems {
		resp[i] = systemDto.NewResponse(s)
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}

// UpdateSystem updates a playground system.
// @Tags playground-systems
// @Accept json
// @Produce json
// @Param id path string true "System ID"
// @Param system body systemDto.RequestDto true "System details"
// @Security Bearer
// @Success 200 {object} systemDto.ResponseDto
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /playground/systems/{id} [put]
func (h *SystemsHandlers) UpdateSystem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var dto systemDto.RequestDto
	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	value, err := dto.ToUpdateValue(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	system, err := h.Service.UpdateSystem(r.Context(), value)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, systemDto.NewResponse(system), http.StatusOK)
}

// DeleteSystem deletes a playground system.
// @Tags playground-systems
// @Param id path string true "System ID"
// @Security Bearer
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /playground/systems/{id} [delete]
func (h *SystemsHandlers) DeleteSystem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	if err = h.Service.DeleteSystem(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
