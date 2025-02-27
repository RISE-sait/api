package barber

import (
	values "api/internal/domains/barber/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"log"
	"time"
)

type RequestDto struct {
	BeginDateTime string    `json:"begin_time" validate:"required" example:"2023-10-05T07:00:00Z"`
	EndDateTime   string    `json:"end_time" validate:"required" example:"2023-10-05T07:00:00Z"`
	BarberID      uuid.UUID `json:"barber_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	CustomerID    uuid.UUID `json:"customer_id" example:"00000000-0000-0000-0000-000000000000"`
}

func (dto RequestDto) validate() (time.Time, time.Time, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return time.Time{}, time.Time{}, err
	}

	beginDateTime, err := validators.ParseDateTime(dto.BeginDateTime)

	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endTime, err := validators.ParseDateTime(dto.EndDateTime)

	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return beginDateTime, endTime, nil
}

func (dto RequestDto) ToCreateEventValue() (values.CreateEventValues, *errLib.CommonError) {

	beginTime, endTime, err := dto.validate()

	if err != nil {

		return values.CreateEventValues{}, err
	}

	return values.CreateEventValues{
		EventValuesBase: values.EventValuesBase{
			BeginDateTime: beginTime,
			EndDateTime:   endTime,
			BarberID:      dto.BarberID,
			CustomerID:    dto.CustomerID,
		},
	}, nil
}

func (dto RequestDto) ToUpdateEventValue(idStr string) (values.UpdateEventValues, *errLib.CommonError) {

	var updatedEvent values.UpdateEventValues

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return updatedEvent, err
	}

	beginTime, endTime, err := dto.validate()

	if err != nil {

		log.Println("Error: ", err)
		return updatedEvent, err
	}

	return values.UpdateEventValues{
		ID: id,
		EventValuesBase: values.EventValuesBase{
			BeginDateTime: beginTime,
			EndDateTime:   endTime,
			BarberID:      dto.BarberID,
			CustomerID:    dto.CustomerID,
		},
	}, nil
}
