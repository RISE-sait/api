package middlewares

import (
	errLib "api/internal/libs/errors"
	jwtLib "api/internal/libs/jwt"
	"log"

	responseHandlers "api/internal/libs/responses"
	"context"
	"net/http"
	"strings"
)

// JWTAuthMiddleware validates JWT tokens and checks user roles.
// It allows SUPERADMIN access to all routes and grants access if the user's role matches any allowed role (case-insensitive).
// Responds with 401 for missing/invalid tokens and 403 for unauthorized roles.
// Adds token claims to the request context.
//
// Example:
// router.Use(JWTAuthMiddleware("admin", "manager"))
func JWTAuthMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
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

			if claims.StaffInfo == nil || !hasRequiredRole(claims.StaffInfo.Role, allowedRoles) {
				responseHandlers.RespondWithError(w, errLib.New("You do not have permission to access this resource", http.StatusForbidden))
				return
			}

			log.Println(claims.UserID)

			// Add the claims to the request context for use in handlers
			ctx := context.WithValue(r.Context(), "userId", claims.UserID)
			ctx = context.WithValue(ctx, "hubspotId", claims.HubspotID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken extracts the JWT token from the Authorization header or cookie.
func extractToken(r *http.Request) (string, *errLib.CommonError) {
	// Check the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			return "", errLib.New("Invalid token format", http.StatusUnauthorized)
		}
		return tokenParts[1], nil
	}

	// Check the cookie
	tokenCookie, err := r.Cookie("access_token")
	if err == nil && tokenCookie.Value != "" {
		return tokenCookie.Value, nil
	}

	// No token found
	return "", errLib.New("Authorization token is required", http.StatusUnauthorized)
}

// hasRequiredRole checks if the user's role matches any of the allowed roles or is SUPERADMIN.
func hasRequiredRole(userRole string, allowedRoles []string) bool {
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
