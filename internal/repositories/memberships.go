package repositories

import (
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type MembershipsRepository struct {
	Queries *db.Queries
}

func (r *MembershipsRepository) CreateMembership(c context.Context, membership *db.CreateMembershipParams) *utils.HTTPError {
	row, err := r.Queries.CreateMembership(c, *membership)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to create membership: %+v", *membership)
		return utils.CreateHTTPError("Failed to create membership", http.StatusInternalServerError)
	}

	return nil
}

func (r *MembershipsRepository) GetMembershipById(c context.Context, id uuid.UUID) (*db.Membership, *utils.HTTPError) {
	membership, err := r.Queries.GetMembershipById(c, id)

	if err != nil {
		log.Printf("Failed to retrieve membership with ID: %s", id)
		return nil, utils.MapDatabaseError(err)
	}

	return &membership, nil
}

func (r *MembershipsRepository) GetAllMemberships(c context.Context) (*[]db.Membership, *utils.HTTPError) {
	memberships, err := r.Queries.GetAllMemberships(c)

	if err != nil {
		return &[]db.Membership{}, utils.MapDatabaseError(err)
	}

	return &memberships, nil
}

func (r *MembershipsRepository) UpdateMembership(c context.Context, membership *db.UpdateMembershipParams) *utils.HTTPError {
	row, err := r.Queries.UpdateMembership(c, *membership)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to update membership: %+v", *membership)
		return utils.CreateHTTPError("Membership not found", http.StatusNotFound)
	}

	return nil
}

func (r *MembershipsRepository) DeleteMembership(c context.Context, id uuid.UUID) *utils.HTTPError {
	row, err := r.Queries.DeleteFacilityType(c, id)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to delete membership id: %+v", id)
		return utils.CreateHTTPError("Membership not found", http.StatusNotFound)
	}

	return nil
}
