package auth

import (
	"api/internal/di"
	dto "api/internal/domains/identity/dto"
	service "api/internal/domains/identity/services"
	"api/internal/libs/jwt"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
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

// Login authenticates a user and returns a JWT token.
// @Summary Authenticate a user and return a JWT token
// @Description Authenticates a user using credentials and returns a JWT token
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body dto.LoginCredentialsDto true "User login credentials"
// @Success 200 {object} entity.UserInfo "User authenticated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /auth/login [post]
func (h *AuthenticationController) Login(w http.ResponseWriter, r *http.Request) {

	var credentialsDto dto.LoginCredentialsDto
	if err := validators.ParseJSON(r.Body, &credentialsDto); err != nil {
		response_handlers.RespondWithError(w, err)
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
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: false, // Prevent JavaScript access
		Secure:   false, // Use HTTPS in production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour), // Set expiration to 24 hours
	})

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
	response_handlers.RespondWithSuccess(w, *userInfo, http.StatusOK)
}
