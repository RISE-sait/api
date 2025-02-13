package values

type UserInfo struct {
	Email     string
	FirstName *string
	LastName  *string
	Phone     *string
}

type CustomerRegistrationInfo struct {
	UserInfo
	Waivers  []CustomerWaiverSigning
	Password *string
}

type StaffRegistrationInfo struct {
	UserInfo
	StaffDetails
}
