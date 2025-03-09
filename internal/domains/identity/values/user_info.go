package identity

import (
	"api/internal/domains/user/values/staff"
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

type RegularCustomerRegistrationRequestInfo struct {
	UserRegistrationRequestNecessaryInfo
	Email                      string
	Phone                      string
	HasConsentToSms            bool
	HasConsentToEmailMarketing bool
	Waivers                    []CustomerWaiverSigning
}

type ChildRegistrationRequestInfo struct {
	UserRegistrationRequestNecessaryInfo
	ParentEmail string
	Waivers     []CustomerWaiverSigning
}

type StaffRegistrationRequestInfo struct {
	HubSpotID string
	staff.CreateValues
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
