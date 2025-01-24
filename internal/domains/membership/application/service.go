package membership

import (
	entity "api/internal/domains/membership/entities"
	"api/internal/domains/membership/infra/persistence"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type MembershipService struct {
	Repo *persistence.MembershipsRepository
}

func NewMembershipService(repo *persistence.MembershipsRepository) *MembershipService {
	return &MembershipService{Repo: repo}
}

func (s *MembershipService) Create(ctx context.Context, membership *entity.Membership) *errLib.CommonError {

	return s.Repo.Create(ctx, membership)
}

func (s *MembershipService) GetById(ctx context.Context, id uuid.UUID) (*entity.Membership, *errLib.CommonError) {
	return s.Repo.GetByID(ctx, id)
}

func (s *MembershipService) GetAll(ctx context.Context) ([]entity.Membership, *errLib.CommonError) {
	return s.Repo.List(ctx, "")
}

func (s *MembershipService) Update(ctx context.Context, membership *entity.Membership) *errLib.CommonError {

	return s.Repo.Update(ctx, membership)
}

func (s *MembershipService) Delete(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.Delete(ctx, id)
}
