package values

type CreatePendingChildAccountValueObject struct {
	RegisterCredentials
	ParentEmail string
	Waivers     []CustomerWaiverSigning
}
