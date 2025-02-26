package course

import (
	entity "api/internal/domains/course/entity"
	values "api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type RequestDto struct {
	Name        string `json:"name" validate:"notwhitespace"`
	Description string `json:"description"`
}

func (dto *RequestDto) ToDetails() (*values.Details, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	return &values.Details{
		Name:        dto.Name,
		Description: dto.Description,
	}, nil
}

func (dto *RequestDto) ToEntity(idStr string) (*entity.Course, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	return &entity.Course{
		ID:          id,
		Name:        dto.Name,
		Description: dto.Description,
	}, nil
}
