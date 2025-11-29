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

// lookupNames fetches human-readable names for team and location
func (s *Service) lookupNames(ctx context.Context, teamID, locationID uuid.UUID) (teamName, locationName string) {
	teamName = "Unknown Team"
	locationName = "Unknown Location"

	if teamID != uuid.Nil {
		var name sql.NullString
		_ = s.db.QueryRowContext(ctx, "SELECT name FROM athletic.teams WHERE id = $1", teamID).Scan(&name)
		if name.Valid {
			teamName = name.String
		}
	}

	if locationID != uuid.Nil {
		var name sql.NullString
		_ = s.db.QueryRowContext(ctx, "SELECT name FROM facilities.locations WHERE id = $1", locationID).Scan(&name)
		if name.Valid {
			locationName = name.String
		}
	}

	return teamName, locationName
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
		
		// Log the activity with human-readable format
		loc, _ := time.LoadLocation("America/Edmonton")
		if loc == nil {
			loc = time.FixedZone("MST", -7*60*60)
		}
		teamName, locationName := s.lookupNames(ctx, val.TeamID, val.LocationID)
		activityDesc := fmt.Sprintf("Created practice for %s at %s on %s",
			teamName,
			locationName,
			val.StartTime.In(loc).Format("Jan 2, 2006 at 3:04 PM"))
		if err := s.staffActivityLogsService.InsertStaffActivity(ctx, r.GetTx(), staffID, activityDesc); err != nil {
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
		// The incoming time already has the correct timezone from the request
		// Use America/Edmonton for Calgary timezone (handles MST/MDT automatically)
		loc, err := time.LoadLocation("America/Edmonton")
		if err != nil {
			// Fallback: use fixed MST offset (UTC-7) if timezone data unavailable
			loc = time.FixedZone("MST", -7*60*60)
		}
		localTime := practice.StartTime.In(loc)

		startTime = localTime.Format("January 2, 2006 at 3:04 PM")
		startTimeISO = localTime.Format(time.RFC3339)
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
		loc, _ := time.LoadLocation("America/Edmonton")
		if loc == nil {
			loc = time.FixedZone("MST", -7*60*60)
		}
		teamName, locationName := s.lookupNames(ctx, val.TeamID, val.LocationID)
		activityDesc := fmt.Sprintf("Updated practice for %s at %s on %s",
			teamName,
			locationName,
			val.StartTime.In(loc).Format("Jan 2, 2006 at 3:04 PM"))
		return s.staffActivityLogsService.InsertStaffActivity(ctx, r.GetTx(), staffID, activityDesc)
	})
}

func (s *Service) DeletePractice(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	// Fetch practice details before deletion for audit log
	practice, fetchErr := s.repo.GetByID(ctx, id)
	if fetchErr != nil {
		return fetchErr
	}

	teamName, locationName := s.lookupNames(ctx, practice.TeamID, practice.LocationID)

	loc, _ := time.LoadLocation("America/Edmonton")
	if loc == nil {
		loc = time.FixedZone("MST", -7*60*60)
	}

	return s.executeInTx(ctx, func(r *repo.Repository) *errLib.CommonError {
		if err := r.Delete(ctx, id); err != nil {
			return err
		}
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}

		activityDesc := fmt.Sprintf("Deleted practice for %s at %s on %s",
			teamName,
			locationName,
			practice.StartTime.In(loc).Format("Jan 2, 2006 at 3:04 PM"))

		return s.staffActivityLogsService.InsertStaffActivity(ctx, r.GetTx(), staffID, activityDesc)
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
	case contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist:
		// Admin/receptionist can see all teams
		rows, err := s.db.QueryContext(ctx, `SELECT id FROM athletic.teams`)
		if err != nil {
			return nil, errLib.New("failed to get all teams", http.StatusInternalServerError)
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
	default:
		return nil, errLib.New("role not supported", http.StatusForbidden)
	}
}
