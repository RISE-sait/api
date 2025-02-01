package dto

import (
	"api/internal/domains/membership/plans/values"
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

func (dto *MembershipPlanRequestDto) ToCreateValueObjects() (*values.MembershipPlanCreate, *errLib.CommonError) {
	err := validators.ValidateDto(dto)
	if err != nil {
		return nil, err
	}

	return &values.MembershipPlanCreate{
		Name:             dto.Name,
		Price:            dto.Price,
		MembershipID:     dto.MembershipID,
		PaymentFrequency: dto.PaymentFrequency,
		AmtPeriods:       dto.AmtPeriods,
	}, nil
}

func (dto *MembershipPlanRequestDto) ToUpdateValueObjects(membershipIdStr, planIdStr string) (*values.MembershipPlanUpdate, *errLib.CommonError) {

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

	return &values.MembershipPlanUpdate{
		ID:               planId,
		Name:             dto.Name,
		Price:            dto.Price,
		MembershipID:     membershipId,
		PaymentFrequency: dto.PaymentFrequency,
		AmtPeriods:       dto.AmtPeriods,
	}, nil
}
