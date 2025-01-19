package user

import (
	db "api/sqlc"

	"github.com/google/uuid"
)

type UpdateUserEmailRequest struct {
	Id    uuid.UUID `json:"id" validate:"required"`
	Email string    `json:"email" validate:"email"`
}

func (r *UpdateUserEmailRequest) ToDBParams() *db.UpdateUserEmailParams {

	return &db.UpdateUserEmailParams{
		ID:    r.Id,
		Email: r.Email,
	}
}
