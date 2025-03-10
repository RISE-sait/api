package customer

import (
	"github.com/google/uuid"
	"time"
)

type Response struct {
	UserID     uuid.UUID `json:"user_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      *string   `json:"email,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	HubspotId  string    `json:"hubspot_id"`
	ProfilePic *string   `json:"profile_pic,omitempty"`
}

type AthleteResponseDto struct {
	ID            uuid.UUID `json:"id"`
	ProfilePicUrl *string   `json:"profile_pic_url,omitempty"`
	Wins          int32     `json:"wins"`
	Losses        int32     `json:"losses"`
	Points        int32     `json:"points"`
	Steals        int32     `json:"steals"`
	Assists       int32     `json:"assists"`
	Rebounds      int32     `json:"rebounds"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
