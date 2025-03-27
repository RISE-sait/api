package purchase

import (
	"api/internal/domains/payment/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/google/uuid"
)

type MembershipPlanRequestDto struct {
	MembershipPlanID uuid.UUID `json:"membership_plan_id"`
	Status           string    `json:"status"`
}

func (dto *MembershipPlanRequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}
	return nil
}

func (dto *MembershipPlanRequestDto) ToPurchaseRequestInfo(customerId uuid.UUID) (values.MembershipPlanPurchaseInfo, *errLib.CommonError) {

	var vo values.MembershipPlanPurchaseInfo

	if err := dto.validate(); err != nil {
		return vo, err
	}

	return values.MembershipPlanPurchaseInfo{
		CustomerId:       customerId,
		MembershipPlanId: dto.MembershipPlanID,
		Status:           dto.Status,
	}, nil
}
