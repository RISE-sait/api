package dto

import (
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type CreatePendingChildAccountDto struct {
	Child       CustomerRegistrationDto `json:"child" validate:"required,structonly"`
	ParentEmail string                  `json:"parent_email" validate:"required,email"`
}

func (dto *CreatePendingChildAccountDto) ToValueObjects() (*values.CreatePendingChildAccountValueObject, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	if err := validators.ValidateDto(&dto.Child); err != nil {
		return nil, err
	}

	child := dto.Child

	var waiversVo []values.CustomerWaiverSigning
	for _, waiver := range child.CustomerWaiverSigningDtos {
		vo, err := waiver.ToValueObjects()
		if err != nil {
			return nil, err
		}
		waiversVo = append(waiversVo, *vo)
	}

	pendingChildAccountCreateVo := values.CreatePendingChildAccountValueObject{
		ChildEmail:  child.Email,
		ParentEmail: dto.ParentEmail,
		Waivers:     waiversVo,
	}

	if child.Password != "" {
		pendingChildAccountCreateVo.Password = &child.Password
	}

	return &pendingChildAccountCreateVo, nil
}
