package repository

import (
	database_errors "api/internal/constants"
	"api/internal/di"
	entity "api/internal/domains/facility/entities"
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

type FacilityTypesRepository struct {
	Queries *db.Queries
}

func NewFacilityTypesRepository(container *di.Container) *FacilityTypesRepository {
	return &FacilityTypesRepository{
		Queries: container.Queries.FacilityDb,
	}
}

func (r *FacilityTypesRepository) CreateFacilityType(c context.Context, name string) *errLib.CommonError {

	row, err := r.Queries.CreateFacilityType(c, name)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle specific Postgres errors
			switch pqErr.Code {
			case database_errors.UniqueViolation:
				return errLib.New("Facility type with the given name already exists", http.StatusConflict)
			}
		}
		log.Printf("Error creating facility type : %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	if row == 0 {
		return errLib.New("Facility type not created", http.StatusInternalServerError)
	}

	return nil
}

func (r *FacilityTypesRepository) GetFacilityType(c context.Context, id uuid.UUID) (*string, *errLib.CommonError) {
	name, err := r.Queries.GetFacilityTypeById(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Facility type not found", http.StatusNotFound)
		}
		log.Printf("Failed to retrieve facility type with ID: %s, error: %v", id, err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &name, nil
}

func (r *FacilityTypesRepository) GetAllFacilityTypes(c context.Context, filter string) ([]entity.FacilityType, *errLib.CommonError) {
	dbFacilityTypes, err := r.Queries.GetAllFacilityTypes(c)

	if err != nil {
		log.Printf("Error retrieving facility types: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	facilityTypes := make([]entity.FacilityType, len(dbFacilityTypes))
	for i, facilityType := range dbFacilityTypes {
		facilityTypes[i] = entity.FacilityType{
			ID:   facilityType.ID,
			Name: facilityType.Name,
		}
	}

	return facilityTypes, nil
}

func (r *FacilityTypesRepository) UpdateFacilityType(c context.Context, facilityType *values.FacilityType) *errLib.CommonError {

	dbParams := db.UpdateFacilityTypeParams{
		ID:   facilityType.ID,
		Name: facilityType.Name,
	}

	row, err := r.Queries.UpdateFacilityType(c, dbParams)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle specific Postgres errors
			switch pqErr.Code {
			case database_errors.UniqueViolation:
				return errLib.New("Facility type with the given name already exists", http.StatusConflict)
			}
		}
		log.Printf("Error updating facility type with ID: %s, error: %v", facilityType.ID, err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		log.Printf("Facility type not found with ID: %s", facilityType.ID)
		return errLib.New("Facility type not found", http.StatusNotFound)
	}

	return nil
}

func (r *FacilityTypesRepository) DeleteFacilityType(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteFacilityType(c, id)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			log.Printf("Error deleting facility type with ID: %s, error: %v", id, err)
			return errLib.New("Internal server error", http.StatusInternalServerError)
		}
	}

	if row == 0 {
		log.Printf("Failed to delete facility type id: %+v", id)
		return errLib.New("Facility not found", http.StatusNotFound)
	}

	return nil
}
