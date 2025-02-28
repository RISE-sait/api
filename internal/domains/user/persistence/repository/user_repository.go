package user

import (
	db "api/internal/domains/user/persistence/sqlc/generated"
	values "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"
)

// Repository provides methods to interact with the user data in the database.
type Repository struct {
	Queries *db.Queries
}

var _ RepositoryInterface = (*Repository)(nil)

// NewUserRepository creates a new instance of UserRepository with the provided dependency injection container.
func NewUserRepository(queries *db.Queries) *Repository {
	return &Repository{
		Queries: queries,
	}
}

func (r *Repository) GetUserIDByHubSpotId(ctx context.Context, id string) (*uuid.UUID, *errLib.CommonError) {

	dbUserId, err := r.Queries.GetUserIDByHubSpotId(ctx, sql.NullString{
		String: id,
		Valid:  true,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("User not found", http.StatusNotFound)
		}

		log.Printf("Unhandled error: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &dbUserId, nil
}

func (r *Repository) GetUsers(ctx context.Context) ([]values.ReadValue, *errLib.CommonError) {

	dbUsers, err := r.Queries.GetUsers(ctx)

	if err != nil {

		log.Printf("Unhandled error: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	users := make([]values.ReadValue, len(dbUsers))

	for i, dbUser := range dbUsers {
		user := values.ReadValue{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
		}

		if dbUser.HubspotID.Valid {
			user.HubspotID = &dbUser.HubspotID.String
		}

		if dbUser.ProfilePicUrl.Valid {
			user.ProfilePicUrl = &dbUser.ProfilePicUrl.String
		}

		users[i] = user
	}

	return users, nil
}

func (r *Repository) UpdateStats(ctx context.Context, valuesToUpdate values.StatsUpdateValue) *errLib.CommonError {

	var args db.UpdateUserStatsParams

	if valuesToUpdate.Wins != nil {
		args.Wins = sql.NullInt32{
			Int32: *valuesToUpdate.Wins,
			Valid: true,
		}
	}

	if valuesToUpdate.Losses != nil {
		args.Losses = sql.NullInt32{
			Int32: *valuesToUpdate.Losses,
			Valid: true,
		}
	}

	if valuesToUpdate.Points != nil {
		args.Points = sql.NullInt32{
			Int32: *valuesToUpdate.Points,
			Valid: true,
		}
	}

	if valuesToUpdate.Steals != nil {
		args.Steals = sql.NullInt32{
			Int32: *valuesToUpdate.Steals,
			Valid: true,
		}
	}

	if valuesToUpdate.Assists != nil {
		args.Assists = sql.NullInt32{
			Int32: *valuesToUpdate.Assists,
			Valid: true,
		}
	}

	if valuesToUpdate.Rebounds != nil {
		args.Rebounds = sql.NullInt32{
			Int32: *valuesToUpdate.Rebounds,
			Valid: true,
		}
	}

	updatedRows, err := r.Queries.UpdateUserStats(ctx, args)

	if err != nil {

		log.Printf("Unhandled error: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if updatedRows == 0 {
		return errLib.New("Person with the associated ID not found", http.StatusNotFound)
	}

	return nil
}
