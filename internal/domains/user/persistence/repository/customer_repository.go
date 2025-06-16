package user

import (
	"api/internal/di"
	db "api/internal/domains/user/persistence/sqlc/generated"
	userValues "api/internal/domains/user/values"
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

type CustomerRepository struct {
	Queries *db.Queries
}

func NewCustomerRepository(container *di.Container) *CustomerRepository {
	return &CustomerRepository{
		Queries: container.Queries.UserDb,
	}
}

func (r *CustomerRepository) GetCustomers(ctx context.Context, limit, offset int32, parentID uuid.UUID, search string) ([]userValues.ReadValue, *errLib.CommonError) {

	dbCustomers, err := r.Queries.GetCustomers(ctx, db.GetCustomersParams{
		ParentID: uuid.NullUUID{
			UUID:  parentID,
			Valid: parentID != uuid.Nil,
		},
		Search: sql.NullString{
			String: search,
			Valid:  search != "",
		},
		Offset: offset,
		Limit:  limit,
	})

	if err != nil {
		log.Printf("Error getting dbCustomers: %s", err)
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	customers := make([]userValues.ReadValue, len(dbCustomers))

	for i, dbCustomer := range dbCustomers {
		customer := userValues.ReadValue{
			ID:          dbCustomer.ID,
			DOB:         dbCustomer.Dob,
			FirstName:   dbCustomer.FirstName,
			LastName:    dbCustomer.LastName,
			CountryCode: dbCustomer.CountryAlpha2Code,
			CreatedAt:   dbCustomer.CreatedAt,
			UpdatedAt:   dbCustomer.UpdatedAt,
		}

		if dbCustomer.HubspotID.Valid {
			customer.HubspotID = &dbCustomer.HubspotID.String
		}

		if dbCustomer.Phone.Valid {
			customer.Phone = &dbCustomer.Phone.String
		}

		if dbCustomer.Email.Valid {
			customer.Email = &dbCustomer.Email.String
		}

		if dbCustomer.MembershipName.Valid && dbCustomer.MembershipPlanName.Valid && dbCustomer.MembershipStartDate.Valid && dbCustomer.MembershipPlanID.Valid {

			customer.MembershipInfo = &userValues.MembershipReadValue{
				MembershipPlanID:      dbCustomer.MembershipPlanID.UUID,
				MembershipPlanName:    dbCustomer.MembershipPlanName.String,
				MembershipName:        dbCustomer.MembershipName.String,
				MembershipStartDate:   dbCustomer.MembershipStartDate.Time,
				MembershipRenewalDate: dbCustomer.MembershipPlanRenewalDate.Time,
			}
		}

		if dbCustomer.Rebounds.Valid && dbCustomer.Wins.Valid && dbCustomer.Points.Valid && dbCustomer.Steals.Valid && dbCustomer.Assists.Valid && dbCustomer.Losses.Valid {
			customer.AthleteInfo = &userValues.AthleteReadValue{
				Wins:     dbCustomer.Wins.Int32,
				Losses:   dbCustomer.Losses.Int32,
				Points:   dbCustomer.Points.Int32,
				Steals:   dbCustomer.Steals.Int32,
				Assists:  dbCustomer.Assists.Int32,
				Rebounds: dbCustomer.Rebounds.Int32,
				PhotoURL: func(n sql.NullString) *string {
					if n.Valid {
						return &n.String
					}
					return nil
				}(dbCustomer.PhotoUrl),
			}
		}

		customers[i] = customer
	}

	return customers, nil
}

func (r *CustomerRepository) GetCustomer(ctx context.Context, id uuid.UUID, email string) (userValues.ReadValue, *errLib.CommonError) {

	dbCustomer, err := r.Queries.GetCustomer(ctx, db.GetCustomerParams{
		ID: uuid.NullUUID{
			UUID:  id,
			Valid: id != uuid.Nil,
		},
		Email: sql.NullString{
			String: email,
			Valid:  email != "",
		},
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userValues.ReadValue{}, errLib.New("Customer not found", http.StatusNotFound)
		}
		log.Printf("Error getting dbCustomer: %s", err)
		return userValues.ReadValue{}, errLib.New("internal error", http.StatusInternalServerError)
	}

	customer := userValues.ReadValue{
		ID:          dbCustomer.ID,
		DOB:         dbCustomer.Dob,
		FirstName:   dbCustomer.FirstName,
		LastName:    dbCustomer.LastName,
		CountryCode: dbCustomer.CountryAlpha2Code,
		CreatedAt:   dbCustomer.CreatedAt,
		UpdatedAt:   dbCustomer.UpdatedAt,
	}

	if dbCustomer.HubspotID.Valid {
		customer.HubspotID = &dbCustomer.HubspotID.String
	}

	if dbCustomer.Phone.Valid {
		customer.Phone = &dbCustomer.Phone.String
	}

	if dbCustomer.Email.Valid {
		customer.Email = &dbCustomer.Email.String
	}

	if dbCustomer.MembershipName.Valid && dbCustomer.MembershipPlanName.Valid && dbCustomer.MembershipStartDate.Valid && dbCustomer.MembershipPlanID.Valid {

		customer.MembershipInfo = &userValues.MembershipReadValue{
			MembershipPlanID:      dbCustomer.MembershipPlanID.UUID,
			MembershipPlanName:    dbCustomer.MembershipPlanName.String,
			MembershipName:        dbCustomer.MembershipName.String,
			MembershipStartDate:   dbCustomer.MembershipStartDate.Time,
			MembershipRenewalDate: dbCustomer.MembershipPlanRenewalDate.Time,
		}
	}

	if dbCustomer.Rebounds.Valid && dbCustomer.Wins.Valid && dbCustomer.Points.Valid && dbCustomer.Steals.Valid && dbCustomer.Assists.Valid && dbCustomer.Losses.Valid {
		customer.AthleteInfo = &userValues.AthleteReadValue{
			Wins:     dbCustomer.Wins.Int32,
			Losses:   dbCustomer.Losses.Int32,
			Points:   dbCustomer.Points.Int32,
			Steals:   dbCustomer.Steals.Int32,
			Assists:  dbCustomer.Assists.Int32,
			Rebounds: dbCustomer.Rebounds.Int32,
			PhotoURL: func(n sql.NullString) *string {
				if n.Valid {
					return &n.String
				}
				return nil
			}(dbCustomer.PhotoUrl),
		}
	}

	return customer, nil
}

func (r *CustomerRepository) UpdateAthleteTeam(ctx context.Context, athleteID, teamID uuid.UUID) *errLib.CommonError {

	args := db.UpdateAthleteTeamParams{
		AthleteID: athleteID,
		TeamID: uuid.NullUUID{
			UUID:  teamID,
			Valid: true,
		},
	}

	updatedRows, err := r.Queries.UpdateAthleteTeam(ctx, args)

	if err != nil {

		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "fk_team" {
				return errLib.New("Team not found", http.StatusNotFound)
			}
		}

		log.Printf("Unhandled error when updating athlete's team: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if updatedRows == 0 {
		return errLib.New("Athlete not found", http.StatusNotFound)
	}

	return nil
}

func (r *CustomerRepository) UpdateStats(ctx context.Context, valuesToUpdate userValues.StatsUpdateValue) *errLib.CommonError {

	var args db.UpdateAthleteStatsParams

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

	updatedRows, err := r.Queries.UpdateAthleteStats(ctx, args)

	if err != nil {

		log.Printf("Unhandled error: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if updatedRows == 0 {
		return errLib.New("Person with the associated ID not found", http.StatusNotFound)
	}

	return nil
}
func (r *CustomerRepository) ListAthletes(ctx context.Context, limit, offset int32) ([]userValues.AthleteReadValue, *errLib.CommonError) {
	dbAthletes, err := r.Queries.GetAthletes(ctx, db.GetAthletesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Printf("Error getting athletes: %s", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	athletes := make([]userValues.AthleteReadValue, len(dbAthletes))
	for i, a := range dbAthletes {
		athletes[i] = userValues.AthleteReadValue{
			ID:        a.ID,
			FirstName: a.FirstName,
			LastName:  a.LastName,
			Points:    a.Points,
			Wins:      a.Wins,
			Losses:    a.Losses,
			Assists:   a.Assists,
			Rebounds:  a.Rebounds,
			Steals:    a.Steals,
			PhotoURL:  ToStringPtr(a.PhotoUrl),
			TeamID: func(n uuid.NullUUID) *uuid.UUID {
				if n.Valid {
					return &n.UUID
				}
				return nil
			}(a.TeamID),
		}
	}

	return athletes, nil
}
func (r *CustomerRepository) CountCustomers(ctx context.Context, parentID uuid.UUID, search string) (int64, *errLib.CommonError) {
	searchArg := sql.NullString{Valid: false}
	if search != "" {
		searchArg = sql.NullString{String: search, Valid: true}
	}

	parentArg := uuid.NullUUID{Valid: false}
	if parentID != uuid.Nil {
		parentArg = uuid.NullUUID{UUID: parentID, Valid: true}
	}

	count, err := r.Queries.CountCustomers(ctx, db.CountCustomersParams{
		ParentID: parentArg,
		Search:   searchArg,
	})
	if err != nil {
		return 0, errLib.New("Failed to count customers: "+err.Error(), http.StatusInternalServerError)
	}

	return count, nil
}

func (r *CustomerRepository) GetActiveMembershipInfo(ctx context.Context, customerID uuid.UUID) (*userValues.MembershipPlansReadValue, *errLib.CommonError) {
	row, err := r.Queries.GetActiveMembershipInfo(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("active membership not found", http.StatusNotFound)
		}
		log.Printf("error fetching active membership info: %v", err)
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	info := userValues.MembershipPlansReadValue{
		ID:         row.ID,
		CustomerID: row.CustomerID,
		StartDate:  row.StartDate,
		Status:     string(row.Status),
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		PhotoURL: func(n sql.NullString) *string {
			if n.Valid {
				return &n.String
			}
			return nil
		}(row.PhotoUrl),
		MembershipID:       row.MembershipID,
		MembershipPlanID:   row.MembershipPlanID,
		MembershipName:     row.MembershipName,
		MembershipPlanName: row.MembershipPlanName,
	}

	if row.RenewalDate.Valid {
		info.RenewalDate = &row.RenewalDate.Time
	}

	return &info, nil
}
func (r *CustomerRepository) ListMembershipHistory(ctx context.Context, customerID uuid.UUID) ([]userValues.MembershipHistoryValue, *errLib.CommonError) {
	dbRows, err := r.Queries.ListMembershipHistory(ctx, customerID)
	if err != nil {
		log.Printf("error querying membership history: %v", err)
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	results := make([]userValues.MembershipHistoryValue, len(dbRows))
	for i, row := range dbRows {
		var renewal *time.Time
		if row.RenewalDate.Valid {
			renewal = &row.RenewalDate.Time
		}

		results[i] = userValues.MembershipHistoryValue{
			ID:             row.ID,
			CustomerID:     row.CustomerID,
			StartDate:      row.StartDate,
			RenewalDate:    renewal,
			Status:         string(row.Status),
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
			MembershipID:   row.MembershipID,
			MembershipName: row.MembershipName,
			MembershipDescription: row.MembershipDescription,
			MembershipPlanID:   row.MembershipPlanID,
			MembershipPlanName: row.MembershipPlanName,
			UnitAmount: func(n sql.NullInt32) int {
				if n.Valid {
					return int(n.Int32)
				}
				return 0
			}(row.UnitAmount),
			Currency: func(n sql.NullString) string {
				if n.Valid {
					return n.String
				}
				return ""
			}(row.Currency),
			Interval: func(n sql.NullString) string {
				if n.Valid {
					return n.String
				}
				return ""
			}(row.Interval),
		}
	}

	return results, nil
}
