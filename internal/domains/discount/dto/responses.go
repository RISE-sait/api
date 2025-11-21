package discount

import (
	"time"

	"github.com/google/uuid"
)

type ResponseDto struct {
	ID                    uuid.UUID  `json:"id"`
	Name                  string     `json:"name"`
	Description           string     `json:"description,omitempty"`
	DiscountPercent       int        `json:"discount_percent,omitempty"`
	DiscountAmount        *float64   `json:"discount_amount,omitempty"`
	DiscountType          string     `json:"discount_type"`
	IsUseUnlimited        bool       `json:"is_use_unlimited"`
	UsePerClient          int        `json:"use_per_client"`
	IsActive              bool       `json:"is_active"`
	ValidFrom             time.Time  `json:"valid_from"`
	ValidTo               *time.Time `json:"valid_to,omitempty"`
	DurationType          string     `json:"duration_type"`
	DurationMonths        *int       `json:"duration_months,omitempty"`
	AppliesTo             string     `json:"applies_to"`
	MaxRedemptions        *int       `json:"max_redemptions,omitempty"`
	TimesRedeemed         int        `json:"times_redeemed"`
	StripeCouponID        *string    `json:"stripe_coupon_id,omitempty"`
	StripePromotionCodeID *string    `json:"stripe_promotion_code_id,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
