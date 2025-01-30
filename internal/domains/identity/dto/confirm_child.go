package identity

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"regexp"
)

type ConfirmChildDto struct {
	ChildEmail  string `json:"child_email"`
	ParentEmail string `json:"parent_email"`
}

func NewConfirmChildDto(childEmail, parentEmail string) *ConfirmChildDto {
	return &ConfirmChildDto{
		ChildEmail:  childEmail,
		ParentEmail: parentEmail,
	}
}

func (c *ConfirmChildDto) Validate() *errLib.CommonError {

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if !emailRegex.MatchString(c.ChildEmail) {
		return errLib.New("Invalid email format for field 'child_email'", http.StatusBadRequest)
	}

	if !emailRegex.MatchString(c.ParentEmail) {
		return errLib.New("Invalid email format for field 'parent_email'", http.StatusBadRequest)
	}

	return nil
}
