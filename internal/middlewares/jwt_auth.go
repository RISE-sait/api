package middlewares

import (
	"context"
	"net/http"
	"strings"

	errLib "api/internal/libs/errors"
	jwtLib "api/internal/libs/jwt"
	responseHandlers "api/internal/libs/responses"
	contextUtils "api/utils/context"
)

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

// OptionalJWTAuthMiddleware parses the JWT token if present and injects the claims into the request context.
// It does not require the token; requests without a token proceed anonymously.
func OptionalJWTAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := jwtLib.VerifyToken(tokenParts[1])
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

			ctx = context.WithValue(ctx, contextUtils.RoleKey, userRole)
			ctx = context.WithValue(ctx, contextUtils.UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
