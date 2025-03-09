package customer

import (
	db "api/internal/domains/user/persistence/sqlc/generated"
	values "api/internal/domains/user/values/customer"
	"api/internal/domains/user/values/user"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
)

// Repository provides methods to interact with the user data in the database.
type Repository struct {
	Queries *db.Queries
}

// NewCustomerRepository creates a new instance of UserRepository with the provided dependency injection container.
func NewCustomerRepository(queries *db.Queries) *Repository {
	return &Repository{
		Queries: queries,
	}
}

func (r *Repository) GetCustomers(ctx context.Context, hubspotIds []string) ([]user.ReadValue, *errLib.CommonError) {

	dbCustomers, err := r.Queries.GetCustomers(ctx, hubspotIds)

	if err != nil {
		log.Println(fmt.Sprintf("Error getting dbCustomers: %s", err))
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	customers := make([]user.ReadValue, len(dbCustomers))

	for i, dbCustomer := range dbCustomers {
		customer := user.ReadValue{
			ID:        dbCustomer.ID,
			HubspotID: dbCustomer.HubspotID,
			CreatedAt: dbCustomer.CreatedAt,
			UpdatedAt: dbCustomer.UpdatedAt,
		}

		if dbCustomer.ProfilePicUrl.Valid {
			customer.ProfilePicUrl = &dbCustomer.ProfilePicUrl.String
		}

		customers[i] = customer
	}

	return customers, nil
}

func (r *Repository) GetMembershipPlansByCustomer(ctx context.Context, customerID uuid.UUID) ([]user.CustomerMembershipPlansReadValue, *errLib.CommonError) {

	dbPlans, err := r.Queries.GetMembershipPlansByCustomer(ctx, customerID)

	if err != nil {
		log.Println(fmt.Sprintf("Error getting membership plans by customer: %s", err))
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	plans := make([]user.CustomerMembershipPlansReadValue, len(dbPlans))

	for i, dbPlan := range dbPlans {
		plan := user.CustomerMembershipPlansReadValue{
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

func (r *Repository) UpdateStats(ctx context.Context, valuesToUpdate values.StatsUpdateValue) *errLib.CommonError {

	var args db.UpdateCustomerStatsParams

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

	updatedRows, err := r.Queries.UpdateCustomerStats(ctx, args)

	if err != nil {

		log.Printf("Unhandled error: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if updatedRows == 0 {
		return errLib.New("Person with the associated ID not found", http.StatusNotFound)
	}

	return nil
}
