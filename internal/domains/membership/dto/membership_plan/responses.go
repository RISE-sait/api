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
	CreditAllocation    *int32    `json:"credit_allocation,omitempty"`
	WeeklyCreditLimit   *int32    `json:"weekly_credit_limit,omitempty"`
	UnitAmount          int       `json:"unit_amount"`
	Currency            string    `json:"currency"`
	Interval            string    `json:"interval"`
	Price               string    `json:"price"`
	JoiningFeePrice     string    `json:"joining_fee_price"`
	IsVisible           bool      `json:"is_visible"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func NewPlanResponse(plan values.PlanReadValues) PlanResponse {
	displayPrice := fmt.Sprintf("$%.2f", float64(plan.UnitAmount)/100) // convert unit_amount to dollars and format as string
	joiningFeePrice := fmt.Sprintf("$%.2f", float64(plan.JoiningFee)/100) // convert joining_fee to dollars and format as string

	return PlanResponse{
		MembershipID:        plan.MembershipID,
		ID:                  plan.ID,
		Name:                plan.Name,
		StripePriceID:       plan.StripePriceID,
		StripeJoiningFeesID: getPtrIfNotEmpty(plan.StripeJoiningFeesID),
		AmtPeriods:          plan.AmtPeriods,
		CreditAllocation:    plan.CreditAllocation,
		WeeklyCreditLimit:   plan.WeeklyCreditLimit,
		UnitAmount:          plan.UnitAmount,
		Currency:            strings.ToUpper(plan.Currency), // e.g. "USD", "CAD"
		Interval:            plan.Interval, // e.g. "month", "year", weekly, etc.
		Price:               displayPrice, // e.g. "$10.00" display price
		JoiningFeePrice:     joiningFeePrice, // e.g. "$130.00" display price
		IsVisible:           plan.IsVisible,
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
