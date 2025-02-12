package staff

import (
	"api/internal/domains/staff/values"
	errLib "api/internal/libs/errors"
	"net/http"
)

type UpdateStaffDto struct {
	Role          string `json:"role"`
	IsActiveStaff bool   `json:"is_active_staff"`
}

func (sc *UpdateStaffDto) ToUpdateStaffValueObject() (*values.StaffDetails, *errLib.CommonError) {
	if len(sc.Role) > 100 {
		return nil, errLib.New("Role is too long", http.StatusBadRequest)
	}

	return &values.StaffDetails{
		RoleName: sc.Role,
		IsActive: sc.IsActiveStaff,
	}, nil
}
