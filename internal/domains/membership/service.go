package membership

import (
	"api/cmd/server/di"
	persistence "api/internal/domains/membership/persistence"
	"api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type MembershipService struct {
	Repo *persistence.MembershipsRepository
}

func NewMembershipService(container *di.Container) *MembershipService {
	return &MembershipService{Repo: persistence.NewMembershipsRepository(container)}
}

func (s *MembershipService) Create(ctx context.Context, input *values.MembershipDetails) *errLib.CommonError {

	membership := &values.MembershipDetails{
		Name:        input.Name,
		Description: input.Description,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
	}

	return s.Repo.Create(ctx, membership)
}

func (s *MembershipService) GetById(ctx context.Context, id uuid.UUID) (*values.MembershipAllFields, *errLib.CommonError) {
	return s.Repo.GetByID(ctx, id)
}

func (s *MembershipService) GetAll(ctx context.Context) ([]values.MembershipAllFields, *errLib.CommonError) {
	return s.Repo.List(ctx, "")
}

func (s *MembershipService) Update(ctx context.Context, membership *values.MembershipAllFields) *errLib.CommonError {
	return s.Repo.Update(ctx, membership)
}

func (s *MembershipService) Delete(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.Delete(ctx, id)
}
