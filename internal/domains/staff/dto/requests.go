package staff

import (
	entity "api/internal/domains/staff/entity"
	values "api/internal/domains/staff/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

// RequestDto represents the request dto for creating and updating a staff member.
type RequestDto struct {
	IsActive bool   `json:"is_active" validate:"required"`
	RoleName string `json:"role_name" validate:"required"`
}

// ToDetails converts the CreateStaffRequest DTO to the domain value object.
func (dto *RequestDto) ToDetails() (*values.Details, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	return &values.Details{
		IsActive: dto.IsActive,
		RoleName: dto.RoleName,
	}, nil
}

// ToEntity converts the UpdateStaffRequest DTO to the domain entity.
func (dto *RequestDto) ToEntity(idStr string) (*entity.Staff, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	return &entity.Staff{
		ID: id,
		Details: values.Details{
			IsActive: dto.IsActive,
			RoleName: dto.RoleName,
		},
	}, nil
}
