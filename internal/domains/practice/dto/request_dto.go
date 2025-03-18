package practice

import (
	"api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type RequestDto struct {
	Name        string `json:"name" validate:"required,notwhitespace"`
	Description string `json:"description"`
	Level       string `json:"level" validate:"required,notwhitespace"`
	Capacity    int32  `json:"capacity" validate:"required,gt=0"`
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

	return values.CreatePracticeValues{
		PracticeDetails: values.PracticeDetails{
			Name:        dto.Name,
			Description: dto.Description,
			Level:       dto.Level,
			Capacity:    dto.Capacity,
		},
	}, nil
}

func (dto RequestDto) ToUpdateValueObjects(idStr string) (values.UpdatePracticeValues, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return values.UpdatePracticeValues{}, err
	}

	if err = dto.validate(); err != nil {
		return values.UpdatePracticeValues{}, err
	}

	return values.UpdatePracticeValues{
		ID: id,
		PracticeDetails: values.PracticeDetails{
			Name:        dto.Name,
			Description: dto.Description,
			Level:       dto.Level,
			Capacity:    dto.Capacity,
		},
	}, nil
}
