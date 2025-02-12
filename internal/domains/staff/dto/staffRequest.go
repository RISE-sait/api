package dto

import (
	"api/internal/domains/staff/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"time"

	"github.com/google/uuid"
)

type StaffRequestDto struct {
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RoleID    uuid.UUID `json:"role_id"`
	RoleName  string    `json:"role_name"`
}

func (dto *StaffRequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}
	return nil
}

func (dto *StaffRequestDto) ToUpdateValueObjects(idStr string) (*values.StaffAllFields, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.StaffAllFields{
		ID: id,
		StaffDetails: values.StaffDetails{
			IsActive:  dto.IsActive,
			CreatedAt: dto.CreatedAt,
			UpdatedAt: dto.UpdatedAt,
			RoleID:    dto.RoleID,
			RoleName:  dto.RoleName,
		},
	}, nil
}
