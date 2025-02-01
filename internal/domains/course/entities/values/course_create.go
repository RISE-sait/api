package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"
	"time"
)

type CourseCreate struct {
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}

func NewCourseCreate(name, description string, startDate, endDate time.Time) *CourseCreate {
	return &CourseCreate{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		StartDate:   startDate,
		EndDate:     endDate,
	}
}

func (cc *CourseCreate) Validate() *errLib.CommonError {
	// Validate name
	if cc.Name == "" {
		return errLib.New(errEmptyName, http.StatusBadRequest)
	}
	if len(cc.Name) > 100 {
		return errLib.New(errNameTooLong, http.StatusBadRequest)
	}

	// Validate description
	if cc.Description == "" {
		return errLib.New(errEmptyDescription, http.StatusBadRequest)
	}
	if len(cc.Description) > 100 {
		return errLib.New(errDescriptionTooLong, http.StatusBadRequest)
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
