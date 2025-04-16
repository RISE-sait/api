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

// GetEvent retrieves a single event by its ID from the repository.
//
// Parameters:
//   - ctx: The context for the request, used for cancellation and deadlines
//   - eventID: The UUID of the event to retrieve
//
// Returns:
//   - values.ReadEventValues: The event data if found
//   - *errLib.CommonError:
//   - nil if the operation was successful
//   - http.StatusNotFound (wrapped in CommonError) if no event was found with the given ID
//   - other repository errors if the operation fails
//
// Example:
//
//	event, err := svc.GetEvent(ctx, eventID)
//	if err.HTTPCode == http.StatusNotFound {
//		// Handle not found error
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
// - Returns *errLib.CommonError if recurrence is being extended
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

const timeLayout = "15:04:05Z07:00"

func generateEventsFromRecurrence(recurrence values.CreateEventsRecurrenceValues) ([]values.CreateEventsSpecificValues, *errLib.CommonError) {

	if recurrence.RecurrenceStartAt.After(recurrence.RecurrenceEndAt) {
		return nil, errLib.New("Recurrence start date must be before the end date", http.StatusBadRequest)
	}

	if recurrence.Capacity <= 0 {
		return nil, errLib.New("Capacity must be greater than zero", http.StatusBadRequest)
	}

	startTime, err := time.Parse(timeLayout, recurrence.EventStartTime)
	if err != nil {
		return nil, errLib.New(fmt.Sprintf("Invalid start time format - must be HH:MM:SS±HH:MM (e.g. 09:00:00+00:00)"), http.StatusBadRequest)
	}

	endTime, err := time.Parse(timeLayout, recurrence.EventEndTime)
	if err != nil {
		return nil, errLib.New(fmt.Sprintf("Invalid end time format - must be HH:MM:SS±HH:MM (e.g. 17:00:00+00:00)"), http.StatusBadRequest)
	}

	// If the end time is before start time (crosses midnight), add a day to end time
	if endTime.Before(startTime) {
		endTime = endTime.AddDate(0, 0, 1)
	}

	adjustTime := func(date time.Time, t time.Time) time.Time {
		return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location())
	}

	var events []values.CreateEventsSpecificValues
	addEvent := func(date time.Time) {
		start := adjustTime(date, startTime)
		end := adjustTime(date, endTime)
		if end.Before(start) {
			end = end.AddDate(0, 0, 1)
		}
		events = append(events, values.CreateEventsSpecificValues{
			CreatedBy:  recurrence.CreatedBy,
			StartAt:    start,
			EndAt:      end,
			ProgramID:  recurrence.ProgramID,
			LocationID: recurrence.LocationID,
			TeamID:     recurrence.TeamID,
			Capacity:   recurrence.Capacity,
		})
	}

	if recurrence.Day == nil {
		addEvent(recurrence.RecurrenceStartAt)
		return events, nil
	}

	for date := recurrence.RecurrenceStartAt; !date.After(recurrence.RecurrenceEndAt); date = date.AddDate(0, 0, 1) {
		if date.Weekday() == *recurrence.Day {
			addEvent(date)
			date = date.AddDate(0, 0, 6) // Skip to next week
		}
	}

	return events, nil
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
	startTime, err := time.Parse(timeLayout, timeUpdate.NewStartTime)
	if err != nil {
		return nil, errLib.New(fmt.Sprintf("Invalid start time format - must be HH:MM:SS±HH:MM (e.g. 09:00:00+00:00)"), http.StatusBadRequest)
	}

	endTime, err := time.Parse(timeLayout, timeUpdate.NewEndTime)
	if err != nil {
		return nil, errLib.New(fmt.Sprintf("Invalid end time format - must be HH:MM:SS±HH:MM (e.g. 17:00:00+00:00)"), http.StatusBadRequest)
	}

	if newCapacity <= 0 {
		return nil, errLib.New("capacity must be positive", http.StatusBadRequest)
	}

	if len(existingEvents) == 0 {
		return nil, errLib.New("no events to update", http.StatusBadRequest)
	}

	// Pre-allocate slice
	updates := make([]values.UpdateEventValues, len(existingEvents))

	for i, evt := range existingEvents {
		newStart := updateTimeKeepingDate(evt.StartAt, startTime)
		newEnd := updateTimeKeepingDate(evt.EndAt, endTime)

		if newEnd.Before(newStart) {
			newEnd = newEnd.AddDate(0, 0, 1) // Handle midnight crossing
		}

		updates[i] = values.UpdateEventValues{
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
		}
	}

	return updates, nil
}

func updateTimeKeepingDate(original time.Time, newTime time.Time) time.Time {
	return time.Date(
		original.Year(), original.Month(), original.Day(),
		newTime.Hour(), newTime.Minute(), 0, 0, original.Location(),
	)
}
