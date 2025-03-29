package program

import (
	"api/internal/domains/program/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/shopspring/decimal"
	"net/http"
)

type RequestDto struct {
	Name        string `json:"name" validate:"required,notwhitespace"`
	Description string `json:"description"`
	Level       string `json:"level" validate:"required,notwhitespace"`
	Type        string `json:"type" validate:"required,notwhitespace"`
	PaygPrice   string `json:"payg_price"`
}

func (dto RequestDto) validate() (decimal.NullDecimal, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return decimal.NullDecimal{
			Valid: false,
		}, err
	}

	if dto.PaygPrice != "" {
		if price, err := decimal.NewFromString(dto.PaygPrice); err != nil {
			return decimal.NullDecimal{
				Valid: false,
			}, errLib.New("Invalid price format", http.StatusBadRequest)
		} else if price.LessThanOrEqual(decimal.Zero) {
			return decimal.NullDecimal{
				Valid: false,
			}, errLib.New("Price must be greater than zero", http.StatusBadRequest)
		} else {
			return decimal.NullDecimal{
				Decimal: price,
				Valid:   true,
			}, nil
		}
	}

	return decimal.NullDecimal{}, nil
}

func (dto RequestDto) ToCreateValueObjects() (values.CreateProgramValues, *errLib.CommonError) {

	price, err := dto.validate()

	if err != nil {
		return values.CreateProgramValues{}, err
	}

	return values.CreateProgramValues{
		ProgramDetails: values.ProgramDetails{
			Name:        dto.Name,
			Description: dto.Description,
			Level:       dto.Level,
			Type:        dto.Type,
			PayGPrice:   price,
		},
	}, nil
}

func (dto RequestDto) ToUpdateValueObjects(idStr string) (values.UpdateProgramValues, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return values.UpdateProgramValues{}, err
	}

	price, err := dto.validate()

	if err != nil {
		return values.UpdateProgramValues{}, err
	}

	return values.UpdateProgramValues{
		ID: id,
		ProgramDetails: values.ProgramDetails{
			Name:        dto.Name,
			Description: dto.Description,
			Level:       dto.Level,
			Type:        dto.Type,
			PayGPrice:   price,
		},
	}, nil
}
