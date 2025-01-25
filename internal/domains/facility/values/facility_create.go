package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type FacilityCreate struct {
	Name           string
	Location       string
	FacilityTypeID uuid.UUID
}

func NewFacilityCreate(name, location string, facilityTypeID uuid.UUID) *FacilityCreate {
	return &FacilityCreate{
		Name:           strings.TrimSpace(name),
		Location:       strings.TrimSpace(location),
		FacilityTypeID: facilityTypeID,
	}
}

func (cc FacilityCreate) Validate() *errLib.CommonError {
	cc.Name = strings.TrimSpace(cc.Name)

	// Validate name
	if cc.Name == "" {
		return errLib.New(errEmptyName, http.StatusBadRequest)
	}
	if len(cc.Name) > 100 {
		return errLib.New(errNameTooLong, http.StatusBadRequest)
	}

	// Validate location
	if cc.Location == "" {
		return errLib.New(errEmptyName, http.StatusBadRequest)
	}

	if len(cc.Location) > 100 {
		return errLib.New(errLocationTooLong, http.StatusBadRequest)
	}

	// Validate facility type ID
	if cc.FacilityTypeID == uuid.Nil {
		return errLib.New(errFacilityTypeEmpty, http.StatusBadRequest)
	}

	return nil
}
