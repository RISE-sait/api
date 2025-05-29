package dto

type ContactRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
	Phone   string `json:"phone" validate:"required"`
	Message string `json:"message" validate:"required"`
	Token   string `json:"token" validate:"required"` // reCAPTCHA token
}
