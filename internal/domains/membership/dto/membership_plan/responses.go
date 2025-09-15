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
	JoiningFee          int       `json:"joining_fee"`
	JoiningFeeDisplay   string    `json:"joining_fee_display"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func NewPlanResponse(plan values.PlanReadValues) PlanResponse {
	displayPrice := fmt.Sprintf("$%.2f", float64(plan.UnitAmount)/100) // convert unit_amount to dollars and format as string
	joiningFeeDisplay := fmt.Sprintf("$%.2f", float64(plan.JoiningFee)/100) // convert joining_fee to dollars and format as string

	return PlanResponse{
		MembershipID:        plan.MembershipID,
		ID:                  plan.ID,
		Name:                plan.Name,
		StripePriceID:       plan.StripePriceID,
		StripeJoiningFeesID: getPtrIfNotEmpty(plan.StripeJoiningFeesID),
		AmtPeriods:          plan.AmtPeriods,
		UnitAmount:          plan.UnitAmount,
		Currency:            strings.ToUpper(plan.Currency), // e.g. "USD", "CAD"
		Interval:            plan.Interval, // e.g. "month", "year", weekly, etc.
		Price:               displayPrice, // e.g. "$10.00" display price
		JoiningFee:          plan.JoiningFee,
		JoiningFeeDisplay:   joiningFeeDisplay, // e.g. "$130.00" display price
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
