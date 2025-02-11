package membership

import (
	values "api/internal/domains/membership/values/plans"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"

	"github.com/google/uuid"
)

type MembershipPlanRequestDto struct {
	MembershipID     uuid.UUID `json:"membership_id" validate:"required"`
	Name             string    `json:"name" validate:"notwhitespace"`
	Price            int64     `json:"price" validate:"required"`
	PaymentFrequency string    `json:"payment_frequency" validate:"notwhitespace"`
	AmtPeriods       int       `json:"amt_periods" `
}

func (dto *MembershipPlanRequestDto) ToCreateValueObjects() (*values.MembershipPlanDetails, *errLib.CommonError) {
	err := validators.ValidateDto(dto)
	if err != nil {
		return nil, err
	}

	var periods *int

	if dto.AmtPeriods != 0 {
		amtPeriods := int(dto.AmtPeriods)
		periods = &amtPeriods
	}

	return &values.MembershipPlanDetails{
		Name:             dto.Name,
		Price:            dto.Price,
		MembershipID:     dto.MembershipID,
		PaymentFrequency: dto.PaymentFrequency,
		AmtPeriods:       periods,
	}, nil
}

func (dto *MembershipPlanRequestDto) ToUpdateValueObjects(membershipIdStr, planIdStr string) (*values.MembershipPlanAllFields, *errLib.CommonError) {

	membershipId, err := validators.ParseUUID(membershipIdStr)

	if err != nil {
		return nil, err
	}

	planId, err := validators.ParseUUID(planIdStr)

	if err != nil {
		return nil, err
	}

	err = validators.ValidateDto(dto)
	if err != nil {
		return nil, err
	}

	var periods *int

	if dto.AmtPeriods != 0 {
		amtPeriods := int(dto.AmtPeriods)
		periods = &amtPeriods
	}

	return &values.MembershipPlanAllFields{
		ID: planId,
		MembershipPlanDetails: values.MembershipPlanDetails{
			Name:             dto.Name,
			Price:            dto.Price,
			MembershipID:     membershipId,
			PaymentFrequency: dto.PaymentFrequency,
			AmtPeriods:       periods,
		},
	}, nil
}
