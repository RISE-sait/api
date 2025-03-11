package user

import (
	db "api/internal/domains/user/persistence/sqlc/generated"
	customerValues "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"api/internal/services/gcp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
)

// CustomerRepository provides methods to interact with the user data in the database.
type CustomerRepository struct {
	Queries *db.Queries
}

// NewCustomerRepository creates a new instance of UserRepository with the provided dependency injection container.
func NewCustomerRepository(queries *db.Queries) *CustomerRepository {
	return &CustomerRepository{
		Queries: queries,
	}
}

func (r *CustomerRepository) GetCustomers(ctx context.Context) ([]customerValues.ReadValue, *errLib.CommonError) {

	dbCustomers, err := r.Queries.GetCustomers(ctx)

	if err != nil {
		log.Println(fmt.Sprintf("Error getting dbCustomers: %s", err))
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	customers := make([]customerValues.ReadValue, len(dbCustomers))

	for i, dbCustomer := range dbCustomers {
		customer := customerValues.ReadValue{
			ID:          dbCustomer.ID,
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

		customer.ProfilePicUrl = gcp.GeneratePublicFileURL(fmt.Sprintf("athletes/%v", dbCustomer.ID))

		customers[i] = customer
	}

	return customers, nil
}

func (r *CustomerRepository) GetChildrenByCustomerID(ctx context.Context, id uuid.UUID) ([]customerValues.ReadValue, *errLib.CommonError) {

	dbCustomers, err := r.Queries.GetChildren(ctx, id)

	if err != nil {
		log.Println(fmt.Sprintf("Error getting dbCustomers: %s", err))
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	customers := make([]customerValues.ReadValue, len(dbCustomers))

	for i, dbCustomer := range dbCustomers {
		customer := customerValues.ReadValue{
			ID:          dbCustomer.ID,
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

		customer.ProfilePicUrl = gcp.GeneratePublicFileURL(fmt.Sprintf("athletes/%v", dbCustomer.ID))

		customers[i] = customer
	}

	return customers, nil
}

func (r *CustomerRepository) GetMembershipPlansByCustomer(ctx context.Context, customerID uuid.UUID) ([]customerValues.MembershipPlansReadValue, *errLib.CommonError) {

	dbPlans, err := r.Queries.GetMembershipPlansByCustomer(ctx, customerID)

	if err != nil {
		log.Println(fmt.Sprintf("Error getting membership plans by customer: %s", err))
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	plans := make([]customerValues.MembershipPlansReadValue, len(dbPlans))

	for i, dbPlan := range dbPlans {
		plan := customerValues.MembershipPlansReadValue{
			ID:               dbPlan.ID,
			CustomerID:       dbPlan.CustomerID,
			MembershipPlanID: dbPlan.MembershipPlanID,
			StartDate:        dbPlan.StartDate,
			Status:           string(dbPlan.Status),
			CreatedAt:        dbPlan.CreatedAt,
			UpdatedAt:        dbPlan.UpdatedAt,
			MembershipName:   dbPlan.MembershipName,
		}

		if dbPlan.RenewalDate.Valid {
			plan.RenewalDate = &dbPlan.RenewalDate.Time
		}

		plans[i] = plan
	}

	return plans, nil
}

func (r *CustomerRepository) GetAthleteInfo(ctx context.Context, id uuid.UUID) (customerValues.AthleteReadValue, *errLib.CommonError) {

	var response customerValues.AthleteReadValue

	dbAthleteInfo, err := r.Queries.GetAthlete(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, errLib.New("Athlete not found", http.StatusNotFound)
		}

		log.Printf("Unhandled error: %v", err)
		return response, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	response = customerValues.AthleteReadValue{
		ID:        dbAthleteInfo.ID,
		Wins:      dbAthleteInfo.Wins,
		Losses:    dbAthleteInfo.Losses,
		Points:    dbAthleteInfo.Points,
		Steals:    dbAthleteInfo.Steals,
		Assists:   dbAthleteInfo.Assists,
		Rebounds:  dbAthleteInfo.Rebounds,
		CreatedAt: dbAthleteInfo.CreatedAt,
		UpdatedAt: dbAthleteInfo.UpdatedAt,
	}

	response.ProfilePicUrl = gcp.GeneratePublicFileURL(fmt.Sprintf("athletes/%v", response.ID))

	return response, nil
}

func (r *CustomerRepository) UpdateStats(ctx context.Context, valuesToUpdate customerValues.StatsUpdateValue) *errLib.CommonError {

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
