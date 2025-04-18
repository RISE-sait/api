package user

import (
	"api/internal/di"
	db "api/internal/domains/user/persistence/sqlc/generated"
	userValues "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
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
		log.Println(fmt.Sprintf("Error getting dbCustomers: %s", err))
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
		log.Println(fmt.Sprintf("Error getting dbCustomer: %s", err))
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
