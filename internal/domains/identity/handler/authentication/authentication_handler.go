package authentication

import (
	"api/internal/di"
	dto "api/internal/domains/identity/dto/common"
	service "api/internal/domains/identity/service/authentication"
	identityUtils "api/internal/domains/identity/utils"
	identity "api/internal/domains/identity/values"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"
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

	firebaseToken, err := identityUtils.GetFirebaseTokenFromAuthorizationHeader(r)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	jwtToken, userInfo, err := h.AuthService.AuthenticateUser(r.Context(), firebaseToken)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := mapReadValueToResponse(userInfo)

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

	parentID, ctxErr := contextUtils.GetUserID(r.Context())

	if ctxErr != nil {
		responseHandlers.RespondWithError(w, ctxErr)
		return
	}

	jwtToken, userInfo, err := h.AuthService.AuthenticateChild(r.Context(), childId, parentID)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := mapReadValueToResponse(userInfo)

	w.Header().Set("Authorization", "Bearer "+jwtToken)
	w.WriteHeader(http.StatusOK)
	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusOK)

}

// function to convert the user info to a response
func mapReadValueToResponse(userInfo identity.UserReadInfo) dto.UserAuthenticationResponseDto {
	responseBody := dto.UserAuthenticationResponseDto{
		ID:          userInfo.ID,
		FirstName:   userInfo.FirstName,
		LastName:    userInfo.LastName,
		Email:       userInfo.Email,
		Role:        userInfo.Role,
		Phone:       userInfo.Phone,
		Age:         userInfo.Age,
		CountryCode: userInfo.CountryCode,
	}

	if userInfo.MembershipInfo != nil {
		responseBody.MembershipInfo = &dto.MembershipReadResponseDto{
			MembershipName:        userInfo.MembershipInfo.MembershipName,
			MembershipDescription: userInfo.MembershipInfo.MembershipDescription,
			MembershipBenefits:    userInfo.MembershipInfo.MembershipBenefits,
			PlanName:              userInfo.MembershipInfo.PlanName,
			StartDate:             userInfo.MembershipInfo.StartDate,
		}

		if userInfo.MembershipInfo.RenewalDate != nil {
			responseBody.MembershipInfo.RenewalDate = userInfo.MembershipInfo.RenewalDate
		}
	}

	if userInfo.AthleteInfo != nil {
		responseBody.AthleteInfo = &dto.AthleteResponseDto{
			Wins:     userInfo.AthleteInfo.Wins,
			Losses:   userInfo.AthleteInfo.Losses,
			Points:   userInfo.AthleteInfo.Points,
			Steals:   userInfo.AthleteInfo.Steals,
			Assists:  userInfo.AthleteInfo.Assists,
			Rebounds: userInfo.AthleteInfo.Rebounds,
		}
	}

	return responseBody
}
