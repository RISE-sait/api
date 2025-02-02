package persistence

import (
	"api/cmd/server/di"
	database_errors "api/internal/constants"
	db "api/internal/domains/facility/persistence/sqlc/generated"
	"api/internal/domains/facility/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type FacilityRepository struct {
	Queries *db.Queries
}

func NewFacilityRepository(container *di.Container) *FacilityRepository {
	return &FacilityRepository{
		Queries: container.Queries.FacilityDb,
	}
}

func (r *FacilityRepository) CreateFacility(c context.Context, facility *values.FacilityDetails) *errLib.CommonError {

	dbParams := db.CreateFacilityParams{
		Name:           facility.Name,
		Location:       facility.Location,
		FacilityTypeID: facility.FacilityTypeID,
	}

	row, err := r.Queries.CreateFacility(c, dbParams)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle specific Postgres errors
			switch pqErr.Code {
			case database_errors.ForeignKeyViolation:
				return errLib.New("Invalid facility type ID", http.StatusBadRequest)
			case database_errors.UniqueViolation:
				return errLib.New("Facility with the given name already exists", http.StatusConflict)
			}
		}
		log.Printf("Error creating facility: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	if row == 0 {
		return errLib.New("Facility not created", http.StatusInternalServerError)
	}

	return nil
}

func (r *FacilityRepository) GetFacility(c context.Context, id uuid.UUID) (*values.FacilityAllFields, *errLib.CommonError) {
	facility, err := r.Queries.GetFacilityById(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Facility not found", http.StatusNotFound)
		}
		log.Printf("Failed to retrieve facility with ID: %s, error: %v", id, err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &values.FacilityAllFields{
		ID: facility.ID,
		FacilityDetails: values.FacilityDetails{
			Name:           facility.Name,
			Location:       facility.Location,
			FacilityTypeID: facility.FacilityTypeID,
		},
	}, nil
}

func (r *FacilityRepository) GetAllFacilities(c context.Context, filter string) ([]values.FacilityAllFields, *errLib.CommonError) {
	dbFacilities, err := r.Queries.GetAllFacilities(c)

	if err != nil {
		log.Printf("Error retrieving facilities: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	courses := make([]values.FacilityAllFields, len(dbFacilities))
	for i, dbFacility := range dbFacilities {
		courses[i] = values.FacilityAllFields{
			ID: dbFacility.ID,
			FacilityDetails: values.FacilityDetails{
				Name:           dbFacility.Name,
				Location:       dbFacility.Location,
				FacilityTypeID: dbFacility.FacilityTypeID,
			},
		}
	}

	return courses, nil
}

func (r *FacilityRepository) UpdateFacility(c context.Context, facility *values.FacilityAllFields) *errLib.CommonError {

	dbParams := db.UpdateFacilityParams{
		ID:             facility.ID,
		Name:           facility.Name,
		Location:       facility.Location,
		FacilityTypeID: facility.FacilityTypeID,
	}

	row, err := r.Queries.UpdateFacility(c, dbParams)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle specific Postgres errors
			switch pqErr.Code {
			case database_errors.ForeignKeyViolation:
				return errLib.New("Facility type ID not found", http.StatusBadRequest)
			case database_errors.UniqueViolation:
				return errLib.New("Facility with the given name already exists", http.StatusConflict)
			}
		}
		log.Printf("Error updating facility with ID: %s, error: %v", facility.ID, err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		log.Printf("Facility not found with ID: %s", facility.ID)
		return errLib.New("Facility not found", http.StatusNotFound)
	}

	return nil
}

func (r *FacilityRepository) DeleteFacility(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteFacility(c, id)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			log.Printf("Error deleting facility with ID: %s, error: %v", id, err)
			return errLib.New("Internal server error", http.StatusInternalServerError)
		}
	}

	if row == 0 {
		log.Printf("Failed to delete facility id: %+v", id)
		return errLib.New("Facility not found", http.StatusNotFound)
	}

	return nil
}
