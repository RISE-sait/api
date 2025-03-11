package identity

import "github.com/google/uuid"

type UserRegistrationRequestNecessaryInfo struct {
	Age         int32
	FirstName   string
	LastName    string
	CountryCode string
}

type UserReadInfo struct {
	ID          uuid.UUID
	Age         int32
	CountryCode string
	FirstName   string
	LastName    string
	Email       *string
	Role        string
	Phone       *string
}

type StaffReadInfo struct {
	Age         int32
	CountryCode string
	FirstName   string
	LastName    string
	Email       string
	Role        string
	Phone       string
}

type AthleteRegistrationRequestInfo struct {
	UserRegistrationRequestNecessaryInfo
	Email                      string
	Phone                      string
	HasConsentToSms            bool
	HasConsentToEmailMarketing bool
	Waivers                    []CustomerWaiverSigning
}

type ParentRegistrationRequestInfo struct {
	UserRegistrationRequestNecessaryInfo
	Email                      string
	Phone                      string
	HasConsentToSms            bool
	HasConsentToEmailMarketing bool
}

type ChildRegistrationRequestInfo struct {
	UserRegistrationRequestNecessaryInfo
	ParentEmail string
	Waivers     []CustomerWaiverSigning
}

type StaffRegistrationRequestInfo struct {
	UserRegistrationRequestNecessaryInfo
	Email    string
	Phone    string
	IsActive bool
	RoleName string
}

type ApprovedStaffRegistrationRequestInfo struct {
	UserID uuid.UUID
	StaffRegistrationRequestInfo
}

type PendingStaffRegistrationRequestInfo struct {
	StaffRegistrationRequestInfo
}
