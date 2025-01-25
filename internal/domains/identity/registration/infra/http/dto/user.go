package dto

import (
	"github.com/google/uuid"
)

type UpdateUserEmailRequest struct {
	Id    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}
