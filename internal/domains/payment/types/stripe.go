package stripe

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentFrequency string

const (
	Day      PaymentFrequency = "day"
	Week     PaymentFrequency = "week"
	Biweekly PaymentFrequency = "biweekly"
	Month    PaymentFrequency = "month"
)

func IsPaymentFrequencyValid(frequency PaymentFrequency) bool {
	switch frequency {
	case Day, Week, Biweekly, Month:
		return true
	default:
		return false
	}
}

type CheckoutItem struct {
	ID       uuid.UUID
	Name     string
	Price    decimal.Decimal
	Quantity int
}

type OneTimePaymentCheckoutItemType string

const (
	Program        OneTimePaymentCheckoutItemType = "program"
	Event          OneTimePaymentCheckoutItemType = "event"
	MembershipPlan OneTimePaymentCheckoutItemType = "membership_plan"
)

func IsOneTimePaymentCheckoutItemTypeValid(itemType OneTimePaymentCheckoutItemType) bool {
	switch itemType {
	case Program, Event, MembershipPlan:
		return true
	default:
		return false
	}
}
