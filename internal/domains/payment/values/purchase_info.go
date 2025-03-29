package values

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type MembershipPlanPurchaseInfo struct {
	CustomerId       uuid.UUID
	MembershipPlanId uuid.UUID
	Status           string
	RenewalDate      *time.Time
}

type ProgramRegistrationInfo struct {
	ProgramName string
	Price       decimal.NullDecimal
}

type MembershipPlanJoiningRequirement struct {
	ID               uuid.UUID
	Name             string
	Price            decimal.Decimal
	JoiningFee       decimal.Decimal
	AutoRenew        bool
	MembershipID     uuid.UUID
	PaymentFrequency string
	AmtPeriods       *int32
	IsOneTimePayment bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
