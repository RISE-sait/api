package identity

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"regexp"
)

type CreateChildAccountDto struct {
	ParentEmail string `json:"parent_email"`
}

func NewChildAccountCreateDto(parentEmail string) *CreateChildAccountDto {
	return &CreateChildAccountDto{
		ParentEmail: parentEmail,
	}
}

func (vu *CreateChildAccountDto) Validate() *errLib.CommonError {

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(vu.ParentEmail) {
		return errLib.New("Invalid email format for field 'parent_email'", http.StatusBadRequest)
	}

	return nil
}
