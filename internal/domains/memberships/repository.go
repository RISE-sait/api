package memberships

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

func (r *Repository) CreateMembership(c context.Context, membership *db.CreateMembershipParams) *types.HTTPError {
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

func (r *Repository) GetMembershipById(c context.Context, id uuid.UUID) (*db.Membership, *types.HTTPError) {
	membership, err := r.Queries.GetMembershipById(c, id)

	if err != nil {
		log.Printf("Failed to retrieve membership with ID: %s", id)
		return nil, utils.MapDatabaseError(err)
	}

	return &membership, nil
}

func (r *Repository) GetAllMemberships(c context.Context) (*[]db.Membership, *types.HTTPError) {
	memberships, err := r.Queries.GetAllMemberships(c)

	if err != nil {
		return &[]db.Membership{}, utils.MapDatabaseError(err)
	}

	return &memberships, nil
}

func (r *Repository) UpdateMembership(c context.Context, membership *db.UpdateMembershipParams) *types.HTTPError {
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

func (r *Repository) DeleteMembership(c context.Context, id uuid.UUID) *types.HTTPError {
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
