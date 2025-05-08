package game

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/game/persistence/sqlc/generated"
	values "api/internal/domains/game/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Repository wraps SQL queries and transaction context for the Game domain.
type Repository struct {
	Queries *db.Queries
	Tx      *sql.Tx
}

// GetTx returns the current SQL transaction, if any.
func (r *Repository) GetTx() *sql.Tx {
	return r.Tx
}

// WithTx returns a new Repository instance bound to the given SQL transaction.
func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{
		Queries: r.Queries.WithTx(tx),
		Tx:      tx,
	}
}

// NewGameRepository initializes a Repository using the provided DI container.
func NewGameRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.GameDb,
	}
}

// Helper: Converts sql.NullTime to *time.Time.
func nullableTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// Helper: Converts sql.NullInt32 to *int32.
func nullableInt32ToPtr(n sql.NullInt32) *int32 {
	if n.Valid {
		return &n.Int32
	}
	return nil
}

// Helper: Converts *int32 to sql.NullInt32.
func toNullInt32(ptr *int32) sql.NullInt32 {
	if ptr != nil {
		return sql.NullInt32{Int32: *ptr, Valid: true}
	}
	return sql.NullInt32{Valid: false}
}

// Helper: Converts *time.Time to sql.NullTime.
func toNullTime(ptr *time.Time) sql.NullTime {
	if ptr != nil {
		return sql.NullTime{Time: *ptr, Valid: true}
	}
	return sql.NullTime{Valid: false}
}

// Helper: Converts a non-empty string to sql.NullString.
func toNullString(s string) sql.NullString {
	if s != "" {
		return sql.NullString{String: s, Valid: true}
	}
	return sql.NullString{Valid: false}
}

// GetGameById fetches a single game record by ID and maps it to a domain value.
// Returns a 404 error if not found.
func (r *Repository) GetGameById(ctx context.Context, id uuid.UUID) (values.ReadGameValue, *errLib.CommonError) {
	dbGame, err := r.Queries.GetGameById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadGameValue{}, errLib.New("Game not found", http.StatusNotFound)
		}
		log.Printf("Error getting game: %v", err)
		return values.ReadGameValue{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.ReadGameValue{
		ID:           dbGame.ID,
		HomeTeamID:   dbGame.HomeTeamID,
		HomeTeamName: dbGame.HomeTeamName,
		AwayTeamID:   dbGame.AwayTeamID,
		AwayTeamName: dbGame.AwayTeamName,
		HomeScore:    nullableInt32ToPtr(dbGame.HomeScore),
		AwayScore:    nullableInt32ToPtr(dbGame.AwayScore),
		StartTime:    dbGame.StartTime,
		EndTime:      nullableTimeToPtr(dbGame.EndTime),
		LocationID:   dbGame.LocationID,
		LocationName: dbGame.LocationName,
		Status:       dbGame.Status.String,
		CreatedAt:    nullableTimeToPtr(dbGame.CreatedAt),
		UpdatedAt:    nullableTimeToPtr(dbGame.UpdatedAt),
	}, nil
}

// UpdateGame updates an existing game in the database using the given value object.
// Returns 404 if no rows were affected.
func (r *Repository) UpdateGame(ctx context.Context, value values.UpdateGameValue) *errLib.CommonError {
	params := db.UpdateGameParams{
		ID:         value.ID,
		HomeScore:  toNullInt32(value.HomeScore),
		AwayScore:  toNullInt32(value.AwayScore),
		StartTime:  value.StartTime,
		EndTime:    toNullTime(value.EndTime),
		LocationID: value.LocationID,
		Status:     toNullString(value.Status),
	}

	affectedRows, err := r.Queries.UpdateGame(ctx, params)
	if err != nil {
		log.Println("Error updating game:", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	if affectedRows == 0 {
		return errLib.New("Game not found", http.StatusNotFound)
	}
	return nil
}

// GetGames fetches all games and maps them to domain values.
func (r *Repository) GetGames(ctx context.Context) ([]values.ReadGameValue, *errLib.CommonError) {
	dbGames, err := r.Queries.GetGames(ctx)
	if err != nil {
		log.Println("Error getting games:", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	games := make([]values.ReadGameValue, len(dbGames))
	for i, dbGame := range dbGames {
		games[i] = values.ReadGameValue{
			ID:           dbGame.ID,
			HomeTeamID:   dbGame.HomeTeamID,
			HomeTeamName: dbGame.HomeTeamName,
			AwayTeamID:   dbGame.AwayTeamID,
			AwayTeamName: dbGame.AwayTeamName,
			HomeScore:    nullableInt32ToPtr(dbGame.HomeScore),
			AwayScore:    nullableInt32ToPtr(dbGame.AwayScore),
			StartTime:    dbGame.StartTime,
			EndTime:      nullableTimeToPtr(dbGame.EndTime),
			LocationID:   dbGame.LocationID,
			LocationName: dbGame.LocationName,
			Status:       dbGame.Status.String,
			CreatedAt:    nullableTimeToPtr(dbGame.CreatedAt),
			UpdatedAt:    nullableTimeToPtr(dbGame.UpdatedAt),
		}
	}

	return games, nil
}

// DeleteGame deletes a game by its ID. Returns a 404 error if no rows were deleted.
func (r *Repository) DeleteGame(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	rowCount, err := r.Queries.DeleteGame(ctx, id)
	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	if rowCount == 0 {
		return errLib.New("Game not found", http.StatusNotFound)
	}
	return nil
}

// CreateGame inserts a new game into the database.
// Handles unique constraint violations and wraps errors in common format.
func (r *Repository) CreateGame(ctx context.Context, details values.CreateGameValue) *errLib.CommonError {
	params := db.CreateGameParams{
		HomeTeamID: details.HomeTeamID,
		AwayTeamID: details.AwayTeamID,
		HomeScore:  toNullInt32(details.HomeScore),
		AwayScore:  toNullInt32(details.AwayScore),
		StartTime:  details.StartTime,
		EndTime:    toNullTime(details.EndTime),
		LocationID: details.LocationID,
		Status:     toNullString(details.Status),
	}

	err := r.Queries.CreateGame(ctx, params)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("Duplicate game entry", http.StatusConflict)
		}
		log.Println("Error creating game:", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
