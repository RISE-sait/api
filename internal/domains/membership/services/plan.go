package membership

import (
	"api/internal/di"
	persistence "api/internal/domains/membership/persistence/repositories"
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type PlansService struct {
	Repo *persistence.PlansRepository
}

func NewMembershipPlansService(container *di.Container) *PlansService {
	return &PlansService{Repo: persistence.NewMembershipPlansRepository(container)}
}

func (s *PlansService) CreateMembershipPlan(ctx context.Context, plan *values.PlanCreateValues) *errLib.CommonError {

	return s.Repo.CreateMembershipPlan(ctx, plan)

}

func (s *PlansService) GetMembershipPlans(ctx context.Context, customerId, membershipId uuid.UUID) ([]values.PlanReadValues, *errLib.CommonError) {
	return s.Repo.GetMembershipPlans(ctx, membershipId, customerId)
}

func (s *PlansService) GetMembershipPlanById(ctx context.Context, id uuid.UUID) (*values.PlanReadValues, *errLib.CommonError) {
	return s.Repo.GetMembershipPlanById(ctx, id)
}

func (s *PlansService) UpdateMembershipPlan(ctx context.Context, plan *values.PlanUpdateValues) *errLib.CommonError {

	return s.Repo.UpdateMembershipPlan(ctx, plan)
}

func (s *PlansService) DeleteMembershipPlan(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteMembershipPlan(ctx, id)
}
