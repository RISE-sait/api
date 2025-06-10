package playground

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/playground/persistence/sqlc/generated"
	values "api/internal/domains/playground/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
)
// Repository provides methods to interact with the database for the playground domain.
type Repository struct {
	Queries *db.Queries
	Tx      *sql.Tx
}
// GetTx returns the current transaction of the repository.
func (r *Repository) GetTx() *sql.Tx {
	return r.Tx
}
// WithTx returns a new Repository with the provided transaction.
func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{
		Queries: r.Queries.WithTx(tx),
		Tx:      tx,
	}
}
// NewRepository initializes a new Repository with the provided DI container.
func NewRepository(container *di.Container) *Repository {
	return &Repository{Queries: container.Queries.PlaygroundDb}
}
// CreateSession creates a new session in the database.
func (r *Repository) CreateSession(ctx context.Context, v values.CreateSessionValue) (values.Session, *errLib.CommonError) {
	params := db.CreateSessionParams{
		SystemID:   v.SystemID,
		CustomerID: v.CustomerID,
		StartTime:  v.StartTime,
		EndTime:    v.EndTime,
	}
	row, err := r.Queries.CreateSession(ctx, params)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return values.Session{}, errLib.New("A session at this schedule overlaps with an existing session", http.StatusConflict)
		}
		if errors.As(err, &pqErr) {
			switch pqErr.Constraint {
			case "fk_system_id", "sessions_system_id_fkey":
				return values.Session{}, errLib.New("System with the associated ID doesn't exist", http.StatusNotFound)
			case "fk_customer", "sessions_customer_id_fkey":
				return values.Session{}, errLib.New("Customer with the associated ID doesn't exist", http.StatusNotFound)
			case "check_end_time":
				return values.Session{}, errLib.New("end_time must be after start_time", http.StatusBadRequest)
			}
		}
		log.Println("Failed to create session:", err)
		return values.Session{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return mapDbCreateSessionToValue(row), nil
}
// GetSessions retrieves all sessions.
func (r *Repository) GetSessions(ctx context.Context) ([]values.Session, *errLib.CommonError) {
	dbSessions, err := r.Queries.GetSessions(ctx)
	if err != nil {
		log.Println("Failed to get sessions:", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	list := make([]values.Session, len(dbSessions))
	for i, s := range dbSessions {
		list[i] = mapDbSessionsRowToValue(s)
	}
	return list, nil
}
// GetSession retrieves a session by its ID.
func (r *Repository) GetSession(ctx context.Context, id uuid.UUID) (values.Session, *errLib.CommonError) {
	row, err := r.Queries.GetSession(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.Session{}, errLib.New("Session not found", http.StatusNotFound)
		}
		log.Println("Failed to get session:", err)
		return values.Session{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return mapDbSessionToValue(row), nil
}
// DeleteSession deletes a session by its ID.
func (r *Repository) DeleteSession(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	count, err := r.Queries.DeleteSession(ctx, id)
	if err != nil {
		log.Println("Failed to delete session:", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	if count == 0 {
		return errLib.New("Session not found", http.StatusNotFound)
	}
	return nil
}
// mapDbSessionToValue maps a database row to a Session value.
func mapDbSessionToValue(dbRow db.GetSessionRow) values.Session {
	return values.Session{
		ID:                dbRow.ID,
		SystemID:          dbRow.SystemID,
		SystemName:        dbRow.SystemName,
		CustomerID:        dbRow.CustomerID,
		CustomerFirstName: dbRow.CustomerFirstName,
		CustomerLastName:  dbRow.CustomerLastName,
		StartTime:         dbRow.StartTime,
		EndTime:           dbRow.EndTime,
		CreatedAt:         dbRow.CreatedAt,
		UpdatedAt:         dbRow.UpdatedAt,
	}
}
// mapDbCreateSessionToValue maps a database row to a CreateSession value.
func mapDbCreateSessionToValue(dbRow db.CreateSessionRow) values.Session {
	return values.Session{
		ID:         dbRow.ID,
		SystemID:   dbRow.SystemID,
		CustomerID: dbRow.CustomerID,
		StartTime:  dbRow.StartTime,
		EndTime:    dbRow.EndTime,
		CreatedAt:  dbRow.CreatedAt,
		UpdatedAt:  dbRow.UpdatedAt,
	}
}
// mapDbSessionsRowToValue maps a database row to a Session value.
func mapDbSessionsRowToValue(dbRow db.GetSessionsRow) values.Session {
	return values.Session{
		ID:                dbRow.ID,
		SystemID:          dbRow.SystemID,
		SystemName:        dbRow.SystemName,
		CustomerID:        dbRow.CustomerID,
		CustomerFirstName: dbRow.CustomerFirstName,
		CustomerLastName:  dbRow.CustomerLastName,
		StartTime:         dbRow.StartTime,
		EndTime:           dbRow.EndTime,
		CreatedAt:         dbRow.CreatedAt,
		UpdatedAt:         dbRow.UpdatedAt,
	}
}
