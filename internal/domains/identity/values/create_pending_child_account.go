package values

type CreatePendingChildAccountValueObject struct {
	ChildEmail  string
	Password    *string
	ParentEmail string
	Waivers     []CustomerWaiverSigning
}
