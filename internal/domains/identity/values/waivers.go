package identity

import (
	"github.com/google/uuid"
	"time"
)

type Waiver struct {
	ID        uuid.UUID
	URL       string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CustomerWaiverSigning struct {
	WaiverUrl      string
	IsWaiverSigned bool
	UpdatedAt      time.Time
}
