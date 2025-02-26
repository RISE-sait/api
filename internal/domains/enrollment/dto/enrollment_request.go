package dto

import (
	"api/internal/domains/enrollment/entity"
	"api/internal/domains/enrollment/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/google/uuid"
)

type EnrollmentRequestDto struct {
	CustomerId uuid.UUID `json:"customer_id" validate:"required"`
	EventId    uuid.UUID `json:"event_id" validate:"required"`
}

func (dto *EnrollmentRequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}
	return nil
}

func (dto *EnrollmentRequestDto) ToCreateValueObjects() (*values.EnrollmentDetails, *errLib.CommonError) {

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.EnrollmentDetails{
		CustomerId: dto.CustomerId,
		EventId:    dto.EventId,
	}, nil
}

func (dto *EnrollmentRequestDto) ToUpdateValueObjects(idStr string) (*entity.Enrollment, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &entity.Enrollment{
		ID:         id,
		CustomerID: dto.CustomerId,
		EventID:    dto.EventId,
	}, nil
}
