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

// ParseTime parses a time string in the format "15:04:05+00:00" (HH:MM:SS+00:00) and returns it in the same format.
// If the input time string is invalid or does not match the expected format, an error is returned.
//
// @param timeStr The time string to parse, in the format "15:04:05+00:00" (hours:minutes:seconds timezone).
func ParseTime(timeStr string) (string, *errLib.CommonError) {
	// Define the expected time format
	expectedFormat := "15:04:05+00:00"

	// Parse the time string with the expected format
	timeParsed, err := time.Parse("15:04:05-07:00", timeStr)
	if err != nil {
		errMsg := fmt.Sprintf("Invalid time format. Expected format (%s), got: %s", expectedFormat, timeStr)
		return "", errLib.New(errMsg, http.StatusBadRequest)
	}

	// Return the time in the same format (15:04:05+00:00)
	return timeParsed.Format(expectedFormat), nil
}
