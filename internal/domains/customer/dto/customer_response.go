package dto

import (
	"github.com/google/uuid"
)

type CustomerResponse struct {
	CustomerInfo CustomerInfo `json:"customer_info"`
	// MembershipInfo CustomerMembershipInfo `json:"membership_info,omitempty"`
	EventDetails CustomerEventDetails `json:"event_details,omitempty"`
}

type CustomerInfo struct {
	CustomerId uuid.UUID `json:"customer_id"`
	FirstName  string    `json:"first_name,omitempty"`
	LastName   string    `json:"last_name,omitempty"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
}

// type CustomerMembershipInfo struct {
// 	Name     string    `json:"name"`
// 	PlanInfo *PlanInfo `json:"plan_info,omitempty"`
// }

// type PlanInfo struct {
// 	Id               uuid.UUID `json:"id"`
// 	Name             string    `json:"name"`
// 	RenewalDate      string    `json:"plan_renewal_date"`
// 	Status           string    `json:"status"`
// 	StartDate        string    `json:"start_date"`
// 	UpdatedAt        string    `json:"updated_at"`
// 	PaymentFrequency string    `json:"payment_frequency"`
// 	AmtPeriods       *int32    `json:"amt_periods,omitempty"`
// 	Price            int32     `json:"price"`
// }

type CustomerEventDetails struct {
	IsCancelled bool   `json:"is_cancelled"`
	CheckedInAt string `json:"checked_in_at,omitempty"`
}
