package entity

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	CustomerID            uuid.UUID
	Name                  *string
	Email                 string
	MembershipName        string
	MembershipRenewalDate *time.Time
	Attendance            int64
}
