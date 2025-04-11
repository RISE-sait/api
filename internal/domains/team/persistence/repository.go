package team

import (
	databaseErrors "api/internal/constants"
	db "api/internal/domains/team/persistence/sqlc/generated"
	values "api/internal/domains/team/values"
	errLib "api/internal/libs/errors"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
)

type Repository struct {
	Queries *db.Queries
}

func NewTeamRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) Update(ctx context.Context, team values.UpdateTeamValues) *errLib.CommonError {

	params := db.UpdateTeamParams{
		ID:       team.ID,
		Name:     team.TeamDetails.Name,
		Capacity: team.TeamDetails.Capacity,
		CoachID: uuid.NullUUID{
			UUID:  team.TeamDetails.CoachID,
			Valid: team.TeamDetails.CoachID != uuid.Nil,
		},
	}

	_, err := r.Queries.UpdateTeam(ctx, params)

	if err != nil {
		// Check if the error is a unique violation (duplicate name)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == databaseErrors.UniqueViolation {
				return errLib.New("Team name already exists", http.StatusConflict)
			}
			log.Println(fmt.Sprintf("Database error when updating team: %v", err.Error()))
			return errLib.New("Database error when updating team:", http.StatusInternalServerError)
		}
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]values.GetTeamValues, *errLib.CommonError) {

	dbTeams, err := r.Queries.GetTeams(ctx)

	if err != nil {

		log.Println("Error getting teams: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)

		return nil, dbErr
	}

	teams := make([]values.GetTeamValues, len(dbTeams))

	for i, dbPractice := range dbTeams {
		team := values.GetTeamValues{
			ID:        dbPractice.ID,
			CreatedAt: dbPractice.CreatedAt,
			UpdatedAt: dbPractice.UpdatedAt,
			TeamDetails: values.Details{
				Name:     dbPractice.Name,
				Capacity: dbPractice.Capacity,
			},
		}

		if dbPractice.CoachID.Valid {
			team.TeamDetails.CoachID = dbPractice.CoachID.UUID
		}

		teams[i] = team
	}

	return teams, nil
}

func (r *Repository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteTeam(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Team not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) Create(c context.Context, teamDetails values.CreateTeamValues) *errLib.CommonError {

	params := db.CreateTeamParams{
		Name:     teamDetails.Name,
		Capacity: teamDetails.Capacity,
		CoachID: uuid.NullUUID{
			UUID:  teamDetails.CoachID,
			Valid: teamDetails.CoachID != uuid.Nil,
		},
	}

	_, err := r.Queries.CreateTeam(c, params)

	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			// Return a custom error for unique violation
			return errLib.New("Team name already exists", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
