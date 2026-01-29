package user

import (
	"time"

	"github.com/google/uuid"
)

type ReadValue struct {
	ID                           uuid.UUID
	DOB                          time.Time
	HubspotID                    *string
	CountryCode                  string
	FirstName                    string
	LastName                     string
	Email                        *string
	Phone                        *string
	Notes                        *string
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
	IsArchived                   bool
	DeletedAt                    *time.Time
	ScheduledDeletionAt          *time.Time
	EmergencyContactName         *string
	EmergencyContactPhone        *string
	EmergencyContactRelationship *string
	LastMobileLoginAt            *time.Time
	PendingEmail                 *string
	MembershipInfo               *MembershipReadValue
	AthleteInfo                  *AthleteReadValue
}

type UpdateValue struct {
	ParentID                     uuid.UUID
	FirstName                    string
	LastName                     string
	Email                        string
	Phone                        string
	Dob                          time.Time
	CountryAlpha2Code            string
	HasMarketingEmailConsent     bool
	HasSmsConsent                bool
	Gender                       string
	ID                           uuid.UUID
	EmergencyContactName         string
	EmergencyContactPhone        string
	EmergencyContactRelationship string
}

type CustomerFilter struct {
	ParentID  uuid.UUID
	FirstName string
	LastName  string
	Email     string
	Phone     string
	ID        uuid.UUID
}
