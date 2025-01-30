package entities

import (
	"time"

	"github.com/google/uuid"
)

type PendingChildAccount struct {
	ID          uuid.UUID
	ParentEmail string
	UserEmail   string
	Password    string
	CreatedAt   time.Time
}
