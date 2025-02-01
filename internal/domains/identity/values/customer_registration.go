package values

type CustomerRegistrationValueObject struct {
	Email    string
	Password *string
	Waivers  []CustomerWaiverSigning
}
