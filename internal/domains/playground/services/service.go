package playground

import (
	"api/internal/di"
	repo "api/internal/domains/playground/persistence"
	values "api/internal/domains/playground/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo *repo.Repository
}
// NewService initializes a new Service for the playground domain.
func NewService(container *di.Container) *Service {
	return &Service{repo: repo.NewRepository(container)}
}
// CreateSession creates a new session in the playground domain.
func (s *Service) CreateSession(ctx context.Context, v values.CreateSessionValue) (values.Session, *errLib.CommonError) {
	return s.repo.CreateSession(ctx, v)
}
// GetSessions retrieves all sessions in the playground domain.
func (s *Service) GetSessions(ctx context.Context) ([]values.Session, *errLib.CommonError) {
	return s.repo.GetSessions(ctx)
}
// GetSession retrieves a specific session by its ID.
func (s *Service) GetSession(ctx context.Context, id uuid.UUID) (values.Session, *errLib.CommonError) {
	return s.repo.GetSession(ctx, id)
}
// UpdateSession updates an existing session in the playground domain.
func (s *Service) DeleteSession(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.repo.DeleteSession(ctx, id)
}
