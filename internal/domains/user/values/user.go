package user

import (
	"github.com/google/uuid"
	"time"
)

type ReadValue struct {
	ID             uuid.UUID
	Age            int32
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
