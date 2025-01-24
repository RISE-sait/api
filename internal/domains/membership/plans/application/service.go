package membership_plan

import (
	entity "api/internal/domains/membership/plans/entities"
	"api/internal/domains/membership/plans/infra/persistence"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type MembershipPlansService struct {
	Repo *persistence.MembershipPlansRepository
}

func NewFacilityManager(repo *persistence.MembershipPlansRepository) *MembershipPlansService {
	return &MembershipPlansService{Repo: repo}
}

func (s *MembershipPlansService) CreateMembershipPlan(ctx context.Context, plan *entity.MembershipPlan) *errLib.CommonError {

	return s.Repo.CreateMembershipPlan(ctx, plan)

}

func (s *MembershipPlansService) GetMembershipPlansByMembershipId(ctx context.Context, id uuid.UUID) ([]entity.MembershipPlan, *errLib.CommonError) {
	return s.Repo.GetMembershipPlansByMembershipId(ctx, id)
}

func (s *MembershipPlansService) UpdateMembershipPlan(ctx context.Context, plan *entity.MembershipPlan) *errLib.CommonError {

	return s.Repo.UpdateMembershipPlan(ctx, plan)
}

func (s *MembershipPlansService) DeleteMembershipPlan(ctx context.Context, membershipId, planId uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteMembershipPlan(ctx, membershipId, planId)
}
