package persistence

import (
	entity "api/internal/domains/membership/entities"
	db "api/internal/domains/membership/infra/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"net/http"

	"github.com/google/uuid"
)

type MembershipsRepository struct {
	Queries *db.Queries
}

func (r *MembershipsRepository) Create(c context.Context, membership *entity.Membership) *errLib.CommonError {

	dbParams := db.CreateMembershipParams{
		Name: membership.Name, Description: sql.NullString{
			String: membership.Description,
			Valid:  membership.Description != "",
		},
		StartDate: membership.StartDate,
		EndDate:   membership.EndDate,
	}

	row, err := r.Queries.CreateMembership(c, dbParams)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		return errLib.New("Membership not created", http.StatusInternalServerError)
	}

	return nil
}

func (r *MembershipsRepository) GetByID(c context.Context, id uuid.UUID) (*entity.Membership, *errLib.CommonError) {
	membership, err := r.Queries.GetMembershipById(c, id)

	if err != nil {
		return nil, errLib.TranslateDBErrorToCommonError(err)
	}
	return &entity.Membership{
		ID:          membership.ID,
		Name:        membership.Name,
		Description: membership.Description.String,
		StartDate:   membership.StartDate,
		EndDate:     membership.EndDate,
	}, nil
}

func (r *MembershipsRepository) List(c context.Context, after string) ([]entity.Membership, *errLib.CommonError) {
	dbMemberships, err := r.Queries.GetAllMemberships(c)

	if err != nil {

		dbErr := errLib.TranslateDBErrorToCommonError(err)
		return []entity.Membership{}, dbErr
	}

	memebrships := make([]entity.Membership, len(dbMemberships))
	for i, dbCourse := range dbMemberships {
		memebrships[i] = entity.Membership{
			ID:          dbCourse.ID,
			Name:        dbCourse.Name,
			Description: dbCourse.Description.String,
			StartDate:   dbCourse.StartDate,
			EndDate:     dbCourse.EndDate,
		}
	}

	return memebrships, nil
}

func (r *MembershipsRepository) Update(c context.Context, membership *entity.Membership) *errLib.CommonError {

	dbMembershipParams := db.UpdateMembershipParams{
		ID:   membership.ID,
		Name: membership.Name,
		Description: sql.NullString{
			String: membership.Description,
			Valid:  membership.Description != "",
		},
		StartDate: membership.StartDate,
		EndDate:   membership.EndDate,
	}

	row, err := r.Queries.UpdateMembership(c, dbMembershipParams)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		return errLib.New("Membership not found", http.StatusNotFound)
	}

	return nil
}

func (r *MembershipsRepository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteMembership(c, id)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		return errLib.New("Membership not found", http.StatusNotFound)
	}

	return nil
}
