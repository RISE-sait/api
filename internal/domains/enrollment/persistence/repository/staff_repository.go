package enrollment

import (
	db "api/internal/domains/enrollment/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"

	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type StaffsRepository struct {
	Queries *db.Queries
}

func NewEventStaffsRepository(dbQueries *db.Queries) *StaffsRepository {
	return &StaffsRepository{
		Queries: dbQueries,
	}
}

func (r *StaffsRepository) AssignStaffToEvent(c context.Context, eventId, staffId uuid.UUID) *errLib.CommonError {

	dbParams := db.AssignStaffToEventParams{
		EventID: eventId,
		StaffID: staffId,
	}

	if _, err := r.Queries.AssignStaffToEvent(c, dbParams); err != nil {
		log.Printf("Failed to assign staff %+v to event: %+v. Error: %v", staffId, eventId, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
func (r *StaffsRepository) UnassignedStaffFromEvent(c context.Context, eventId, staffId uuid.UUID) *errLib.CommonError {

	dbParams := db.UnassignStaffFromEventParams{
		EventID: eventId,
		StaffID: staffId,
	}

	if _, err := r.Queries.UnassignStaffFromEvent(c, dbParams); err != nil {
		log.Printf("Failed to unassign staff: %+v from event %+v. Error: %v", staffId, eventId, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil

}
