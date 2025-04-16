package user

import (
	values "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"api/utils/countries"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type UpdateRequestDto struct {
	ParentID                 uuid.UUID `json:"parent_id"`
	FirstName                string    `json:"first_name" validate:"required,notwhitespace"`
	LastName                 string    `json:"last_name" validate:"required,notwhitespace"`
	Email                    string    `json:"email" validate:"omitempty,email"`
	Phone                    string    `json:"phone" validate:"omitempty,e164"`
	Dob                      string    `json:"dob" validate:"required,notwhitespace" example:"2000-01-01"`
	CountryAlpha2Code        string    `json:"country_alpha2_code" validate:"required,notwhitespace" example:"US"`
	HasMarketingEmailConsent *bool     `json:"has_marketing_email_consent" validate:"required"`
	HasSmsConsent            *bool     `json:"has_sms_consent" validate:"required"`
	Gender                   string    `json:"gender" validate:"omitempty,oneof=M F"`
}

func (dto UpdateRequestDto) ToUpdateValue(id uuid.UUID) (values.UpdateValue, *errLib.CommonError) {

	dob, err := time.Parse("2006-01-02", dto.Dob)

	if err != nil {
		return values.UpdateValue{}, errLib.New(fmt.Sprintf("Invalid dob date format '%s'. Must be YYYY-MM-DD", dto.Dob), http.StatusBadRequest)
	}

	if dto.CountryAlpha2Code != "" && !countries.IsValidAlpha2Code(dto.CountryAlpha2Code) {
		return values.UpdateValue{}, errLib.New("Invalid country code", http.StatusBadRequest)
	}

	if valErr := validators.ValidateDto(&dto); valErr != nil {
		return values.UpdateValue{}, valErr
	}

	return values.UpdateValue{
		ParentID:                 dto.ParentID,
		FirstName:                dto.FirstName,
		LastName:                 dto.LastName,
		Email:                    dto.Email,
		Phone:                    dto.Phone,
		Dob:                      dob,
		CountryAlpha2Code:        dto.CountryAlpha2Code,
		HasMarketingEmailConsent: *dto.HasMarketingEmailConsent,
		HasSmsConsent:            *dto.HasSmsConsent,
		Gender:                   dto.Gender,
		ID:                       id,
	}, nil
}
