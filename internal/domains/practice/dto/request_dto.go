package practice

import (
	"api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/shopspring/decimal"
	"net/http"
)

type RequestDto struct {
	Name        string `json:"name" validate:"required,notwhitespace"`
	Description string `json:"description"`
	Level       string `json:"level" validate:"required,notwhitespace"`
	PayGPrice   string `json:"pay_as_u_go_price"`
}

func (dto RequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(&dto); err != nil {
		return err
	}
	return nil
}

func (dto RequestDto) ToCreateValueObjects() (values.CreatePracticeValues, *errLib.CommonError) {

	if err := dto.validate(); err != nil {
		return values.CreatePracticeValues{}, err
	}

	vo := values.CreatePracticeValues{
		PracticeDetails: values.PracticeDetails{
			Name:        dto.Name,
			Description: dto.Description,
			Level:       dto.Level,
		},
	}

	if dto.PayGPrice != "" {
		if priceDecimal, decimalErr := decimal.NewFromString(dto.PayGPrice); decimalErr != nil {
			return values.CreatePracticeValues{}, errLib.New("pay_as_u_go_price: Invalid price format", http.StatusBadRequest)
		} else {
			vo.PayGPrice = &priceDecimal
		}
	}

	return vo, nil
}

func (dto RequestDto) ToUpdateValueObjects(idStr string) (values.UpdatePracticeValues, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return values.UpdatePracticeValues{}, err
	}

	if err = dto.validate(); err != nil {
		return values.UpdatePracticeValues{}, err
	}

	details := values.UpdatePracticeValues{
		ID: id,
		PracticeDetails: values.PracticeDetails{
			Name:        dto.Name,
			Description: dto.Description,
			Level:       dto.Level,
		},
	}

	if dto.PayGPrice != "" {
		if priceDecimal, decimalErr := decimal.NewFromString(dto.PayGPrice); decimalErr != nil {
			return values.UpdatePracticeValues{}, errLib.New("pay_as_u_go_price: Invalid price format", http.StatusBadRequest)
		} else {
			details.PayGPrice = &priceDecimal
		}
	}

	return details, nil
}
