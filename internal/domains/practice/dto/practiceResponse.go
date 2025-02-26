package dto

import (
	"github.com/google/uuid"
)

type PracticeResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
}
