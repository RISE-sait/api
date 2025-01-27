package registration

import (
	"api/internal/domains/identity/lib"
	"api/internal/domains/identity/registration/dto"
	"api/internal/domains/identity/registration/values"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type Handler struct {
	RegistrationService *AccountRegistrationService
}

func NewHandler(RegistrationService *AccountRegistrationService) *Handler {
	return &Handler{RegistrationService: RegistrationService}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var dto dto.CreateUserRequest
	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	userCredentialsCreate := values.NewUserPasswordCreate(dto.Email, dto.Password)
	staffCreate := values.NewStaffCreate(dto.Role, dto.IsActiveStaff)
	waiverCreate := values.NewWaiverCreate(dto.Email, dto.WaiverUrl, dto.IsSignedWaiver)

	userInfo, err := h.RegistrationService.CreateAccount(r.Context(), userCredentialsCreate, staffCreate, waiverCreate)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	token, err := lib.SignJWT(*userInfo)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusNoContent)
}
