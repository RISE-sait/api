package haircut

import (
	"github.com/google/uuid"
	"time"
)

type BarberServiceValuesBase struct {
	ServiceTypeID uuid.UUID
	BarberID      uuid.UUID
}

type CreateBarberServiceValues struct {
	BarberServiceValuesBase
}

type ReadBarberServicesValues struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	BarberServiceValuesBase
	HaircutName string
	BarberName  string
}

type UpdateBarberServicesValues struct {
	ID uuid.UUID
	BarberServiceValuesBase
}
