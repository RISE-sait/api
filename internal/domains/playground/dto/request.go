package playground

import (
	values "api/internal/domains/playground/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"time"

	"github.com/google/uuid"
)
// RequestDto represents the data transfer object for creating a new session.
type RequestDto struct {
	SystemID  uuid.UUID `json:"system_id" validate:"required"`
	StartTime string    `json:"start_time" validate:"required"`
	EndTime   string    `json:"end_time" validate:"required"`
}
// Validate validates the RequestDto using the validators package.
func (dto RequestDto) toTimes() (time.Time, time.Time, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return time.Time{}, time.Time{}, err
	}
	start, err := validators.ParseDateTime(dto.StartTime)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := validators.ParseDateTime(dto.EndTime)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}
// ToCreateValue converts the RequestDto to a CreateSessionValue, which is used to create a new session.
func (dto RequestDto) ToCreateValue(customerID uuid.UUID) (values.CreateSessionValue, *errLib.CommonError) {
	start, end, err := dto.toTimes()
	if err != nil {
		return values.CreateSessionValue{}, err
	}
	return values.CreateSessionValue{
		SystemID:   dto.SystemID,
		CustomerID: customerID,
		StartTime:  start,
		EndTime:    end,
	}, nil
}
