package customer

import (
	db "api/internal/domains/customer/persistence/sqlc/generated"
	values "api/internal/domains/customer/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

// Repository provides methods to interact with the user data in the database.
type Repository struct {
	Queries *db.Queries
}

var _ RepositoryInterface = (*Repository)(nil)

// NewCustomerRepository creates a new instance of UserRepository with the provided dependency injection container.
func NewCustomerRepository(queries *db.Queries) *Repository {
	return &Repository{
		Queries: queries,
	}
}

func (r *Repository) GetCustomers(ctx context.Context, hubspotIds []string) ([]values.ReadValue, *errLib.CommonError) {

	dbCustomers, err := r.Queries.GetCustomers(ctx, hubspotIds)

	if err != nil {
		log.Println(fmt.Sprintf("Error getting dbCustomers: %s", err))
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	customers := make([]values.ReadValue, len(dbCustomers))

	for i, dbCustomer := range dbCustomers {
		customer := values.ReadValue{
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
