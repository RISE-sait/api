package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"
)

type Credentials struct {
	Email    string
	Password string
}

func NewCredentials(email, password string) *Credentials {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	return &Credentials{
		Email:    strings.TrimSpace(email),
		Password: strings.TrimSpace(password),
	}
}

func (c *Credentials) Validate() *errLib.CommonError {
	if c.Email == "" {
		return errLib.New("email cannot be empty", http.StatusBadRequest)
	}
	if !strings.Contains(c.Email, "@") {
		return errLib.New("invalid email format", http.StatusBadRequest)
	}

	if c.Password == "" {
		return errLib.New("password cannot be empty", http.StatusBadRequest)
	}
	if len(c.Password) < 5 {
		return errLib.New("password must be at least 6 characters", http.StatusBadRequest)
	}
	return nil
}
