package user

import (
	"github.com/google/uuid"
	"time"
)

type ReadValue struct {
	ID             uuid.UUID
	DOB            time.Time
	HubspotID      *string
	CountryCode    string
	FirstName      string
	LastName       string
	Email          *string
	Phone          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	MembershipInfo *MembershipReadValue
	AthleteInfo    *AthleteReadValue
}

type UpdateValue struct {
	ParentID                 uuid.UUID
	FirstName                string
	LastName                 string
	Email                    string
	Phone                    string
	Dob                      time.Time
	CountryAlpha2Code        string
	HasMarketingEmailConsent bool
	HasSmsConsent            bool
	Gender                   string
	ID                       uuid.UUID
}
