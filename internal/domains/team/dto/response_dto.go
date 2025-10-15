package team

import (
	"github.com/google/uuid"
	"time"
)

type Response struct {
	ID         uuid.UUID           `json:"id"`
	Name       string              `json:"name"`
	Capacity   int32               `json:"capacity"`
	Coach      *Coach              `json:"coach,omitempty"`
	LogoURL    *string             `json:"logo_url,omitempty"`
	IsExternal bool                `json:"is_external"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
	Roster     *[]RosterMemberInfo `json:"roster,omitempty"`
}

type Coach struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type RosterMemberInfo struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email,omitempty"`
	Country  string    `json:"country"`
	Points   int32     `json:"points"`
	Wins     int32     `json:"wins"`
	Losses   int32     `json:"losses"`
	Assists  int32     `json:"assists"`
	Rebounds int32     `json:"rebounds"`
	Steals   int32     `json:"steals"`
}
