package entity

import (
	staffValues "api/internal/domains/staff/values"
	"github.com/google/uuid"
)

type UserInfo struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
	StaffInfo *staffValues.Details
}
