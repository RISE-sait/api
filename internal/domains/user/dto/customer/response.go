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
	MembershipInfo               *MembershipResponseDto  `json:"membership_info,omitempty"`
	Memberships                  []MembershipResponseDto `json:"memberships,omitempty"`
	PhotoURL                     *string                `json:"photo_url,omitempty"`
	IsArchived                   bool                   `json:"is_archived"`
	ArchivedAt                   *time.Time             `json:"archived_at,omitempty"`
	DeletedAt                    *time.Time             `json:"deleted_at,omitempty"`
	ScheduledDeletionAt          *time.Time             `json:"scheduled_deletion_at,omitempty"`
	DaysUntilDeletion            *int                   `json:"days_until_deletion,omitempty"`
}

type MembershipResponseDto struct {
	MembershipName        *string    `json:"membership_name,omitempty"`
	MembershipPlanID      *uuid.UUID `json:"membership_plan_id,omitempty"`
	MembershipPlanName    *string    `json:"membership_plan_name,omitempty"`
	MembershipStartDate   *time.Time `json:"membership_start_date,omitempty"`
	MembershipRenewalDate *time.Time `json:"membership_renewal_date,omitempty"`
	Status                *string    `json:"status,omitempty"`
	StripeSubscriptionID  *string    `json:"stripe_subscription_id,omitempty"`
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
		ArchivedAt:                   customer.ArchivedAt,
		DeletedAt:                    customer.DeletedAt,
		ScheduledDeletionAt:          customer.ScheduledDeletionAt,
	}

	// Calculate days until deletion
	response.DaysUntilDeletion = calculateDaysUntilDeletion(customer)

	if len(customer.Memberships) > 0 {
		// Set membership_info to the primary (first) membership for backwards compatibility
		primary := customer.Memberships[0]
		response.MembershipInfo = &MembershipResponseDto{
			MembershipName:        &primary.MembershipName,
			MembershipStartDate:   &primary.MembershipStartDate,
			MembershipRenewalDate: &primary.MembershipRenewalDate,
			MembershipPlanID:      &primary.MembershipPlanID,
			MembershipPlanName:    &primary.MembershipPlanName,
			Status:                &primary.Status,
			StripeSubscriptionID:  primary.StripeSubscriptionID,
		}

		// Map all memberships into the memberships array
		response.Memberships = make([]MembershipResponseDto, len(customer.Memberships))
		for i, m := range customer.Memberships {
			response.Memberships[i] = MembershipResponseDto{
				MembershipName:        &customer.Memberships[i].MembershipName,
				MembershipStartDate:   &m.MembershipStartDate,
				MembershipRenewalDate: &m.MembershipRenewalDate,
				MembershipPlanID:      &m.MembershipPlanID,
				MembershipPlanName:    &customer.Memberships[i].MembershipPlanName,
				Status:                &customer.Memberships[i].Status,
				StripeSubscriptionID:  m.StripeSubscriptionID,
			}
		}
	}

	if customer.AthleteInfo != nil && customer.AthleteInfo.PhotoURL != nil {
		response.PhotoURL = customer.AthleteInfo.PhotoURL
	}

	return response
}

// calculateDaysUntilDeletion returns the number of days until the account is permanently deleted
// Returns nil if the account is not scheduled for deletion
func calculateDaysUntilDeletion(customer values.ReadValue) *int {
	now := time.Now().UTC()
	var deletionDate time.Time

	// Check soft-deleted accounts first (they have explicit scheduled_deletion_at)
	if customer.ScheduledDeletionAt != nil {
		deletionDate = *customer.ScheduledDeletionAt
	} else if customer.ArchivedAt != nil {
		// Archived accounts are deleted 30 days after being archived
		deletionDate = customer.ArchivedAt.Add(30 * 24 * time.Hour)
	} else {
		return nil
	}

	days := int(deletionDate.Sub(now).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return &days
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
	StripeSubscriptionID  *string    `json:"stripe_subscription_id,omitempty"`
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
		StripeSubscriptionID:  v.StripeSubscriptionID,
	}
}
