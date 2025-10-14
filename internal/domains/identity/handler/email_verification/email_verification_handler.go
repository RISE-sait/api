package email_verification

import (
	"api/config"
	"api/internal/di"
	"api/internal/domains/identity/service/email_verification"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"api/utils/email"
	"log"
	"net/http"
	"strings"
)

type EmailVerificationHandler struct {
	VerificationService *email_verification.EmailVerificationService
	FrontendBaseURL     string
}

func NewEmailVerificationHandler(container *di.Container) *EmailVerificationHandler {
	return &EmailVerificationHandler{
		VerificationService: email_verification.NewEmailVerificationService(container),
		FrontendBaseURL:     config.Env.FrontendBaseURL,
	}
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// VerifyEmail verifies a user's email address using the provided token
// @Summary Verify email address
// @Description Verifies a user's email address using the verification token sent to their email
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Verification token"
// @Success 200 {object} map[string]interface{} "Email verified successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid or missing token"
// @Failure 409 {object} map[string]interface{} "Conflict: Email already verified"
// @Failure 410 {object} map[string]interface{} "Gone: Token expired"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /verify-email [post]
func (h *EmailVerificationHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var requestDto VerifyEmailRequest

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	token := strings.TrimSpace(requestDto.Token)
	if token == "" {
		responseHandlers.RespondWithError(w, errLib.New("Token is required", http.StatusBadRequest))
		return
	}

	log.Printf("Processing email verification for token: %s...", token[:10])

	if err := h.VerificationService.VerifyEmailToken(r.Context(), token); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"message": "Email verified successfully. You can now log in to your account.",
		"verified": true,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// ResendVerificationEmail resends the verification email to a user
// @Summary Resend verification email
// @Description Generates a new verification token and resends the verification email
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body ResendVerificationRequest true "User email"
// @Success 200 {object} map[string]interface{} "Verification email sent"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid email"
// @Failure 404 {object} map[string]interface{} "Not Found: User not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Email already verified"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /resend-verification [post]
func (h *EmailVerificationHandler) ResendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	var requestDto ResendVerificationRequest

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	userEmail := strings.TrimSpace(requestDto.Email)
	if userEmail == "" {
		responseHandlers.RespondWithError(w, errLib.New("Email is required", http.StatusBadRequest))
		return
	}

	log.Printf("Resending verification email to: %s", userEmail)

	// Generate new token and get user info
	userID, firstName, token, err := h.VerificationService.ResendVerificationEmail(r.Context(), userEmail)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Generate verification URL
	verificationURL := h.VerificationService.GetVerificationURL(token, h.FrontendBaseURL)

	// Send the verification email
	if emailErr := email.SendEmailVerification(userEmail, firstName, verificationURL); emailErr != nil {
		log.Printf("Failed to send verification email to %s for user %s: %v", userEmail, userID, emailErr)
		responseHandlers.RespondWithError(w, emailErr)
		return
	}

	response := map[string]interface{}{
		"message": "Verification email sent successfully. Please check your inbox.",
		"email":   userEmail,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}
