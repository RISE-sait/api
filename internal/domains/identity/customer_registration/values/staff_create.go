package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"
)

type StaffCreate struct {
	Role     string
	IsActive bool
}

func NewStaffCreate(role string, isActive bool) *StaffCreate {
	return &StaffCreate{
		Role:     strings.TrimSpace(role),
		IsActive: isActive,
	}
}

func (sc *StaffCreate) Validate() *errLib.CommonError {
	if len(sc.Role) > 100 {
		return errLib.New("Role is too long", http.StatusBadRequest)
	}

	return nil
}
