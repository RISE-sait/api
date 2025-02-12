package dto

import (
	entity "api/internal/domains/course/entities"
	"api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type CourseRequestDto struct {
	Name        string `json:"name" validate:"notwhitespace"`
	Description string `json:"description"`
}

func (dto *CourseRequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}
	return nil
}

func (dto *CourseRequestDto) ToCreateValueObjects() (*values.CourseDetails, *errLib.CommonError) {

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.CourseDetails{
		Name:        dto.Name,
		Description: dto.Description,
	}, nil
}

func (dto *CourseRequestDto) ToUpdateValueObjects(idStr string) (*entity.Course, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &entity.Course{
		ID:          id,
		Name:        dto.Name,
		Description: dto.Description,
	}, nil
}
