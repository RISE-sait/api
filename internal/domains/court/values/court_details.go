package court

import "github.com/google/uuid"

// BaseDetails holds common court fields
type BaseDetails struct {
	Name       string
	LocationID uuid.UUID
}

// CreateDetails represents the data needed to create a court
type CreateDetails struct {
	BaseDetails
}

// UpdateDetails represents data for updating a court
type UpdateDetails struct {
	ID uuid.UUID
	BaseDetails
}

// ReadValues represents court data returned from queries
type ReadValues struct {
	ID uuid.UUID
	BaseDetails
}
