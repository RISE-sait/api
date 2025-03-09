package registration

import (
	"api/internal/di"
	dto "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/service/registration"
	values "api/internal/domains/identity/values"
	responsehandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type StaffHandlers struct {
	StaffRegistrationService *service.StaffsRegistrationService
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
// @Param staff body dto.StaffRegistrationRequestDto true "Staff registration details"
// @Success 201 {object} map[string]interface{} "Staff registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized: Invalid or missing authentication token"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to register staff"
// @Router /register/staff [post]
func (h *StaffHandlers) CreateStaff(w http.ResponseWriter, r *http.Request) {

	var requestDto dto.StaffRegistrationRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responsehandlers.RespondWithError(w, err)
		return
	}

	valueObject := values.StaffRegistrationRequestInfo{
		UserRegistrationRequestNecessaryInfo: values.UserRegistrationRequestNecessaryInfo{
			Age:       requestDto.Age,
			FirstName: requestDto.FirstName,
			LastName:  requestDto.LastName,
		},
		StaffCreateValues: values.StaffCreateValues{
			IsActive: requestDto.IsActiveStaff,
			RoleName: requestDto.Role,
		},
	}

	err := h.StaffRegistrationService.RegisterStaff(r.Context(), valueObject)
	if err != nil {
		responsehandlers.RespondWithError(w, err)
		return
	}

	responsehandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
