package identity

import (
	"github.com/google/uuid"
	"time"
)

type MembershipReadResponseDto struct {
	MembershipName string     `json:"membership_name"`
	PlanName       string     `json:"plan_name"`
	StartDate      time.Time  `json:"start_date"`
	RenewalDate    *time.Time `json:"renewal_date,omitempty"`
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
	Age            int32                      `json:"age"`
	CountryCode    string                     `json:"country_code,omitempty"`
	MembershipInfo *MembershipReadResponseDto `json:"membership_info,omitempty"`
}
