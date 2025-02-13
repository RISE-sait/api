package values

type CreatePendingChildAccountValueObject struct {
	CustomerRegistrationInfo
	ParentEmail string
	Waivers     []CustomerWaiverSigning
}
