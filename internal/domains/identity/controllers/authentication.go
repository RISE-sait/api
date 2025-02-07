package identity

import (
	"api/internal/di"
	dto "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
	errLib "api/internal/libs/errors"
	"api/internal/libs/jwt"
	response_handlers "api/internal/libs/responses"
	"encoding/json"
	"io"
	"net/http"
	"time"
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

	var credentialsDto dto.LoginCredentialsDto

	if ioErr := json.Unmarshal(body, &credentialsDto); ioErr != nil {
		response_handlers.RespondWithError(w, errLib.New("Invalid JSON format for credentials", http.StatusBadRequest))
		return
	}

	credentials, err := credentialsDto.ToValueObjects()

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	userInfo, err := h.AuthService.AuthenticateUser(r.Context(), *credentials)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	token, err := jwt.SignJWT(*userInfo)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwtToken",
		Value:    token,
		Path:     "/",
		HttpOnly: false, // Prevent JavaScript access
		Secure:   false, // Use HTTPS in production
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour), // Set expiration to 24 hours
	})

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
	response_handlers.RespondWithSuccess(w, *userInfo, http.StatusOK)
}
