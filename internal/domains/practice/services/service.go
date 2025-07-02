package practice

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/practice/persistence"
	values "api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	txUtils "api/utils/db"

	"github.com/google/uuid"
)

type Service struct {
	repo                     *repo.Repository
	staffActivityLogsService *staffActivityLogs.Service
	db                       *sql.DB
}

func NewService(container *di.Container) *Service {
	return &Service{
		repo:                     repo.NewRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		db:                       container.DB,
	}
}

func (s *Service) executeInTx(ctx context.Context, fn func(repo *repo.Repository) *errLib.CommonError) *errLib.CommonError {
	return txUtils.ExecuteInTx(ctx, s.db, func(tx *sql.Tx) *errLib.CommonError {
		txRepo := s.repo.WithTx(tx)
		return fn(txRepo)
	})
}

func (s *Service) CreatePractice(ctx context.Context, val values.CreatePracticeValue) *errLib.CommonError {
	return s.executeInTx(ctx, func(r *repo.Repository) *errLib.CommonError {
		if err := r.Create(ctx, val); err != nil {
			return err
		}
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}
		return s.staffActivityLogsService.InsertStaffActivity(ctx, r.GetTx(), staffID, fmt.Sprintf("Created practice %+v", val))
	})
}

func (s *Service) UpdatePractice(ctx context.Context, val values.UpdatePracticeValue) *errLib.CommonError {
	return s.executeInTx(ctx, func(r *repo.Repository) *errLib.CommonError {
		if err := r.Update(ctx, val); err != nil {
			return err
		}
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}
		return s.staffActivityLogsService.InsertStaffActivity(ctx, r.GetTx(), staffID, fmt.Sprintf("Updated practice %+v", val))
	})
}

func (s *Service) DeletePractice(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.executeInTx(ctx, func(r *repo.Repository) *errLib.CommonError {
		if err := r.Delete(ctx, id); err != nil {
			return err
		}
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}
		return s.staffActivityLogsService.InsertStaffActivity(ctx, r.GetTx(), staffID, fmt.Sprintf("Deleted practice %s", id))
	})
}

func (s *Service) GetPractice(ctx context.Context, id uuid.UUID) (values.ReadPracticeValue, *errLib.CommonError) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetPractices(ctx context.Context, teamID uuid.UUID, limit, offset int32) ([]values.ReadPracticeValue, *errLib.CommonError) {
	return s.repo.List(ctx, teamID, limit, offset)
}

func (s *Service) CreateRecurringPractices(ctx context.Context, rec values.RecurrenceValues, base values.CreatePracticeValue) *errLib.CommonError {
	practices, err := generatePracticesFromRecurrence(rec, base)
	if err != nil {
		return err
	}
	return s.executeInTx(ctx, func(r *repo.Repository) *errLib.CommonError {
		for _, p := range practices {
			if err := r.Create(ctx, p); err != nil {
				return err
			}
		}
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}
		return s.staffActivityLogsService.InsertStaffActivity(ctx, r.GetTx(), staffID, "Created recurring practices")
	})
}

func generatePracticesFromRecurrence(rec values.RecurrenceValues, base values.CreatePracticeValue) ([]values.CreatePracticeValue, *errLib.CommonError) {
	const layout = "15:04:05Z07:00"
	startTime, err := time.Parse(layout, rec.StartTime)
	if err != nil {
		return nil, errLib.New("invalid start time", http.StatusBadRequest)
	}
	endTime, err := time.Parse(layout, rec.EndTime)
	if err != nil {
		return nil, errLib.New("invalid end time", http.StatusBadRequest)
	}
	if endTime.Before(startTime) {
		endTime = endTime.AddDate(0, 0, 1)
	}
	adjust := func(date time.Time, t time.Time) time.Time {
		return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location())
	}
	var practices []values.CreatePracticeValue
	for date := rec.FirstOccurrence; !date.After(rec.LastOccurrence); date = date.AddDate(0, 0, 1) {
		if date.Weekday() == rec.DayOfWeek {
			st := adjust(date, startTime)
			et := adjust(date, endTime)
			if et.Before(st) {
				et = et.AddDate(0, 0, 1)
			}
			p := base
			p.StartTime = st
			p.EndTime = &et
			practices = append(practices, p)
			date = date.AddDate(0, 0, 6)
		}
	}
	return practices, nil
}
