package waivers

import (
	"api/internal/types"
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"log"
	"net/http"
)

type Repository struct {
	Queries *db.Queries
}

func (r *Repository) GetWaiverByEmailAndDocLink(c context.Context, params *db.GetWaiverByEmailAndDocLinkParams) (*db.Waiver, *types.HTTPError) {
	waiver, err := r.Queries.GetWaiverByEmailAndDocLink(c, *params)

	if err != nil {
		log.Printf("Failed to retrieve waiver: %+v", *params)
		return nil, utils.MapDatabaseError(err)
	}

	return &waiver, nil
}

func (r *Repository) UpdateWaiverStatus(c context.Context, waiver *db.UpdateWaiverSignedStatusByEmailParams) *types.HTTPError {
	row, err := r.Queries.UpdateWaiverSignedStatusByEmail(c, *waiver)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {

		log.Printf("Failed to update membership: %+v", *waiver)
		return utils.CreateHTTPError("Waiver not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) GetAllUniqueWaivers(c context.Context) (*[]db.Waiver, *types.HTTPError) {
	waivers, err := r.Queries.GetAllUniqueWaiverDocs(c)

	if err != nil {
		return &[]db.Waiver{}, utils.CreateHTTPError(err.Error(), http.StatusInternalServerError)
	}
	return &waivers, nil
}
