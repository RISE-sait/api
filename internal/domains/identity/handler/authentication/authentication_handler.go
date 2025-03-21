package authentication

import (
	"api/internal/di"
	dto "api/internal/domains/identity/dto/common"
	service "api/internal/domains/identity/service/authentication"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"api/internal/middlewares"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type Handlers struct {
	AuthService *service.Service
}

func NewHandlers(container *di.Container) *Handlers {

	authService := service.NewAuthenticationService(container)

	return &Handlers{AuthService: authService}
}

// Login authenticates a user's firebase token and returns a JWT token.
// @Tags authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Firebase token for user verification"
// @Success 200 {object} dto.UserAuthenticationResponseDto "User authenticated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid Firebase token"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /auth [post]
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {

	firebaseToken := r.Header.Get("Authorization")

	if firebaseToken == "" {

		responseHandlers.RespondWithError(w, errLib.New("Missing Firebase token", http.StatusBadRequest))
		return
	}

	// Firebase token starts with bearer

	extractedTokenParts := strings.Split(firebaseToken, " ")

	if extractedTokenParts[0] != "Bearer" {
		responseHandlers.RespondWithError(w, errLib.New("Invalid Firebase token. Missing Bearer", http.StatusUnauthorized))
		return
	}

	extractedFirebaseToken := extractedTokenParts[1]

	jwtToken, userInfo, err := h.AuthService.AuthenticateUser(r.Context(), extractedFirebaseToken)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.UserAuthenticationResponseDto{
		ID:          userInfo.ID,
		Gender:      userInfo.Gender,
		FirstName:   userInfo.FirstName,
		LastName:    userInfo.LastName,
		Email:       userInfo.Email,
		Role:        userInfo.Role,
		Phone:       userInfo.Phone,
		Age:         userInfo.Age,
		CountryCode: userInfo.CountryCode,
	}

	if userInfo.IsActiveStaff != nil {
		responseBody.IsActiveStaff = userInfo.IsActiveStaff
	}

	if userInfo.MembershipInfo != nil {
		responseBody.MembershipInfo = &dto.MembershipReadResponseDto{
			MembershipName: userInfo.MembershipInfo.MembershipName,
			PlanName:       userInfo.MembershipInfo.PlanName,
			StartDate:      userInfo.MembershipInfo.StartDate,
		}

		if userInfo.MembershipInfo.RenewalDate != nil {
			responseBody.MembershipInfo.RenewalDate = userInfo.MembershipInfo.RenewalDate
		}
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
// @Param id path string true "Child ID"
// @Success 200 {object} dto.UserAuthenticationResponseDto "User authenticated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid Firebase token"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /auth/child/{id} [post]
func (h *Handlers) LoginAsChild(w http.ResponseWriter, r *http.Request) {

	childIdStr := chi.URLParam(r, "id")

	childId, err := validators.ParseUUID(childIdStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	parentId := r.Context().Value(middlewares.UserIDKey).(uuid.UUID)

	jwtToken, userInfo, err := h.AuthService.AuthenticateChild(r.Context(), childId, parentId)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.UserAuthenticationResponseDto{
		FirstName:   userInfo.FirstName,
		LastName:    userInfo.LastName,
		Email:       userInfo.Email,
		Role:        userInfo.Role,
		Phone:       userInfo.Phone,
		Age:         userInfo.Age,
		CountryCode: userInfo.CountryCode,
	}

	w.Header().Set("Authorization", "Bearer "+jwtToken)
	w.WriteHeader(http.StatusOK)
	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusOK)

}
