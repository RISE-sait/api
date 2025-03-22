package haircut

import (
	values "api/internal/domains/haircut/values"
	"github.com/google/uuid"
	"time"
)

type BarberServiceResponseDto struct {
	ID            uuid.UUID `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ServiceTypeID uuid.UUID `json:"service_type_id"`
	BarberID      uuid.UUID `json:"barber_id"`
	HaircutName   string    `json:"haircut_name"`
	BarberName    string    `json:"barber_name"`
}

func NewServiceResponseDto(service values.ReadBarberServicesValues) BarberServiceResponseDto {
	return BarberServiceResponseDto{
		ID:            service.ID,
		CreatedAt:     service.CreatedAt,
		UpdatedAt:     service.UpdatedAt,
		ServiceTypeID: service.ServiceTypeID,
		BarberID:      service.BarberID,
		HaircutName:   service.HaircutName,
		BarberName:    service.BarberName,
	}
}
