package game

import (
	values "api/internal/domains/game/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"

	"github.com/google/uuid"
)

type RequestDto struct {
	Name        string    `json:"name" validate:"required,notwhitespace"`
	Description string    `json:"description"`
	WinTeam     uuid.UUID `json:"win_team"`
	LoseTeam    uuid.UUID `json:"lose_team"`
	WinScore    int32     `json:"win_score"`
	LoseScore   int32     `json:"lose_score"`
}

func (dto *RequestDto) ToCreateGameName() (values.CreateGameValue, *errLib.CommonError) {

	var details values.CreateGameValue

	if err := validators.ValidateDto(dto); err != nil {
		return details, err
	}

	details = values.CreateGameValue{
		BaseValue: values.BaseValue{
			Name:        dto.Name,
			Description: dto.Description,
			WinTeamID:   dto.WinTeam,
			LoseTeamID:  dto.LoseTeam,
			WinScore:    dto.WinScore,
			LoseScore:   dto.LoseScore,
		},
	}

	return details, nil
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

	details = values.UpdateGameValue{
		ID: id,
		BaseValue: values.BaseValue{
			Name:        dto.Name,
			Description: dto.Description,
			WinTeamID:   dto.WinTeam,
			LoseTeamID:  dto.LoseTeam,
			WinScore:    dto.WinScore,
			LoseScore:   dto.LoseScore,
		},
	}

	return details, nil
}
