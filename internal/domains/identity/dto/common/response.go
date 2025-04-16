package identity

import (
	"github.com/google/uuid"
	"time"
)

type MembershipReadResponseDto struct {
	MembershipName        string     `json:"membership_name"`
	MembershipDescription string     `json:"membership_description"`
	MembershipBenefits    string     `json:"membership_benefits"`
	PlanName              string     `json:"plan_name"`
	StartDate             time.Time  `json:"start_date"`
	RenewalDate           *time.Time `json:"renewal_date,omitempty"`
}

type AthleteResponseDto struct {
	Wins     int32 `json:"wins"`
	Losses   int32 `json:"losses"`
	Points   int32 `json:"points"`
	Steals   int32 `json:"steals"`
	Assists  int32 `json:"assists"`
	Rebounds int32 `json:"rebounds"`
}

type UserAuthenticationResponseDto struct {
	ID             uuid.UUID                  `json:"id"`
	FirstName      string                     `json:"first_name"`
	LastName       string                     `json:"last_name"`
	Gender         *string                    `json:"gender,omitempty"`
	Email          *string                    `json:"email,omitempty"`
	Role           string                     `json:"role"`
	IsActiveStaff  *bool                      `json:"is_active_staff,omitempty"`
	Phone          *string                    `json:"phone,omitempty"`
	DOB            string                     `json:"age"`
	CountryCode    string                     `json:"country_code,omitempty"`
	MembershipInfo *MembershipReadResponseDto `json:"membership_info,omitempty"`
	AthleteInfo    *AthleteResponseDto        `json:"athlete_info,omitempty"`
}
