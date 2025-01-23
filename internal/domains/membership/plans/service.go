package membership_plan

import (
	"api/internal/domains/membership/plans/dto"
	infra "api/internal/domains/membership/plans/infra"
	db "api/internal/domains/membership/plans/infra/sqlc/generated"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	Repo *infra.Repo
}

func NewService(repo *infra.Repo) *Service {
	return &Service{Repo: repo}
}

func (s *Service) CreateMembershipPlan(ctx context.Context, body dto.CreateMembershipPlanRequest) *errLib.CommonError {

	if err := validators.ValidateDto(&body); err != nil {
		return err
	}

	if err := s.Repo.CreateMembershipPlan(ctx, &body); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetPlansMembershipById(ctx context.Context, id uuid.UUID) (*[]dto.MembershipPlanResponse, *errLib.CommonError) {
	plans, err := s.Repo.GetMembershipPlansByMembershipId(ctx, id)
	if err != nil {
		return nil, err
	}

	result := []dto.MembershipPlanResponse{}
	for _, plan := range plans {
		result = append(result, dto.MembershipPlanResponse{
			ID:               plan.ID,
			Name:             plan.Name,
			MembershipID:     plan.MembershipID,
			Price:            plan.Price,
			PaymentFrequency: string(plan.PaymentFrequency.PaymentFrequency),
			AmtPeriods:       int(plan.AmtPeriods.Int32),
		})
	}

	return &result, nil
}

func (s *Service) UpdatePlan(ctx context.Context, body dto.UpdateMembershipPlanRequest) *errLib.CommonError {

	if err := validators.ValidateDto(&body); err != nil {
		return err
	}

	if err := s.Repo.UpdateMembershipPlan(ctx, &body); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeletePlan(ctx context.Context, membershipId, planId uuid.UUID) *errLib.CommonError {

	plan := &db.DeleteMembershipPlanParams{
		MembershipID: membershipId,
		ID:           planId,
	}

	if err := s.Repo.DeleteMembershipPlan(ctx, plan); err != nil {
		return err
	}

	return nil
}
