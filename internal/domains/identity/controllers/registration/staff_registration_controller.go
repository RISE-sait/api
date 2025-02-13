package registration

import (
	"api/internal/di"
	identity "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type staffRegistrationController struct {
	StaffRegistrationService *service.StaffRegistrationService
}

func NewStaffRegistrationController(container *di.Container) *staffRegistrationController {

	staffRegistrationService := service.NewStaffRegistrationService(container)

	return &staffRegistrationController{
		StaffRegistrationService: staffRegistrationService,
	}
}

// CreateStaff creates a new staff member account.
// @Summary Create a new staff member account
// @Description Registers a new staff member with the provided details
// @Tags registration
// @Accept json
// @Produce json
// @Param staff body identity.StaffRegistrationDto true "Staff registration details"
// @Success 201 {object} identity.StaffRegistrationDto "Staff registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /register/staff [post]
func (c *staffRegistrationController) CreateStaff(w http.ResponseWriter, r *http.Request) {

	var dto identity.StaffRegistrationDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	valueObject, err := dto.ToValueObjects()

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	userInfo, err := c.StaffRegistrationService.RegisterStaff(r.Context(), valueObject)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, *userInfo, http.StatusCreated)
}
