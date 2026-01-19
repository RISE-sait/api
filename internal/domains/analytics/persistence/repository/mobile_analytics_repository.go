package repository

import (
	"context"
	"database/sql"
	"time"

	"api/internal/di"
	db "api/internal/domains/analytics/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
)

type MobileAnalyticsRepository struct {
	queries *db.Queries
}

func NewMobileAnalyticsRepository(container *di.Container) *MobileAnalyticsRepository {
	return &MobileAnalyticsRepository{
		queries: container.Queries.AnalyticsDb,
	}
}

func (r *MobileAnalyticsRepository) RecordMobileLogin(ctx context.Context, userID uuid.UUID) *errLib.CommonError {
	rowsAffected, err := r.queries.RecordMobileLogin(ctx, userID)
	if err != nil {
		return errLib.New("Failed to record mobile login", 500)
	}
	if rowsAffected == 0 {
		return errLib.New("No user found with the provided ID", 404)
	}
	return nil
}

type MobileUsageStats struct {
	TotalMobileUsers  int64 `json:"total_mobile_users"`
	ActiveToday       int64 `json:"active_today"`
	ActiveLast7Days   int64 `json:"active_last_7_days"`
	ActiveLast30Days  int64 `json:"active_last_30_days"`
	TotalUsers        int64 `json:"total_users"`
	MobileAdoptionPct float64 `json:"mobile_adoption_pct"`
}

func (r *MobileAnalyticsRepository) GetMobileUsageStats(ctx context.Context) (MobileUsageStats, *errLib.CommonError) {
	row, err := r.queries.GetMobileUsageStats(ctx)
	if err != nil {
		return MobileUsageStats{}, errLib.New("Failed to get mobile usage stats", 500)
	}

	stats := MobileUsageStats{
		TotalMobileUsers: row.TotalMobileUsers,
		ActiveToday:      row.ActiveToday,
		ActiveLast7Days:  row.ActiveLast7Days,
		ActiveLast30Days: row.ActiveLast30Days,
		TotalUsers:       row.TotalUsers,
	}

	if stats.TotalUsers > 0 {
		stats.MobileAdoptionPct = float64(stats.TotalMobileUsers) / float64(stats.TotalUsers) * 100
	}

	return stats, nil
}

type RecentMobileLogin struct {
	ID              uuid.UUID  `json:"id"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	Email           *string    `json:"email"`
	LastMobileLogin *time.Time `json:"last_mobile_login_at"`
}

func (r *MobileAnalyticsRepository) GetRecentMobileLogins(ctx context.Context, limit, offset int32) ([]RecentMobileLogin, *errLib.CommonError) {
	rows, err := r.queries.GetRecentMobileLogins(ctx, db.GetRecentMobileLoginsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, errLib.New("Failed to get recent mobile logins", 500)
	}

	logins := make([]RecentMobileLogin, 0, len(rows))
	for _, row := range rows {
		var email *string
		if row.Email.Valid {
			email = &row.Email.String
		}

		var lastLogin *time.Time
		if row.LastMobileLoginAt.Valid {
			t := row.LastMobileLoginAt.Time
			lastLogin = &t
		}

		logins = append(logins, RecentMobileLogin{
			ID:              row.ID,
			FirstName:       row.FirstName,
			LastName:        row.LastName,
			Email:           email,
			LastMobileLogin: lastLogin,
		})
	}

	return logins, nil
}

type DailyMobileLogins struct {
	Date        time.Time `json:"date"`
	UniqueUsers int64     `json:"unique_users"`
}

func (r *MobileAnalyticsRepository) GetMobileLoginsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]DailyMobileLogins, *errLib.CommonError) {
	rows, err := r.queries.GetMobileLoginsByDateRange(ctx, db.GetMobileLoginsByDateRangeParams{
		LastMobileLoginAt:   sql.NullTime{Time: startDate, Valid: true},
		LastMobileLoginAt_2: sql.NullTime{Time: endDate, Valid: true},
	})
	if err != nil {
		return nil, errLib.New("Failed to get mobile logins by date range", 500)
	}

	logins := make([]DailyMobileLogins, 0, len(rows))
	for _, row := range rows {
		logins = append(logins, DailyMobileLogins{
			Date:        row.LoginDate,
			UniqueUsers: row.UniqueUsers,
		})
	}

	return logins, nil
}
