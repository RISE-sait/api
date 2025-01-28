package staff

import (
	"api/internal/domains/identity/customer_registration/infra/http/dto"
	"api/internal/domains/identity/customer_registration/values"
	"api/internal/domains/staff"
	"api/internal/libs/jwt"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type Handler struct {
	StaffService *staff.StaffService
}

func NewHandler(accountRegistrationService *staff.StaffService) *Handler {
	return &Handler{
		StaffService: accountRegistrationService,
	}
}

func (c *Handler) CreateTraditionalAccount(w http.ResponseWriter, r *http.Request) {
	var dto dto.CreateUserRequest

	// Step 1: Decode and validate the request body.
	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	userPasswordCreate := values.NewUserPasswordCreate(dto.Email, dto.Password)
	staffCreate := values.NewStaffCreate(dto.Role, dto.IsActiveStaff)

	// Step 2: Call the service to create the account
	userInfo, err := c.StaffService.CreateAccount(r.Context(), userPasswordCreate, staffCreate)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	// Step 3: Create JWT claims
	signedToken, err := jwt.SignJWT(*userInfo)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	// Step 4: Set Authorization header and respond
	w.Header().Set("Authorization", "Bearer "+signedToken)
	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
