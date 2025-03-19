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
	MembershipName string
	PlanName       string
	StartDate      time.Time
	RenewalDate    *time.Time
}

type UserReadInfo struct {
	ID             uuid.UUID
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
