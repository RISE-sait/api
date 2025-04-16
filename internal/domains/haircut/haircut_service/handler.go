package haircut_service

import (
	"api/internal/di"
	dto "api/internal/domains/haircut/haircut_service/dto"
	repository "api/internal/domains/haircut/haircut_service/persistence"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/go-chi/chi"
	"net/http"
)

// ServicesHandler provides HTTP handlers for managing events.
type ServicesHandler struct {
	Repo *repository.BarberServiceRepository
}

func NewBarberServicesHandler(container *di.Container) *ServicesHandler {
	return &ServicesHandler{Repo: repository.NewBarberServiceRepository(container)}
}

// GetBarberServices gets all haircut services
// @Tags haircuts
// @Produce json
// @Success 200 {array} dto.BarberServiceResponseDto
// @Failure 500 {object} map[string]interface{}
// @Router /haircuts/services [get]
func (h *ServicesHandler) GetBarberServices(w http.ResponseWriter, r *http.Request) {

	services, err := h.Repo.GetBarberServices(r.Context())

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.BarberServiceResponseDto, len(services))

	for i, service := range services {
		result[i] = dto.NewServiceResponseDto(service)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// CreateBarberService creates a new barber service
// @Tags haircuts
// @Accept json
// @Produce json
// @Param request body dto.CreateBarberServiceRequestDto true "Create Barber Service Request"
// @Success 201 {object} nil
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /haircuts/services [post]
func (h *ServicesHandler) CreateBarberService(w http.ResponseWriter, r *http.Request) {

	var requestDto dto.CreateBarberServiceRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	createValues, err := requestDto.ToCreateBarberServiceValue()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.CreateBarberService(r.Context(), createValues.BarberID, createValues.ServiceTypeID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// DeleteBarberService deletes a barber service by its ID
// @Tags haircuts
// @Produce json
// @Param id path string true "Barber Service ID"
// @Success 204 "No Content: Updated successfully"
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /haircuts/services/{id} [delete]
func (h *ServicesHandler) DeleteBarberService(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
	}

	if err = h.Repo.DeleteBarberService(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
