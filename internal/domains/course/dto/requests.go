package course

import (
	values "api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/shopspring/decimal"
	"net/http"
)

type RequestDto struct {
	Name        string `json:"name" validate:"required,notwhitespace"`
	Description string `json:"description"`
	PayGPrice   string `json:"pay_as_u_go_price"`
}

func (dto RequestDto) ToCreateCourseDetails() (values.CreateCourseDetails, *errLib.CommonError) {

	if err := validators.ValidateDto(&dto); err != nil {
		return values.CreateCourseDetails{}, err
	}

	vo := values.CreateCourseDetails{
		Details: values.Details{
			Name:        dto.Name,
			Description: dto.Description,
		},
	}

	if dto.PayGPrice != "" {
		if priceDecimal, decimalErr := decimal.NewFromString(dto.PayGPrice); decimalErr != nil {
			return values.CreateCourseDetails{}, errLib.New("pay_as_u_go_price: Invalid price format", http.StatusBadRequest)
		} else {
			vo.PayGPrice = &priceDecimal
		}
	}

	return vo, nil
}

func (dto RequestDto) ToUpdateCourseDetails(idStr string) (values.UpdateCourseDetails, *errLib.CommonError) {

	var details values.UpdateCourseDetails

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return details, err
	}

	if err = validators.ValidateDto(&dto); err != nil {
		return details, err
	}

	details = values.UpdateCourseDetails{
		ID: id,
		Details: values.Details{
			Name:        dto.Name,
			Description: dto.Description,
		},
	}

	if dto.PayGPrice != "" {
		if priceDecimal, decimalErr := decimal.NewFromString(dto.PayGPrice); decimalErr != nil {
			return values.UpdateCourseDetails{}, errLib.New("pay_as_u_go_price: Invalid price format", http.StatusBadRequest)
		} else {
			details.PayGPrice = &priceDecimal
		}
	}

	return details, nil
}
