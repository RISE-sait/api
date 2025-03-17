package course

import (
	values "api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type RequestDto struct {
	Name        string `json:"name" validate:"required,notwhitespace"`
	Description string `json:"description"`
	Capacity    int32  `json:"capacity" validate:"required,gt=0"`
}

func (dto RequestDto) ToCreateCourseDetails() (values.CreateCourseDetails, *errLib.CommonError) {

	if err := validators.ValidateDto(&dto); err != nil {
		return values.CreateCourseDetails{}, err
	}

	return values.CreateCourseDetails{
		Details: values.Details{
			Name:        dto.Name,
			Description: dto.Description,
			Capacity:    dto.Capacity,
		},
	}, nil
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
			Capacity:    dto.Capacity,
		},
	}

	return details, nil
}
