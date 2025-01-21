package facility_types

import (
	"api/internal/types"
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
}

func (r *Repository) CreateFacilityType(c context.Context, name string) *types.HTTPError {
	row, err := r.Queries.CreateFacilityType(c, name)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to create facility type: %+v", name)
		return utils.CreateHTTPError("Internal Server Error", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) GetFacilityType(c context.Context, id uuid.UUID) (*db.FacilityType, *types.HTTPError) {
	facilityType, err := r.Queries.GetFacilityTypeById(c, id)

	if err != nil {
		return nil, utils.MapDatabaseError(err)
	}

	return &facilityType, nil
}

func (r *Repository) GetAllFacilityTypes(c context.Context) (*[]db.FacilityType, *types.HTTPError) {
	facilityTypes, err := r.Queries.GetAllFacilityTypes(c)

	if err != nil {
		return &[]db.FacilityType{}, utils.MapDatabaseError(err)
	}

	return &facilityTypes, nil
}

func (r *Repository) UpdateFacilityType(c context.Context, facilityType *db.UpdateFacilityTypeParams) *types.HTTPError {
	row, err := r.Queries.UpdateFacilityType(c, *facilityType)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to update facility type: %+v", *facilityType)
		return utils.CreateHTTPError("Facility type not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) DeleteFacilityType(c context.Context, id uuid.UUID) *types.HTTPError {
	row, err := r.Queries.DeleteFacilityType(c, id)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to delete facility type id: %+v", id)
		return utils.CreateHTTPError("Facility not found", http.StatusNotFound)
	}

	return nil
}
