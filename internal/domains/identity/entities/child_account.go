package entity

import (
	"time"

	"github.com/google/uuid"
)

type PendingChildAccount struct {
	ID          uuid.UUID
	ParentEmail string
	UserEmail   string
	FirstName   *string
	LastName    *string
	Password    *string
	CreatedAt   time.Time
}
