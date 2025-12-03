package court

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/court/persistence/sqlc/generated"
	values "api/internal/domains/court/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository struct {
	Queries *db.Queries
	Tx      *sql.Tx
}

func NewRepository(container *di.Container) *Repository {
	return &Repository{Queries: container.Queries.CourtDb}
}

func (r *Repository) GetTx() *sql.Tx { return r.Tx }

func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{Queries: r.Queries.WithTx(tx), Tx: tx}
}

func (r *Repository) Create(ctx context.Context, d values.CreateDetails) (values.ReadValues, *errLib.CommonError) {
	params := db.CreateCourtParams{LocationID: d.LocationID, Name: d.Name}
	row, err := r.Queries.CreateCourt(ctx, params)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return values.ReadValues{}, errLib.New("court already exists", http.StatusConflict)
		}
		log.Printf("error creating court: %v", err)
		return values.ReadValues{}, errLib.New("internal server error", http.StatusInternalServerError)
	}
	return values.ReadValues{
		ID: row.ID,
		BaseDetails: values.BaseDetails{
			Name:       row.Name,
			LocationID: row.LocationID,
		},
	}, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError) {
	row, err := r.Queries.GetCourtById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadValues{}, errLib.New("court not found", http.StatusNotFound)
		}
		log.Printf("error getting court: %v", err)
		return values.ReadValues{}, errLib.New("internal server error", http.StatusInternalServerError)
	}
	return values.ReadValues{
		ID:           row.ID,
		LocationName: row.LocationName,
		BaseDetails:  values.BaseDetails{LocationID: row.LocationID, Name: row.Name},
	}, nil
}

func (r *Repository) List(ctx context.Context) ([]values.ReadValues, *errLib.CommonError) {
	dbCourts, err := r.Queries.GetCourts(ctx)
	if err != nil {
		log.Printf("error listing courts: %v", err)
		return nil, errLib.New("internal server error", http.StatusInternalServerError)
	}
	courts := make([]values.ReadValues, len(dbCourts))
	for i, c := range dbCourts {
		courts[i] = values.ReadValues{
			ID:           c.ID,
			LocationName: c.LocationName,
			BaseDetails:  values.BaseDetails{LocationID: c.LocationID, Name: c.Name},
		}
	}
	return courts, nil
}

func (r *Repository) Update(ctx context.Context, d values.UpdateDetails) *errLib.CommonError {
	params := db.UpdateCourtParams{ID: d.ID, LocationID: d.LocationID, Name: d.Name}
	count, err := r.Queries.UpdateCourt(ctx, params)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("court already exists", http.StatusConflict)
		}
		log.Printf("error updating court: %v", err)
		return errLib.New("internal server error", http.StatusInternalServerError)
	}
	if count == 0 {
		return errLib.New("court not found", http.StatusNotFound)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	count, err := r.Queries.DeleteCourt(ctx, id)
	if err != nil {
		log.Printf("error deleting court: %v", err)
		return errLib.New("internal server error", http.StatusInternalServerError)
	}
	if count == 0 {
		return errLib.New("court not found", http.StatusNotFound)
	}
	return nil
}
