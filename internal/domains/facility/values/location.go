package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"
)

type FacilityLocation struct {
	value string
}

func NewFacilityLocation(location string) (*FacilityLocation, *errLib.CommonError) {
	location = strings.TrimSpace(location)
	if location == "" {
		return nil, errLib.New("'location' cannot be empty or whitespace", http.StatusBadRequest)
	}
	return &FacilityLocation{value: location}, nil
}

func (c FacilityLocation) String() string {
	return c.value
}
