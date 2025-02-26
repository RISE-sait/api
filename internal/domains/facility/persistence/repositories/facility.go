package repository

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	entity "api/internal/domains/facility/entity"
	db "api/internal/domains/facility/persistence/sqlc/generated"
	values "api/internal/domains/facility/values"
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

func (r *FacilityRepository) CreateFacility(c context.Context, facility *values.Details) (*entity.Facility, *errLib.CommonError) {

	dbParams := db.CreateFacilityParams{
		Name:               facility.Name,
		Address:            facility.Address,
		FacilityCategoryID: facility.FacilityCategoryID,
	}

	dbFacility, err := r.Queries.CreateFacility(c, dbParams)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle specific Postgres errors
			switch pqErr.Code {
			case databaseErrors.ForeignKeyViolation:
				return nil, errLib.New("Invalid facility type HubSpotId", http.StatusBadRequest)
			case databaseErrors.UniqueViolation:
				return nil, errLib.New("Facility with the given name already exists", http.StatusConflict)
			}
		}
		log.Printf("Error creating facility: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entity.Facility{
		ID: dbFacility.ID,
		Details: values.Details{
			Name:                 dbFacility.Name,
			Address:              dbFacility.Address,
			FacilityCategoryName: dbFacility.FacilityCategoryName,
			FacilityCategoryID:   dbFacility.FacilityCategoryID,
		},
	}, nil
}

func (r *FacilityRepository) GetFacility(c context.Context, id uuid.UUID) (*entity.Facility, *errLib.CommonError) {
	facility, err := r.Queries.GetFacilityById(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Facility not found", http.StatusNotFound)
		}
		log.Printf("Failed to retrieve facility with HubSpotId: %s, error: %v", id, err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entity.Facility{
		ID: facility.ID,
		Details: values.Details{
			Name:                 facility.Name,
			Address:              facility.Address,
			FacilityCategoryName: facility.FacilityCategoryName,
			FacilityCategoryID:   facility.FacilityCategoryID,
		},
	}, nil
}

func (r *FacilityRepository) GetFacilities(c context.Context, name string) ([]entity.Facility, *errLib.CommonError) {
	dbFacilities, err := r.Queries.GetFacilities(c, sql.NullString{String: name, Valid: name != ""})

	if err != nil {
		log.Printf("Error retrieving facilities: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	courses := make([]entity.Facility, len(dbFacilities))
	for i, dbFacility := range dbFacilities {
		courses[i] = entity.Facility{
			ID: dbFacility.ID,
			Details: values.Details{
				Name:                 dbFacility.Name,
				Address:              dbFacility.Address,
				FacilityCategoryName: dbFacility.FacilityCategoryName,
				FacilityCategoryID:   dbFacility.FacilityCategoryID,
			},
		}
	}

	return courses, nil
}

func (r *FacilityRepository) UpdateFacility(c context.Context, facility *entity.Facility) *errLib.CommonError {

	dbParams := db.UpdateFacilityParams{
		ID:                 facility.ID,
		Name:               facility.Name,
		Address:            facility.Address,
		FacilityCategoryID: facility.FacilityCategoryID,
	}

	row, err := r.Queries.UpdateFacility(c, dbParams)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle specific Postgres errors
			switch pqErr.Code {
			case databaseErrors.ForeignKeyViolation:
				return errLib.New("Facility type HubSpotId not found", http.StatusBadRequest)
			case databaseErrors.UniqueViolation:
				return errLib.New("Facility with the given name already exists", http.StatusConflict)
			}
		}
		log.Printf("Error updating facility with HubSpotId: %s, error: %v", facility.ID, err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		log.Printf("Facility not found with HubSpotId: %s", facility.ID)
		return errLib.New("Facility not found", http.StatusNotFound)
	}

	return nil
}

func (r *FacilityRepository) DeleteFacility(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteFacility(c, id)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			log.Printf("Error deleting facility with HubSpotId: %s, error: %v", id, err)
			return errLib.New("Internal server error", http.StatusInternalServerError)
		}
	}

	if row == 0 {
		log.Printf("Failed to delete facility id: %+v", id)
		return errLib.New("Facility not found", http.StatusNotFound)
	}

	return nil
}
