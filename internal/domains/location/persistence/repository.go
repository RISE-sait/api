package location

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/location/persistence/sqlc/generated"
	values "api/internal/domains/location/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

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

func NewLocationRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.LocationDb,
	}
}

func (r *Repository) CreateLocation(c context.Context, Location values.CreateDetails) (values.ReadValues, *errLib.CommonError) {

	var response values.ReadValues

	dbParams := db.CreateLocationParams{
		Name:    Location.Name,
		Address: Location.Address,
	}

	dbLocation, err := r.Queries.CreateLocation(c, dbParams)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle specific Postgres errors
			if pqErr.Code == databaseErrors.UniqueViolation {
				return response, errLib.New("Location with the given name already exists", http.StatusConflict)
			}
		}
		log.Printf("Error creating Location: %v", err)
		return response, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.ReadValues{
		ID: dbLocation.ID,
		BaseDetails: values.BaseDetails{
			Name:    dbLocation.Name,
			Address: dbLocation.Address,
		},
	}, nil
}

func (r *Repository) GetLocationByID(c context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError) {
	Location, err := r.Queries.GetLocationById(c, id)

	var response values.ReadValues

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, errLib.New("Location not found", http.StatusNotFound)
		}
		log.Printf("Failed to retrieve Location with HubSpotId: %s, error: %v", id, err)
		return response, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.ReadValues{
		ID: Location.ID,
		BaseDetails: values.BaseDetails{
			Name:    Location.Name,
			Address: Location.Address,
		},
	}, nil
}

func (r *Repository) GetLocations(ctx context.Context) ([]values.ReadValues, *errLib.CommonError) {
	dbFacilities, err := r.Queries.GetLocations(ctx)

	if err != nil {
		log.Printf("Error retrieving facilities: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	courses := make([]values.ReadValues, len(dbFacilities))
	for i, dbLocation := range dbFacilities {
		courses[i] = values.ReadValues{
			ID: dbLocation.ID,
			BaseDetails: values.BaseDetails{
				Name:    dbLocation.Name,
				Address: dbLocation.Address,
			},
		}
	}

	return courses, nil
}

func (r *Repository) UpdateLocation(c context.Context, Location values.UpdateDetails) *errLib.CommonError {

	dbParams := db.UpdateLocationParams{
		ID:      Location.ID,
		Name:    Location.Name,
		Address: Location.Address,
	}

	row, err := r.Queries.UpdateLocation(c, dbParams)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle specific Postgres errors
			if pqErr.Code == databaseErrors.UniqueViolation {
				return errLib.New("Location with the given name already exists", http.StatusConflict)
			}
		}
		log.Printf("Error updating Location with HubSpotId: %s, error: %v", Location.ID, err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		log.Printf("Location not found with ID: %s", Location.ID)
		return errLib.New("Location not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) DeleteLocation(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteLocation(c, id)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			log.Printf("Error deleting Location with HubSpotId: %s, error: %v", id, err)
			return errLib.New("Internal server error", http.StatusInternalServerError)
		}
	}

	if row == 0 {
		log.Printf("Failed to delete Location id: %+v", id)
		return errLib.New("Location not found", http.StatusNotFound)
	}

	return nil
}
