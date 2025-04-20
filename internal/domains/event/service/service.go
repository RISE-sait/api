package service

import (
	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/event/persistence/repository"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
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
	eventsRepository         *repo.EventsRepository
	recurrencesRepository    *repo.RecurrencesRepository
	staffActivityLogsService *staffActivityLogs.Service
	db                       *sql.DB
}

func NewEventService(container *di.Container) *Service {
	return &Service{
		eventsRepository:         repo.NewEventsRepository(container),
		recurrencesRepository:    repo.NewRecurrencesRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		db:                       container.DB,
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

	if txErr := fn(s.eventsRepository.WithTx(tx)); txErr != nil {
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
	return s.eventsRepository.GetEvent(ctx, eventID)
}

func (s *Service) GetEvents(ctx context.Context, filter values.GetEventsFilter) ([]values.ReadEventValues, *errLib.CommonError) {
	return s.eventsRepository.GetEvents(ctx, filter)
}

func (s *Service) CreateEvents(ctx context.Context, details values.RecurrenceValues) *errLib.CommonError {

	events, err := generateEventsFromRecurrence(details)

	if err != nil {
		return err
	}

	return s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {
		// Create events
		if err = txRepo.CreateEvents(ctx, events); err != nil {
			return err
		}

		return s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			details.UpdatedBy,
			fmt.Sprintf("Created events with details: %+v", details),
		)
	})
}

func (s *Service) CreateEvent(ctx context.Context, details values.CreateEventValues) *errLib.CommonError {

	return s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {

		if err := txRepo.CreateEvents(ctx, []values.CreateEventValues{details}); err != nil {
			return err
		}

		isStaff, err := contextUtils.IsStaff(ctx)

		if err != nil {
			return err
		}

		if isStaff {

			return s.staffActivityLogsService.InsertStaffActivity(
				ctx,
				txRepo.GetTx(),
				details.CreatedBy,
				fmt.Sprintf("Created event with details: %+v", details),
			)
		}
		return nil
	})
}

func (s *Service) UpdateEvent(ctx context.Context, details values.UpdateEventValues) *errLib.CommonError {

	return s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {

		if err := txRepo.UpdateEvent(ctx, details); err != nil {
			return err
		}

		isStaff, err := contextUtils.IsStaff(ctx)
		if err != nil {
			return err
		}

		if isStaff {
			return s.staffActivityLogsService.InsertStaffActivity(
				ctx,
				txRepo.GetTx(),
				details.UpdatedBy,
				fmt.Sprintf("Updated event with ID and new details: %+v", details),
			)
		} else {
			return nil
		}

	})
}

// UpdateRecurringEvents updates existing events within a recurrence schedule
//
// - Updates events within the new recurrence period
// - Deletes events outside the new period
// - Returns *errLib.CommonError if recurrence is being extended
func (s *Service) UpdateRecurringEvents(ctx context.Context, details values.RecurrenceValues) *errLib.CommonError {

	return s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {

		if err := txRepo.DeleteUnmodifiedEventsByRecurrenceID(ctx, details.RecurrenceID); err != nil {
			return err
		}

		eventsToCreate, err := generateEventsFromRecurrence(details)

		if err != nil {
			return err
		}

		if err = txRepo.CreateEvents(ctx, eventsToCreate); err != nil {
			return err
		}

		return s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			details.UpdatedBy,
			fmt.Sprintf("Updated events with details: %+v", details),
		)
	})
}

func (s *Service) DeleteEvents(ctx context.Context, ids []uuid.UUID) *errLib.CommonError {

	var (
		err     *errLib.CommonError
		isStaff bool
		staffID uuid.UUID
	)

	return s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {

		if err = s.eventsRepository.DeleteEvents(ctx, ids); err != nil {
			return err
		}

		isStaff, err = contextUtils.IsStaff(ctx)
		if err != nil {
			return err
		}

		if isStaff {

			staffID, err = contextUtils.GetUserID(ctx)

			if err != nil {
				return err
			}

			return s.staffActivityLogsService.InsertStaffActivity(
				ctx,
				txRepo.GetTx(),
				staffID,
				fmt.Sprintf("Deleted events with ids: %+v", ids),
			)
		} else {
			return nil
		}
	})
}

func (s *Service) GetEventsSchedules(ctx context.Context, filter values.GetEventsFilter) ([]values.Schedule, *errLib.CommonError) {
	return s.recurrencesRepository.GetEventsRecurrences(ctx, filter.ProgramType, filter.ProgramID, filter.LocationID, filter.ParticipantID, filter.TeamID, filter.CreatedBy, filter.UpdatedBy, filter.Before, filter.After)
}

func generateEventsFromRecurrence(recurrence values.RecurrenceValues) ([]values.CreateEventValues, *errLib.CommonError) {

	const timeLayout = "15:04:05Z07:00"

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

	var events []values.CreateEventValues
	addEvent := func(date time.Time) {
		start := adjustTime(date, startTime)
		end := adjustTime(date, endTime)
		if end.Before(start) {
			end = end.AddDate(0, 0, 1)
		}
		events = append(events, values.CreateEventValues{
			CreatedBy: recurrence.UpdatedBy,
			EventDetails: values.EventDetails{
				StartAt:    start,
				EndAt:      end,
				ProgramID:  recurrence.ProgramID,
				LocationID: recurrence.LocationID,
				TeamID:     recurrence.TeamID,
				Capacity:   recurrence.Capacity,
			},
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
