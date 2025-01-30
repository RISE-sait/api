package identity

import (
	"api/cmd/server/di"
	identity "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
	errLib "api/internal/libs/errors"
	"api/internal/libs/jwt"
	response_handlers "api/internal/libs/responses"
	"encoding/json"
	"io"
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

	body, ioErr := io.ReadAll(r.Body)

	if ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Failed to read request body", http.StatusBadRequest))
		return
	}

	var credentials identity.Credentials

	if ioErr := json.Unmarshal(body, &credentials); ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Invalid JSON format for credentials", http.StatusBadRequest))
		return
	}

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
