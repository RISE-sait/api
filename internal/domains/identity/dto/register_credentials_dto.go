package dto

type UserInfoDto struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,omitempty"`
	LastName  string `json:"last_name" validate:"required,omitempty"`
}
