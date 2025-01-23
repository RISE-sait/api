package membership

import (
	dto "api/internal/domains/membership/dto"
	membership "api/internal/domains/membership/infra"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	Repo *membership.MembershipsRepository
}

func NewService(repository *membership.MembershipsRepository) *Service {
	return &Service{Repo: repository}
}

func (s *Service) CreateMembership(ctx context.Context, body dto.CreateMembershipRequest) *errLib.CommonError {
	if err := validators.ValidateDto(&body); err != nil {
		return err
	}

	if err := s.Repo.CreateMembership(ctx, &body); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetMembershipById(ctx context.Context, id uuid.UUID) (*dto.MembershipResponse, *errLib.CommonError) {
	membership, err := s.Repo.GetMembershipById(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.MembershipResponse{
		ID:          membership.ID,
		Name:        membership.Name,
		Description: membership.Description.String,
		StartDate:   membership.StartDate,
		EndDate:     membership.EndDate,
	}, nil
}

func (s *Service) GetAllMemberships(ctx context.Context) (*[]dto.MembershipResponse, *errLib.CommonError) {
	memberships, err := s.Repo.GetAllMemberships(ctx)
	if err != nil {
		return nil, err
	}

	var results []dto.MembershipResponse
	for _, membership := range memberships {
		results = append(results, dto.MembershipResponse{
			ID:          membership.ID,
			Name:        membership.Name,
			Description: membership.Description.String,
			StartDate:   membership.StartDate,
			EndDate:     membership.EndDate,
		})
	}

	return &results, nil
}

func (s *Service) UpdateMembership(ctx context.Context, body dto.UpdateMembershipRequest) *errLib.CommonError {

	if err := validators.ValidateDto(&body); err != nil {
		return err
	}

	if err := s.Repo.UpdateMembership(ctx, &body); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteMembership(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	if err := s.Repo.DeleteMembership(ctx, id); err != nil {
		return err
	}

	return nil
}
