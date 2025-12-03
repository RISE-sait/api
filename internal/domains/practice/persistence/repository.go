package practice

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"api/internal/di"
	db "api/internal/domains/practice/persistence/sqlc/generated"
	values "api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
)

// Repository provides DB operations for practices using sqlc generated queries.
type Repository struct {
	Queries *db.Queries
	Tx      *sql.Tx
}

// NewRepository creates a Repository using the container's sqlc queries.
func NewRepository(container *di.Container) *Repository {
	return &Repository{Queries: container.Queries.PracticeDb}
}

// WithTx returns a repository operating in the given transaction.
func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{Queries: r.Queries.WithTx(tx), Tx: tx}
}

// GetTx exposes the current transaction, if any.
func (r *Repository) GetTx() *sql.Tx {
	return r.Tx
}

func toNullTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{Valid: false}
}

func toNullString(s string) sql.NullString {
	if s != "" {
		return sql.NullString{String: s, Valid: true}
	}
	return sql.NullString{Valid: false}
}

func toNullUUID(u *uuid.UUID) uuid.NullUUID {
	if u != nil {
		return uuid.NullUUID{UUID: *u, Valid: true}
	}
	return uuid.NullUUID{Valid: false}
}

func unwrapNullUUID(nu uuid.NullUUID) *uuid.UUID {
	if nu.Valid {
		return &nu.UUID
	}
	return nil
}

func unwrapNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func unwrapNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func unwrapInterfaceToString(i interface{}) string {
	if i == nil {
		return ""
	}
	if str, ok := i.(string); ok {
		return str
	}
	return ""
}

// Create inserts a new practice.
func (r *Repository) Create(ctx context.Context, val values.CreatePracticeValue) *errLib.CommonError {
	params := db.CreatePracticeParams{
		TeamID:     val.TeamID,
		StartTime:  val.StartTime,
		EndTime:    toNullTime(val.EndTime),
		CourtID:    uuid.NullUUID{UUID: val.CourtID, Valid: val.CourtID != uuid.Nil},
		LocationID: val.LocationID,
		Status:     toNullString(val.Status),
		BookedBy:   toNullUUID(val.BookedBy),
	}
	if err := r.Queries.CreatePractice(ctx, params); err != nil {
		// Check if it's an exclusion constraint violation (double booking within practices)
		if strings.Contains(err.Error(), "no_overlapping_practices") {
			return errLib.New("This court is already booked during the selected time slot", http.StatusConflict)
		}
		// Check if it's a cross-table booking conflict (from trigger)
		if strings.Contains(err.Error(), "Court is already booked") {
			return errLib.New(err.Error(), http.StatusConflict)
		}
		return errLib.New("failed to create practice", http.StatusInternalServerError)
	}
	return nil
}

// Update modifies an existing practice.
func (r *Repository) Update(ctx context.Context, val values.UpdatePracticeValue) *errLib.CommonError {
	params := db.UpdatePracticeParams{
		TeamID:     val.TeamID,
		StartTime:  val.StartTime,
		EndTime:    toNullTime(val.EndTime),
		CourtID:    uuid.NullUUID{UUID: val.CourtID, Valid: val.CourtID != uuid.Nil},
		LocationID: val.LocationID,
		Status:     toNullString(val.Status),
		BookedBy:   toNullUUID(val.BookedBy),
		ID:         val.ID,
	}
	if err := r.Queries.UpdatePractice(ctx, params); err != nil {
		// Check if it's an exclusion constraint violation (double booking)
		if strings.Contains(err.Error(), "no_overlapping_practices") {
			return errLib.New("This court is already booked during the selected time slot", http.StatusConflict)
		}
		return errLib.New("failed to update practice", http.StatusInternalServerError)
	}
	return nil
}

// Delete removes a practice by ID.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	if err := r.Queries.DeletePractice(ctx, id); err != nil {
		return errLib.New("failed to delete practice", http.StatusInternalServerError)
	}
	return nil
}

// GetByID fetches a practice by ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (values.ReadPracticeValue, *errLib.CommonError) {
	row, err := r.Queries.GetPracticeByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadPracticeValue{}, errLib.New("practice not found", http.StatusNotFound)
		}
		return values.ReadPracticeValue{}, errLib.New("failed to get practice", http.StatusInternalServerError)
	}
	return mapDbPracticeToValue(row), nil
}

func mapDbPracticeToValue(row db.GetPracticeByIDRow) values.ReadPracticeValue {
	return values.ReadPracticeValue{
		ID:           row.ID,
		TeamID:       row.TeamID,
		TeamName:     row.TeamName,
		TeamLogoUrl:  unwrapNullString(row.TeamLogoUrl),
		StartTime:    row.StartTime,
		EndTime:      unwrapNullTime(row.EndTime),
		LocationID:   row.LocationID,
		LocationName: row.LocationName,
		CourtID:      row.CourtID.UUID,
		CourtName:    unwrapNullString(row.CourtName),
		Status:       unwrapNullString(row.Status),
		BookedBy:     unwrapNullUUID(row.BookedBy),
		BookedByName: unwrapInterfaceToString(row.BookedByName),
		CreatedAt:    unwrapNullTime(row.CreatedAt),
		UpdatedAt:    unwrapNullTime(row.UpdatedAt),
	}
}

// List returns practices filtered by team ID. If teamID is uuid.Nil all practices are returned.
func (r *Repository) List(ctx context.Context, teamID uuid.UUID, limit, offset int32) ([]values.ReadPracticeValue, *errLib.CommonError) {
	param := db.ListPracticesParams{
		TeamID: uuid.NullUUID{UUID: teamID, Valid: teamID != uuid.Nil},
		Limit:  limit,
		Offset: offset,
	}
	rows, err := r.Queries.ListPractices(ctx, param)
	if err != nil {
		return nil, errLib.New("failed to list practices", http.StatusInternalServerError)
	}
	res := make([]values.ReadPracticeValue, len(rows))
	for i, row := range rows {
		res[i] = mapDbPracticeToValue(db.GetPracticeByIDRow(row))
	}
	return res, nil
}
