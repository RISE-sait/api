package team

import (
	"api/internal/di"
	db "api/internal/domains/team/persistence/sqlc/generated"
	values "api/internal/domains/team/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var constraintErrors = map[string]struct {
	Message string
	Status  int
}{
	"fk_coach": {
		Message: "The referenced coach doesn't exist",
		Status:  http.StatusNotFound,
	},
	"unique_team_name": {
		Message: "The team name already exists",
		Status:  http.StatusConflict,
	},
}

type Repository struct {
	Queries *db.Queries
	Tx      *sql.Tx
}

func (r *Repository) GetTx() *sql.Tx {
	return r.Tx
}

func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{
		Queries: r.Queries.WithTx(tx),
		Tx:      tx,
	}
}

func NewTeamRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.TeamDb,
	}
}

func (r *Repository) Update(ctx context.Context, team values.UpdateTeamValues) *errLib.CommonError {

	var logoURL sql.NullString
	if team.TeamDetails.LogoURL != nil {
		logoURL = sql.NullString{String: *team.TeamDetails.LogoURL, Valid: true}
	}

	params := db.UpdateTeamParams{
		ID:       team.ID,
		Name:     team.TeamDetails.Name,
		Capacity: team.TeamDetails.Capacity,
		CoachID: uuid.NullUUID{
			UUID:  team.TeamDetails.CoachID,
			Valid: team.TeamDetails.CoachID != uuid.Nil,
		},
		LogoUrl: logoURL,
	}

	_, err := r.Queries.UpdateTeam(ctx, params)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

			if errInfo, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(errInfo.Message, errInfo.Status)
			}
		}
		log.Printf("Database error when updating team: %v", err.Error())
		return errLib.New("Database error when updating team:", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) UpdateAthletesTeam(ctx context.Context, athleteID, teamID uuid.UUID) *errLib.CommonError {

	params := db.UpdateAthleteTeamParams{
		TeamID: uuid.NullUUID{
			UUID:  teamID,
			Valid: teamID != uuid.Nil,
		},
		ID: athleteID,
	}

	impactedRows, err := r.Queries.UpdateAthleteTeam(ctx, params)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

			if pqErr.Constraint == "fk_team" {
				return errLib.New("The referenced team doesn't exist", http.StatusNotFound)
			}
		}
		log.Printf("Database error when updating athlete's team: %v", err.Error())
		return errLib.New("Database error when updating athlete's team:", http.StatusInternalServerError)
	}

	if impactedRows == 0 {
		return errLib.New("Athlete not found", http.StatusNotFound)
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
		var logoURL *string
		if dbPractice.LogoUrl.Valid {
			logoURL = &dbPractice.LogoUrl.String
		}

		team := values.GetTeamValues{
			ID:         dbPractice.ID,
			IsExternal: dbPractice.IsExternal,
			CreatedAt:  dbPractice.CreatedAt,
			UpdatedAt:  dbPractice.UpdatedAt,
			TeamDetails: values.Details{
				Name:     dbPractice.Name,
				Capacity: dbPractice.Capacity,
				LogoURL:  logoURL,
			},
		}

		if dbPractice.CoachID.Valid {
			team.TeamDetails.CoachID = dbPractice.CoachID.UUID
			if coachName, ok := dbPractice.CoachName.(string); ok {
				team.TeamDetails.CoachName = coachName
			}
			team.TeamDetails.CoachEmail = dbPractice.CoachEmail
		}

		// Fetch roster for each team
		roster, rosterErr := r.getRosterMembers(ctx, dbPractice.ID)
		if rosterErr != nil {
			return nil, rosterErr
		}
		team.Roster = roster

		teams[i] = team
	}

	return teams, nil
}

func (r *Repository) ListByCoach(ctx context.Context, coachID uuid.UUID) ([]values.GetTeamValues, *errLib.CommonError) {

	dbTeams, err := r.Queries.GetTeamsByCoach(ctx, uuid.NullUUID{UUID: coachID, Valid: true})

	if err != nil {
		log.Println("Error getting teams by coach: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)
		return nil, dbErr
	}

	teams := make([]values.GetTeamValues, len(dbTeams))

	for i, dbTeam := range dbTeams {
		var logoURL *string
		if dbTeam.LogoUrl.Valid {
			logoURL = &dbTeam.LogoUrl.String
		}

		team := values.GetTeamValues{
			ID:         dbTeam.ID,
			IsExternal: dbTeam.IsExternal,
			CreatedAt:  dbTeam.CreatedAt,
			UpdatedAt:  dbTeam.UpdatedAt,
			TeamDetails: values.Details{
				Name:     dbTeam.Name,
				Capacity: dbTeam.Capacity,
				LogoURL:  logoURL,
			},
		}

		if dbTeam.CoachID.Valid {
			team.TeamDetails.CoachID = dbTeam.CoachID.UUID
			team.TeamDetails.CoachName = dbTeam.CoachName
			team.TeamDetails.CoachEmail = dbTeam.CoachEmail.String
		}

		// Fetch roster for each team
		roster, rosterErr := r.getRosterMembers(ctx, dbTeam.ID)
		if rosterErr != nil {
			return nil, rosterErr
		}
		team.Roster = roster

		teams[i] = team
	}

	return teams, nil
}

func (r *Repository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteTeam(c, id)

	if err != nil {
		log.Printf("Failed to delete team with ID: %s. Error: %v", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Team not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) getRosterMembers(ctx context.Context, teamID uuid.UUID) ([]values.RosterMemberInfo, *errLib.CommonError) {
	dbMembers, err := r.Queries.GetTeamRoster(ctx, teamID)

	if err != nil {
		log.Printf("Failed to get team roster: %v", err)
		return nil, errLib.New("Internal server error when getting team roster", http.StatusInternalServerError)
	}

	members := make([]values.RosterMemberInfo, len(dbMembers))

	for i, dbMember := range dbMembers {
		member := values.RosterMemberInfo{
			ID:       dbMember.ID,
			Email:    dbMember.Email.String,
			Country:  dbMember.CountryAlpha2Code,
			Name:     dbMember.Name,
			Points:   dbMember.Points,
			Wins:     dbMember.Wins,
			Losses:   dbMember.Losses,
			Assists:  dbMember.Assists,
			Rebounds: dbMember.Rebounds,
			Steals:   dbMember.Steals,
		}

		if dbMember.PhotoUrl.Valid {
			member.PhotoURL = &dbMember.PhotoUrl.String
		}

		members[i] = member
	}

	return members, nil
}

func (r *Repository) Create(c context.Context, teamDetails values.CreateTeamValues) *errLib.CommonError {

	var logoURL sql.NullString
	if teamDetails.LogoURL != nil {
		logoURL = sql.NullString{String: *teamDetails.LogoURL, Valid: true}
	}

	params := db.CreateTeamParams{
		Name:       teamDetails.Name,
		Capacity:   teamDetails.Capacity,
		CoachID:    uuid.NullUUID{
			UUID:  teamDetails.CoachID,
			Valid: teamDetails.CoachID != uuid.Nil,
		},
		LogoUrl:    logoURL,
		IsExternal: false, // Internal RISE team
	}

	_, err := r.Queries.CreateTeam(c, params)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {

			if errInfo, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(errInfo.Message, errInfo.Status)
			}
		}

		log.Printf("Database error when creating team: %v", err.Error())
		return errLib.New("Database error when creating team:", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (values.GetTeamValues, *errLib.CommonError) {

	dbTeam, dbErr := r.Queries.GetTeamById(ctx, id)

	if dbErr != nil {
		if errors.Is(dbErr, sql.ErrNoRows) {
			return values.GetTeamValues{}, errLib.New("Team not found", http.StatusNotFound)
		}

		log.Println("Error getting team: ", dbErr)
		return values.GetTeamValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	var logoURL *string
	if dbTeam.LogoUrl.Valid {
		logoURL = &dbTeam.LogoUrl.String
	}

	team := values.GetTeamValues{
		ID:         dbTeam.ID,
		IsExternal: dbTeam.IsExternal,
		CreatedAt:  dbTeam.CreatedAt,
		UpdatedAt:  dbTeam.UpdatedAt,
		TeamDetails: values.Details{
			Name:     dbTeam.Name,
			Capacity: dbTeam.Capacity,
			LogoURL:  logoURL,
		},
	}

	if dbTeam.CoachID.Valid {
		team.TeamDetails.CoachID = dbTeam.CoachID.UUID
		if coachName, ok := dbTeam.CoachName.(string); ok {
			team.TeamDetails.CoachName = coachName
		}
		team.TeamDetails.CoachEmail = dbTeam.CoachEmail
	}

	roster, err := r.getRosterMembers(ctx, id)

	if err != nil {
		return values.GetTeamValues{}, err
	}

	team.Roster = roster

	return team, nil
}

func (r *Repository) CreateExternal(c context.Context, teamDetails values.CreateExternalTeamValues) *errLib.CommonError {

	var logoURL sql.NullString
	if teamDetails.LogoURL != nil {
		logoURL = sql.NullString{String: *teamDetails.LogoURL, Valid: true}
	}

	params := db.CreateExternalTeamParams{
		Name:     teamDetails.Name,
		Capacity: teamDetails.Capacity,
		LogoUrl:  logoURL,
	}

	_, err := r.Queries.CreateExternalTeam(c, params)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {

			if errInfo, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(errInfo.Message, errInfo.Status)
			}
		}

		log.Printf("Database error when creating external team: %v", err.Error())
		return errLib.New("Database error when creating external team:", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) ListExternal(ctx context.Context) ([]values.GetTeamValues, *errLib.CommonError) {

	dbTeams, err := r.Queries.GetExternalTeams(ctx)

	if err != nil {
		log.Println("Error getting external teams: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)
		return nil, dbErr
	}

	teams := make([]values.GetTeamValues, len(dbTeams))

	for i, dbTeam := range dbTeams {
		var logoURL *string
		if dbTeam.LogoUrl.Valid {
			logoURL = &dbTeam.LogoUrl.String
		}

		team := values.GetTeamValues{
			ID:         dbTeam.ID,
			IsExternal: true, // External teams
			CreatedAt:  dbTeam.CreatedAt,
			UpdatedAt:  dbTeam.UpdatedAt,
			TeamDetails: values.Details{
				Name:     dbTeam.Name,
				Capacity: dbTeam.Capacity,
				LogoURL:  logoURL,
			},
		}

		teams[i] = team
	}

	return teams, nil
}

func (r *Repository) ListInternal(ctx context.Context) ([]values.GetTeamValues, *errLib.CommonError) {

	dbTeams, err := r.Queries.GetInternalTeams(ctx)

	if err != nil {
		log.Println("Error getting internal teams: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)
		return nil, dbErr
	}

	teams := make([]values.GetTeamValues, len(dbTeams))

	for i, dbTeam := range dbTeams {
		var logoURL *string
		if dbTeam.LogoUrl.Valid {
			logoURL = &dbTeam.LogoUrl.String
		}

		team := values.GetTeamValues{
			ID:         dbTeam.ID,
			IsExternal: false, // Internal RISE teams
			CreatedAt:  dbTeam.CreatedAt,
			UpdatedAt:  dbTeam.UpdatedAt,
			TeamDetails: values.Details{
				Name:     dbTeam.Name,
				Capacity: dbTeam.Capacity,
				LogoURL:  logoURL,
			},
		}

		if dbTeam.CoachID.Valid {
			team.TeamDetails.CoachID = dbTeam.CoachID.UUID
			team.TeamDetails.CoachName = dbTeam.CoachName
			team.TeamDetails.CoachEmail = dbTeam.CoachEmail.String
		}

		teams[i] = team
	}

	return teams, nil
}

func (r *Repository) SearchByName(ctx context.Context, query string, limit int32) ([]values.GetTeamValues, *errLib.CommonError) {

	// Add wildcard for LIKE query
	searchQuery := "%" + query + "%"

	dbTeams, err := r.Queries.SearchTeamsByName(ctx, db.SearchTeamsByNameParams{
		Lower: searchQuery,
		Limit: limit,
	})

	if err != nil {
		log.Println("Error searching teams: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)
		return nil, dbErr
	}

	teams := make([]values.GetTeamValues, len(dbTeams))

	for i, dbTeam := range dbTeams {
		var logoURL *string
		if dbTeam.LogoUrl.Valid {
			logoURL = &dbTeam.LogoUrl.String
		}

		team := values.GetTeamValues{
			ID:         dbTeam.ID,
			IsExternal: dbTeam.IsExternal,
			CreatedAt:  dbTeam.CreatedAt,
			UpdatedAt:  dbTeam.UpdatedAt,
			TeamDetails: values.Details{
				Name:     dbTeam.Name,
				Capacity: dbTeam.Capacity,
				LogoURL:  logoURL,
			},
		}

		if dbTeam.CoachID.Valid {
			team.TeamDetails.CoachID = dbTeam.CoachID.UUID
			if coachName, ok := dbTeam.CoachName.(string); ok {
				team.TeamDetails.CoachName = coachName
			}
			team.TeamDetails.CoachEmail = dbTeam.CoachEmail
		}

		teams[i] = team
	}

	return teams, nil
}

func (r *Repository) CheckNameExists(ctx context.Context, name string) (bool, *errLib.CommonError) {

	exists, err := r.Queries.CheckTeamNameExists(ctx, name)

	if err != nil {
		log.Printf("Error checking team name existence: %v", err)
		return false, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return exists, nil
}
