package haircut

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type ServiceValuesBase struct {
	Name          string
	Description   *string
	Price         decimal.Decimal
	DurationInMin int32
}

type CreateHaircutServiceValues struct {
	ServiceValuesBase
}

type ReadHaircutServicesValues struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	ServiceValuesBase
}

type UpdateHaircutServicesValues struct {
	ID uuid.UUID
	ServiceValuesBase
}
