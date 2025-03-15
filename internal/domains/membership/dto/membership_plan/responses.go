package membership_plan

import (
	values "api/internal/domains/membership/values"
	"github.com/google/uuid"
	"time"
)

type PlanResponse struct {
	MembershipID     uuid.UUID `json:"membership_id"`
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Price            string    `json:"price"`
	PaymentFrequency string    `json:"payment_frequency"`
	AmtPeriods       *int32    `json:"amt_periods,omitempty"`
	JoiningFees      string    `json:"joining_fees"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func NewPlanResponse(plan values.PlanReadValues) PlanResponse {

	return PlanResponse{
		MembershipID:     plan.MembershipID,
		ID:               plan.ID,
		Name:             plan.Name,
		Price:            plan.Price.String(),
		PaymentFrequency: plan.PaymentFrequency,
		AmtPeriods:       plan.AmtPeriods,
		JoiningFees:      plan.JoiningFees.String(),
		CreatedAt:        plan.CreatedAt,
		UpdatedAt:        plan.UpdatedAt,
	}
}
