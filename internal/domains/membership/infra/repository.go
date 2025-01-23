package membership

import (
	"api/internal/domains/membership/dto"
	db "api/internal/domains/membership/infra/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"net/http"

	"github.com/google/uuid"
)

type MembershipsRepository struct {
	Queries *db.Queries
}

func (r *MembershipsRepository) CreateMembership(c context.Context, membership *dto.CreateMembershipRequest) *errLib.CommonError {

	dbParams := membership.ToDBParams()

	row, err := r.Queries.CreateMembership(c, *dbParams)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		return errLib.New("Membership not created", http.StatusInternalServerError)
	}

	return nil
}

func (r *MembershipsRepository) GetMembershipById(c context.Context, id uuid.UUID) (*db.Membership, *errLib.CommonError) {
	membership, err := r.Queries.GetMembershipById(c, id)

	if err != nil {
		return nil, errLib.TranslateDBErrorToCommonError(err)
	}
	return &membership, nil
}

func (r *MembershipsRepository) GetAllMemberships(c context.Context) ([]db.Membership, *errLib.CommonError) {
	memberships, err := r.Queries.GetAllMemberships(c)

	if err != nil {

		dbErr := errLib.TranslateDBErrorToCommonError(err)
		return []db.Membership{}, dbErr
	}

	return memberships, nil
}

func (r *MembershipsRepository) UpdateMembership(c context.Context, membership *dto.UpdateMembershipRequest) *errLib.CommonError {

	dbMembershipParams := membership.ToDBParams()

	row, err := r.Queries.UpdateMembership(c, *dbMembershipParams)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		return errLib.New("Membership not found", http.StatusNotFound)
	}

	return nil
}

func (r *MembershipsRepository) DeleteMembership(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteMembership(c, id)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		return errLib.New("Membership not found", http.StatusNotFound)
	}

	return nil
}
