package registration

import (
	"api/internal/di"
	"api/internal/domains/identity/dto/staff"
	service "api/internal/domains/identity/service/staff"
	responsehandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type StaffHandlers struct {
	StaffRegistrationService *service.RegistrationService
}

func NewStaffRegistrationHandlers(container *di.Container) *StaffHandlers {

	staffRegistrationService := service.NewStaffRegistrationService(container)

	return &StaffHandlers{
		StaffRegistrationService: staffRegistrationService,
	}
}

// CreateStaff registers a new staff member account.
// @Summary Register a new staff member
// @Description Creates a new staff account in the system using the provided registration details.
// @Tags registration
// @Accept json
// @Produce json
// @Param staff body staff.RegistrationRequestDto true "Staff registration details"
// @Success 201 {object} map[string]interface{} "Staff registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized: Invalid or missing authentication token"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to register staff"
// @Router /register/staff [post]
func (c *StaffHandlers) CreateStaff(w http.ResponseWriter, r *http.Request) {

	var dto staff.RegistrationRequestDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		responsehandlers.RespondWithError(w, err)
		return
	}

	valueObject, err := dto.ToDetails()

	if err != nil {
		responsehandlers.RespondWithError(w, err)
		return
	}

	err = c.StaffRegistrationService.RegisterStaff(r.Context(), valueObject)
	if err != nil {
		responsehandlers.RespondWithError(w, err)
		return
	}

	responsehandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
