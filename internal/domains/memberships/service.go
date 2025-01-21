package memberships

import (
	dto "api/internal/shared/dto/memberships"
	"api/internal/types"
	db "api/sqlc"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) CreateMembership(ctx context.Context, req dto.CreateMembershipRequest) *types.HTTPError {
	params := req.ToDBParams()

	err := s.Repo.CreateMembership(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetMembershipById(ctx context.Context, id uuid.UUID) (*db.Membership, *types.HTTPError) {
	membership, err := s.Repo.GetMembershipById(ctx, id)
	if err != nil {
		return nil, err
	}
	return membership, nil
}

func (s *Service) GetAllMemberships(ctx context.Context) (*[]db.Membership, *types.HTTPError) {
	memberships, err := s.Repo.GetAllMemberships(ctx)
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (s *Service) UpdateMembership(ctx context.Context, req dto.UpdateMembershipRequest) *types.HTTPError {
	params := req.ToDBParams()

	err := s.Repo.UpdateMembership(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteMembership(ctx context.Context, id uuid.UUID) *types.HTTPError {
	err := s.Repo.DeleteMembership(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
