package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MembershipUpdate struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}

func NewMembershipUpdate(id uuid.UUID, name, description string, startDate, endDate time.Time) *MembershipUpdate {
	return &MembershipUpdate{
		ID:          id,
		Name:        strings.TrimSpace(name),
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
	}
}

func (mu MembershipUpdate) Validate() *errLib.CommonError {
	if mu.ID == uuid.Nil {
		return errLib.New(errEmptyName, http.StatusBadRequest)
	}

	trimmedName := strings.TrimSpace(mu.Name)
	if trimmedName == "" {
		return errLib.New(errEmptyName, http.StatusBadRequest)
	}
	if len(trimmedName) > 100 {
		return errLib.New(errNameTooLong, http.StatusBadRequest)
	}

	trimmedDescription := strings.TrimSpace(mu.Description)
	if trimmedDescription == "" {
		return errLib.New(errEmptyDescription, http.StatusBadRequest)
	}
	if len(trimmedDescription) > 100 {
		return errLib.New(errDescriptionTooLong, http.StatusBadRequest)
	}

	if mu.EndDate.Before(mu.StartDate) {
		return errLib.New(errInvalidDateRange, http.StatusBadRequest)

	}

	return nil
}
