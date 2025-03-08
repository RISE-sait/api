package authentication

import (
	"api/internal/di"
	dto "api/internal/domains/identity/dto/common"
	service "api/internal/domains/identity/service/authentication"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/middlewares"
	"github.com/go-chi/chi"
	"net/http"
)

type Handlers struct {
	AuthService *service.Service
}

func NewHandlers(container *di.Container) *Handlers {

	authService := service.NewAuthenticationService(container)

	return &Handlers{AuthService: authService}
}

// Login authenticates a user and returns a JWT token.
// @Summary Authenticate a user and return a JWT token
// @Description Authenticates a user using Firebase token and returns a JWT token for the authenticated user
// @Tags authentication
// @Accept json
// @Produce json
// @Param firebase_token header string true "Firebase token for user verification" // Firebase token in the Authorization header
// @Success 200 {object} dto.UserAuthenticationResponseDto "User authenticated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid Firebase token"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /auth [post]
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {

	firebaseToken := r.Header.Get("firebase_token")

	if firebaseToken == "" {

		responseHandlers.RespondWithError(w, errLib.New("Missing Firebase token", http.StatusBadRequest))
		return
	}

	jwtToken, userInfo, err := h.AuthService.AuthenticateUser(r.Context(), firebaseToken)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.UserAuthenticationResponseDto{
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
		Role:      userInfo.Role,
		Age:       0,
	}

	w.Header().Set("Authorization", "Bearer "+jwtToken)
	w.WriteHeader(http.StatusOK)
	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusOK)

}

// LoginAsChild authenticates a user and returns a JWT token.
// @Summary Authenticate a user and return a JWT token
// @Description Authenticates a user using Firebase token and returns a JWT token for the authenticated user
// @Tags authentication
// @Accept json
// @Produce json
// @Security Bearer
// @Param hubspot_id path string true "Child HubSpotId"
// @Success 200 {object} identity.UserNecessaryInfoRequestDto "User authenticated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid Firebase token"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /auth/child/{hubspot_id} [post]
func (h *Handlers) LoginAsChild(w http.ResponseWriter, r *http.Request) {

	childHubspotId := chi.URLParam(r, "hubspot_id")

	parentHubspotId := r.Context().Value(middlewares.HubspotIDKey).(string)

	jwtToken, userInfo, err := h.AuthService.AuthenticateChild(r.Context(), childHubspotId, parentHubspotId)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.UserNecessaryInfoRequestDto{
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
		Age:       0,
	}

	w.Header().Set("Authorization", "Bearer "+jwtToken)
	w.WriteHeader(http.StatusOK)
	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusOK)

}
