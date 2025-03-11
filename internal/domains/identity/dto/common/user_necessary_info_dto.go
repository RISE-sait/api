package identity

import (
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"api/utils/countries"
	"net/http"
)

type UserNecessaryInfoRequestDto struct {
	CountryCode string `json:"country_code"`
	FirstName   string `json:"first_name" validate:"required,notwhitespace"`
	LastName    string `json:"last_name" validate:"required,notwhitespace"`
	Age         int    `json:"age" validate:"required,gt=0"`
}

func (dto UserNecessaryInfoRequestDto) Validate() *errLib.CommonError {
	if err := validators.ValidateDto(&dto); err != nil {
		return err
	}

	if dto.CountryCode != "" && !countries.IsValidAlpha2Code(dto.CountryCode) {
		return errLib.New("Invalid country code", http.StatusBadRequest)
	}

	return nil
}
