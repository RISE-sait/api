package dto

import (
	"api/internal/domains/practice/entity"
	"api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type PracticeRequestDto struct {
	Name        string `json:"name" validate:"notwhitespace"`
	Description string `json:"description"`
}

func (dto PracticeRequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(&dto); err != nil {
		return err
	}
	return nil
}

func (dto PracticeRequestDto) ToCreateValueObjects() (*values.PracticeDetails, *errLib.CommonError) {

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.PracticeDetails{
		Name:        dto.Name,
		Description: dto.Description,
	}, nil
}

func (dto PracticeRequestDto) ToUpdateValueObjects(idStr string) (entity.Practice, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return entity.Practice{}, err
	}

	if err = dto.validate(); err != nil {
		return entity.Practice{}, err
	}

	return entity.Practice{
		ID:          id,
		Name:        dto.Name,
		Description: dto.Description,
	}, nil
}
