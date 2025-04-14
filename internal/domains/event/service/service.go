package service

import (
	"api/internal/di"
	repo "api/internal/domains/event/persistence/repository"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type Service struct {
	repo *repo.EventsRepository
}

func NewEventService(container *di.Container) *Service {
	return &Service{
		repo: repo.NewEventsRepository(container),
	}
}

func (s *Service) GetEvent(ctx context.Context, eventID uuid.UUID) (values.ReadEventValues, *errLib.CommonError) {
	return s.repo.GetEvent(ctx, eventID)
}

func (s *Service) GetEvents(ctx context.Context, filter values.GetEventsFilter) ([]values.ReadEventValues, *errLib.CommonError) {
	return s.repo.GetEvents(ctx, filter)
}

func (s *Service) CreateEvents(ctx context.Context, details values.CreateEventsRecurrenceValues) *errLib.CommonError {

	specificDates, err := generateEventsFromRecurrence(details)

	if err != nil {
		return err
	}

	return s.repo.CreateEvents(ctx, specificDates)
}

func (s *Service) UpdateEvent(ctx context.Context, details values.UpdateEventValues) *errLib.CommonError {
	return s.repo.UpdateEvent(ctx, details)
}

func (s *Service) DeleteEvents(ctx context.Context, ids []uuid.UUID) *errLib.CommonError {
	return s.repo.DeleteEvent(ctx, ids)
}

func generateEventsFromRecurrence(recurrence values.CreateEventsRecurrenceValues) ([]values.CreateEventsSpecificValues, *errLib.CommonError) {
	var specificEvents []values.CreateEventsSpecificValues

	if recurrence.RecurrenceStartAt.After(recurrence.RecurrenceEndAt) {
		return nil, errLib.New("Recurrence start date must be before the end date", http.StatusBadRequest)
	}

	if recurrence.Capacity <= 0 {
		return nil, errLib.New("Capacity must be greater than zero", http.StatusBadRequest)
	}

	// Parse the time strings into time.Time objects for the current day
	eventDate := recurrence.RecurrenceStartAt
	layout := "15:04" // Assuming time format is HH:MM

	startTime, err := time.Parse(layout, recurrence.EventStartTime)
	if err != nil {
		return nil, errLib.New("Invalid start time format", http.StatusBadRequest)
	}

	endTime, err := time.Parse(layout, recurrence.EventEndTime)
	if err != nil {
		return nil, errLib.New("Invalid end time format", http.StatusBadRequest)
	}

	// Adjust the times to be on the event date
	startTime = time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, eventDate.Location())

	endTime = time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, eventDate.Location())

	if recurrence.Day == nil {
		// non recurring
		specificEvent := values.CreateEventsSpecificValues{
			CreatedBy:  recurrence.CreatedBy,
			StartAt:    startTime,
			EndAt:      endTime,
			ProgramID:  recurrence.ProgramID,
			LocationID: recurrence.LocationID,
			TeamID:     recurrence.TeamID,
			Capacity:   recurrence.Capacity,
		}
		specificEvents = append(specificEvents, specificEvent)

		return specificEvents, nil
	}

	// If the end time is before start time (crosses midnight), add a day to end time
	if endTime.Before(startTime) {
		endTime = endTime.AddDate(0, 0, 1)
	}

	// Iterate through each week between start and end dates
	for currentDate := recurrence.RecurrenceStartAt; !currentDate.After(recurrence.RecurrenceEndAt); currentDate = currentDate.AddDate(0, 0, 7) {

		// Find the next occurrence of the specified weekday
		eventDate = currentDate
		for eventDate.Weekday() != *recurrence.Day {
			eventDate = eventDate.AddDate(0, 0, 1)

			// If we've moved past the end date, break
			if eventDate.After(recurrence.RecurrenceEndAt) {
				break
			}
		}

		// Check if we're still within the recurrence period
		if eventDate.After(recurrence.RecurrenceEndAt) {
			continue
		}

		// Set the specific times for this event
		eventStart := time.Date(
			eventDate.Year(), eventDate.Month(), eventDate.Day(),
			startTime.Hour(), startTime.Minute(), 0, 0, eventDate.Location(),
		)

		eventEnd := time.Date(
			eventDate.Year(), eventDate.Month(), eventDate.Day(),
			endTime.Hour(), endTime.Minute(), 0, 0, eventDate.Location(),
		)

		// Handle midnight crossing
		if eventEnd.Before(eventStart) {
			eventEnd = eventEnd.AddDate(0, 0, 1)
		}

		// Create the specific event
		specificEvent := values.CreateEventsSpecificValues{
			CreatedBy:  recurrence.CreatedBy,
			StartAt:    eventStart,
			EndAt:      eventEnd,
			ProgramID:  recurrence.ProgramID,
			LocationID: recurrence.LocationID,
			TeamID:     recurrence.TeamID,
			Capacity:   recurrence.Capacity,
		}

		specificEvents = append(specificEvents, specificEvent)
	}

	return specificEvents, nil
}
