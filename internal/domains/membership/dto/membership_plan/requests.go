package membership_plan

import (
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/google/uuid"
)

type PlanRequestDto struct {
	MembershipID        uuid.UUID `json:"membership_id" validate:"required"`
	Name                string    `json:"name" validate:"notwhitespace"`
	AmtPeriods          *int32    `json:"amt_periods" validate:"omitempty,gt=0"`
	StripePriceID       string    `json:"stripe_price_id" validate:"required,notwhitespace"`
	StripeJoiningFeesID string    `json:"stripe_joining_fees_id"`
}

func (dto PlanRequestDto) ToCreateValueObjects() (values.PlanCreateValues, *errLib.CommonError) {

	var vo values.PlanCreateValues

	err := validators.ValidateDto(dto)
	if err != nil {
		return vo, err
	}

	value := values.PlanCreateValues{
		PlanDetails: values.PlanDetails{
			Name:                dto.Name,
			MembershipID:        dto.MembershipID,
			AmtPeriods:          dto.AmtPeriods,
			StripeJoiningFeesID: dto.StripeJoiningFeesID,
			StripePriceID:       dto.StripePriceID,
		},
	}

	return value, nil
}

func (dto PlanRequestDto) ToUpdateValueObjects(planIdStr string) (values.PlanUpdateValues, *errLib.CommonError) {

	var vo values.PlanUpdateValues

	planId, err := validators.ParseUUID(planIdStr)

	if err != nil {
		return vo, err
	}

	err = validators.ValidateDto(dto)
	if err != nil {
		return vo, err
	}

	return values.PlanUpdateValues{
		ID: planId,
		PlanDetails: values.PlanDetails{
			Name:                dto.Name,
			MembershipID:        dto.MembershipID,
			AmtPeriods:          dto.AmtPeriods,
			StripeJoiningFeesID: dto.StripeJoiningFeesID,
			StripePriceID:       dto.StripePriceID,
		},
	}, nil
}
