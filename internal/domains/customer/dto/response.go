package customer

import (
	"github.com/google/uuid"
)

type Response struct {
	UserID     uuid.UUID `json:"user_id"`
	HubspotId  string    `json:"hubspot_id"`
	ProfilePic *string   `json:"profile_pic,omitempty"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      *string   `json:"email,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
}
