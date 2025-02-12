package membership

import (
	values "api/internal/domains/membership/values/memberships"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type MembershipRequestDto struct {
	Name        string `json:"name" validate:"notwhitespace" example:"Premium Membership"`
	Description string `json:"description" validate:"omitempty,notwhitespace" example:"Access to all premium features"`
}

func (dto *MembershipRequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}
	return nil
}

func (dto *MembershipRequestDto) ToMembershipCreateValueObject() (*values.MembershipDetails, *errLib.CommonError) {

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.MembershipDetails{
		Name:        dto.Name,
		Description: dto.Description,
	}, nil
}

func (dto *MembershipRequestDto) ToMembershipUpdateValueObject(idStr string) (*values.MembershipAllFields, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.MembershipAllFields{
		ID: id,
		MembershipDetails: values.MembershipDetails{
			Name:        dto.Name,
			Description: dto.Description,
		},
	}, nil
}
