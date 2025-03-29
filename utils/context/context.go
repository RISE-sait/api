package context

import (
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
	"net/http"
)

// Key is a custom type for context keys to avoid collisions.
type Key string
type CtxRole string

const (
	UserIDKey Key = "userId"
	RoleKey   Key = "role"
)

const (
	RoleSuperAdmin CtxRole = "SUPERADMIN"
	RoleAdmin      CtxRole = "ADMIN"
	RoleInstructor CtxRole = "INSTRUCTOR"
	RoleParent     CtxRole = "PARENT"
	RoleChild      CtxRole = "CHILD"
	RoleAthlete    CtxRole = "ATHLETE"
	RoleCoach      CtxRole = "COACH"
	RoleBarber     CtxRole = "BARBER"
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

func GetUserRole(ctx context.Context) (CtxRole, *errLib.CommonError) {

	if ctx == nil {
		return "", errLib.New("context cannot be nil", http.StatusBadRequest)
	}

	userRole, ok := ctx.Value(RoleKey).(CtxRole)

	if !ok || userRole == "" {
		return "", errLib.New("user ID not found in context", http.StatusUnauthorized)
	}

	return userRole, nil
}
