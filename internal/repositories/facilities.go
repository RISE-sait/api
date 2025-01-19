package repositories

import (
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type FacilityRepository struct {
	Queries *db.Queries
}

func (r *FacilityRepository) CreateFacility(c context.Context, params *db.CreateFacilityParams) *utils.HTTPError {
	row, err := r.Queries.CreateFacility(c, *params)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No facility created", http.StatusInternalServerError)
	}

	return nil
}

func (r *FacilityRepository) GetFacility(c context.Context, id uuid.UUID) (*db.Facility, *utils.HTTPError) {
	facility, err := r.Queries.GetFacilityById(c, id)

	if err != nil {
		log.Printf("Failed to retrieve facility with ID: %s", id)
		return nil, utils.MapDatabaseError(err)
	}

	return &facility, nil
}

func (r *FacilityRepository) GetAllFacilities(c context.Context) (*[]db.Facility, *utils.HTTPError) {
	facilities, err := r.Queries.GetAllFacilities(c)

	if err != nil {
		return &[]db.Facility{}, utils.CreateHTTPError(err.Error(), http.StatusInternalServerError)
	}

	return &facilities, nil
}

func (r *FacilityRepository) UpdateFacility(c context.Context, facility *db.UpdateFacilityParams) *utils.HTTPError {
	row, err := r.Queries.UpdateFacility(c, *facility)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("Facility not found", http.StatusNotFound)
	}

	return nil
}

func (r *FacilityRepository) DeleteFacility(c context.Context, id uuid.UUID) *utils.HTTPError {
	row, err := r.Queries.DeleteFacility(c, id)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to delete facility id: %+v", id)
		return utils.CreateHTTPError("Facility not found", http.StatusNotFound)
	}

	return nil
}
