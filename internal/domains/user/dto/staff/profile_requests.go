package staff

import (
	values "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

// StaffProfileUpdateRequestDto represents the request dto for updating staff profile information.
type StaffProfileUpdateRequestDto struct {
	PhotoURL *string `json:"photo_url,omitempty"`
}

// ToUpdateValue converts the staff profile update request DTO to domain values.
func (dto StaffProfileUpdateRequestDto) ToUpdateValue(idStr string) (values.UpdateStaffProfileValues, *errLib.CommonError) {
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		return values.UpdateStaffProfileValues{}, err
	}

	return values.UpdateStaffProfileValues{
		ID:       id,
		PhotoURL: dto.PhotoURL,
	}, nil
}