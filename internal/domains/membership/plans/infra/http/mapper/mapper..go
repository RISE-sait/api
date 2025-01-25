package mapper

import (
	entity "api/internal/domains/membership/plans/entities"
	"api/internal/domains/membership/plans/infra/http/dto"
)

func MapEntityToResponse(membership *entity.MembershipPlan) *dto.MembershipPlanResponse {
	return &dto.MembershipPlanResponse{
		ID:               membership.ID,
		Name:             membership.Name,
		MembershipID:     membership.MembershipID,
		Price:            membership.Price,
		PaymentFrequency: membership.PaymentFrequency,
		AmtPeriods:       membership.AmtPeriods,
	}
}
