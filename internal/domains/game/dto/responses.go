package game

import (
	values "api/internal/domains/game/values"
	"time"

	"github.com/google/uuid"
)
// ResponseDto defines the structure of the JSON response for a game.
// It includes both IDs and human-readable names for teams and locations,
// along with game metadata such as scores, timing, and status.
type ResponseDto struct {
	ID           uuid.UUID  `json:"id"`
	HomeTeamID   uuid.UUID  `json:"home_team_id"`
	HomeTeamName string     `json:"home_team_name"`
	AwayTeamID   uuid.UUID  `json:"away_team_id"`
	AwayTeamName string     `json:"away_team_name"`
	HomeScore    *int32     `json:"home_score,omitempty"`
	AwayScore    *int32     `json:"away_score,omitempty"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	LocationID   uuid.UUID  `json:"location_id"`
	LocationName string     `json:"location_name"`
	Status       string     `json:"status"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}
// NewGameResponse maps a ReadGameValue from the domain layer into a ResponseDto
// used for API responses. It flattens and formats data for client consumption.
func NewGameResponse(details values.ReadGameValue) ResponseDto {
	return ResponseDto{
		ID:           details.ID,
		HomeTeamID:   details.HomeTeamID,
		HomeTeamName: details.HomeTeamName,
		AwayTeamID:   details.AwayTeamID,
		AwayTeamName: details.AwayTeamName,
		HomeScore:    details.HomeScore,
		AwayScore:    details.AwayScore,
		StartTime:    details.StartTime,
		EndTime:      details.EndTime,
		LocationID:   details.LocationID,
		LocationName: details.LocationName,
		Status:       details.Status,
		CreatedAt:    details.CreatedAt,
		UpdatedAt:    details.UpdatedAt,
	}
}
