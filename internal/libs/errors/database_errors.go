package errLib

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"log"
	"net/http"
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
	}

	// logging the error as the error may contain sensitive information
	log.Printf("Unhandled database error: %v", err)

	return New("Internal server error", http.StatusInternalServerError)
}
