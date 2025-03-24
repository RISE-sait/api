package game

import (
	values "api/internal/domains/game/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type RequestDto struct {
	Name string `json:"name" validate:"notwhitespace"`
}

func (dto *RequestDto) ToCreateGameName() (string, *errLib.CommonError) {

	var details values.CreateGameValue

	if err := validators.ValidateDto(dto); err != nil {
		return details.Name, err
	}

	return details.Name, nil
}

func (dto *RequestDto) ToUpdateGameValue(idStr string) (values.UpdateGameValue, *errLib.CommonError) {

	var details values.UpdateGameValue

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return details, err
	}

	if err = validators.ValidateDto(dto); err != nil {
		return details, err
	}

	details.ID = id

	details.Name = dto.Name

	return details, nil
}
