package membership

import (
	entity "api/internal/domains/membership/entities"
	membership "api/internal/domains/membership/infra/persistence"
	"api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type MembershipService struct {
	Repo *membership.MembershipsRepository
}

func NewMembershipService(repo *membership.MembershipsRepository) *MembershipService {
	return &MembershipService{Repo: repo}
}

func (s *MembershipService) Create(ctx context.Context, input *values.MembershipCreate) *errLib.CommonError {

	if err := input.Validate(); err != nil {
		return err
	}

	membership := &entity.Membership{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
	}

	return s.Repo.Create(ctx, membership)
}

func (s *MembershipService) GetById(ctx context.Context, id uuid.UUID) (*entity.Membership, *errLib.CommonError) {
	return s.Repo.GetByID(ctx, id)
}

func (s *MembershipService) GetAll(ctx context.Context) ([]entity.Membership, *errLib.CommonError) {
	return s.Repo.List(ctx, "")
}

func (s *MembershipService) Update(ctx context.Context, membership *values.MembershipUpdate) *errLib.CommonError {

	if err := membership.Validate(); err != nil {
		return err
	}

	return s.Repo.Update(ctx, membership)
}

func (s *MembershipService) Delete(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.Delete(ctx, id)
}
