package barber

import (
	values "api/internal/domains/barber/values"
	"github.com/google/uuid"
	"time"
)

type ResponseDto struct {
	ID            uuid.UUID `json:"id"`
	BeginDateTime string    `json:"begin_time"`
	EndDateTime   string    `json:"end_time"`
	BarberID      uuid.UUID `json:"barber_id,omitempty"`
	CustomerID    uuid.UUID `json:"customer_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func NewEventResponse(event values.EventReadValues) ResponseDto {
	return ResponseDto{
		ID:            event.ID,
		BeginDateTime: event.BeginDateTime.String(),
		EndDateTime:   event.EndDateTime.String(),
		BarberID:      event.BarberID,
		CustomerID:    event.CustomerID,
		CreatedAt:     event.CreatedAt,
		UpdatedAt:     event.UpdatedAt,
	}
}
