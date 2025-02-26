package membership_plan

import (
	values "api/internal/domains/membership/values"
	"github.com/google/uuid"
)

type PlanResponse struct {
	MembershipID     uuid.UUID `json:"membership_id"`
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Price            int64     `json:"price"`
	PaymentFrequency *string   `json:"payment_frequency"`
	AmtPeriods       *int      `json:"amt_periods,omitempty"`
}

func NewPlanResponse(plan values.PlanReadValues) *PlanResponse {
	return &PlanResponse{
		ID:               plan.ID,
		Name:             plan.Name,
		MembershipID:     plan.MembershipID,
		Price:            plan.Price,
		PaymentFrequency: plan.PaymentFrequency,
		AmtPeriods:       plan.AmtPeriods,
	}
}
