package membership_plan

import (
	"fmt"
	"strings"
	"time"

	values "api/internal/domains/membership/values"
	"github.com/google/uuid"
)

type PlanResponse struct {
	MembershipID        uuid.UUID `json:"membership_id"`
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	StripePriceID       string    `json:"stripe_price_id"`
	StripeJoiningFeesID *string   `json:"stripe_joining_fees_id,omitempty"`
	AmtPeriods          *int32    `json:"amt_periods,omitempty"`
	UnitAmount          int       `json:"unit_amount"`
	Currency            string    `json:"currency"`
	Interval            string    `json:"interval"`
	Price               string    `json:"price"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func NewPlanResponse(plan values.PlanReadValues) PlanResponse {
	displayPrice := fmt.Sprintf("$%.2f", float64(plan.UnitAmount)/100)

	return PlanResponse{
		MembershipID:        plan.MembershipID,
		ID:                  plan.ID,
		Name:                plan.Name,
		StripePriceID:       plan.StripePriceID,
		StripeJoiningFeesID: getPtrIfNotEmpty(plan.StripeJoiningFeesID),
		AmtPeriods:          plan.AmtPeriods,
		UnitAmount:          plan.UnitAmount,
		Currency:            strings.ToUpper(plan.Currency),
		Interval:            plan.Interval,
		Price:               displayPrice,
		CreatedAt:           plan.CreatedAt,
		UpdatedAt:           plan.UpdatedAt,
	}
}

func getPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
