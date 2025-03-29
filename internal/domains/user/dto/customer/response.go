package customer

import (
	values "api/internal/domains/user/values"
	"github.com/google/uuid"
	"time"
)

type Response struct {
	UserID         uuid.UUID              `json:"user_id"`
	Age            int32                  `json:"age"`
	FirstName      string                 `json:"first_name"`
	LastName       string                 `json:"last_name"`
	Email          *string                `json:"email,omitempty"`
	Phone          *string                `json:"phone,omitempty"`
	HubspotId      *string                `json:"hubspot_id,omitempty"`
	CountryCode    string                 `json:"country_code"`
	AthleteInfo    *AthleteResponseDto    `json:"athlete_info,omitempty"`
	MembershipInfo *MembershipResponseDto `json:"membership_info,omitempty"`
}

type MembershipResponseDto struct {
	MembershipName        *string    `json:"membership_name,omitempty"`
	MembershipPlanID      *uuid.UUID `json:"membership_plan_id,omitempty"`
	MembershipPlanName    *string    `json:"membership_plan_name,omitempty"`
	MembershipStartDate   *time.Time `json:"membership_start_date,omitempty"`
	MembershipRenewalDate *time.Time `json:"membership_renewal_date,omitempty"`
}

type AthleteResponseDto struct {
	Wins     int32 `json:"wins"`
	Losses   int32 `json:"losses"`
	Points   int32 `json:"points"`
	Steals   int32 `json:"steals"`
	Assists  int32 `json:"assists"`
	Rebounds int32 `json:"rebounds"`
}

func UserReadValueToResponse(customer values.ReadValue) Response {
	response := Response{
		UserID:      customer.ID,
		Age:         customer.Age,
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

	if athleteInfo := customer.AthleteInfo; athleteInfo != nil {
		response.AthleteInfo = &AthleteResponseDto{
			Wins:     athleteInfo.Wins,
			Losses:   athleteInfo.Losses,
			Points:   athleteInfo.Points,
			Steals:   athleteInfo.Steals,
			Assists:  athleteInfo.Assists,
			Rebounds: athleteInfo.Rebounds,
		}
	}

	return response
}
