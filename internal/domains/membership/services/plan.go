package membership

import (
	"api/cmd/server/di"
	entity "api/internal/domains/membership/entities"
	persistence "api/internal/domains/membership/persistence/repositories"

	values "api/internal/domains/membership/values/plans"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type MembershipPlansService struct {
	Repo *persistence.MembershipPlansRepository
}

func NewMembershipPlansService(container *di.Container) *MembershipPlansService {
	return &MembershipPlansService{Repo: persistence.NewMembershipPlansRepository(container)}
}

func (s *MembershipPlansService) CreateMembershipPlan(ctx context.Context, plan *values.MembershipPlanCreate) *errLib.CommonError {

	return s.Repo.CreateMembershipPlan(ctx, plan)

}

func (s *MembershipPlansService) GetMembershipPlansByMembershipId(ctx context.Context, id uuid.UUID) ([]entity.MembershipPlan, *errLib.CommonError) {
	return s.Repo.GetMembershipPlansByMembershipId(ctx, id)
}

func (s *MembershipPlansService) UpdateMembershipPlan(ctx context.Context, plan *values.MembershipPlanUpdate) *errLib.CommonError {

	return s.Repo.UpdateMembershipPlan(ctx, plan)
}

func (s *MembershipPlansService) DeleteMembershipPlan(ctx context.Context, membershipId, planId uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteMembershipPlan(ctx, membershipId, planId)
}
