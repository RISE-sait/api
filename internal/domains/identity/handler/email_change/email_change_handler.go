package email_change

import (
	"api/config"
	"api/internal/di"
	"api/internal/domains/identity/dto"
	"api/internal/domains/identity/service/email_change"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"
	"api/utils/email"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type EmailChangeHandler struct {
	EmailChangeService *email_change.EmailChangeService
	FrontendBaseURL    string
}

func NewEmailChangeHandler(container *di.Container) *EmailChangeHandler {
	return &EmailChangeHandler{
		EmailChangeService: email_change.NewEmailChangeService(container),
		FrontendBaseURL:    config.Env.FrontendBaseURL,
	}
}

// InitiateEmailChange starts the email change process for a user
// @Summary Initiate email change
// @Description Initiates an email change by sending a verification link to the new email address
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.InitiateEmailChangeRequest true "New email address"
// @Security BearerAuth
// @Success 200 {object} dto.EmailChangeResponse "Email change initiated"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid email"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Cannot change another user's email"
// @Failure 404 {object} map[string]interface{} "Not Found: User not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Email already in use"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id}/email/change [post]
func (h *EmailChangeHandler) InitiateEmailChange(w http.ResponseWriter, r *http.Request) {
	// Get target user ID from URL
	userIDStr := chi.URLParam(r, "id")
	targetUserID, parseErr := uuid.Parse(userIDStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid user ID", http.StatusBadRequest))
		return
	}

	// Get authenticated user from context
	authUserID, userIDErr := contextUtils.GetUserID(r.Context())
	if userIDErr != nil {
		responseHandlers.RespondWithError(w, userIDErr)
		return
	}

	role, roleErr := contextUtils.GetUserRole(r.Context())
	if roleErr != nil {
		responseHandlers.RespondWithError(w, roleErr)
		return
	}

	// Check authorization: user can only change their own email, or admin can change any
	if authUserID != targetUserID && !isAdmin(role) {
		responseHandlers.RespondWithError(w, errLib.New("Not authorized to change this user's email", http.StatusForbidden))
		return
	}

	// Parse request body
	var requestDto dto.InitiateEmailChangeRequest
	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	newEmail := strings.TrimSpace(strings.ToLower(requestDto.NewEmail))
	if newEmail == "" {
		responseHandlers.RespondWithError(w, errLib.New("New email is required", http.StatusBadRequest))
		return
	}

	log.Printf("Initiating email change for user %s to %s", targetUserID, newEmail)

	// Initiate the email change
	token, err := h.EmailChangeService.InitiateEmailChange(r.Context(), targetUserID, newEmail)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Get user's first name for the email (this also generates a new token)
	firstName, pendingEmail, newToken, resendErr := h.EmailChangeService.ResendEmailChangeVerification(r.Context(), targetUserID)
	if resendErr != nil {
		// Don't fail the request, just log it - use the original token
		log.Printf("Warning: Could not get user info for email, using original token: %v", resendErr)
		firstName = "User"
		pendingEmail = newEmail
		newToken = token
	}

	// Generate verification URL and send email (use newToken since ResendEmailChangeVerification overwrites the original)
	verificationURL := h.EmailChangeService.GetEmailChangeVerificationURL(newToken, h.FrontendBaseURL)
	if emailErr := email.SendEmailChangeVerification(pendingEmail, firstName, pendingEmail, verificationURL); emailErr != nil {
		log.Printf("Failed to send email change verification to %s: %v", pendingEmail, emailErr)
		responseHandlers.RespondWithError(w, emailErr)
		return
	}

	response := dto.EmailChangeResponse{
		Message:      "Verification email sent to your new email address. Please check your inbox and click the link to confirm the change.",
		PendingEmail: newEmail,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// VerifyEmailChange verifies a token and completes the email change
// @Summary Verify and complete email change
// @Description Verifies the token sent to the new email and completes the email change
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body dto.VerifyEmailChangeRequest true "Verification token"
// @Success 200 {object} map[string]interface{} "Email changed successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid or missing token"
// @Failure 410 {object} map[string]interface{} "Gone: Token expired"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/email/verify [post]
func (h *EmailChangeHandler) VerifyEmailChange(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.VerifyEmailChangeRequest

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	token := strings.TrimSpace(requestDto.Token)
	if token == "" {
		responseHandlers.RespondWithError(w, errLib.New("Token is required", http.StatusBadRequest))
		return
	}

	log.Printf("Processing email change verification for token: %s...", token[:10])

	if err := h.EmailChangeService.VerifyAndCompleteEmailChange(r.Context(), token); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"message": "Email changed successfully. You can now log in with your new email address.",
		"success": true,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// ResendEmailChangeVerification resends the verification email to the pending email address
// @Summary Resend email change verification
// @Description Generates a new token and resends the verification email to the pending email address
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} dto.EmailChangeResponse "Verification email resent"
// @Failure 400 {object} map[string]interface{} "Bad Request: No pending email change"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found: User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id}/email/resend [post]
func (h *EmailChangeHandler) ResendEmailChangeVerification(w http.ResponseWriter, r *http.Request) {
	// Get target user ID from URL
	userIDStr := chi.URLParam(r, "id")
	targetUserID, parseErr := uuid.Parse(userIDStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid user ID", http.StatusBadRequest))
		return
	}

	// Get authenticated user from context
	authUserID, userIDErr := contextUtils.GetUserID(r.Context())
	if userIDErr != nil {
		responseHandlers.RespondWithError(w, userIDErr)
		return
	}

	role, roleErr := contextUtils.GetUserRole(r.Context())
	if roleErr != nil {
		responseHandlers.RespondWithError(w, roleErr)
		return
	}

	// Check authorization
	if authUserID != targetUserID && !isAdmin(role) {
		responseHandlers.RespondWithError(w, errLib.New("Not authorized", http.StatusForbidden))
		return
	}

	log.Printf("Resending email change verification for user %s", targetUserID)

	// Resend verification
	firstName, pendingEmail, token, err := h.EmailChangeService.ResendEmailChangeVerification(r.Context(), targetUserID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Generate verification URL and send email
	verificationURL := h.EmailChangeService.GetEmailChangeVerificationURL(token, h.FrontendBaseURL)
	if emailErr := email.SendEmailChangeVerification(pendingEmail, firstName, pendingEmail, verificationURL); emailErr != nil {
		log.Printf("Failed to send email change verification to %s: %v", pendingEmail, emailErr)
		responseHandlers.RespondWithError(w, emailErr)
		return
	}

	response := dto.EmailChangeResponse{
		Message:      "Verification email resent. Please check your inbox.",
		PendingEmail: pendingEmail,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// CancelEmailChange cancels any pending email change for the user
// @Summary Cancel pending email change
// @Description Cancels any pending email change request
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Email change cancelled"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found: User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id}/email/cancel [delete]
func (h *EmailChangeHandler) CancelEmailChange(w http.ResponseWriter, r *http.Request) {
	// Get target user ID from URL
	userIDStr := chi.URLParam(r, "id")
	targetUserID, parseErr := uuid.Parse(userIDStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid user ID", http.StatusBadRequest))
		return
	}

	// Get authenticated user from context
	authUserID, userIDErr := contextUtils.GetUserID(r.Context())
	if userIDErr != nil {
		responseHandlers.RespondWithError(w, userIDErr)
		return
	}

	role, roleErr := contextUtils.GetUserRole(r.Context())
	if roleErr != nil {
		responseHandlers.RespondWithError(w, roleErr)
		return
	}

	// Check authorization
	if authUserID != targetUserID && !isAdmin(role) {
		responseHandlers.RespondWithError(w, errLib.New("Not authorized", http.StatusForbidden))
		return
	}

	log.Printf("Cancelling email change for user %s", targetUserID)

	if err := h.EmailChangeService.CancelPendingEmailChange(r.Context(), targetUserID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"message": "Pending email change cancelled.",
		"success": true,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// isAdmin checks if the role is an admin role
func isAdmin(role contextUtils.CtxRole) bool {
	return role == contextUtils.RoleAdmin ||
		role == contextUtils.RoleSuperAdmin ||
		role == contextUtils.RoleIT
}
