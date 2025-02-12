package persistence

import (
	"api/internal/di"
	entity "api/internal/domains/events/entities"
	db "api/internal/domains/events/persistence/sqlc/generated"
	"api/internal/domains/events/values"
	errLib "api/internal/libs/errors"

	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type EventsRepository struct {
	Queries *db.Queries
}

func NewEventsRepository(container *di.Container) *EventsRepository {
	return &EventsRepository{
		Queries: container.Queries.EventDb,
	}
}

func (r *EventsRepository) CreateEvent(c context.Context, event *values.EventDetails) *errLib.CommonError {

	dbParams := db.CreateEventParams{
		BeginTime: event.BeginTime,
		EndTime:   event.EndTime,
		CourseID: uuid.NullUUID{
			UUID:  event.CourseID,
			Valid: event.CourseID != uuid.Nil,
		},
		FacilityID: event.FacilityID,
		Day:        db.DayEnum(event.Day),
	}

	row, err := r.Queries.CreateEvent(c, dbParams)

	if err != nil {
		log.Printf("Failed to create event: %+v. Error: %v", event, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Course or facility not found", http.StatusNotFound)
	}

	return nil
}

func (r *EventsRepository) GetEvents(ctx context.Context, fields values.EventDetails) ([]entity.Event, *errLib.CommonError) {

	dbParams := db.GetEventsParams{
		BeginTime:  fields.BeginTime,
		EndTime:    fields.EndTime,
		FacilityID: fields.FacilityID,
		CourseID: uuid.NullUUID{
			UUID:  fields.CourseID,
			Valid: fields.CourseID != uuid.Nil,
		},
	}

	dbevents, err := r.Queries.GetEvents(ctx, dbParams)

	if err != nil {
		log.Println("Failed to get events: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	events := make([]entity.Event, len(dbevents))
	for i, dbevent := range dbevents {

		events[i] = entity.Event{
			ID: dbevent.ID,
			Course: &entity.Course{
				ID:   dbevent.CourseID,
				Name: dbevent.Course,
			},
			Facility:  dbevent.Facility,
			BeginTime: dbevent.BeginTime,
			EndTime:   dbevent.EndTime,
			Day:       string(dbevent.Day),
		}

	}

	return events, nil
}

func (r *EventsRepository) UpdateEvent(c context.Context, event *entity.Event) (*entity.Event, *errLib.CommonError) {
	dbEventParams := db.UpdateEventParams{
		BeginTime: event.BeginTime,
		EndTime:   event.EndTime,
		CourseID: uuid.NullUUID{
			UUID:  event.Course.ID,
			Valid: event.Course.ID != uuid.Nil,
		},
		FacilityID: event.FacilityID,
		Day:        db.DayEnum(event.Day),
		ID:         event.ID,
	}

	dbEvent, err := r.Queries.UpdateEvent(c, dbEventParams)

	if err != nil {
		log.Printf("Failed to update event: %+v. Error: %v", event, err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entity.Event{
		ID: dbEvent.ID,
		Course: &entity.Course{
			ID:   dbEvent.CourseID.UUID,
			Name: dbEvent.CourseName,
		},
		Facility:   dbEvent.FacilityName,
		FacilityID: dbEvent.FacilityID,
		BeginTime:  dbEvent.BeginTime,
		EndTime:    dbEvent.EndTime,
		Day:        string(dbEvent.Day),
	}, nil

}

func (r *EventsRepository) DeleteEvent(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteEvent(c, id)

	if err != nil {
		log.Printf("Failed to delete event with ID: %s. Error: %s", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Event not found", http.StatusNotFound)
	}

	return nil
}

func (r *EventsRepository) GetEventDetails(ctx context.Context, id uuid.UUID) (*entity.Event, *errLib.CommonError) {

	eventDetails, err := r.Queries.GetEventById(ctx, id)

	if err != nil {
		log.Println("Failed to get event details: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	event := &entity.Event{
		ID:        eventDetails.ID,
		Course:    &entity.Course{ID: eventDetails.CourseID, Name: eventDetails.Course},
		Facility:  eventDetails.Facility,
		BeginTime: eventDetails.BeginTime,
		EndTime:   eventDetails.EndTime,
		Day:       string(eventDetails.Day),
	}

	return event, nil
}
