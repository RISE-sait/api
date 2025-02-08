package events

import (
	"api/internal/di"
	dto "api/internal/domains/events/dto"
	entity "api/internal/domains/events/entities"
	"api/internal/domains/events/values"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// EventsController provides HTTP handlers for managing events.
type EventsController struct {
	Service *EventsService
}

// NewEventsController creates a new instance of EventsController.
func NewEventsController(container *di.Container) *EventsController {
	return &EventsController{Service: NewEventsService(container)}
}

// GetAllEvents retrieves all events from the database.
func (c *EventsController) GetEvents(w http.ResponseWriter, r *http.Request) {

	courseIdStr := r.URL.Query().Get("course_id")
	facilityIdStr := r.URL.Query().Get("facility_id")

	beginTimeStr := r.URL.Query().Get("begin_datetime")
	endTimeStr := r.URL.Query().Get("end_datetime")

	var courseId uuid.UUID
	var facilityId uuid.UUID
	var beginDatetime time.Time
	var endDatetime time.Time

	if courseIdStr != "" {
		id, err := validators.ParseUUID(courseIdStr)

		if err != nil {
			response_handlers.RespondWithError(w, err)
			return
		}

		courseId = id
	}

	if facilityIdStr != "" {

		id, err := validators.ParseUUID(facilityIdStr)

		if err != nil {
			response_handlers.RespondWithError(w, err)
			return
		}

		facilityId = id
	}

	if beginTimeStr != "" {
		datetime, err := validators.ParseTime(beginTimeStr)
		if err != nil {
			response_handlers.RespondWithError(w, err)
			return
		}

		beginDatetime = datetime
	}

	if endTimeStr != "" {
		datetime, err := validators.ParseTime(endTimeStr)
		if err != nil {
			response_handlers.RespondWithError(w, err)
			return
		}

		endDatetime = datetime
	}

	details := values.EventDetails{
		BeginTime:  beginDatetime,
		EndTime:    endDatetime,
		CourseID:   courseId,
		FacilityID: facilityId,
	}

	events, err := c.Service.GetEvents(r.Context(), details)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.EventResponse, len(events))

	for i, event := range events {
		result[i] = mapEntityToResponse(&event)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

func (c *EventsController) CreateEvent(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.EventRequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	eventCreate, err := targetBody.ToEventDetails()

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := c.Service.CreateEvent(r.Context(), eventCreate); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (c *EventsController) UpdateEvent(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var targetBody dto.EventRequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	params, err := (&targetBody).ToEventAllFields(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := c.Service.UpdateEvent(r.Context(), params); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (c *EventsController) DeleteEvent(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
	}

	if err = c.Service.DeleteEvent(r.Context(), id); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (c *EventsController) GetCustomersCountByEventId(w http.ResponseWriter, r *http.Request) {

	eventIdStr := chi.URLParam(r, "id")

	var eventId uuid.UUID

	if eventIdStr != "" {
		id, err := validators.ParseUUID(eventIdStr)

		if err != nil {
			response_handlers.RespondWithError(w, err)
			return
		}

		eventId = id
	}

	count, err := c.Service.GetCustomersCountByEventId(r.Context(), eventId)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, count, http.StatusOK)
}

func mapEntityToResponse(event *entity.Event) dto.EventResponse {
	return dto.EventResponse{
		ID:        event.ID,
		BeginTime: event.BeginTime.Format("15:04"), // Convert to "HH:MM:SS"
		EndTime:   event.EndTime.Format("15:04"),
		Course:    event.Course,
		Facility:  event.Facility,
		Day:       event.Day,
	}
}
