package user

import (
	"github.com/google/uuid"
	"time"
)

type ReadValue struct {
	ID            uuid.UUID
	HubspotID     *string
	ProfilePicUrl *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type StatsUpdateValue struct {
	ID       uuid.UUID
	Wins     *int32
	Losses   *int32
	Points   *int32
	Steals   *int32
	Assists  *int32
	Rebounds *int32
}
