package customer

import (
	values "api/internal/domains/user/values"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Response struct {
	UserID                       uuid.UUID              `json:"user_id"`
	DOB                          string                 `json:"dob"`
	FirstName                    string                 `json:"first_name"`
	LastName                     string                 `json:"last_name"`
	Email                        *string                `json:"email,omitempty"`
	Phone                        *string                `json:"phone,omitempty"`
	HubspotId                    *string                `json:"hubspot_id,omitempty"`
	CountryCode                  string                 `json:"country_code"`
	Notes                        *string                `json:"notes,omitempty"`
	EmergencyContactName         *string                `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone        *string                `json:"emergency_contact_phone,omitempty"`
	EmergencyContactRelationship *string                `json:"emergency_contact_relationship,omitempty"`
	LastMobileLoginAt            *time.Time             `json:"last_mobile_login_at,omitempty"`
	PendingEmail                 *string                `json:"pending_email,omitempty"`
	MembershipInfo               *MembershipResponseDto `json:"membership_info,omitempty"`
	PhotoURL                     *string                `json:"photo_url,omitempty"`
	IsArchived                   bool                   `json:"is_archived"`
	DeletedAt                    *time.Time             `json:"deleted_at,omitempty"`
	ScheduledDeletionAt          *time.Time             `json:"scheduled_deletion_at,omitempty"`
}

type MembershipResponseDto struct {
	MembershipName        *string    `json:"membership_name,omitempty"`
	MembershipPlanID      *uuid.UUID `json:"membership_plan_id,omitempty"`
	MembershipPlanName    *string    `json:"membership_plan_name,omitempty"`
	MembershipStartDate   *time.Time `json:"membership_start_date,omitempty"`
	MembershipRenewalDate *time.Time `json:"membership_renewal_date,omitempty"`
	Status                *string    `json:"status,omitempty"`
}

func UserReadValueToResponse(customer values.ReadValue) Response {
	response := Response{
		UserID:                       customer.ID,
		DOB:                          customer.DOB.Format("2006-01-02"),
		FirstName:                    customer.FirstName,
		LastName:                     customer.LastName,
		Email:                        customer.Email,
		Phone:                        customer.Phone,
		CountryCode:                  customer.CountryCode,
		HubspotId:                    customer.HubspotID,
		Notes:                        customer.Notes,
		EmergencyContactName:         customer.EmergencyContactName,
		EmergencyContactPhone:        customer.EmergencyContactPhone,
		EmergencyContactRelationship: customer.EmergencyContactRelationship,
		LastMobileLoginAt:            customer.LastMobileLoginAt,
		PendingEmail:                 customer.PendingEmail,
		IsArchived:                   customer.IsArchived,
		DeletedAt:                    customer.DeletedAt,
		ScheduledDeletionAt:          customer.ScheduledDeletionAt,
	}

	if customer.MembershipInfo != nil {
		response.MembershipInfo = &MembershipResponseDto{
			MembershipName:        &customer.MembershipInfo.MembershipName,
			MembershipStartDate:   &customer.MembershipInfo.MembershipStartDate,
			MembershipRenewalDate: &customer.MembershipInfo.MembershipRenewalDate,
			MembershipPlanID:      &customer.MembershipInfo.MembershipPlanID,
			MembershipPlanName:    &customer.MembershipInfo.MembershipPlanName,
			Status:                &customer.MembershipInfo.Status,
		}
	}

	if customer.AthleteInfo != nil && customer.AthleteInfo.PhotoURL != nil {
		response.PhotoURL = customer.AthleteInfo.PhotoURL
	}

	return response
}

type MembershipHistoryResponse struct {
	MembershipName        string     `json:"membership_name"`
	MembershipDescription string     `json:"membership_description"`
	MembershipPlanName    string     `json:"membership_plan_name"`
	MembsershipBenefits   string     `json:"membership_benefits"`
	Price                 string     `json:"price"`
	StartDate             time.Time  `json:"start_date"`
	RenewalDate           *time.Time `json:"renewal_date,omitempty"`
	NextPaymentDate       *time.Time `json:"next_payment_date,omitempty"`
	Status                string     `json:"status"`
}

func MembershipHistoryValueToResponse(v values.MembershipHistoryValue) MembershipHistoryResponse {
	price := fmt.Sprintf("$%.2f", float64(v.UnitAmount)/100)
	return MembershipHistoryResponse{
		MembershipName:        v.MembershipName,
		MembershipDescription: v.MembershipDescription,
		MembershipPlanName:    v.MembershipPlanName,
		Price:                 price,
		StartDate:             v.StartDate,
		RenewalDate:           v.RenewalDate,
		NextPaymentDate:       v.NextPaymentDate,
		Status:                v.Status,
		MembsershipBenefits:   v.MembershipBenefits,
	}
}
