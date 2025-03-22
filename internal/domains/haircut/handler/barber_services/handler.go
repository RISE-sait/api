package haircut

import (
	dto "api/internal/domains/haircut/dto/barber_services"
	repository "api/internal/domains/haircut/persistence/repository"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/go-chi/chi"
	"net/http"
)

// BarberServicesHandler provides HTTP handlers for managing events.
type BarberServicesHandler struct {
	Repo *repository.BarberServiceRepository
}

func NewBarberServicesHandler(repo *repository.BarberServiceRepository) *BarberServicesHandler {
	return &BarberServicesHandler{Repo: repo}
}

// GetBarberServices godoc
// @Summary Get all barber services
// @Description Retrieve a list of all barber services
// @Tags barber
// @Produce json
// @Success 200 {array} dto.BarberServiceResponseDto
// @Failure 500 {object} map[string]interface{}
// @Router /barbers/services [get]
func (h *BarberServicesHandler) GetBarberServices(w http.ResponseWriter, r *http.Request) {

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

// CreateBarberService godoc
// @Summary Create a new barber service
// @Description Create a new barber service with the provided details
// @Tags barber
// @Accept json
// @Produce json
// @Param request body dto.CreateBarberServiceRequestDto true "Create Barber Service Request"
// @Success 201 {object} nil
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /barbers/services [post]
func (h *BarberServicesHandler) CreateBarberService(w http.ResponseWriter, r *http.Request) {

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

// DeleteBarberService godoc
// @Summary Delete a barber service
// @Description Delete a barber service by its ID
// @Tags barber
// @Produce json
// @Param id path string true "Barber Service ID"
// @Success 204 "No Content: Updated successfully"
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /barbers/services/{id} [delete]
func (h *BarberServicesHandler) DeleteBarberService(w http.ResponseWriter, r *http.Request) {

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
