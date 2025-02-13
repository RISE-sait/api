package dto

import (
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type CreatePendingChildAccountDto struct {
	Child       *CustomerRegistrationDto `json:"child" validate:"required,structonly"`
	ParentEmail string                   `json:"parent_email" validate:"required,email"`
}

func (dto *CreatePendingChildAccountDto) ToValueObjects() (*values.CreatePendingChildAccountValueObject, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {

		return nil, err
	}

	childVo, err := dto.Child.ToValueObjects()

	if err != nil {
		return nil, err
	}

	pendingChildAccountCreateVo := values.CreatePendingChildAccountValueObject{
		CustomerRegistrationInfo: values.CustomerRegistrationInfo{
			UserInfo: values.UserInfo{
				Email: dto.Child.Email,
			},
		},
		ParentEmail: dto.ParentEmail,
		Waivers:     childVo.Waivers,
	}

	if dto.Child.FirstName != "" {
		pendingChildAccountCreateVo.FirstName = &dto.Child.FirstName
	}

	if dto.Child.LastName != "" {
		pendingChildAccountCreateVo.LastName = &dto.Child.LastName
	}

	if *(childVo.Password) != "" {
		pendingChildAccountCreateVo.Password = childVo.Password
	}

	return &pendingChildAccountCreateVo, nil
}
