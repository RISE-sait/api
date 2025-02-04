package validators

import (
	errLib "api/internal/libs/errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func ParseUUID(str string) (uuid.UUID, *errLib.CommonError) {

	// Now try to parse the UUID using the uuid library
	parsedID, err := uuid.Parse(str)
	if err != nil {
		errMsg := fmt.Sprintf("invalid UUID: %s, error: %v", str, err)
		return uuid.Nil, errLib.New(errMsg, http.StatusBadRequest)
	}

	return parsedID, nil
}

func ParseDateTime(str string) (time.Time, *errLib.CommonError) {

	// Now try to parse the DateTime using the time library
	datetime, err := time.Parse(time.RFC3339, str)
	if err != nil {
		errMsg := fmt.Sprintf("invalid DateTime: %s, error: %v", str, err)
		return time.Time{}, errLib.New(errMsg, http.StatusBadRequest)
	}

	return datetime, nil
}
