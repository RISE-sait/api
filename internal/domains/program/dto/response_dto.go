package program

import (
	"time"

	"github.com/google/uuid"
)

type Response struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"type"`
	Capacity    *int32    `json:"capacity,omitempty"`
	PhotoURL    *string   `json:"photo_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
