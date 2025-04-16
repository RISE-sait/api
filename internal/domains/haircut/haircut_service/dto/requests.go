package haircut_service

import (
	values "api/internal/domains/haircut/haircut_service/values"
	errLib "api/internal/libs/errors"
	"github.com/google/uuid"
)

type CreateBarberServiceRequestDto struct {
	BarberID         uuid.UUID `json:"barber_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	HaircutServiceID uuid.UUID `json:"haircut_service_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
}

func (dto CreateBarberServiceRequestDto) ToCreateBarberServiceValue() (values.CreateBarberServiceValues, *errLib.CommonError) {

	return values.CreateBarberServiceValues{
		BarberServiceValuesBase: values.BarberServiceValuesBase{
			ServiceTypeID: dto.HaircutServiceID,
			BarberID:      dto.BarberID,
		},
	}, nil
}
