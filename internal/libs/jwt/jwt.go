package jwtLib

import (
	"api/config"
	errLib "api/internal/libs/errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RoleInfo struct {
	Role          string `json:"role"`
	IsActiveStaff *bool  `json:"isActiveStaff,omitempty"`
}

type CustomClaims struct {
	UserID uuid.UUID `json:"user_id"`
	*RoleInfo
}

type JwtClaims struct {
	CustomClaims
	jwt.RegisteredClaims
}

func SignJWT(customClaims CustomClaims) (string, *errLib.CommonError) {

	claims := JwtClaims{
		CustomClaims: customClaims,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.Env.JwtConfig.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // 24 hours (reduced from 15 days)
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.Env.JwtConfig.Secret))
	if err != nil {
		log.Printf("Error signing JWT token: %v", err)
		return "", errLib.New("Error signing token", http.StatusInternalServerError)
	}

	return signedToken, nil
}

func VerifyToken(tokenString string) (*JwtClaims, *errLib.CommonError) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Env.JwtConfig.Secret), nil
	})

	if err != nil {
		return nil, errLib.New(err.Error(), http.StatusUnauthorized)
	}

	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errLib.New("invalid token", http.StatusUnauthorized)
}
