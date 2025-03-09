package identity

import (
	"github.com/google/uuid"
)

type UserRegistrationRequestNecessaryInfo struct {
	Age       int
	FirstName string
	LastName  string
}

type UserAuthenticationResponseInfo struct {
	Age                        int
	HasConsentToSms            bool
	HasConsentToEmailMarketing bool
	FirstName                  string
	LastName                   string
	Email                      *string
	Role                       string
	Phone                      *string
}

type AdultCustomerRegistrationRequestInfo struct {
	UserRegistrationRequestNecessaryInfo
	Waivers                    []CustomerWaiverSigning
	Email                      string
	Phone                      string
	HasConsentToSms            bool
	HasConsentToEmailMarketing bool
}

type AthleteRegistrationRequestInfo struct {
	AdultCustomerRegistrationRequestInfo
	Waivers []CustomerWaiverSigning
}

type ChildRegistrationRequestInfo struct {
	UserRegistrationRequestNecessaryInfo
	ParentEmail string
	Waivers     []CustomerWaiverSigning
}

type StaffCreateValues struct {
	IsActive bool
	RoleName string
}

type StaffRegistrationRequestInfo struct {
	UserRegistrationRequestNecessaryInfo
	HubSpotID string
	StaffCreateValues
}

type PendingUserReadValues struct {
	ID                         uuid.UUID
	FirstName                  string
	LastName                   string
	Phone                      *string
	HasConsentToSms            bool
	HasConsentToEmailMarketing bool
	Email                      *string
}

type AthleteInfo struct {
}
