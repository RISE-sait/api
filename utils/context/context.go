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
	RoleSuperAdmin   CtxRole = "SUPERADMIN"
	RoleIT           CtxRole = "IT"
	RoleAdmin        CtxRole = "ADMIN"
	RoleInstructor   CtxRole = "INSTRUCTOR"
	RoleAthlete      CtxRole = "ATHLETE"
	RoleCoach        CtxRole = "COACH"
	RoleBarber       CtxRole = "BARBER"
	RoleReceptionist CtxRole = "RECEPTIONIST"
)

// GetUserID retrieves the user ID from the context. Returns an error if the context is nil
// or the user ID is not found or invalid.
//
// Returns:
//   - uuid.UUID: The user ID if found.
//   - *errLib.CommonError: Error if context is nil or user ID is not found.
//
// Example usage:
//
//	userID, err := GetUserID(ctx)  // Retrieves user ID from the context.
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

// GetUserRole retrieves the user role from the context. Returns an error if the context is nil
// or the user role is not found or invalid.
//
// Returns:
//   - CtxRole: The user role if found.
//   - *errLib.CommonError: Error if context is nil or user role is not found.
//
// Example usage:
//
//	userRole, err := GetUserRole(ctx)  // Retrieves user role from the context.
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

func IsStaff(ctx context.Context) (bool, *errLib.CommonError) {
	role, err := GetUserRole(ctx)
	if err != nil {
		return false, err
	}

	switch role {
	case RoleSuperAdmin, RoleIT, RoleAdmin, RoleInstructor, RoleCoach, RoleBarber, RoleReceptionist:
		return true, nil
	default:
		return false, nil
	}
}
