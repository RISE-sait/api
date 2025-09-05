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
func unwrapNullableString(n sql.NullString) string {
	if n.Valid {
		return n.String
	}
	return ""
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
		CourtID:    uuid.NullUUID{UUID: value.CourtID, Valid: value.CourtID != uuid.Nil},
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
func (r *Repository) GetGames(ctx context.Context, filter values.GetGamesFilter) ([]values.ReadGameValue, *errLib.CommonError) {
	var courtID, locationID uuid.NullUUID
	
	if filter.CourtID != nil {
		courtID = uuid.NullUUID{UUID: *filter.CourtID, Valid: true}
	}
	if filter.LocationID != nil {
		locationID = uuid.NullUUID{UUID: *filter.LocationID, Valid: true}
	}
	
	params := db.GetGamesParams{
		CourtID:    courtID,
		LocationID: locationID,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
	}

	log.Printf("DEBUG: GetGames params - CourtID: %v, LocationID: %v, Limit: %d, Offset: %d", 
		params.CourtID, params.LocationID, params.Limit, params.Offset)
	
	dbGames, err := r.Queries.GetGames(ctx, params)
	if err != nil {
		log.Println("Error getting games:", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	
	log.Printf("DEBUG: Retrieved %d games from database", len(dbGames))

	games := make([]values.ReadGameValue, len(dbGames))
	for i, dbGame := range dbGames {
		games[i] = values.ReadGameValue{
			ID:              dbGame.ID,
			HomeTeamID:      dbGame.HomeTeamID,
			HomeTeamName:    dbGame.HomeTeamName,
			HomeTeamLogoUrl: unwrapNullableString(dbGame.HomeTeamLogoUrl),
			AwayTeamID:      dbGame.AwayTeamID,
			AwayTeamName:    dbGame.AwayTeamName,
			AwayTeamLogoUrl: unwrapNullableString(dbGame.AwayTeamLogoUrl),
			HomeScore:       nullableInt32ToPtr(dbGame.HomeScore),
			AwayScore:       nullableInt32ToPtr(dbGame.AwayScore),
			StartTime:       dbGame.StartTime,
			EndTime:         nullableTimeToPtr(dbGame.EndTime),
			LocationID:      dbGame.LocationID,
			LocationName:    dbGame.LocationName,
			CourtID:         dbGame.CourtID.UUID,
			CourtName:       unwrapNullableString(dbGame.CourtName),
			Status:          dbGame.Status.String,
			CreatedAt:       nullableTimeToPtr(dbGame.CreatedAt),
			UpdatedAt:       nullableTimeToPtr(dbGame.UpdatedAt),
		}
	}

	return games, nil
}

// GetUpcomingGames fetches all upcoming games and maps them to domain values.
func (r *Repository) GetUpcomingGames(ctx context.Context, limit, offset int32) ([]values.ReadGameValue, *errLib.CommonError) {
	params := db.GetUpcomingGamesParams{
		Limit:  limit,
		Offset: offset,
	}

	dbGames, err := r.Queries.GetUpcomingGames(ctx, params)
	if err != nil {
		log.Println("Error getting upcoming games:", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return mapDbUpcomingGamesToValues(dbGames), nil
}

// GetPastGames fetches all past games and maps them to domain values.
func (r *Repository) GetPastGames(ctx context.Context, limit, offset int32) ([]values.ReadGameValue, *errLib.CommonError) {
	params := db.GetPastGamesParams{
		Limit:  limit,
		Offset: offset,
	}

	dbGames, err := r.Queries.GetPastGames(ctx, params)
	if err != nil {
		log.Println("Error getting past games:", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return mapDbPastGamesToValues(dbGames), nil
}

// GetGamesByTeams fetches games involving any of the provided team IDs.
func (r *Repository) GetGamesByTeams(ctx context.Context, teamIDs []uuid.UUID, limit, offset int32) ([]values.ReadGameValue, *errLib.CommonError) {
	params := db.GetGamesByTeamsParams{
		TeamIds: teamIDs,
		Limit:   limit,
		Offset:  offset,
	}
	// Fetch games from the database using the provided team IDs, limit, and offset.
	dbGames, err := r.Queries.GetGamesByTeams(ctx, params)
	if err != nil {
		log.Println("Error getting games by teams:", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	// Map the database rows to domain values.
	games := make([]values.ReadGameValue, len(dbGames))
	for i, dbGame := range dbGames {
		games[i] = values.ReadGameValue{
			ID:              dbGame.ID,
			HomeTeamID:      dbGame.HomeTeamID,
			HomeTeamName:    dbGame.HomeTeamName,
			HomeTeamLogoUrl: unwrapNullableString(dbGame.HomeTeamLogoUrl),
			AwayTeamID:      dbGame.AwayTeamID,
			AwayTeamName:    dbGame.AwayTeamName,
			AwayTeamLogoUrl: unwrapNullableString(dbGame.AwayTeamLogoUrl),
			HomeScore:       nullableInt32ToPtr(dbGame.HomeScore),
			AwayScore:       nullableInt32ToPtr(dbGame.AwayScore),
			StartTime:       dbGame.StartTime,
			EndTime:         nullableTimeToPtr(dbGame.EndTime),
			LocationID:      dbGame.LocationID,
			LocationName:    dbGame.LocationName,
			Status:          dbGame.Status.String,
			CreatedAt:       nullableTimeToPtr(dbGame.CreatedAt),
			UpdatedAt:       nullableTimeToPtr(dbGame.UpdatedAt),
		}
	}
	return games, nil
}

// GetUpcomingGamesByTeams fetches upcoming games involving any of the provided team IDs.
func (r *Repository) GetUpcomingGamesByTeams(ctx context.Context, teamIDs []uuid.UUID, limit, offset int32) ([]values.ReadGameValue, *errLib.CommonError) {
	params := db.GetUpcomingGamesByTeamsParams{
		TeamIds: teamIDs,
		Limit:   limit,
		Offset:  offset,
	}
	dbGames, err := r.Queries.GetUpcomingGamesByTeams(ctx, params)
	if err != nil {
		log.Println("Error getting upcoming games by teams:", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return mapDbUpcomingGamesByTeamsToValues(dbGames), nil
}

// GetPastGamesByTeams fetches past games involving any of the provided team IDs.
func (r *Repository) GetPastGamesByTeams(ctx context.Context, teamIDs []uuid.UUID, limit, offset int32) ([]values.ReadGameValue, *errLib.CommonError) {
	params := db.GetPastGamesByTeamsParams{
		TeamIds: teamIDs,
		Limit:   limit,
		Offset:  offset,
	}
	dbGames, err := r.Queries.GetPastGamesByTeams(ctx, params)
	if err != nil {
		log.Println("Error getting past games by teams:", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return mapDbPastGamesByTeamsToValues(dbGames), nil
}

func mapDbPastGamesToValues(dbGames []db.GetPastGamesRow) []values.ReadGameValue {
	games := make([]values.ReadGameValue, len(dbGames))
	for i, dbGame := range dbGames {
		games[i] = values.ReadGameValue{
			ID:              dbGame.ID,
			HomeTeamID:      dbGame.HomeTeamID,
			HomeTeamName:    dbGame.HomeTeamName,
			HomeTeamLogoUrl: unwrapNullableString(dbGame.HomeTeamLogoUrl),
			AwayTeamID:      dbGame.AwayTeamID,
			AwayTeamName:    dbGame.AwayTeamName,
			AwayTeamLogoUrl: unwrapNullableString(dbGame.AwayTeamLogoUrl),
			HomeScore:       nullableInt32ToPtr(dbGame.HomeScore),
			AwayScore:       nullableInt32ToPtr(dbGame.AwayScore),
			StartTime:       dbGame.StartTime,
			EndTime:         nullableTimeToPtr(dbGame.EndTime),
			LocationID:      dbGame.LocationID,
			LocationName:    dbGame.LocationName,
			Status:          dbGame.Status.String,
			CreatedAt:       nullableTimeToPtr(dbGame.CreatedAt),
			UpdatedAt:       nullableTimeToPtr(dbGame.UpdatedAt),
		}
	}
	return games
}

func mapDbUpcomingGamesToValues(dbGames []db.GetUpcomingGamesRow) []values.ReadGameValue {
	games := make([]values.ReadGameValue, len(dbGames))
	for i, dbGame := range dbGames {
		games[i] = values.ReadGameValue{
			ID:              dbGame.ID,
			HomeTeamID:      dbGame.HomeTeamID,
			HomeTeamName:    dbGame.HomeTeamName,
			HomeTeamLogoUrl: unwrapNullableString(dbGame.HomeTeamLogoUrl),
			AwayTeamID:      dbGame.AwayTeamID,
			AwayTeamName:    dbGame.AwayTeamName,
			AwayTeamLogoUrl: unwrapNullableString(dbGame.AwayTeamLogoUrl),
			HomeScore:       nullableInt32ToPtr(dbGame.HomeScore),
			AwayScore:       nullableInt32ToPtr(dbGame.AwayScore),
			StartTime:       dbGame.StartTime,
			EndTime:         nullableTimeToPtr(dbGame.EndTime),
			LocationID:      dbGame.LocationID,
			LocationName:    dbGame.LocationName,
			Status:          dbGame.Status.String,
			CreatedAt:       nullableTimeToPtr(dbGame.CreatedAt),
			UpdatedAt:       nullableTimeToPtr(dbGame.UpdatedAt),
		}
	}
	return games
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
		CourtID:    uuid.NullUUID{UUID: details.CourtID, Valid: details.CourtID != uuid.Nil},
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

func mapDbUpcomingGamesByTeamsToValues(dbGames []db.GetUpcomingGamesByTeamsRow) []values.ReadGameValue {
	games := make([]values.ReadGameValue, len(dbGames))
	for i, dbGame := range dbGames {
		games[i] = values.ReadGameValue{
			ID:              dbGame.ID,
			HomeTeamID:      dbGame.HomeTeamID,
			HomeTeamName:    dbGame.HomeTeamName,
			HomeTeamLogoUrl: unwrapNullableString(dbGame.HomeTeamLogoUrl),
			AwayTeamID:      dbGame.AwayTeamID,
			AwayTeamName:    dbGame.AwayTeamName,
			AwayTeamLogoUrl: unwrapNullableString(dbGame.AwayTeamLogoUrl),
			HomeScore:       nullableInt32ToPtr(dbGame.HomeScore),
			AwayScore:       nullableInt32ToPtr(dbGame.AwayScore),
			StartTime:       dbGame.StartTime,
			EndTime:         nullableTimeToPtr(dbGame.EndTime),
			LocationID:      dbGame.LocationID,
			LocationName:    dbGame.LocationName,
			Status:          dbGame.Status.String,
			CreatedAt:       nullableTimeToPtr(dbGame.CreatedAt),
			UpdatedAt:       nullableTimeToPtr(dbGame.UpdatedAt),
		}
	}
	return games
}

func mapDbPastGamesByTeamsToValues(dbGames []db.GetPastGamesByTeamsRow) []values.ReadGameValue {
	games := make([]values.ReadGameValue, len(dbGames))
	for i, dbGame := range dbGames {
		games[i] = values.ReadGameValue{
			ID:              dbGame.ID,
			HomeTeamID:      dbGame.HomeTeamID,
			HomeTeamName:    dbGame.HomeTeamName,
			HomeTeamLogoUrl: unwrapNullableString(dbGame.HomeTeamLogoUrl),
			AwayTeamID:      dbGame.AwayTeamID,
			AwayTeamName:    dbGame.AwayTeamName,
			AwayTeamLogoUrl: unwrapNullableString(dbGame.AwayTeamLogoUrl),
			HomeScore:       nullableInt32ToPtr(dbGame.HomeScore),
			AwayScore:       nullableInt32ToPtr(dbGame.AwayScore),
			StartTime:       dbGame.StartTime,
			EndTime:         nullableTimeToPtr(dbGame.EndTime),
			LocationID:      dbGame.LocationID,
			LocationName:    dbGame.LocationName,
			Status:          dbGame.Status.String,
			CreatedAt:       nullableTimeToPtr(dbGame.CreatedAt),
			UpdatedAt:       nullableTimeToPtr(dbGame.UpdatedAt),
		}
	}
	return games
}
