package values

type CustomerRegistrationValueObject struct {
	RegisterCredentials
	Waivers []CustomerWaiverSigning
}
