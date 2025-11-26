package middlewares

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	errLib "api/internal/libs/errors"
	jwtLib "api/internal/libs/jwt"
	responseHandlers "api/internal/libs/responses"
	contextUtils "api/utils/context"
)

var db *sql.DB

// SetDB sets the database connection for suspension checks
func SetDB(database *sql.DB) {
	db = database
}

// JWTAuthMiddleware validates JWT tokens and checks user roles.
// It allows superadmin access to all routes and grants access if the user's role matches any allowed role (case-insensitive).
// If isAllowAnyoneWithValidToken is true, any user with a valid token is allowed, regardless of roles.
// Responds with 401 for missing/invalid tokens and 403 for unauthorized roles.
// Adds token claims to the request context.
//
// Example:
// router.Use(JWTAuthMiddleware(false, "admin", "manager"))
func JWTAuthMiddleware(isAllowAnyoneWithValidToken bool, allowedRoles ...contextUtils.CtxRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractToken(r)
			if err != nil {
				responseHandlers.RespondWithError(w, err)
				return
			}

			// Verify the token and extract claims
			claims, err := jwtLib.VerifyToken(token)
			if err != nil {
				responseHandlers.RespondWithError(w, errLib.New("Invalid or expired token", http.StatusUnauthorized))
				return
			}

			ctx := r.Context()

			userRole, err := extractRole(claims.RoleInfo)
			if err != nil {
				responseHandlers.RespondWithError(w, err)
				return
			}

			// Check if user is suspended or deleted (skip for superadmin)
			if userRole != contextUtils.RoleSuperAdmin {
				suspended, deleted, statusErr := checkUserSuspensionOrDeletion(ctx, claims.UserID.String())
				if statusErr != nil {
					log.Printf("Error checking account status for user %s: %v", claims.UserID, statusErr)
					// Continue despite error - don't block legitimate users if DB query fails
				} else if deleted {
					responseHandlers.RespondWithError(w, errLib.New("Your account has been deleted. Please use the app to recover your account if within the recovery period.", http.StatusUnauthorized))
					return
				} else if suspended {
					responseHandlers.RespondWithError(w, errLib.New("Your account has been suspended. Please contact support for more information.", http.StatusForbidden))
					return
				}
			}

			isAuthorized := hasRequiredRole(userRole, allowedRoles)

			if isAuthorized || isAllowAnyoneWithValidToken {
				// Add the claims to the request context for use in handlers
				ctx = context.WithValue(ctx, contextUtils.RoleKey, userRole)
				ctx = context.WithValue(ctx, contextUtils.UserIDKey, claims.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				responseHandlers.RespondWithError(w, errLib.New("You do not have permission to access this resource", http.StatusForbidden))
				return
			}
		})
	}
}

// extractToken extracts the JWT token from the Authorization header or cookie.
func extractToken(r *http.Request) (string, *errLib.CommonError) {
	// Check the Authorization header
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return "", errLib.New("Authorization token is required", http.StatusUnauthorized)
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		return "", errLib.New("Invalid token format", http.StatusUnauthorized)
	}
	return tokenParts[1], nil
}

func extractRole(userRoleInfo *jwtLib.RoleInfo) (contextUtils.CtxRole, *errLib.CommonError) {
	if userRoleInfo == nil {
		return "", errLib.New("No role found in jwt", http.StatusUnauthorized)
	}

	userRole := userRoleInfo.Role

	roles := []contextUtils.CtxRole{
		contextUtils.RoleAdmin,
		contextUtils.RoleInstructor,
		contextUtils.RoleCoach,
		contextUtils.RoleSuperAdmin,
		contextUtils.RoleParent,
		contextUtils.RoleBarber,
		contextUtils.RoleAthlete,
		contextUtils.RoleChild,
		contextUtils.RoleReceptionist,
	}

	for _, role := range roles {
		if strings.EqualFold(userRole, string(role)) {
			return role, nil
		}
	}

	return "", errLib.New("Invalid role", http.StatusUnauthorized)
}

// hasRequiredRole checks if the user's role matches any of the allowed roles or is SUPERADMIN.
func hasRequiredRole(userRole contextUtils.CtxRole, allowedRoles []contextUtils.CtxRole) bool {
	if userRole == "" {
		return false
	}

	// SUPER ADMIN has access to everything
	if userRole == contextUtils.RoleSuperAdmin {
		return true
	}

	// Check if the user's role matches any allowed role
	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}

	return false
}

// checkUserSuspensionOrDeletion checks if a user is currently suspended or has deleted their account
// Returns (isSuspended, isDeleted, error)
func checkUserSuspensionOrDeletion(ctx context.Context, userID string) (bool, bool, error) {
	if db == nil {
		log.Printf("Warning: Database connection not set in JWT middleware")
		return false, false, nil // Fail open if DB not configured
	}

	query := `
		SELECT suspended_at, suspension_expires_at, deleted_at
		FROM users.users
		WHERE id = $1
	`

	var suspendedAt, suspensionExpiresAt, deletedAt sql.NullTime
	err := db.QueryRowContext(ctx, query, userID).Scan(&suspendedAt, &suspensionExpiresAt, &deletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// User not found - let it proceed, will fail at authorization
			return false, false, nil
		}
		return false, false, err
	}

	// Check if user account is deleted
	if deletedAt.Valid {
		return false, true, nil
	}

	// Check if user is suspended
	if !suspendedAt.Valid {
		// Not suspended
		return false, false, nil
	}

	// Check if suspension has expired
	if suspensionExpiresAt.Valid {
		now := time.Now().UTC()
		if now.After(suspensionExpiresAt.Time) {
			// Suspension has expired - user should be unsuspended
			// But we'll return false here and let them access (auto-unsuspension)
			// Ideally you'd want a cron job to clear expired suspensions
			return false, false, nil
		}
	}

	// User is currently suspended
	return true, false, nil
}
