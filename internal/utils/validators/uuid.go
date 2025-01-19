package validators

import (
	"api/internal/utils"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// ParseUUID attempts to parse a UUID
func ParseUUID(str string) (uuid.UUID, *utils.HTTPError) {

	// Now try to parse the UUID using the uuid library
	parsedID, err := uuid.Parse(str)
	if err != nil {
		errMsg := fmt.Sprintf("invalid UUID: %s, error: %v", str, err)
		return uuid.Nil, utils.CreateHTTPError(errMsg, http.StatusBadRequest)
	}

	return parsedID, nil
}
