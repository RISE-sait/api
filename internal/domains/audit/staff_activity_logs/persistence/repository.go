package staff_activity_logs

import (
	"api/internal/di"
	db "api/internal/domains/audit/staff_activity_logs/persistence/sqlc/generated"
	values "api/internal/domains/audit/staff_activity_logs/values"
	errLib "api/internal/libs/errors"
	"database/sql"

	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
}

func NewRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.StaffActivityLogsDb,
	}
}

func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{
		Queries: r.Queries.WithTx(tx),
	}
}

func (r *Repository) InsertStaffActivity(ctx context.Context, tx *sql.Tx, staffId uuid.UUID, activityDescription string) *errLib.CommonError {

	if err := r.Queries.WithTx(tx).InsertStaffActivity(ctx, db.InsertStaffActivityParams{
		StaffID:             staffId,
		ActivityDescription: activityDescription,
	}); err != nil {
		log.Printf("Failed to insert staff activity for staff: %+v. Error: %v", staffId, err.Error())
		return errLib.New("Internal server error when inserting staff activity", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) GetStaffActivityLogs(ctx context.Context, staffId uuid.UUID, searchDescription string, limit, offset int32) ([]values.StaffActivityLog, *errLib.CommonError) {

	activities, err := r.Queries.GetStaffActivityLogs(ctx, db.GetStaffActivityLogsParams{
		StaffID: uuid.NullUUID{
			UUID:  staffId,
			Valid: staffId != uuid.Nil,
		},
		SearchDescription: sql.NullString{
			String: searchDescription,
			Valid:  searchDescription != "",
		},
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		log.Printf("Failed to get staff activity logs for staff: %+v. Error: %v", staffId, err.Error())
		return nil, errLib.New("Internal server error when getting staff activity logs", http.StatusInternalServerError)
	}

	activityLogs := make([]values.StaffActivityLog, len(activities))

	for i, activity := range activities {
		activityLogs[i] = values.StaffActivityLog{
			ID:                  activity.ID,
			StaffID:             activity.StaffID,
			ActivityDescription: activity.ActivityDescription,
			CreatedAt:           activity.CreatedAt,
			FirstName:           activity.FirstName,
			LastName:            activity.LastName,
			Email:               activity.Email.String,
		}
	}

	return activityLogs, nil

}
