package membership_plan

import (
	"net/http"

	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"

	"github.com/google/uuid"
)

// PlanRequestDto represents the request body for creating/updating a membership plan.
// You can either provide an existing Stripe Price ID (Option 1) or pricing details to auto-create in Stripe (Option 2).
// @Description Membership plan creation request. Provide either stripe_price_id OR (unit_amount + billing_interval) to auto-create Stripe product.
type PlanRequestDto struct {
	MembershipID      uuid.UUID `json:"membership_id" validate:"required" example:"f47ac10b-58cc-4372-a567-0e02b2c3d479"`
	Name              string    `json:"name" validate:"notwhitespace" example:"Monthly Gold Plan"`
	AmtPeriods        *int32    `json:"amt_periods" validate:"omitempty,gt=0" example:"12"`
	CreditAllocation  *int32    `json:"credit_allocation" validate:"omitempty,gte=0" example:"100"`
	WeeklyCreditLimit *int32    `json:"weekly_credit_limit" validate:"omitempty,gte=0" example:"10"`

	// Option 1: Provide existing Stripe Price ID (legacy/manual way)
	StripePriceID       string `json:"stripe_price_id" example:"price_1ABC123..."`
	StripeJoiningFeesID string `json:"stripe_joining_fees_id" example:"price_1XYZ789..."`

	// Option 2: Provide pricing details to auto-create in Stripe (new way)
	UnitAmount       *int64 `json:"unit_amount" example:"5000"`        // Price in cents (e.g., 5000 = $50.00)
	Currency         string `json:"currency" example:"cad"`            // "cad" or "usd", defaults to "cad"
	BillingInterval  string `json:"billing_interval" example:"month"`  // "month", "year", "week", "day"
	IntervalCount    *int64 `json:"interval_count" example:"1"`        // defaults to 1
	JoiningFeeAmount *int64 `json:"joining_fee_amount" example:"15000"` // Optional one-time fee in cents (e.g., 15000 = $150.00)
}

func (dto PlanRequestDto) ToCreateValueObjects() (values.PlanCreateValues, *errLib.CommonError) {

	var vo values.PlanCreateValues

	err := validators.ValidateDto(&dto)
	if err != nil {
		return vo, err
	}

	// Validate: Either provide StripePriceID OR pricing details, not neither
	hasStripePriceID := dto.StripePriceID != ""
	hasPricingDetails := dto.UnitAmount != nil && dto.BillingInterval != ""

	if !hasStripePriceID && !hasPricingDetails {
		return vo, errLib.New("Either stripe_price_id or pricing details (unit_amount, billing_interval) required", http.StatusBadRequest)
	}

	// Validate billing interval if provided
	if dto.BillingInterval != "" {
		validIntervals := map[string]bool{"day": true, "week": true, "month": true, "year": true}
		if !validIntervals[dto.BillingInterval] {
			return vo, errLib.New("Invalid billing_interval: must be day, week, month, or year", http.StatusBadRequest)
		}
	}

	value := values.PlanCreateValues{
		PlanDetails: values.PlanDetails{
			Name:                dto.Name,
			MembershipID:        dto.MembershipID,
			AmtPeriods:          dto.AmtPeriods,
			StripeJoiningFeesID: dto.StripeJoiningFeesID,
			StripePriceID:       dto.StripePriceID,
			CreditAllocation:    dto.CreditAllocation,
			WeeklyCreditLimit:   dto.WeeklyCreditLimit,
		},
		UnitAmount:       dto.UnitAmount,
		Currency:         dto.Currency,
		BillingInterval:  dto.BillingInterval,
		IntervalCount:    dto.IntervalCount,
		JoiningFeeAmount: dto.JoiningFeeAmount,
	}

	return value, nil
}

func (dto PlanRequestDto) ToUpdateValueObjects(planIdStr string) (values.PlanUpdateValues, *errLib.CommonError) {

	var vo values.PlanUpdateValues

	planId, err := validators.ParseUUID(planIdStr)

	if err != nil {
		return vo, err
	}

	err = validators.ValidateDto(&dto)
	if err != nil {
		return vo, err
	}

	return values.PlanUpdateValues{
		ID: planId,
		PlanDetails: values.PlanDetails{
			Name:                dto.Name,
			MembershipID:        dto.MembershipID,
			AmtPeriods:          dto.AmtPeriods,
			StripeJoiningFeesID: dto.StripeJoiningFeesID,
			StripePriceID:       dto.StripePriceID,
			CreditAllocation:    dto.CreditAllocation,
			WeeklyCreditLimit:   dto.WeeklyCreditLimit,
		},
	}, nil
}
