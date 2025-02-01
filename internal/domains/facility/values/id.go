package values

import (
	errLib "api/internal/libs/errors"
	"net/http"

	"github.com/google/uuid"
)

func NewId(id uuid.UUID, fieldName string) (uuid.UUID, *errLib.CommonError) {
	if id == uuid.Nil {
		return uuid.Nil, errLib.New("'"+fieldName+"' cannot be empty", http.StatusBadRequest)
	}
	return id, nil
}
