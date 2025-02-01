package dto

import (
	"api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"time"
)

type CourseRequestDto struct {
	Name        string    `json:"name" validate:"notwhitespace"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required,gtcsfield=StartDate"`
}

func (dto *CourseRequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}
	return nil
}

func (dto *CourseRequestDto) ToCreateValueObjects() (*values.CourseCreate, *errLib.CommonError) {

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.CourseCreate{
		Name:        dto.Name,
		Description: dto.Description,
		StartDate:   dto.StartDate,
		EndDate:     dto.EndDate,
	}, nil
}

func (dto *CourseRequestDto) ToUpdateValueObjects(idStr string) (*values.CourseUpdate, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.CourseUpdate{
		ID:          id,
		Name:        dto.Name,
		Description: dto.Description,
		StartDate:   dto.StartDate,
		EndDate:     dto.EndDate,
	}, nil
}
