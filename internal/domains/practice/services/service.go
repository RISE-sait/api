package practice

import (
	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	notificationService "api/internal/domains/notification/services"
	notificationValues "api/internal/domains/notification/values"
	repo "api/internal/domains/practice/persistence"
	values "api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	txUtils "api/utils/db"
	"context"
	"database/sql"
	"fmt"

	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo                     *repo.Repository
	staffActivityLogsService *staffActivityLogs.Service
	notificationService      *notificationService.NotificationService
	db                       *sql.DB
}

func NewService(container *di.Container) *Service {
	return &Service{
		repo:                     repo.NewRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		notificationService:      notificationService.NewNotificationService(container),
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
		
		// Log the activity
		if err := s.staffActivityLogsService.InsertStaffActivity(ctx, r.GetTx(), staffID, fmt.Sprintf("Created practice %+v", val)); err != nil {
			return err
		}
		
		// Send automatic notification to team if TeamID is provided
		if val.TeamID != uuid.Nil {
			go s.sendPracticeNotification(context.Background(), val)
		}
		
		return nil
	})
}

func (s *Service) sendPracticeNotification(ctx context.Context, practice values.CreatePracticeValue) {
	startTime := "TBD"
	startTimeISO := ""
	if !practice.StartTime.IsZero() {
		// Convert UTC time from database to Mountain Time (MST/MDT)
		// Use fixed zone MDT (UTC-6) since we're currently in daylight saving time
		mdt := time.FixedZone("MDT", -6*3600) // MDT is UTC-6
		mountainTime := practice.StartTime.In(mdt)
		
		startTime = mountainTime.Format("January 2, 2006 at 3:04 PM MST")
		startTimeISO = mountainTime.Format(time.RFC3339)
	}
	
	notification := notificationValues.TeamNotification{
		Type:   "practice",
		Title:  "New Practice Scheduled",
		Body:   fmt.Sprintf("Practice on %s", startTime),
		TeamID: practice.TeamID,
		Data: map[string]interface{}{
			"practiceId": "new-practice",  // CreatePracticeValue doesn't have ID yet
			"type":       "practice",
			"startAt":    startTimeISO,
		},
	}
	
	// Send notification (errors are logged by the notification service)
	s.notificationService.SendTeamNotification(ctx, practice.TeamID, notification)
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

// GetUserPractices retrieves practices for the teams associated with the given user.
// Coaches see practices for teams they coach, athletes for the team they belong to.
func (s *Service) GetUserPractices(ctx context.Context, userID uuid.UUID, role contextUtils.CtxRole, limit, offset int32) ([]values.ReadPracticeValue, *errLib.CommonError) {
	teamIDs, err := s.getUserTeamIDs(ctx, userID, role)
	if err != nil {
		return nil, err
	}
	if len(teamIDs) == 0 {
		return []values.ReadPracticeValue{}, nil
	}
	var result []values.ReadPracticeValue
	for _, id := range teamIDs {
		practices, err := s.repo.List(ctx, id, limit, offset)
		if err != nil {
			return nil, err
		}
		result = append(result, practices...)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartTime.Before(result[j].StartTime) })
	return result, nil
}

func (s *Service) getUserTeamIDs(ctx context.Context, userID uuid.UUID, role contextUtils.CtxRole) ([]uuid.UUID, *errLib.CommonError) {
	switch role {
	case contextUtils.RoleCoach:
		rows, err := s.db.QueryContext(ctx, `SELECT id FROM athletic.teams WHERE coach_id = $1`, userID)
		if err != nil {
			return nil, errLib.New("failed to get coach teams", http.StatusInternalServerError)
		}
		defer rows.Close()
		var ids []uuid.UUID
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				return nil, errLib.New("failed to scan team id", http.StatusInternalServerError)
			}
			ids = append(ids, id)
		}
		return ids, nil
	case contextUtils.RoleAthlete:
		var id uuid.UUID
		err := s.db.QueryRowContext(ctx, `SELECT team_id FROM athletic.athletes WHERE id = $1`, userID).Scan(&id)
		if err == sql.ErrNoRows {
			return []uuid.UUID{}, nil
		}
		if err != nil {
			return nil, errLib.New("failed to get athlete team", http.StatusInternalServerError)
		}
		if id == uuid.Nil {
			return []uuid.UUID{}, nil
		}
		return []uuid.UUID{id}, nil
	default:
		return nil, errLib.New("role not supported", http.StatusForbidden)
	}
}
