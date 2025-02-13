package auth

import (
	"api/internal/di"
	"time"

	entity "api/internal/domains/identity/entities"
	service "api/internal/domains/identity/services"
	errLib "api/internal/libs/errors"
	"api/internal/libs/jwt"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"log"
	"net/http"
)

type OauthController struct {
	AuthService *service.OauthService
}

func NewOauthController(container *di.Container) *OauthController {

	authService := service.NewOauthService(container)
	return &OauthController{AuthService: authService}
}

// TokenResponse Structure to store the access token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token" validate:"omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

func (h *OauthController) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")

	if code == "" {
		response_handlers.RespondWithError(w, errLib.New("Authorization code is missing", http.StatusBadRequest))
		return
	}

	token, err := service.ExchangeCodeForToken(r.Context(), code)
	if err != nil {
		response_handlers.RespondWithError(w, errLib.New("Invalid request body", http.StatusBadRequest))
		return
	}

	userInfoRespBody, err := h.AuthService.GetUserInfoRespBodyFromGoogleAPI(token.AccessToken)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	var userInfo *entity.UserInfo

	if err := validators.ParseJSON(userInfoRespBody, &userInfo); err != nil {
		log.Println("Error getting user info", err)
		response_handlers.RespondWithError(w, errLib.New("Failed to parse user info from Google", http.StatusInternalServerError))
		return
	}

	userInfo, err = h.AuthService.SetUserInfoWithStaffDetails(r.Context(), *userInfo)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	signedToken, err := jwt.SignJWT(*userInfo)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    signedToken,
		Path:     "/",
		HttpOnly: true, // Prevent JavaScript access
		Secure:   true, // Use HTTPS in production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour), // Set expiration to 24 hours
	})

	w.Header().Set("Authorization", "Bearer "+signedToken)
	w.WriteHeader(http.StatusOK)
}
