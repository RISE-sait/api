package dto

type InitiateEmailChangeRequest struct {
	NewEmail string `json:"new_email" validate:"required,email"`
}

type VerifyEmailChangeRequest struct {
	Token string `json:"token" validate:"required"`
}

type EmailChangeResponse struct {
	Message      string `json:"message"`
	PendingEmail string `json:"pending_email,omitempty"`
}
