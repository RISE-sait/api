package customer

import (
	values "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type StatsUpdateRequestDto struct {
	Wins     *int32 `json:"wins" validate:"omitempty,gt=0"`
	Losses   *int32 `json:"losses" validate:"omitempty,gt=0"`
	Points   *int32 `json:"points" validate:"omitempty,gt=0"`
	Steals   *int32 `json:"steals" validate:"omitempty,gt=0"`
	Assists  *int32 `json:"assists" validate:"omitempty,gt=0"`
	Rebounds *int32 `json:"rebounds" validate:"omitempty,gt=0"`
}

func (dto StatsUpdateRequestDto) ToUpdateValue(idStr string) (values.StatsUpdateValue, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return values.StatsUpdateValue{}, err
	}

	if err := validators.ValidateDto(dto); err != nil {
		return values.StatsUpdateValue{}, err
	}

	return values.StatsUpdateValue{
		ID:       id,
		Wins:     dto.Wins,
		Losses:   dto.Losses,
		Points:   dto.Points,
		Steals:   dto.Steals,
		Assists:  dto.Assists,
		Rebounds: dto.Rebounds,
	}, nil

}
