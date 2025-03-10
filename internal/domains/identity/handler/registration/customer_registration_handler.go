package registration

import (
	"api/internal/di"
	dto "api/internal/domains/identity/dto/customer"
	service "api/internal/domains/identity/service/firebase"
	"api/internal/domains/identity/service/registration"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type CustomerRegistrationHandlers struct {
	CustomerRegistrationService *registration.CustomerRegistrationService
	FirebaseService             *service.Service
	StaffRegistrationService    *registration.StaffsRegistrationService
}

func NewCustomerRegistrationHandlers(container *di.Container) *CustomerRegistrationHandlers {

	return &CustomerRegistrationHandlers{
		CustomerRegistrationService: registration.NewCustomerRegistrationService(container),
		StaffRegistrationService:    registration.NewStaffRegistrationService(container),
		FirebaseService:             service.NewFirebaseService(container),
	}
}

// RegisterCustomer registers a new customer based on their role (athlete or parent).
// It verifies the Firebase token and uses the provided customer details to create a new customer account.
//
// @Summary Register a new customer
// @Description Registers a new customer by verifying the Firebase token and creating an account based on the provided details. The registration can either be for an athlete or a parent, depending on the specified role in the request.
// @Tags registration
// @Accept json
// @Produce json
// @Param customer body dto.RegistrationRequestDto true "Customer registration details" // The customer registration data, including name, age, email, phone, consent, etc.
// @Param firebase_token header string true "Firebase token for user verification" // Firebase token required for verifying the user
// @Success 201 {object} map[string]interface{} "Customer registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Missing or invalid Firebase token or request body"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to register customer"
// @Router /register/customer [post]
func (h *CustomerRegistrationHandlers) RegisterCustomer(w http.ResponseWriter, r *http.Request) {

	firebaseToken := r.Header.Get("firebase_token")

	if firebaseToken == "" {
		responseHandlers.RespondWithError(w, errLib.New("Missing Firebase token", http.StatusBadRequest))
		return
	}

	var requestDto dto.RegistrationRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	email, err := h.FirebaseService.GetUserEmail(r.Context(), firebaseToken)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	switch requestDto.Role {
	case "athlete":
		{

			vo, athleteErr := requestDto.ToAthlete(email)

			if athleteErr != nil {
				responseHandlers.RespondWithError(w, athleteErr)
				return
			}

			err = h.CustomerRegistrationService.RegisterAthlete(r.Context(), vo)
		}

	case "parent":
		{
			vo, parentErr := requestDto.ToParent(email)

			if parentErr != nil {
				responseHandlers.RespondWithError(w, parentErr)
				return
			}

			err = h.CustomerRegistrationService.RegisterParent(r.Context(), vo)
		}

	default:
		responseHandlers.RespondWithError(w, errLib.New("Unsupported Role", http.StatusBadRequest))
		return
	}

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}
