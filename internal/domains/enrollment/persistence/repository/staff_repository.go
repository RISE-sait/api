package enrollment

import (
	"api/internal/di"
	db "api/internal/domains/enrollment/persistence/sqlc/generated"
	"api/internal/domains/event/service"
	errLib "api/internal/libs/errors"

	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type StaffsRepository struct {
	Queries      *db.Queries
	EventService *service.Service
}

func NewEventStaffsRepository(container *di.Container) *StaffsRepository {
	return &StaffsRepository{
		Queries:      container.Queries.EnrollmentDb,
		EventService: service.NewEventService(container),
	}
}

func (r *StaffsRepository) AssignStaffToEvent(ctx context.Context, eventId, staffId uuid.UUID) *errLib.CommonError {

	if _, err := r.EventService.GetEvent(ctx, eventId); err != nil {
		return err
	}

	dbParams := db.AssignStaffToEventParams{
		EventID: eventId,
		StaffID: staffId,
	}

	if _, err := r.Queries.AssignStaffToEvent(ctx, dbParams); err != nil {
		log.Printf("Failed to assign staff %+v to event: %+v. Error: %v", staffId, eventId, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *StaffsRepository) UnassignedStaffFromEvent(ctx context.Context, eventId, staffId uuid.UUID) *errLib.CommonError {

	if _, err := r.EventService.GetEvent(ctx, eventId); err != nil {
		return err
	}

	dbParams := db.UnassignStaffFromEventParams{
		EventID: eventId,
		StaffID: staffId,
	}

	if _, err := r.Queries.UnassignStaffFromEvent(ctx, dbParams); err != nil {
		log.Printf("Failed to unassign staff: %+v from event %+v. Error: %v", staffId, eventId, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil

}
