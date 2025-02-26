package dto

import (
	"github.com/google/uuid"
	"time"
)

type EnrollmentResponse struct {
	ID          uuid.UUID `json:"id"`
	CustomerID  uuid.UUID `json:"customer_id"`
	EventID     uuid.UUID `json:"event_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CheckedInAt time.Time `json:"checked_in_at"`
	IsCancelled bool      `json:"is_cancelled"`
}
