package membership

import (
	"github.com/google/uuid"
)

type MembershipPlanResponse struct {
	MembershipID     uuid.UUID `json:"membership_id"`
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Price            int64     `json:"price"`
	PaymentFrequency string    `json:"payment_frequency"`
	AmtPeriods       *int      `json:"amt_periods,omitempty"`
}
