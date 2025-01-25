package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CourseUpdate struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}

func NewCourseUpdate(id uuid.UUID, name, description string, startDate, endDate time.Time) *CourseUpdate {
	return &CourseUpdate{
		ID:          id,
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		StartDate:   startDate,
		EndDate:     endDate,
	}
}

func (cu *CourseUpdate) Validate() *errLib.CommonError {
	// Validate ID
	if cu.ID == uuid.Nil {
		return errLib.New("invalid course ID", http.StatusBadRequest)
	}

	// Validate Name
	if cu.Name == "" {
		return errLib.New("name cannot be empty or whitespace", http.StatusBadRequest)
	}
	if len(cu.Name) > 100 {
		return errLib.New("name cannot exceed 100 characters", http.StatusBadRequest)
	}

	// Validate Description
	if cu.Description == "" {
		return errLib.New("description cannot be empty or whitespace", http.StatusBadRequest)
	}
	if len(cu.Description) > 100 {
		return errLib.New("description cannot exceed 100 characters", http.StatusBadRequest)
	}

	// Validate Dates
	if cu.StartDate.IsZero() {
		return errLib.New("start date is required", http.StatusBadRequest)
	}
	if cu.EndDate.IsZero() {
		return errLib.New("end date is required", http.StatusBadRequest)
	}
	if cu.EndDate.Before(cu.StartDate) {
		return errLib.New("end date cannot be before start date", http.StatusBadRequest)
	}

	return nil
}
