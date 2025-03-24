package game

import (
	"time"

	"github.com/google/uuid"
)

type BaseValue struct {
	Name         string
	Description  string
	WinTeamID    uuid.UUID
	WinTeamName  string
	LoseTeamID   uuid.UUID
	LoseTeamName string
	WinScore     int32
	LoseScore    int32
}
type CreateGameValue struct {
	BaseValue
}

type UpdateGameValue struct {
	ID uuid.UUID
	BaseValue
}

type ReadValue struct {
	ID uuid.UUID
	BaseValue
	CreatedAt time.Time
	UpdatedAt time.Time
}
