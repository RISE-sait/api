package practice

import (
	values "api/internal/domains/practice/values"
	"time"

	"github.com/google/uuid"
)

type ResponseDto struct {
	ID           uuid.UUID  `json:"id"`
	TeamID       uuid.UUID  `json:"team_id"`
	TeamName     string     `json:"team_name"`
	TeamLogoUrl  string     `json:"team_logo_url"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	LocationID   uuid.UUID  `json:"location_id"`
	LocationName string     `json:"location_name"`
	CourtID      uuid.UUID  `json:"court_id"`
	CourtName    string     `json:"court_name"`
	Status       string     `json:"status"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

func NewResponse(v values.ReadPracticeValue) ResponseDto {
	return ResponseDto{
		ID:           v.ID,
		TeamID:       v.TeamID,
		TeamName:     v.TeamName,
		TeamLogoUrl:  v.TeamLogoUrl,
		StartTime:    v.StartTime,
		EndTime:      v.EndTime,
		LocationID:   v.LocationID,
		LocationName: v.LocationName,
		CourtID:      v.CourtID,
		CourtName:    v.CourtName,
		Status:       v.Status,
		CreatedAt:    v.CreatedAt,
		UpdatedAt:    v.UpdatedAt,
	}
}