package errLib

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

const (
	PGEnumInvalidInput = "22P02" // Postgres error code for invalid enum
)

func TranslateDBErrorToCommonError(err error) *CommonError {
	if errors.Is(err, sql.ErrNoRows) {
		return New("Resource not found", http.StatusNotFound)
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code.Class() == "23" {
			return New(pqErr.Code.Name(), http.StatusBadRequest)
		}

		if string(pqErr.Code) == PGEnumInvalidInput {
			if strings.Contains(pqErr.Message, "enum") {
				return New("Invalid enum value provided. Error: "+pqErr.Message, http.StatusBadRequest)
			}
		}
	}

	// logging the error as the error may contain sensitive information
	log.Printf("Unhandled database error: %v", err)

	return New("Internal server error", http.StatusInternalServerError)
}
