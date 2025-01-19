package middlewares

import (
	"api/config"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// Define your secret key
var secretKey = []byte(config.Envs.JwtConfig.Secret)

// JWT middleware to validate JWT token
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// The token format is typically "Bearer <token>"
		tokenString := strings.Split(authHeader, " ")
		if len(tokenString) != 2 || tokenString[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		// Parse and validate the token
		token, err := jwt.Parse(tokenString[1], func(token *jwt.Token) (interface{}, error) {
			// Ensure the token's signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Check if the token is valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Extract roles from the claims
			role, ok := claims["role"].(string)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Store user data and roles in the context
			type contextKey string
			userKey := contextKey("user")
			roleKey := contextKey("role")
			ctx := context.WithValue(r.Context(), userKey, claims)
			ctx = context.WithValue(ctx, roleKey, role)
			// Pass the context to the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}
	})
}
