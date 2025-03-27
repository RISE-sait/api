package jwtLib

import (
	"api/config"
	errLib "api/internal/libs/errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
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
	jwt.StandardClaims
}

func SignJWT(customClaims CustomClaims) (string, *errLib.CommonError) {

	claims := JwtClaims{
		CustomClaims: customClaims,
		StandardClaims: jwt.StandardClaims{
			Issuer:    config.Env.JwtConfig.Issuer,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 15).Unix(), // 15 days
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.Env.JwtConfig.Secret))
	if err != nil {
		fmt.Println("Error signing token: ", err)
		return "", errLib.New("Error signing token. Check Azure for logs", 500)
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
