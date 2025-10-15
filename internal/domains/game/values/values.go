package values

import (
	"time"

	"github.com/google/uuid"
)

// CreateGameValue represents the data required to create a new game.
// This struct is used in service and repository layers as a domain transfer object.
type CreateGameValue struct {
	HomeTeamID uuid.UUID  // ID of the home team
	AwayTeamID uuid.UUID  // ID of the away team
	HomeScore  *int32     // Optional score for the home team
	AwayScore  *int32     // Optional score for the away team
	StartTime  time.Time  // Start time of the game (required)
	EndTime    *time.Time // Optional end time of the game
	LocationID uuid.UUID  // ID of the location where the game is held
	CourtID    uuid.UUID  // ID of the court where the game is played
	Status     string     // Status of the game (scheduled, completed, canceled)
	CreatedBy  uuid.UUID  // ID of the user (coach/admin) who created the game
}

// UpdateGameValue represents the data required to update an existing game.
// It embeds CreateGameValue and adds an ID field for the game being updated.
type UpdateGameValue struct {
	ID              uuid.UUID // ID of the game to update
	CreateGameValue           // Embedded fields for game details
}

// ReadGameValue represents a full game record retrieved from the database.
// Includes both UUIDs and display names for related entities, as well as timestamps.
type ReadGameValue struct {
	ID              uuid.UUID  // Game ID
	HomeTeamID      uuid.UUID  // UUID of the home team
	HomeTeamName    string     // Name of the home team
	HomeTeamLogoUrl string     // Logo URL of the home team
	AwayTeamID      uuid.UUID  // UUID of the away team
	AwayTeamName    string     // Name of the away team
	AwayTeamLogoUrl string     // Logo URL of the away team
	HomeScore       *int32     // Home team score (nullable)
	AwayScore       *int32     // Away team score (nullable)
	StartTime       time.Time  // Scheduled start time
	EndTime         *time.Time // End time (nullable)
	LocationID      uuid.UUID  // UUID of the game location
	LocationName    string     // Name of the location
	CourtID         uuid.UUID  // UUID of the game court
	CourtName       string     // Name of the court
	Status          string     // Game status
	CreatedBy       *uuid.UUID // UUID of the user who created the game (nullable)
	CreatedByName   string     // Name of the user who created the game
	CreatedAt       *time.Time // Time the record was created (nullable)
	UpdatedAt       *time.Time // Time the record was last updated (nullable)
}

// GetGamesFilter defines optional filters for querying games.
type GetGamesFilter struct {
	CourtID    *uuid.UUID
	LocationID *uuid.UUID
	Limit      int32
	Offset     int32
}
