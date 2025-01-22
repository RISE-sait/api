package oauth

import (
	"api/internal/domains/identity/entities"
	"api/internal/domains/identity/lib"
	"api/internal/libs/errors"
	handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"log"
	"net/http"
)

type Handler struct {
	AuthService *Service
}

func NewHandler(authService *Service) *Handler {
	return &Handler{AuthService: authService}
}

// TokenResponse Structure to store the access token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token" validate:"omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")

	if code == "" {
		handlers.RespondWithError(w, errors.New("Authorization code is missing", http.StatusBadRequest))
		return
	}

	token, err := ExchangeCodeForToken(r.Context(), code)
	if err != nil {
		handlers.RespondWithError(w, errors.New("Invalid request body", http.StatusBadRequest))
		return
	}

	userInfoRespBody, err := h.AuthService.GetUserInfoRespBodyFromGoogleAPI(token.AccessToken)

	if err != nil {
		handlers.RespondWithError(w, err)
		return
	}

	var userInfo *entities.UserInfo

	if err := validators.DecodeRequestBody(userInfoRespBody, &userInfo); err != nil {
		log.Println("Error getting user info", err)
		handlers.RespondWithError(w, errors.New("Failed to parse user info from Google", http.StatusInternalServerError))
		return
	}

	userInfo, err = h.AuthService.SetUserInfoWithStaffDetails(r.Context(), *userInfo)

	signedToken, err := lib.SignJWT(*userInfo)
	if err != nil {
		handlers.RespondWithError(w, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+signedToken)
	w.WriteHeader(http.StatusOK)
}
