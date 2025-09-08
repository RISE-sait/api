package haircut_event

import (
	values "api/internal/domains/haircut/event"
	"github.com/google/uuid"
	"time"
)

type EventResponseDto struct {
	ID            uuid.UUID `json:"id"`
	BeginDateTime time.Time `json:"start_at"`
	EndDateTime   time.Time `json:"end_at"`
	BarberID      uuid.UUID `json:"barber_id"`
	BarberName    string    `json:"barber_name"`
	CustomerName  string    `json:"customer_name"`
	CustomerID    uuid.UUID `json:"customer_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func NewEventResponse(event values.EventReadValues) EventResponseDto {
	return EventResponseDto{
		ID:            event.ID,
		BeginDateTime: event.BeginDateTime,
		EndDateTime:   event.EndDateTime,
		BarberID:      event.BarberID,
		BarberName:    event.BarberName,
		CustomerName:  event.CustomerName,
		CustomerID:    event.CustomerID,
		CreatedAt:     event.CreatedAt,
		UpdatedAt:     event.UpdatedAt,
	}
}
