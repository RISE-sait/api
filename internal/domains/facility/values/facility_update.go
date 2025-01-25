package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type FacilityUpdate struct {
	ID             uuid.UUID
	Name           string
	Location       string
	FacilityTypeID uuid.UUID
}

func NewFacilityUpdate(id uuid.UUID, name, location string, facilityTypeID uuid.UUID) *FacilityUpdate {
	return &FacilityUpdate{
		ID:             id,
		Name:           name,
		Location:       location,
		FacilityTypeID: facilityTypeID,
	}
}

func (fu FacilityUpdate) Validate() *errLib.CommonError {
	if fu.ID == uuid.Nil {
		return errLib.New(errEmptyFacilityId, http.StatusBadRequest)
	}

	trimmedName := strings.TrimSpace(fu.Name)
	if trimmedName == "" {
		return errLib.New(errEmptyName, http.StatusBadRequest)
	}
	if len(trimmedName) > 100 {
		return errLib.New(errNameTooLong, http.StatusBadRequest)

	}

	// Validate location
	trimmedLocation := strings.TrimSpace(fu.Location)

	if trimmedLocation == "" {
		return errLib.New(errEmptyName, http.StatusBadRequest)
	}

	if len(trimmedLocation) > 100 {
		return errLib.New(errLocationTooLong, http.StatusBadRequest)

	}

	// Validate facility type ID
	if fu.FacilityTypeID == uuid.Nil {
		return errLib.New(errFacilityTypeEmpty, http.StatusBadRequest)
	}

	return nil
}
