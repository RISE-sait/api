package identity_utils

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"
)

func GetFirebaseTokenFromAuthorizationHeader(r *http.Request) (string, *errLib.CommonError) {

	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return "", errLib.New("Missing Firebase token", http.StatusBadRequest)
	} else {
		extractedTokenParts := strings.Split(authHeader, " ")

		if extractedTokenParts[0] != "Bearer" {
			return "", errLib.New("Invalid Firebase token. Missing Bearer", http.StatusUnauthorized)
		}

		return extractedTokenParts[1], nil
	}
}
