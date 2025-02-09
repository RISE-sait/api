package dto

type RegisterCredentialsDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"omitempty,min=8"`
}
