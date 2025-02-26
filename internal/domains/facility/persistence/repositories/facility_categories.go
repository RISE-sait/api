package repository

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	entity "api/internal/domains/facility/entity"
	db "api/internal/domains/facility/persistence/sqlc/generated"
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

	_, err := r.Queries.CreateFacilityCategory(c, name)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle specific Postgres errors
			switch pqErr.Code {
			case databaseErrors.UniqueViolation:
				return errLib.New("Facility type with the given name already exists", http.StatusConflict)
			}
		}
		log.Printf("Error creating facility type : %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *FacilityTypesRepository) GetFacilityType(c context.Context, id uuid.UUID) (*string, *errLib.CommonError) {
	name, err := r.Queries.GetFacilityCategoryById(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Facility type not found", http.StatusNotFound)
		}
		log.Printf("Failed to retrieve facility type with HubSpotId: %s, error: %v", id, err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &name, nil
}

func (r *FacilityTypesRepository) GetAllFacilityTypes(c context.Context, filter string) ([]entity.Category, *errLib.CommonError) {
	dbFacilityTypes, err := r.Queries.GetFacilityCategories(c)

	if err != nil {
		log.Printf("Error retrieving facility types: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	facilityTypes := make([]entity.Category, len(dbFacilityTypes))
	for i, facilityType := range dbFacilityTypes {
		facilityTypes[i] = entity.Category{
			ID:   facilityType.ID,
			Name: facilityType.Name,
		}
	}

	return facilityTypes, nil
}

func (r *FacilityTypesRepository) UpdateFacilityType(c context.Context, facilityType *entity.Category) *errLib.CommonError {

	dbParams := db.UpdateFacilityCategoryParams{
		ID:   facilityType.ID,
		Name: facilityType.Name,
	}

	_, err := r.Queries.UpdateFacilityCategory(c, dbParams)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle specific Postgres errors
			switch pqErr.Code {
			case databaseErrors.UniqueViolation:
				return errLib.New("Facility type with the given name already exists", http.StatusConflict)
			}
		}
		log.Printf("Error updating facility type with HubSpotId: %s, error: %v", facilityType.ID, err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *FacilityTypesRepository) DeleteFacilityType(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteFacilityCategory(c, id)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			log.Printf("Error deleting facility type with HubSpotId: %s, error: %v", id, err)
			return errLib.New("Internal server error", http.StatusInternalServerError)
		}
	}

	if row == 0 {
		log.Printf("Failed to delete facility type id: %+v", id)
		return errLib.New("Facility not found", http.StatusNotFound)
	}

	return nil
}
