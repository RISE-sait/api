package membership

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/membership/persistence/sqlc/generated"
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
	Tx      *sql.Tx
}

func (r *Repository) GetTx() *sql.Tx {
	return r.Tx
}

func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{
		Queries: r.Queries.WithTx(tx),
		Tx:      tx,
	}
}

func NewMembershipsRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.MembershipDb,
	}
}

func (r *Repository) Create(c context.Context, membership values.CreateValues) *errLib.CommonError {

	dbParams := db.CreateMembershipParams{
		Name: membership.Name, Description: membership.Description,
		Benefits: membership.Benefits,
	}

	_, err := r.Queries.CreateMembership(c, dbParams)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("There's an existing membership with this name", http.StatusInternalServerError)
		}
		log.Println("Failed to create membership: ", err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) GetByID(c context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError) {
	membership, err := r.Queries.GetMembershipById(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadValues{}, errLib.New("Membership not found", http.StatusNotFound)
		}
		return values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.ReadValues{
		ID: membership.ID,
		BaseValue: values.BaseValue{
			Name:        membership.Name,
			Description: membership.Description,
			Benefits:    membership.Benefits,
		},
		UpdatedAt: membership.UpdatedAt,
		CreatedAt: membership.CreatedAt,
	}, nil
}

func (r *Repository) List(c context.Context) ([]values.ReadValues, *errLib.CommonError) {
	dbMemberships, err := r.Queries.GetMemberships(c)

	if err != nil {
		log.Println("Failed to get memberships: ", err.Error())
		return []values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	memberships := make([]values.ReadValues, len(dbMemberships))
	for i, dbMembership := range dbMemberships {
		memberships[i] = values.ReadValues{
			ID: dbMembership.ID,
			BaseValue: values.BaseValue{
				Name:        dbMembership.Name,
				Description: dbMembership.Description,
				Benefits:    dbMembership.Benefits,
			},
			UpdatedAt: dbMembership.UpdatedAt,
			CreatedAt: dbMembership.CreatedAt,
		}
	}

	return memberships, nil
}

func (r *Repository) Update(c context.Context, membership values.UpdateValues) *errLib.CommonError {

	dbMembershipParams := db.UpdateMembershipParams{
		ID:          membership.ID,
		Name:        membership.Name,
		Description: membership.Description,
		Benefits:    membership.Benefits,
	}

	_, err := r.Queries.UpdateMembership(c, dbMembershipParams)

	if err != nil {
		log.Printf("Internal server error while updating membership: %s", err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteMembership(c, id)

	if err != nil {
		log.Println("Failed to delete membership: ", err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Membership not found", http.StatusNotFound)
	}

	return nil
}
