package membership

import (
	"api/internal/di"
	persistence "api/internal/domains/membership/persistence/repositories"
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	Repo *persistence.Repository
}

func NewMembershipService(container *di.Container) *Service {
	return &Service{Repo: persistence.NewMembershipsRepository(container)}
}

func (s *Service) Create(ctx context.Context, input *values.CreateValues) *errLib.CommonError {

	membership := &values.CreateValues{
		Name:        input.Name,
		Description: input.Description,
	}

	return s.Repo.Create(ctx, membership)
}

func (s *Service) GetById(ctx context.Context, id uuid.UUID) (*values.ReadValues, *errLib.CommonError) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) GetMemberships(ctx context.Context) ([]values.ReadValues, *errLib.CommonError) {
	return s.Repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, membership *values.UpdateValues) *errLib.CommonError {
	return s.Repo.Update(ctx, membership)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.Delete(ctx, id)
}
