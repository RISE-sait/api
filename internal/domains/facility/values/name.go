package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"
)

type FacilityName struct {
	value string
}

func NewFacilityName(name string) (*FacilityLocation, *errLib.CommonError) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errLib.New("'name' cannot be empty or whitespace", http.StatusBadRequest)
	}
	return &FacilityLocation{value: name}, nil
}

func (c FacilityName) String() string {
	return c.value
}
