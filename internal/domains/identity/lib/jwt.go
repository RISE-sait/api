package lib

import (
	"api/config"
	"api/internal/domains/identity/entities"
	errLib "api/internal/libs/errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func SignJWT(userInfo entities.UserInfo) (string, *errLib.CommonError) {

	role := "Athlete"

	isActiveStaff := false

	if userInfo.StaffInfo != nil {
		role = userInfo.StaffInfo.Role
		isActiveStaff = userInfo.StaffInfo.IsActive
	}

	claims := jwt.MapClaims{
		"name":          userInfo.Name,
		"email":         userInfo.Email,
		"iss":           config.Envs.JwtConfig.Issuer,
		"exp":           time.Now().Add(time.Hour * 24 * 15).Unix(), // 15 days
		"role":          role,
		"isActiveStaff": isActiveStaff,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.Envs.JwtConfig.Secret))
	if err != nil {
		fmt.Println("Error signing token: ", err)
		return "", errLib.New("Error signing token. Check Azure for logs", 500)
	}

	return signedToken, nil
}
