package membership_plan

import (
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/shopspring/decimal"
	"net/http"

	"github.com/google/uuid"
)

type PlanRequestDto struct {
	MembershipID     uuid.UUID `json:"membership_id" validate:"required"`
	Name             string    `json:"name" validate:"notwhitespace"`
	Price            string    `json:"price" validate:"required"`
	PaymentFrequency string    `json:"payment_frequency" validate:"required,notwhitespace"`
	AmtPeriods       *int32    `json:"amt_periods" validate:"omitempty,gt=0"`
}

func (dto PlanRequestDto) ToCreateValueObjects() (values.PlanCreateValues, *errLib.CommonError) {

	var vo values.PlanCreateValues

	err := validators.ValidateDto(dto)
	if err != nil {
		return vo, err
	}

	priceDecimal, decimalErr := decimal.NewFromString(dto.Price)
	if decimalErr != nil {
		// Handle the error appropriately, maybe return a custom validation error
		return vo, errLib.New("price: Invalid price format", http.StatusBadRequest)
	}

	value := values.PlanCreateValues{
		PlanDetails: values.PlanDetails{
			Name:             dto.Name,
			Price:            priceDecimal,
			MembershipID:     dto.MembershipID,
			PaymentFrequency: dto.PaymentFrequency,
			AmtPeriods:       dto.AmtPeriods,
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

	priceDecimal, decimalErr := decimal.NewFromString(dto.Price)
	if decimalErr != nil {
		// Handle the error appropriately, maybe return a custom validation error
		return vo, errLib.New("price: Invalid price format", http.StatusBadRequest)
	}

	return values.PlanUpdateValues{
		ID: planId,
		PlanDetails: values.PlanDetails{
			Name:             dto.Name,
			Price:            priceDecimal,
			PaymentFrequency: dto.PaymentFrequency,
			AmtPeriods:       dto.AmtPeriods,
			MembershipID:     dto.MembershipID,
		},
	}, nil
}
