package customer

import (
	values "api/internal/domains/user/values"
	"time"

	"github.com/google/uuid"
)

type Response struct {
	UserID      uuid.UUID `json:"user_id"`
	DOB         string    `json:"dob"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       *string   `json:"email,omitempty"`
	Phone       *string   `json:"phone,omitempty"`
	HubspotId   *string   `json:"hubspot_id,omitempty"`
	CountryCode string    `json:"country_code"`
	MembershipInfo *MembershipResponseDto `json:"membership_info,omitempty"`
}

type MembershipResponseDto struct {
	MembershipName        *string    `json:"membership_name,omitempty"`
	MembershipPlanID      *uuid.UUID `json:"membership_plan_id,omitempty"`
	MembershipPlanName    *string    `json:"membership_plan_name,omitempty"`
	MembershipStartDate   *time.Time `json:"membership_start_date,omitempty"`
	MembershipRenewalDate *time.Time `json:"membership_renewal_date,omitempty"`
}

func UserReadValueToResponse(customer values.ReadValue) Response {
	response := Response{
		UserID:      customer.ID,
		DOB:         customer.DOB.Format("2006-01-02"),
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Email:       customer.Email,
		Phone:       customer.Phone,
		CountryCode: customer.CountryCode,
		HubspotId:   customer.HubspotID,
	}

	if customer.MembershipInfo != nil {
		response.MembershipInfo = &MembershipResponseDto{
			MembershipName:        &customer.MembershipInfo.MembershipName,
			MembershipStartDate:   &customer.MembershipInfo.MembershipStartDate,
			MembershipRenewalDate: &customer.MembershipInfo.MembershipRenewalDate,
			MembershipPlanID:      &customer.MembershipInfo.MembershipPlanID,
			MembershipPlanName:    &customer.MembershipInfo.MembershipPlanName,
		}
	}

	return response
}
