package enrollment

import (
	"api/internal/domains/enrollment/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/google/uuid"
)

type RequestDto struct {
	CustomerId uuid.UUID `json:"customer_id" validate:"required"`
	EventId    uuid.UUID `json:"event_id" validate:"required"`
}

type CreateRequestDto struct {
	RequestDto
}

type UpdateRequestDto struct {
	RequestDto
	ID          uuid.UUID `json:"id" validate:"required"`
	IsCancelled bool      `json:"is_cancelled" validate:"required"`
}

func validate(dto *RequestDto) *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}
	return nil
}

func (dto CreateRequestDto) ToCreateValueObjects() (values.EnrollmentCreateDetails, *errLib.CommonError) {

	if err := validate(&dto.RequestDto); err != nil {
		return values.EnrollmentCreateDetails{}, err
	}

	return values.EnrollmentCreateDetails{
		EnrollmentDetails: values.EnrollmentDetails{
			CustomerId: dto.CustomerId,
			EventId:    dto.EventId,
		},
	}, nil
}

func (dto UpdateRequestDto) ToUpdateValueObjects(idStr string) (values.EnrollmentUpdateDetails, *errLib.CommonError) {

	var updateDetails values.EnrollmentUpdateDetails

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return updateDetails, err
	}

	if err = validate(&dto.RequestDto); err != nil {
		return updateDetails, err
	}

	return values.EnrollmentUpdateDetails{
		ID:          id,
		IsCancelled: dto.IsCancelled,
		EnrollmentDetails: values.EnrollmentDetails{
			CustomerId: dto.CustomerId,
			EventId:    dto.EventId,
		},
	}, nil
}
