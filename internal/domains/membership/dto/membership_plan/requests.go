package membership_plan

import (
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"

	"github.com/google/uuid"
)

type PlanRequestDto struct {
	MembershipID     uuid.UUID `json:"membership_id" validate:"required"`
	Name             string    `json:"name" validate:"notwhitespace"`
	Price            int64     `json:"price" validate:"required"`
	PaymentFrequency *string   `json:"payment_frequency" validate:"omitempty,notwhitespace"`
	AmtPeriods       *int      `json:"amt_periods" validate:"omitempty,gt=0"`
}

func (dto *PlanRequestDto) ToCreateValueObjects() (*values.PlanCreateValues, *errLib.CommonError) {
	err := validators.ValidateDto(dto)
	if err != nil {
		return nil, err
	}

	return &values.PlanCreateValues{
		Name:             dto.Name,
		Price:            dto.Price,
		MembershipID:     dto.MembershipID,
		PaymentFrequency: dto.PaymentFrequency,
		AmtPeriods:       dto.AmtPeriods,
	}, nil
}

func (dto *PlanRequestDto) ToUpdateValueObjects(planIdStr string) (*values.PlanUpdateValues, *errLib.CommonError) {

	planId, err := validators.ParseUUID(planIdStr)

	if err != nil {
		return nil, err
	}

	err = validators.ValidateDto(dto)
	if err != nil {
		return nil, err
	}

	return &values.PlanUpdateValues{
		ID:               planId,
		Name:             dto.Name,
		Price:            dto.Price,
		PaymentFrequency: dto.PaymentFrequency,
		AmtPeriods:       dto.AmtPeriods,
		MembershipID:     dto.MembershipID,
	}, nil
}
