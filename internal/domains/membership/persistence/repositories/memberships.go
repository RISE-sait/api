package membership

import (
	"api/internal/di"
	db "api/internal/domains/membership/persistence/sqlc/generated"
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
}

func NewMembershipsRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.MembershipDb,
	}
}

func (r *Repository) Create(c context.Context, membership *values.CreateValues) *errLib.CommonError {

	dbParams := db.CreateMembershipParams{
		Name: membership.Name, Description: sql.NullString{
			String: membership.Description,
			Valid:  true,
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

func (r *Repository) GetByID(c context.Context, id uuid.UUID) (*values.ReadValues, *errLib.CommonError) {
	membership, err := r.Queries.GetMembershipById(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Membership not found", http.StatusNotFound)
		}
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &values.ReadValues{
		ID:          membership.ID,
		Name:        membership.Name,
		Description: membership.Description.String,
		UpdatedAt:   membership.UpdatedAt.Time,
		CreatedAt:   membership.CreatedAt.Time,
	}, nil
}

func (r *Repository) List(c context.Context) ([]values.ReadValues, *errLib.CommonError) {
	dbMemberships, err := r.Queries.GetAllMemberships(c)

	if err != nil {
		log.Println("Failed to get memberships: ", err.Error())
		return []values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	memberships := make([]values.ReadValues, len(dbMemberships))
	for i, dbMembership := range dbMemberships {
		memberships[i] = values.ReadValues{
			ID:          dbMembership.ID,
			Name:        dbMembership.Name,
			Description: dbMembership.Description.String,
			UpdatedAt:   dbMembership.UpdatedAt.Time,
			CreatedAt:   dbMembership.CreatedAt.Time,
		}
	}

	return memberships, nil
}

func (r *Repository) Update(c context.Context, membership *values.UpdateValues) *errLib.CommonError {

	dbMembershipParams := db.UpdateMembershipParams{
		ID:   membership.ID,
		Name: membership.Name,
		Description: sql.NullString{
			String: membership.Description,
			Valid:  true,
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

func (r *Repository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteMembership(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Membership not found", http.StatusNotFound)
	}

	return nil
}
