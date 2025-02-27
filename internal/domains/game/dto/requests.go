package game

import (
	values "api/internal/domains/game/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type RequestDto struct {
	Name      string `json:"name" validate:"notwhitespace"`
	VideoLink string `json:"video_link" validate:"omitempty,url"`
}

func (dto *RequestDto) ToCreateGameValue() (values.CreateGameValue, *errLib.CommonError) {

	var details values.CreateGameValue

	if err := validators.ValidateDto(dto); err != nil {
		return details, err
	}

	details.BaseValue = values.BaseValue{
		Name: dto.Name,
	}

	if dto.VideoLink != "" {
		details.BaseValue.VideoLink = &dto.VideoLink
	}

	return details, nil
}

func (dto *RequestDto) ToUpdateGameValue(idStr string) (values.UpdateGameValue, *errLib.CommonError) {

	var details values.UpdateGameValue

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return details, err
	}

	if err := validators.ValidateDto(dto); err != nil {
		return details, err
	}

	details.ID = id

	details.BaseValue = values.BaseValue{
		Name: dto.Name,
	}

	if dto.VideoLink != "" {
		details.BaseValue.VideoLink = &dto.VideoLink
	}

	return details, nil
}
