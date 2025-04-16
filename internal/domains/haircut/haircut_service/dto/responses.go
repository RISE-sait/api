package haircut_service

import (
	values "api/internal/domains/haircut/haircut_service/values"
	"github.com/google/uuid"
	"time"
)

type BarberServiceResponseDto struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	HaircutID   uuid.UUID `json:"haircut_id"`
	BarberID    uuid.UUID `json:"barber_id"`
	HaircutName string    `json:"haircut_name"`
	BarberName  string    `json:"barber_name"`
}

func NewServiceResponseDto(service values.ReadBarberServicesValues) BarberServiceResponseDto {
	return BarberServiceResponseDto{
		ID:          service.ID,
		CreatedAt:   service.CreatedAt,
		UpdatedAt:   service.UpdatedAt,
		HaircutID:   service.ServiceTypeID,
		BarberID:    service.BarberID,
		HaircutName: service.HaircutName,
		BarberName:  service.BarberName,
	}
}
