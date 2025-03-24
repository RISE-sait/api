package game

import (
	values "api/internal/domains/game/values"

	"github.com/google/uuid"
)

type ResponseDto struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	WinTeam     uuid.UUID `json:"win_team"`
	LoseTeam    uuid.UUID `json:"lose_team"`
	WinScore    int32     `json:"win_score"`
	LoseScore   int32     `json:"lose_score"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

func NewGameResponse(details values.ReadValue) ResponseDto {
	return ResponseDto{
		ID:          details.ID,
		Name:        details.Name,
		Description: details.Description,
		WinTeam:     details.WinTeamID,
		LoseTeam:    details.LoseTeamID,
		WinScore:    details.WinScore,
		LoseScore:   details.LoseScore,
		CreatedAt:   details.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   details.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
