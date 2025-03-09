package customer

import (
	"github.com/google/uuid"
	"time"
)

type MembershipPlansResponseDto struct {
	ID               uuid.UUID  `json:"id"`
	CustomerID       uuid.UUID  `json:"customer_id"`
	MembershipPlanID uuid.UUID  `json:"membership_plan_id"`
	StartDate        time.Time  `json:"start_date"`
	RenewalDate      *time.Time `json:"renewal_date,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	MembershipName   string     `json:"membership_name"`
}
