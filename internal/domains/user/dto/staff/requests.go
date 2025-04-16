package staff

import (
	values "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

// RequestDto represents the request dto for creating and updating a staff member.
type RequestDto struct {
	IsActive *bool  `json:"is_active" validate:"required"`
	RoleName string `json:"role_name" validate:"required"`
}

// ToUpdateRequestValues converts the UpdateStaffRequest DTO to the domain entity.
// @Description Converts the request DTO to domain values for updating staff details.
// @Param idStr path string true "The UUID of the staff member to update"
// @Param dto body RequestDto true "Request body containing updated staff details"
// @Return values.UpdateStaffValues "Converted domain values for updating a staff member"
// @Return *errLib.CommonError "Validation or processing error"
func (dto RequestDto) ToUpdateRequestValues(idStr string) (values.UpdateStaffValues, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return values.UpdateStaffValues{}, err
	}

	if err = validators.ValidateDto(&dto); err != nil {
		return values.UpdateStaffValues{}, err
	}

	return values.UpdateStaffValues{
		ID:       id,
		IsActive: *dto.IsActive,
		RoleName: dto.RoleName,
	}, nil
}
