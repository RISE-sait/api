package membership_plans

import (
	"api/internal/types"
	db "api/sqlc"
	"context"

	"github.com/google/uuid"

	dto "api/internal/shared/dto/membershipPlans"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) CreateMembershipPlan(ctx context.Context, req dto.CreateMembershipPlanRequest) *types.HTTPError {
	params := req.ToDBParams()

	err := s.Repo.CreateMembershipPlan(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetMembershipPlanDetails(ctx context.Context, membershipID, planID uuid.UUID) (*[]dto.MembershipPlanResponse, *types.HTTPError) {
	plans, err := s.Repo.GetMembershipPlanDetails(ctx, membershipID, planID)
	if err != nil {
		return nil, err
	}
	var responses []dto.MembershipPlanResponse
	for _, plan := range *plans {
		responses = append(responses, *dto.ToMembershipPlanResponse(&plan))
	}

	return &responses, nil
}

func (s *Service) UpdateMembershipPlan(ctx context.Context, req dto.UpdateMembershipPlanRequest) *types.HTTPError {
	params := req.ToDBParams()

	err := s.Repo.UpdateMembershipPlan(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteMembershipPlan(ctx context.Context, membershipID, planID uuid.UUID) *types.HTTPError {
	params := db.DeleteMembershipPlanParams{
		MembershipID: membershipID,
		ID:           planID,
	}

	err := s.Repo.DeleteMembershipPlan(ctx, &params)
	if err != nil {
		return err
	}
	return nil
}
