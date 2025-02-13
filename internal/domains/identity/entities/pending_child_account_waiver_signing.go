package entity

import (
	"time"

	"github.com/google/uuid"
)

type PendingAccountsWaiverSigning struct {
	UserID    uuid.UUID
	WaiverID  uuid.UUID
	WaiverUrl string
	IsSigned  bool
	UpdatedAt time.Time
}
