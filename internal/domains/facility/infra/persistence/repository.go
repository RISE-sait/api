package persistence

import (
	entity "api/internal/domains/facility/entities"
	db "api/internal/domains/facility/infra/persistence/sqlc/generated"
	"api/internal/domains/facility/values"
	errLib "api/internal/libs/errors"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type FacilityRepository struct {
	Queries *db.Queries
}

func (r *FacilityRepository) CreateFacility(c context.Context, facility *values.FacilityCreate) *errLib.CommonError {

	dbParams := db.CreateFacilityParams{
		Name:           facility.Name,
		Location:       facility.Location,
		FacilityTypeID: facility.FacilityTypeID,
	}

	row, err := r.Queries.CreateFacility(c, dbParams)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		return errLib.New("Facility not created", http.StatusInternalServerError)
	}

	return nil
}

func (r *FacilityRepository) GetFacility(c context.Context, id uuid.UUID) (*entity.Facility, *errLib.CommonError) {
	facility, err := r.Queries.GetFacilityById(c, id)

	if err != nil {
		log.Printf("Failed to retrieve facility with ID: %s", id)
		return nil, errLib.New("Facility not found", http.StatusNotFound)
	}

	return &entity.Facility{
		ID:             facility.ID,
		Name:           facility.Name,
		Location:       facility.Location,
		FacilityTypeID: facility.FacilityTypeID,
	}, nil
}

func (r *FacilityRepository) GetAllFacilities(c context.Context, filter string) ([]entity.Facility, *errLib.CommonError) {
	dbFacilities, err := r.Queries.GetAllFacilities(c)

	if err != nil {
		dbErr := errLib.TranslateDBErrorToCommonError(err)
		return []entity.Facility{}, dbErr
	}

	courses := make([]entity.Facility, len(dbFacilities))
	for i, dbFacility := range dbFacilities {
		courses[i] = entity.Facility{
			ID:             dbFacility.ID,
			Name:           dbFacility.Name,
			Location:       dbFacility.Location,
			FacilityTypeID: dbFacility.FacilityTypeID,
		}
	}

	return courses, nil
}

func (r *FacilityRepository) UpdateFacility(c context.Context, facility *values.FacilityUpdate) *errLib.CommonError {

	dbParams := db.UpdateFacilityParams{
		ID:             facility.ID,
		Name:           facility.Name,
		Location:       facility.Location,
		FacilityTypeID: facility.FacilityTypeID,
	}

	row, err := r.Queries.UpdateFacility(c, dbParams)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		return errLib.New("Facility not found", http.StatusNotFound)
	}

	return nil
}

func (r *FacilityRepository) DeleteFacility(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteFacility(c, id)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		log.Printf("Failed to delete facility id: %+v", id)
		return errLib.New("Facility not found", http.StatusNotFound)
	}

	return nil
}
