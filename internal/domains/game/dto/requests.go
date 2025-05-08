package game

import (
	values "api/internal/domains/game/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"time"

	"github.com/google/uuid"
)

// RequestDto represents the expected payload for creating or updating a game.
// It includes team IDs, scores, time info, location, and game status.
type RequestDto struct {
	HomeTeamID uuid.UUID  `json:"home_team_id" validate:"required"`                     // ID of the home team
	AwayTeamID uuid.UUID  `json:"away_team_id" validate:"required"`                     // ID of the away team
	HomeScore  *int32     `json:"home_score"`                                           // Optional score for the home team
	AwayScore  *int32     `json:"away_score"`                                           // Optional score for the away team
	StartTime  time.Time  `json:"start_time" validate:"required"`                       // Required start time of the game
	EndTime    *time.Time `json:"end_time"`                                             // Optional end time of the game
	LocationID uuid.UUID  `json:"location_id" validate:"required"`                      // ID of the location where the game is played
	Status     string     `json:"status" validate:"oneof=scheduled completed canceled"` // Game status must be one of the allowed values
}

// ToCreateGameValue converts a validated RequestDto into a CreateGameValue used in the domain layer.
func (dto *RequestDto) ToCreateGameValue() (values.CreateGameValue, *errLib.CommonError) {
	var details values.CreateGameValue

	// Run validator on the DTO
	if err := validators.ValidateDto(dto); err != nil {
		return details, err
	}

	if dto.HomeTeamID == dto.AwayTeamID {
		return details, errLib.New("home_team_id and away_team_id must be different", 400)
	}

	// Map fields to domain struct
	details = values.CreateGameValue{
		HomeTeamID: dto.HomeTeamID,
		AwayTeamID: dto.AwayTeamID,
		HomeScore:  dto.HomeScore,
		AwayScore:  dto.AwayScore,
		StartTime:  dto.StartTime,
		EndTime:    dto.EndTime,
		LocationID: dto.LocationID,
		Status:     dto.Status,
	}

	return details, nil
}

// ToUpdateGameValue converts a validated RequestDto and a string ID into an UpdateGameValue.
// Useful for PUT/PATCH requests that update an existing game.
func (dto *RequestDto) ToUpdateGameValue(idStr string) (values.UpdateGameValue, *errLib.CommonError) {
	var details values.UpdateGameValue

	// Parse and validate the UUID from the provided string
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		return details, err
	}

	// Validate the rest of the DTO
	if err := validators.ValidateDto(dto); err != nil {
		return details, err
	}

	if dto.HomeTeamID == dto.AwayTeamID {
		return details, errLib.New("home_team_id and away_team_id must be different", 400)
	}

	// Construct the UpdateGameValue object using embedded CreateGameValue
	details = values.UpdateGameValue{
		ID: id,
		CreateGameValue: values.CreateGameValue{
			HomeTeamID: dto.HomeTeamID,
			AwayTeamID: dto.AwayTeamID,
			HomeScore:  dto.HomeScore,
			AwayScore:  dto.AwayScore,
			StartTime:  dto.StartTime,
			EndTime:    dto.EndTime,
			LocationID: dto.LocationID,
			Status:     dto.Status,
		},
	}

	return details, nil
}
