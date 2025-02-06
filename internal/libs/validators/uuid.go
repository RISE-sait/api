package validators

import (
	errLib "api/internal/libs/errors"
	"fmt"
	"net/http"

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
