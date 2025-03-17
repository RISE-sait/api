package middlewares

import (
	errLib "api/internal/libs/errors"
	jwtLib "api/internal/libs/jwt"
	responseHandlers "api/internal/libs/responses"
	"context"
	"net/http"
	"strings"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	UserIDKey    ContextKey = "userId"
	HubspotIDKey ContextKey = "hubspotId"
)

// JWTAuthMiddleware validates JWT tokens and checks user roles.
// It allows SUPERADMIN access to all routes and grants access if the user's role matches any allowed role (case-insensitive).
// If isAllowAnyoneWithValidToken is true, any user with a valid token is allowed, regardless of roles.
// Responds with 401 for missing/invalid tokens and 403 for unauthorized roles.
// Adds token claims to the request context.
//
// Example:
// router.Use(JWTAuthMiddleware(false, "admin", "manager"))
func JWTAuthMiddleware(isAllowAnyoneWithValidToken bool, allowedRoles ...string) func(http.Handler) http.Handler {
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

			if !isAllowAnyoneWithValidToken && !hasRequiredRole(claims.RoleInfo, allowedRoles) {
				responseHandlers.RespondWithError(w, errLib.New("You do not have permission to access this resource", http.StatusForbidden))
				return
			}

			// Add the claims to the request context for use in handlers
			ctx := addClaimsToContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
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

// hasRequiredRole checks if the user's role matches any of the allowed roles or is SUPERADMIN.
func hasRequiredRole(staffInfo *jwtLib.RoleInfo, allowedRoles []string) bool {

	if staffInfo == nil {
		return false
	}

	userRole := staffInfo.Role

	// SUPER ADMIN has access to everything
	if strings.EqualFold(userRole, "SUPERADMIN") {
		return true
	}

	// Check if the user's role matches any allowed role
	for _, role := range allowedRoles {
		if strings.EqualFold(userRole, role) {
			return true
		}
	}

	return false
}

// addClaimsToContext adds the JWT claims to the request context.
func addClaimsToContext(ctx context.Context, claims *jwtLib.JwtClaims) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	return ctx
}
