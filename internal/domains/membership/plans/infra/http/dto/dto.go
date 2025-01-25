package dto

import (
	"github.com/google/uuid"
)

type CreateMembershipPlanRequest struct {
	MembershipID     uuid.UUID `json:"membership_id"`
	Name             string    `json:"name"`
	Price            int64     `json:"price"`
	PaymentFrequency string    `json:"payment_frequency"`
	AmtPeriods       int       `json:"amt_periods"`
}

type UpdateMembershipPlanRequest struct {
	Name             string    `json:"name"`
	Price            int64     `json:"price"`
	PaymentFrequency string    `json:"payment_frequency"`
	AmtPeriods       int       `json:"amt_periods"`
	MembershipID     uuid.UUID `json:"membership_id"`
	ID               uuid.UUID `json:"id"`
}

type MembershipPlanResponse struct {
	MembershipID     uuid.UUID `json:"membership_id"`
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Price            int64     `json:"price"`
	PaymentFrequency string    `json:"payment_frequency"`
	AmtPeriods       int       `json:"amt_periods"`
}
