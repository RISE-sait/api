package valueobjects

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"
)

type Credentials struct {
	email    string
	password string
}

func NewCredentials(email, password string) (*Credentials, *errLib.CommonError) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	return &Credentials{
		email:    email,
		password: password,
	}, nil
}

func (c *Credentials) Email() string {
	return c.email
}

func (c *Credentials) Password() string {
	return c.password
}

func validateEmail(email string) *errLib.CommonError {
	if email == "" {
		return errLib.New("email cannot be empty", http.StatusBadRequest)
	}
	if !strings.Contains(email, "@") {
		return errLib.New("invalid email format", http.StatusBadRequest)
	}
	return nil
}

func validatePassword(password string) *errLib.CommonError {
	if password == "" {
		return errLib.New("password cannot be empty", http.StatusBadRequest)
	}
	if len(password) < 5 {
		return errLib.New("password must be at least 6 characters", http.StatusBadRequest)
	}
	return nil
}
