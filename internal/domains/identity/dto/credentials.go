package identity

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"regexp"
	"strings"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewCredentials(email, password string) *Credentials {
	return &Credentials{
		Email:    strings.TrimSpace(email),
		Password: strings.TrimSpace(password),
	}
}

func (upc *Credentials) Validate() *errLib.CommonError {

	if upc.Email == "" {
		return errLib.New("'Email' cannot be empty or whitespace", http.StatusBadRequest)
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(upc.Email) {
		return errLib.New("Invalid email format", http.StatusBadRequest)
	}

	// user entered a password
	if len(upc.Password) > 0 && len(upc.Password) < 8 {
		return errLib.New("Password must be at least 8 characters", http.StatusBadRequest)
	}

	return nil
}
