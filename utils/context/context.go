package context

import (
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
	"net/http"
)

// Key is a custom type for context keys to avoid collisions.
type Key string

const (
	UserIDKey Key = "userId"
)

func GetUserID(ctx context.Context) (uuid.UUID, *errLib.CommonError) {

	if ctx == nil {
		return uuid.Nil, errLib.New("context cannot be nil", http.StatusBadRequest)
	}

	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)

	if !ok || userID == uuid.Nil {
		return uuid.Nil, errLib.New("user ID not found in context", http.StatusUnauthorized)
	}

	return userID, nil
}
