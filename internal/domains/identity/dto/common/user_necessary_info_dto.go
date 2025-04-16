package identity

import (
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"api/utils/countries"
	"net/http"
)

type UserBaseInfoRequestDto struct {
	CountryCode string `json:"country_code"`
	FirstName   string `json:"first_name" validate:"required,notwhitespace"`
	LastName    string `json:"last_name" validate:"required,notwhitespace"`
	DOB         string `json:"dob" validate:"required" example:"2000-01-01"`
	Gender      string `json:"gender" validate:"omitempty,oneof=M F"`
}

func (dto UserBaseInfoRequestDto) Validate() *errLib.CommonError {
	if err := validators.ValidateDto(&dto); err != nil {
		return err
	}

	if dto.CountryCode != "" && !countries.IsValidAlpha2Code(dto.CountryCode) {
		return errLib.New("Invalid country code", http.StatusBadRequest)
	}

	return nil
}
