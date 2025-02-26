package event_capacity

import (
	db "api/internal/domains/enrollment/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

var _ EventCapacityRepositoryInterface = (*Repository)(nil)

type Repository struct {
	Queries *db.Queries
}

func NewEventCapacityRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) GetEventIsFull(c context.Context, eventId uuid.UUID) (*bool, *errLib.CommonError) {

	isFull, err := r.Queries.GetEventIsFull(c, eventId)

	if err != nil {
		log.Printf("Error getting info: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &isFull, nil
}
