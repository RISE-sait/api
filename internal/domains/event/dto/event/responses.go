package event

import (
	values "api/internal/domains/event/values"
	"github.com/google/uuid"
	"time"
)

//goland:noinspection GoNameStartsWithPackageName
type (
	ProgramInfo struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
		Type string    `json:"type"`
	}

	LocationInfo struct {
		ID      uuid.UUID `json:"id"`
		Name    string    `json:"name"`
		Address string    `json:"address"`
	}

	TeamInfo struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	EventResponseDto struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		StartAt      string    `json:"start_at"`
		EndAt        string    `json:"end_at"`
		LocationInfo `json:"location"`
		*ProgramInfo `json:"program,omitempty"`
		*TeamInfo
	}
)

func NewEventResponseDto(event values.ReadEventValues) EventResponseDto {

	response := EventResponseDto{
		ID:      event.ID,
		StartAt: event.StartAt.String(),
		EndAt:   event.EndAt.String(),
		LocationInfo: LocationInfo{
			ID:      event.Location.ID,
			Name:    event.Location.Name,
			Address: event.Location.Address,
		},
		CreatedAt: event.CreatedAt,
		UpdatedAt: event.UpdatedAt,
	}

	if event.Program != nil {
		response.ProgramInfo = &ProgramInfo{
			ID:   event.Program.ID,
			Name: event.Program.Name,
			Type: event.Program.Type,
		}
	}

	if event.Team != nil {
		response.TeamInfo = &TeamInfo{
			ID:   event.Team.ID,
			Name: event.Team.Name,
		}
	}

	return response
}
