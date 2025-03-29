package identity

import (
	"github.com/google/uuid"
	"time"
)

type UserRegistrationRequestNecessaryInfo struct {
	Age         int32
	FirstName   string
	LastName    string
	CountryCode string
}

type MembershipReadInfo struct {
	MembershipName        string
	MembershipDescription string
	MembershipBenefits    string
	PlanName              string
	StartDate             time.Time
	RenewalDate           *time.Time
}

type AthleteInfo struct {
	Wins     int32
	Losses   int32
	Points   int32
	Steals   int32
	Assists  int32
	Rebounds int32
}

type UserReadInfo struct {
	ID             uuid.UUID
	HubspotID      *string
	Gender         *string
	Age            int32
	CountryCode    string
	FirstName      string
	LastName       string
	Email          *string
	Role           string
	IsActiveStaff  *bool
	Phone          *string
	MembershipInfo *MembershipReadInfo
	AthleteInfo    *AthleteInfo
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
