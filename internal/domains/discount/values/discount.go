package discount

import (
	"time"

	"github.com/google/uuid"
)

type DurationType string

const (
	DurationOnce      DurationType = "once"
	DurationRepeating DurationType = "repeating"
	DurationForever   DurationType = "forever"
)

type DiscountType string

const (
	TypePercentage  DiscountType = "percentage"
	TypeFixedAmount DiscountType = "fixed_amount"
)

type AppliesTo string

const (
	AppliesToSubscription AppliesTo = "subscription"
	AppliesToOneTime      AppliesTo = "one_time"
	AppliesToBoth         AppliesTo = "both"
)

type CreateValues struct {
	Name             string
	Description      string
	DiscountPercent  int
	DiscountAmount   *float64 // For fixed amount discounts
	DiscountType     DiscountType
	IsUseUnlimited   bool
	UsePerClient     int
	IsActive         bool
	ValidFrom        time.Time
	ValidTo          time.Time
	DurationType     DurationType
	DurationMonths   *int // Required when DurationType is "repeating"
	AppliesTo        AppliesTo
	MaxRedemptions   *int // Global max redemptions
	StripeCouponID   *string
}

type UpdateValues struct {
	ID uuid.UUID
	CreateValues
}

type ReadValues struct {
	ID             uuid.UUID
	CreateValues
	TimesRedeemed  int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
