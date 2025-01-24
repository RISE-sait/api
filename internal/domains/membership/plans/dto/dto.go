package dto

import (
	"github.com/google/uuid"
)

type CreateMembershipPlanRequest struct {
	MembershipID     uuid.UUID `json:"membership_id" validate:"required"`
	Name             string    `json:"name" validate:"required_and_notwhitespace"`
	Price            int64     `json:"price" validate:"required"`
	PaymentFrequency string    `json:"payment_frequency" validate:"payment_frequency"`
	AmtPeriods       int       `json:"amt_periods"`
}

type UpdateMembershipPlanRequest struct {
	Name             string    `json:"name" validate:"required_and_notwhitespace"`
	Price            int64     `json:"price" validate:"required"`
	PaymentFrequency string    `json:"payment_frequency" validate:"payment_frequency"`
	AmtPeriods       int       `json:"amt_periods"`
	MembershipID     uuid.UUID `json:"membership_id" validate:"required"`
	ID               uuid.UUID `json:"id" validate:"required"`
}

type MembershipPlanResponse struct {
	MembershipID     uuid.UUID `json:"membership_id"`
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Price            int64     `json:"price"`
	PaymentFrequency string    `json:"payment_frequency"`
	AmtPeriods       int       `json:"amt_periods"`
}
