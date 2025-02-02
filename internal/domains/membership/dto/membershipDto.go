package dto

import (
	"api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"time"
)

type MembershipRequestDto struct {
	Name        string    `json:"name" validate:"notwhitespace"`
	Description string    `json:"description" validate:"omitempty,notwhitespace"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required,gtcsfield=StartDate"`
}

func (dto *MembershipRequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}
	return nil
}

func (dto *MembershipRequestDto) ToMembershipCreateValueObject() (*values.MembershipCreate, *errLib.CommonError) {

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.MembershipCreate{
		Membership: values.Membership{
			Name:        dto.Name,
			Description: dto.Description,
			StartDate:   dto.StartDate,
			EndDate:     dto.EndDate,
		},
	}, nil
}

func (dto *MembershipRequestDto) ToMembershipUpdateValueObject(idStr string) (*values.MembershipUpdate, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.MembershipUpdate{
		ID: id,
		Membership: values.Membership{
			Name:        dto.Name,
			Description: dto.Description,
			StartDate:   dto.StartDate,
			EndDate:     dto.EndDate,
		},
	}, nil
}
