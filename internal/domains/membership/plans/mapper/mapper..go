package mapper

import (
	"api/internal/domains/membership/plans/dto"
	entity "api/internal/domains/membership/plans/entities"
)

func MapCreateRequestToEntity(body dto.CreateMembershipPlanRequest) entity.MembershipPlan {
	return entity.MembershipPlan{
		Name:             body.Name,
		MembershipID:     body.MembershipID,
		Price:            body.Price,
		PaymentFrequency: body.PaymentFrequency,
		AmtPeriods:       body.AmtPeriods,
	}
}

func MapUpdateRequestToEntity(body dto.UpdateMembershipPlanRequest) entity.MembershipPlan {
	return entity.MembershipPlan{
		ID:               body.ID,
		Name:             body.Name,
		MembershipID:     body.MembershipID,
		Price:            body.Price,
		PaymentFrequency: body.PaymentFrequency,
		AmtPeriods:       body.AmtPeriods,
	}
}

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
