package membership_plan

import (
	values "api/internal/domains/membership/values"
	"github.com/google/uuid"
	"time"
)

type PlanResponse struct {
	MembershipID        uuid.UUID `json:"membership_id"`
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	StripePriceID       string    `json:"stripe_price_id"`
	StripeJoiningFeesID *string   `json:"stripe_joining_fees_id,omitempty"`
	AmtPeriods          *int32    `json:"amt_periods,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func NewPlanResponse(plan values.PlanReadValues) PlanResponse {

	response := PlanResponse{
		MembershipID:  plan.MembershipID,
		ID:            plan.ID,
		Name:          plan.Name,
		AmtPeriods:    plan.AmtPeriods,
		CreatedAt:     plan.CreatedAt,
		UpdatedAt:     plan.UpdatedAt,
		StripePriceID: plan.StripePriceID,
	}

	if plan.StripeJoiningFeesID != "" {
		response.StripeJoiningFeesID = &plan.StripeJoiningFeesID
	}

	return response
}
