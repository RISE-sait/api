package team

import (
	"github.com/google/uuid"
	"time"
)

type Response struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Capacity  int32     `json:"capacity"`
	CoachID   uuid.UUID `json:"coach_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
