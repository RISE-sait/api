package validators

import (
	errLib "api/internal/libs/errors"
	"fmt"
	"net/http"
	"time"
)

func ParseDateTime(str string) (time.Time, *errLib.CommonError) {

	// Now try to parse the DateTime using the time library
	datetime, err := time.Parse(time.RFC3339, str)
	if err != nil {
		errMsg := fmt.Sprintf("Invalid DateTime format. Expected RFC3339 (YYYY-MM-DDTHH:MM:SSZ), got: %s", str)
		return time.Time{}, errLib.New(errMsg, http.StatusBadRequest)
	}

	return datetime, nil
}
