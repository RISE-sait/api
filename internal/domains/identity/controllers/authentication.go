package identity

import (
	"api/cmd/server/di"
	identity "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
	"api/internal/libs/jwt"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type AuthenticationController struct {
	AuthService *service.AuthenticationService
}

func NewAuthenticationController(container *di.Container) *AuthenticationController {

	authService := service.NewAuthenticationService(container)

	return &AuthenticationController{AuthService: authService}
}

func (h *AuthenticationController) Login(w http.ResponseWriter, r *http.Request) {
	var dto identity.Credentials
	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	credentials := identity.NewCredentials(dto.Email, dto.Password)

	userInfo, err := h.AuthService.AuthenticateUser(r.Context(), credentials)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	token, err := jwt.SignJWT(*userInfo)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusNoContent)
}
