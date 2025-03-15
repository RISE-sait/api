package membership

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type PlanDetails struct {
	MembershipID     uuid.UUID
	Name             string
	Price            decimal.Decimal
	JoiningFees      decimal.Decimal
	PaymentFrequency string
	AmtPeriods       *int32
	IsAutoRenew      bool
}

type PlanCreateValues struct {
	PlanDetails
}

type PlanUpdateValues struct {
	ID uuid.UUID
	PlanDetails
}

type PlanReadValues struct {
	ID uuid.UUID
	PlanDetails
	CreatedAt time.Time
	UpdatedAt time.Time
}
