package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"
	"time"
)

type MembershipCreate struct {
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}

func NewMembershipCreate(name, description string, startDate time.Time, endDate time.Time) *MembershipCreate {
	return &MembershipCreate{
		Name:        name,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
	}
}

func (cc MembershipCreate) Validate() *errLib.CommonError {
	cc.Name = strings.TrimSpace(cc.Name)

	// Validate name
	if cc.Name == "" {
		return errLib.New(errEmptyName, http.StatusBadRequest)
	}
	if len(cc.Name) > 100 {
		return errLib.New(errNameTooLong, http.StatusBadRequest)
	}

	// Validate dates
	now := time.Now()
	if cc.StartDate.IsZero() {
		return errLib.New(errStartDateRequired, http.StatusBadRequest)
	}
	if cc.EndDate.IsZero() {
		return errLib.New(errEndDateRequired, http.StatusBadRequest)
	}
	if cc.StartDate.Before(now) {
		return errLib.New(errPastStartDate, http.StatusBadRequest)
	}
	if cc.EndDate.Before(cc.StartDate) {
		return errLib.New(errInvalidDateRange, http.StatusBadRequest)
	}

	return nil
}
