package membership

import (
	"api/internal/di"
	db "api/internal/domains/membership/persistence/sqlc/generated"
	values "api/internal/domains/membership/values/memberships"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type MembershipsRepository struct {
	Queries *db.Queries
}

func NewMembershipsRepository(container *di.Container) *MembershipsRepository {
	return &MembershipsRepository{
		Queries: container.Queries.MembershipDb,
	}
}

func (r *MembershipsRepository) Create(c context.Context, membership *values.MembershipDetails) *errLib.CommonError {

	dbParams := db.CreateMembershipParams{
		Name: membership.Name, Description: sql.NullString{
			String: membership.Description,
			Valid:  membership.Description != "",
		},
	}

	row, err := r.Queries.CreateMembership(c, dbParams)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Membership not created", http.StatusInternalServerError)
	}

	return nil
}

func (r *MembershipsRepository) GetByID(c context.Context, id uuid.UUID) (*values.MembershipAllFields, *errLib.CommonError) {
	membership, err := r.Queries.GetMembershipById(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Membership not found", http.StatusNotFound)
		}
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &values.MembershipAllFields{
		ID: membership.ID,
		MembershipDetails: values.MembershipDetails{
			Name:        membership.Name,
			Description: membership.Description.String,
		},
	}, nil
}

func (r *MembershipsRepository) List(c context.Context, after string) ([]values.MembershipAllFields, *errLib.CommonError) {
	dbMemberships, err := r.Queries.GetAllMemberships(c)

	if err != nil {
		log.Println("Failed to get memberships: ", err.Error())
		return []values.MembershipAllFields{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	memebrships := make([]values.MembershipAllFields, len(dbMemberships))
	for i, dbCourse := range dbMemberships {
		memebrships[i] = values.MembershipAllFields{
			ID: dbCourse.ID,
			MembershipDetails: values.MembershipDetails{
				Name:        dbCourse.Name,
				Description: dbCourse.Description.String,
			},
		}
	}

	return memebrships, nil
}

func (r *MembershipsRepository) Update(c context.Context, membership *values.MembershipAllFields) *errLib.CommonError {

	dbMembershipParams := db.UpdateMembershipParams{
		ID:   membership.ID,
		Name: membership.Name,
		Description: sql.NullString{
			String: membership.Description,
			Valid:  membership.Description != "",
		},
	}

	row, err := r.Queries.UpdateMembership(c, dbMembershipParams)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Membership not found", http.StatusNotFound)
	}

	return nil
}

func (r *MembershipsRepository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteMembership(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Membership not found", http.StatusNotFound)
	}

	return nil
}
