package dto

import (
	"github.com/google/uuid"
)

type CustomerResponse struct {
	CustomerID            uuid.UUID `json:"customer_id"`
	Name                  string    `json:"name"`
	Email                 string    `json:"email"`
	Membership            string    `json:"membership"`
	Attendance            int64     `json:"attendance"`
	MembershipRenewalDate string    `json:"membership_renewal_date"`
}
