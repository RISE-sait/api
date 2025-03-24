package registration

import (
	"api/internal/di"
	commonDto "api/internal/domains/identity/dto/common"
	customerDto "api/internal/domains/identity/dto/customer"
	service "api/internal/domains/identity/service/firebase"
	"api/internal/domains/identity/service/registration"
	identityUtils "api/internal/domains/identity/utils"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type AthleteRegistrationHandlers struct {
	CustomerRegistrationService *registration.CustomerRegistrationService
	FirebaseService             *service.Service
	StaffRegistrationService    *registration.StaffsRegistrationService
}

func NewAthleteRegistrationHandlers(container *di.Container) *AthleteRegistrationHandlers {

	return &AthleteRegistrationHandlers{
		CustomerRegistrationService: registration.NewCustomerRegistrationService(container),
		StaffRegistrationService:    registration.NewStaffRegistrationService(container),
		FirebaseService:             service.NewFirebaseService(container),
	}
}

// RegisterAthlete registers a new athlete account.
// It verifies the Firebase token and uses the provided athlete details to create a new account.
//
// @Summary Register a new athlete
// @Description Registers a new athlete by verifying the Firebase token and creating an account based on the provided details.
// @Tags registration
// @Accept json
// @Produce json
// @Param athlete body customerDto.AthleteRegistrationRequestDto true "Athlete registration details" // The athlete registration data, including name, age, email, phone, consent, etc.
// @Param Authorization header string true "Firebase token for user verification" // Firebase token required for verifying the athlete's identity
// @Success 201 {object} map[string]interface{} "Athlete registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Missing or invalid Firebase token or request body"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to register athlete"
// @Router /register/athlete [post]
func (h *AthleteRegistrationHandlers) RegisterAthlete(w http.ResponseWriter, r *http.Request) {

	firebaseToken, err := identityUtils.GetFirebaseTokenFromAuthorizationHeader(r)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var requestDto customerDto.AthleteRegistrationRequestDto

	if err = validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	email, err := h.FirebaseService.GetUserEmail(r.Context(), firebaseToken)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	vo, err := requestDto.ToAthlete(email)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if userInfo, err := h.CustomerRegistrationService.RegisterAthlete(r.Context(), vo); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {

		responseBody := commonDto.UserAuthenticationResponseDto{
			FirstName:   userInfo.FirstName,
			LastName:    userInfo.LastName,
			Email:       userInfo.Email,
			Role:        userInfo.Role,
			Phone:       userInfo.Phone,
			Age:         userInfo.Age,
			CountryCode: userInfo.CountryCode,
		}

		responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
	}
}
