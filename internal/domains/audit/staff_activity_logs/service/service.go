package staff_activity_logs

import (
	"context"
	"database/sql"

	"api/internal/di"
	repo "api/internal/domains/audit/staff_activity_logs/persistence"
	values "api/internal/domains/audit/staff_activity_logs/values"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
)

type Service struct {
	repo *repo.Repository
	db   *sql.DB
}

func NewService(container *di.Container) *Service {
	return &Service{
		repo: repo.NewRepository(container),
		db:   container.DB,
	}
}

// InsertStaffActivity inserts a new staff activity log into the database.
// It is not implemented in HTTP handler currently, only as a service for other domains.
// It should be used in the context of a transaction for atomicity purposes.
func (s *Service) InsertStaffActivity(ctx context.Context, tx *sql.Tx, staffId uuid.UUID, activityDescription string) *errLib.CommonError {
	return s.repo.InsertStaffActivity(ctx, tx, staffId, activityDescription)
}

func (s *Service) GetStaffActivityLogs(ctx context.Context, staffId uuid.UUID, searchDescription string, limit, offset int32) ([]values.StaffActivityLog, *errLib.CommonError) {
	activities, err := s.repo.GetStaffActivityLogs(ctx, staffId, searchDescription, limit, offset)
	if err != nil {
		return nil, err
	}

	return activities, nil
}
