package service

import (
	"api/internal/di"
	repo "api/internal/domains/event/persistence/repository"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

type Service struct {
	repo *repo.EventsRepository
	db   *sql.DB
}

func NewEventService(container *di.Container) *Service {
	return &Service{
		repo: repo.NewEventsRepository(container),
		db:   container.DB,
	}
}

func (s *Service) executeInTx(ctx context.Context, fn func(repo *repo.EventsRepository) *errLib.CommonError) *errLib.CommonError {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("Rollback error (usually harmless): %v", err)
		}
	}()

	if txErr := fn(s.repo.WithTx(tx)); txErr != nil {
		return txErr
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction for events: %v", err)
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}
	return nil
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

// UpdateEvents updates existing events within a recurrence schedule
//
// - Updates events within the new recurrence period
// - Deletes events outside the new period
// - Returns error if recurrence is being extended
func (s *Service) UpdateEvents(ctx context.Context, details values.UpdateEventsValues) *errLib.CommonError {

	if details.NewRecurrenceEndAt.After(details.OriginalRecurrenceEndAt) {
		return errLib.New("Extending recurrence is not supported yet. Please create a new schedule, it will automatically extend the old one", http.StatusBadRequest)
	}

	return s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {

		existingEvents, err := txRepo.GetEvents(ctx, values.GetEventsFilter{
			ProgramID:  details.OriginalProgramID,
			LocationID: details.OriginalLocationID,
			TeamID:     details.OriginalTeamID,
			Before:     details.OriginalRecurrenceEndAt,
			After:      details.OriginalRecurrenceStartAt,
		})

		if err != nil {
			return err
		}

		if len(existingEvents) == 0 {
			return errLib.New("no matching events found to update", http.StatusNotFound)
		}

		var eventsToUpdate []values.ReadEventValues
		var eventsToDelete []uuid.UUID

		for _, event := range existingEvents {
			if event.StartAt.After(details.NewRecurrenceEndAt) {
				eventsToDelete = append(eventsToDelete, event.ID)
			} else {
				eventsToUpdate = append(eventsToUpdate, event)
			}
		}

		// 3. Delete events that are after the new recurrence end
		if len(eventsToDelete) > 0 {
			if err = txRepo.DeleteEvent(ctx, eventsToDelete); err != nil {
				log.Printf("Failed to delete events: %v", err)
				return errLib.New("Failed to delete events", http.StatusInternalServerError)
			}
		}

		updateParams, err := convertEventsForUpdate(
			EventTimeUpdate{
				NewStartTime: details.NewEventStartTime,
				NewEndTime:   details.NewEventEndTime,
			},
			EventIDUpdate{
				NewProgramID:  details.NewProgramID,
				NewLocationID: details.NewLocationID,
				NewTeamID:     details.NewTeamID,
			},
			details.NewCapacity,
			details.UpdatedBy,
			eventsToUpdate, // Use filtered list
		)

		if err != nil {
			return err
		}

		if err = txRepo.UpdateEvents(ctx, details.UpdatedBy, updateParams); err != nil {
			log.Printf("Failed to update events: %v", err)
			return errLib.New("Failed to update events", http.StatusInternalServerError)
		}

		return nil
	})
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

type EventTimeUpdate struct {
	NewStartTime string
	NewEndTime   string
}

type EventIDUpdate struct {
	NewProgramID  uuid.UUID
	NewLocationID uuid.UUID
	NewTeamID     uuid.UUID
}

func convertEventsForUpdate(
	timeUpdate EventTimeUpdate,
	idUpdate EventIDUpdate,
	newCapacity int32,
	updatedBy uuid.UUID,
	existingEvents []values.ReadEventValues,
) ([]values.UpdateEventValues, *errLib.CommonError) {

	// Parse times once
	startTime, err := time.Parse("15:04", timeUpdate.NewStartTime)
	if err != nil {
		return nil, errLib.New(fmt.Sprintf("invalid start time: %v", err), http.StatusBadRequest)
	}

	endTime, err := time.Parse("15:04", timeUpdate.NewEndTime)
	if err != nil {
		return nil, errLib.New(fmt.Sprintf("invalid end time: %v", err), http.StatusBadRequest)
	}

	if newCapacity <= 0 {
		return nil, errLib.New("capacity must be positive", http.StatusBadRequest)
	}

	if len(existingEvents) == 0 {
		return nil, errLib.New("no events to update", http.StatusBadRequest)
	}

	// Pre-allocate slice
	updates := make([]values.UpdateEventValues, 0, len(existingEvents))

	for _, evt := range existingEvents {
		newStart := updateTimeKeepingDate(evt.StartAt, startTime)
		newEnd := updateTimeKeepingDate(evt.EndAt, endTime)

		if newEnd.Before(newStart) {
			newEnd = newEnd.AddDate(0, 0, 1) // Handle midnight crossing
		}

		updates = append(updates, values.UpdateEventValues{
			ID:        evt.ID,
			UpdatedBy: updatedBy,
			Details: values.Details{
				StartAt:    newStart,
				EndAt:      newEnd,
				ProgramID:  idUpdate.NewProgramID,
				LocationID: idUpdate.NewLocationID,
				TeamID:     idUpdate.NewTeamID,
				Capacity:   newCapacity,
			},
		})
	}

	return updates, nil
}

func updateTimeKeepingDate(original time.Time, newTime time.Time) time.Time {
	return time.Date(
		original.Year(), original.Month(), original.Day(),
		newTime.Hour(), newTime.Minute(), 0, 0, original.Location(),
	)
}
