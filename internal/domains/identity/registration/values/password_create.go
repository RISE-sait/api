package values

// import (
// 	errLib "api/internal/libs/errors"
// 	"net/http"
// 	"strings"
// )

// type UserPasswordCreate struct {
// 	Email    string
// 	Password string
// }

// func NewUserPasswordCreate(email, password string) *UserPasswordCreate {
// 	return &UserPasswordCreate{
// 		Email:    strings.TrimSpace(email),
// 		Password: strings.TrimSpace(password),
// 	}
// }

// func (upc *UserPasswordCreate) Validate() *errLib.CommonError {

// 	if upc.Email == "" {
// 		return errLib.New("Email cannot be empty or whitespace", http.StatusBadRequest)
// 	}

// 	if !strings.Contains(upc.Email, "@") {
// 		return errLib.New("Invalid email", http.StatusBadRequest)
// 	}
// 	if len(upc.Password) < 8 {
// 		return errLib.New("Password must be at least 8 characters", http.StatusBadRequest)
// 	}

// 	return nil
// }
