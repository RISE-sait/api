package program

import (
	"api/internal/domains/program/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type RequestDto struct {
	Name        string `json:"name" validate:"required,notwhitespace"`
	Description string `json:"description"`
	Level       string `json:"level" validate:"required,notwhitespace"`
	Type        string `json:"type" validate:"required,notwhitespace"`
	Capacity    *int32 `json:"capacity"`
}

func (dto RequestDto) validate() *errLib.CommonError {
	return validators.ValidateDto(&dto)
}

func (dto RequestDto) ToCreateValueObjects() (values.CreateProgramValues, *errLib.CommonError) {
	err := dto.validate()
	if err != nil {
		return values.CreateProgramValues{}, err
	}

	return values.CreateProgramValues{
		ProgramDetails: values.ProgramDetails{
			Name:        dto.Name,
			Description: dto.Description,
			Level:       dto.Level,
			Type:        dto.Type,
		},
	}, nil
}

func (dto RequestDto) ToUpdateValueObjects(idStr string) (values.UpdateProgramValues, *errLib.CommonError) {
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		return values.UpdateProgramValues{}, err
	}

	err = dto.validate()
	if err != nil {
		return values.UpdateProgramValues{}, err
	}

	return values.UpdateProgramValues{
		ID: id,
		ProgramDetails: values.ProgramDetails{
			Name:        dto.Name,
			Description: dto.Description,
			Level:       dto.Level,
			Capacity:    dto.Capacity,
			Type:        dto.Type,
		},
	}, nil
}
