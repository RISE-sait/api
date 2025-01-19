package utils

import (
	"api/internal/constants"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/lib/pq"
)

func MapDatabaseError(err error) *HTTPError {

	if errors.Is(err, sql.ErrNoRows) {
		return CreateHTTPError("Resource not found", http.StatusNotFound)
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code.Class() == constants.PgBadRequestCodeClass {
			return CreateHTTPError(pqErr.Code.Name(), http.StatusBadRequest)
		}
	}

	// logging the error as the error may contain sensitive information
	log.Printf("Unhandled database error: %v", err)

	return CreateHTTPError("Internal server error", http.StatusInternalServerError)
}
