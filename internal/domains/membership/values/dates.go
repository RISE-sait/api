package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"time"
)

type MembershipDates struct {
	StartDate time.Time
	EndDate   time.Time
}

func NewMembershipDates(startDate, endDate time.Time, startDateFieldName, endDateFieldName string) (*MembershipDates, *errLib.CommonError) {
	if startDate.IsZero() {
		return nil, errLib.New("'"+startDateFieldName+"' is required", http.StatusBadRequest)
	}
	if endDate.IsZero() {
		return nil, errLib.New("'"+endDateFieldName+"' is required", http.StatusBadRequest)
	}
	if endDate.Before(startDate) {
		return nil, errLib.New("'"+endDateFieldName+"' must be after '"+startDateFieldName+"'", http.StatusBadRequest)
	}
	return &MembershipDates{StartDate: startDate, EndDate: endDate}, nil
}
