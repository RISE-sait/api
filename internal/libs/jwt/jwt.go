package jwt

import (
	"api/config"
	entity "api/internal/domains/identity/entities"
	errLib "api/internal/libs/errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type CustomClaims struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"isActive"`
	jwt.StandardClaims
}

func SignJWT(userInfo entity.UserInfo) (string, *errLib.CommonError) {

	claims := CustomClaims{
		FirstName: *userInfo.FirstName,
		LastName:  *userInfo.LastName,
		Email:     userInfo.Email,
		StandardClaims: jwt.StandardClaims{
			Issuer:    config.Envs.JwtConfig.Issuer,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 15).Unix(), // 15 days
		},
	}

	if userInfo.StaffInfo != nil {
		claims.Role = userInfo.StaffInfo.Role
		claims.IsActive = userInfo.StaffInfo.IsActive
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.Envs.JwtConfig.Secret))
	if err != nil {
		fmt.Println("Error signing token: ", err)
		return "", errLib.New("Error signing token. Check Azure for logs", 500)
	}

	return signedToken, nil
}

func VerifyToken(tokenString string) (*CustomClaims, *errLib.CommonError) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Envs.JwtConfig.Secret), nil
	})

	if err != nil {
		return nil, errLib.New(err.Error(), http.StatusUnauthorized)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errLib.New("invalid token", http.StatusUnauthorized)
}
