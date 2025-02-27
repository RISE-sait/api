package game

import (
	values "api/internal/domains/game/values"
	"github.com/google/uuid"
)

type ResponseDto struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	VideoLink *string   `json:"video_link,omitempty"`
}

func NewGameResponse(details values.ReadValue) ResponseDto {
	return ResponseDto{
		ID:        details.ID,
		Name:      details.Name,
		VideoLink: details.VideoLink,
	}
}
