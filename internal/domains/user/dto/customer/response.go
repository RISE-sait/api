package customer

import (
	"github.com/google/uuid"
	"time"
)

type Response struct {
	UserID              uuid.UUID  `json:"user_id"`
	Age                 int32      `json:"age"`
	FirstName           string     `json:"first_name"`
	LastName            string     `json:"last_name"`
	Email               *string    `json:"email,omitempty"`
	Phone               *string    `json:"phone,omitempty"`
	MembershipName      *string    `json:"membership_name,omitempty"`
	MembershipStartDate *time.Time `json:"membership_start_date,omitempty"`
	HubspotId           *string    `json:"hubspot_id,omitempty"`
	CountryCode         string     `json:"country_code"`
	ProfilePic          string     `json:"profile_pic"`
}

type AthleteResponseDto struct {
	ID         uuid.UUID `json:"id"`
	ProfilePic string    `json:"profile_pic"`
	Wins       int32     `json:"wins"`
	Losses     int32     `json:"losses"`
	Points     int32     `json:"points"`
	Steals     int32     `json:"steals"`
	Assists    int32     `json:"assists"`
	Rebounds   int32     `json:"rebounds"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
