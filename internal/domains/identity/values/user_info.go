package values

import (
	"api/internal/domains/staff/values"
)

type UserNecessaryInfo struct {
	Age       int
	FirstName string
	LastName  string
}

type RegularCustomerRegistrationInfo struct {
	UserNecessaryInfo
	Email   string
	Phone   string
	Waivers []CustomerWaiverSigning
}

type ChildRegistrationInfo struct {
	UserNecessaryInfo
	ParentEmail string
	Waivers     []CustomerWaiverSigning
}

type StaffRegistrationInfo struct {
	HubSpotID string
	staff.Details
}
