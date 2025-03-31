package program

import (
	"time"

	"github.com/google/uuid"
)

type Response struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Level       string    `json:"level"`
	Type        string    `json:"type"`
	Capacity    *int32    `json:"capacity,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LevelsResponse struct {
	Name []string `json:"levels"`
}
