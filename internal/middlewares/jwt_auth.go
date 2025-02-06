package middlewares

import (
	errLib "api/internal/libs/errors"
	jwtLib "api/internal/libs/jwt"

	response_handlers "api/internal/libs/responses"
	"context"
	"net/http"
	"strings"
)

// Define a custom type for the context key
type contextKey string

// Create a constant for the claims key
const claimsKey contextKey = "claims"

// JWTAuthMiddleware validates JWT tokens and checks user roles.
// It allows SUPERADMIN access to all routes and grants access if the user's role matches any allowed role (case-insensitive).
// Responds with 401 for missing/invalid tokens and 403 for unauthorized roles.
// Adds token claims to the request context.
//
// Example:
//
//	router.Use(JWTAuthMiddleware("admin", "manager"))
func JWTAuthMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {

				response_handlers.RespondWithError(w, errLib.New("Authorization header is required", http.StatusUnauthorized))
				return
			}

			// The token is typically in the format "Bearer <token>"
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
				response_handlers.RespondWithError(w, errLib.New("Invalid token format", http.StatusUnauthorized))
				return
			}

			token := tokenParts[1]

			// Verify the token and extract claims
			claims, err := jwtLib.VerifyToken(token)
			if err != nil {
				response_handlers.RespondWithError(w, errLib.New("Invalid or expired token", http.StatusUnauthorized))
				return
			}

			hasAccess := false
			for _, role := range allowedRoles {
				if strings.EqualFold(claims.Role, role) || strings.EqualFold(claims.Role, "SUPERADMIN") {
					hasAccess = true
					break
				}
			}

			// Check if the user has the required role
			if !hasAccess {
				response_handlers.RespondWithError(w, errLib.New("You do not have permission to access this resource", http.StatusForbidden))
				return
			}

			// Add the claims to the request context for use in handlers
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
