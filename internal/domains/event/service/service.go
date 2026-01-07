package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	dto "api/internal/domains/event/dto"
	repo "api/internal/domains/event/persistence/repository"
	values "api/internal/domains/event/values"
	stripeService "api/internal/domains/payment/services/stripe"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	txUtils "api/utils/db"

	"github.com/google/uuid"
)

type Service struct {
	eventsRepository         *repo.EventsRepository
	recurrencesRepository    *repo.RecurrencesRepository
	staffActivityLogsService *staffActivityLogs.Service
	productService           *stripeService.ProductService
	notificationService      *EventNotificationService
	db                       *sql.DB
}

func NewEventService(container *di.Container) *Service {
	return &Service{
		eventsRepository:         repo.NewEventsRepository(container),
		recurrencesRepository:    repo.NewRecurrencesRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		productService:           stripeService.NewProductService(),
		notificationService:      NewEventNotificationService(container),
		db:                       container.DB,
	}
}

func (s *Service) executeInTx(ctx context.Context, fn func(repo *repo.EventsRepository) *errLib.CommonError) *errLib.CommonError {
	return txUtils.ExecuteInTx(ctx, s.db, func(tx *sql.Tx) *errLib.CommonError {
		return fn(s.eventsRepository.WithTx(tx))
	})
}

// lookupNames fetches human-readable names for program, location, and team
func (s *Service) lookupNames(ctx context.Context, programID, locationID, teamID uuid.UUID) (programName, locationName, teamName string) {
	programName = ""
	locationName = ""
	teamName = ""

	if programID != uuid.Nil {
		var name sql.NullString
		_ = s.db.QueryRowContext(ctx, "SELECT name FROM program.programs WHERE id = $1", programID).Scan(&name)
		if name.Valid {
			programName = name.String
		}
	}

	if locationID != uuid.Nil {
		var name sql.NullString
		_ = s.db.QueryRowContext(ctx, "SELECT name FROM facilities.locations WHERE id = $1", locationID).Scan(&name)
		if name.Valid {
			locationName = name.String
		}
	}

	if teamID != uuid.Nil {
		var name sql.NullString
		_ = s.db.QueryRowContext(ctx, "SELECT name FROM athletic.teams WHERE id = $1", teamID).Scan(&name)
		if name.Valid {
			teamName = name.String
		}
	}

	return programName, locationName, teamName
}

// formatEventDescription creates a human-readable activity description for an event
func (s *Service) formatEventDescription(ctx context.Context, action string, details values.EventDetails) string {
	loc, _ := time.LoadLocation("America/Edmonton")
	if loc == nil {
		loc = time.FixedZone("MST", -7*60*60)
	}

	programName, locationName, teamName := s.lookupNames(ctx, details.ProgramID, details.LocationID, details.TeamID)

	desc := fmt.Sprintf("%s event", action)
	if programName != "" {
		desc += fmt.Sprintf(" '%s'", programName)
	}
	if locationName != "" {
		desc += fmt.Sprintf(" at %s", locationName)
	}
	if teamName != "" {
		desc += fmt.Sprintf(" for team %s", teamName)
	}
	desc += fmt.Sprintf(" on %s", details.StartAt.In(loc).Format("Jan 2, 2006 at 3:04 PM"))

	return desc
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

func (s *Service) CreateEvents(ctx context.Context, details values.CreateRecurrenceValues) ([]values.ReadEventValues, *errLib.CommonError) {
	// If no PriceID provided but UnitAmount is, create a Stripe price
	if details.PriceID == "" && details.UnitAmount != nil {
		currency := details.Currency
		if currency == "" {
			currency = "cad"
		}

		// Get program name for the Stripe product name
		programName := "Recurring Event"
		if details.ProgramID != uuid.Nil {
			var name sql.NullString
			_ = s.db.QueryRowContext(ctx, "SELECT name FROM program.programs WHERE id = $1", details.ProgramID).Scan(&name)
			if name.Valid {
				programName = name.String + " Recurring Event"
			}
		}

		priceID, productID, err := s.productService.CreateProductWithOneTimePrice(
			programName,
			fmt.Sprintf("Recurring event starting %s", details.FirstOccurrence.Format("Jan 2, 2006")),
			*details.UnitAmount,
			currency,
		)
		if err != nil {
			return nil, err
		}

		details.PriceID = priceID
		log.Printf("[STRIPE] Created Stripe product %s with price %s for recurring event", productID, priceID)
	}

	events, err := generateEventsFromRecurrence(
		details.FirstOccurrence,
		details.LastOccurrence,
		details.StartTime,
		details.EndTime,
		details.CreatedBy,
		details.ProgramID,
		details.LocationID,
		details.CourtID,
		details.TeamID,
		details.RequiredMembershipPlanIDs,
		details.PriceID,
		details.DayOfWeek,
		details.CreditCost,
		details.RegistrationRequired,
	)
	if err != nil {
		return nil, err
	}

	txErr := s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {
		// Create events
		if err = txRepo.CreateEvents(ctx, events); err != nil {
			return err
		}

		return s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			details.CreatedBy,
			s.formatEventDescription(ctx, "Created recurring", values.EventDetails{
				StartAt:    details.FirstOccurrence,
				ProgramID:  details.ProgramID,
				LocationID: details.LocationID,
				TeamID:     details.TeamID,
			}),
		)
	})

	if txErr != nil {
		return nil, txErr
	}

	// Fetch the created events to return them with real IDs
	createdEvents, fetchErr := s.eventsRepository.GetEvents(ctx, values.GetEventsFilter{
		CreatedBy:  details.CreatedBy,
		ProgramID:  details.ProgramID,
		LocationID: details.LocationID,
		After:      details.FirstOccurrence,
		Before:     details.LastOccurrence.AddDate(0, 0, 1), // Add 1 day to include last occurrence
	})
	if fetchErr != nil {
		return nil, fetchErr
	}

	return createdEvents, nil
}

func (s *Service) CreateEvent(ctx context.Context, details values.CreateEventValues) (values.ReadEventValues, *errLib.CommonError) {
	var createdEventData values.ReadEventValues

	// If no PriceID provided but UnitAmount is, create a Stripe price
	if details.PriceID == "" && details.UnitAmount != nil {
		currency := details.Currency
		if currency == "" {
			currency = "cad"
		}

		// Get program name for the Stripe product name
		programName := "Event"
		if details.ProgramID != uuid.Nil {
			var name sql.NullString
			_ = s.db.QueryRowContext(ctx, "SELECT name FROM program.programs WHERE id = $1", details.ProgramID).Scan(&name)
			if name.Valid {
				programName = name.String + " Event"
			}
		}

		priceID, productID, err := s.productService.CreateProductWithOneTimePrice(
			programName,
			fmt.Sprintf("Event on %s", details.StartAt.Format("Jan 2, 2006")),
			*details.UnitAmount,
			currency,
		)
		if err != nil {
			return values.ReadEventValues{}, err
		}

		details.PriceID = priceID
		log.Printf("[STRIPE] Created Stripe product %s with price %s for event", productID, priceID)
	}

	err := s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {
		// Create the event
		if err := txRepo.CreateEvents(ctx, []values.CreateEventValues{details}); err != nil {
			return err
		}

		// Query for the created event to get its ID
		// We search by the unique combination of program, location, start time, and end time
		filter := values.GetEventsFilter{
			
			ProgramID:  details.ProgramID,
			LocationID: details.LocationID,
			After:      details.StartAt,
			Before:     details.EndAt,
			Limit:      1,
		}

		log.Printf("Querying for created event with filter: ProgramID=%s, LocationID=%s, After=%s, Before=%s",
			filter.ProgramID, filter.LocationID, filter.After.Format(time.RFC3339Nano), filter.Before.Format(time.RFC3339Nano))

		events, err := txRepo.GetEvents(ctx, filter)
		if err != nil {
			log.Printf("GetEvents query returned error: %v", err)
			return err
		}

		log.Printf("GetEvents returned %d events", len(events))

		if len(events) == 0 {
			return errLib.New("Failed to retrieve created event", http.StatusInternalServerError)
		}

		createdEvent := events[0]

		// Set membership plan associations
		if len(details.RequiredMembershipPlanIDs) > 0 {
			if err := txRepo.SetEventMembershipPlans(ctx, createdEvent.ID, details.RequiredMembershipPlanIDs); err != nil {
				return err
			}
		}

		isStaff, err := contextUtils.IsStaff(ctx)
		if err != nil {
			return err
		}

		if isStaff {
			if err := s.staffActivityLogsService.InsertStaffActivity(
				ctx,
				txRepo.GetTx(),
				details.CreatedBy,
				s.formatEventDescription(ctx, "Created", details.EventDetails),
			); err != nil {
				return err
			}
		}

		// Capture the created event data to return
		createdEventData = createdEvent
		return nil
	})

	return createdEventData, err
}

func (s *Service) UpdateEvent(ctx context.Context, details values.UpdateEventValues) *errLib.CommonError {
	// Fetch existing event BEFORE update (for change detection)
	var existingEvent values.ReadEventValues
	var fetchErr *errLib.CommonError
	if !details.SkipNotification {
		existingEvent, fetchErr = s.GetEvent(ctx, details.ID)
		if fetchErr != nil {
			// Log but don't fail - we can still update, just won't notify
			log.Printf("[EVENT-UPDATE] Warning: could not fetch existing event for change detection: %v", fetchErr.Message)
		}
	}

	// Perform the actual update
	updateErr := s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {
		if err := txRepo.UpdateEvent(ctx, details); err != nil {
			return err
		}

		// Update membership plan associations
		if err := txRepo.SetEventMembershipPlans(ctx, details.ID, details.RequiredMembershipPlanIDs); err != nil {
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
				s.formatEventDescription(ctx, "Updated", details.EventDetails),
			)
		}
		return nil
	})

	if updateErr != nil {
		return updateErr
	}

	// Send notification asynchronously if significant changes detected
	if !details.SkipNotification && fetchErr == nil {
		changes := s.detectChanges(ctx, existingEvent, details)
		if changes.HasSignificantChanges() {
			go s.notifyEnrolledUsersOfChanges(context.Background(), details.ID, existingEvent, changes)
		}
	}

	return nil
}

// UpdateRecurringEvents updates existing events within a recurrence schedule
//
// - Updates events within the new recurrence period
// - Deletes events outside the new period
// - Returns *errLib.CommonError if recurrence is being extended
func (s *Service) UpdateRecurringEvents(ctx context.Context, details values.UpdateRecurrenceValues) *errLib.CommonError {
	// If no PriceID provided but UnitAmount is, create a Stripe price
	if details.PriceID == "" && details.UnitAmount != nil {
		currency := details.Currency
		if currency == "" {
			currency = "cad"
		}

		// Get program name for the Stripe product name
		programName := "Recurring Event"
		if details.ProgramID != uuid.Nil {
			var name sql.NullString
			_ = s.db.QueryRowContext(ctx, "SELECT name FROM program.programs WHERE id = $1", details.ProgramID).Scan(&name)
			if name.Valid {
				programName = name.String + " Recurring Event"
			}
		}

		priceID, productID, err := s.productService.CreateProductWithOneTimePrice(
			programName,
			fmt.Sprintf("Recurring event starting %s", details.FirstOccurrence.Format("Jan 2, 2006")),
			*details.UnitAmount,
			currency,
		)
		if err != nil {
			return err
		}

		details.PriceID = priceID
		log.Printf("[STRIPE] Created Stripe product %s with price %s for updated recurring event", productID, priceID)
	}

	return s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {
		if err := txRepo.DeleteUnmodifiedEventsByRecurrenceID(ctx, details.ID); err != nil {
			return err
		}

		eventsToCreate, err := generateEventsFromRecurrence(
			details.FirstOccurrence,
			details.LastOccurrence,
			details.StartTime,
			details.EndTime,
			details.UpdatedBy,
			details.ProgramID,
			details.LocationID,
			details.CourtID,
			details.TeamID,
			details.RequiredMembershipPlanIDs,
			details.PriceID,
			details.DayOfWeek,
			details.CreditCost,
			details.RegistrationRequired,
		)
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
			s.formatEventDescription(ctx, "Updated recurring", values.EventDetails{
				StartAt:    details.FirstOccurrence,
				ProgramID:  details.ProgramID,
				LocationID: details.LocationID,
				TeamID:     details.TeamID,
			}),
		)
	})
}

func (s *Service) DeleteUnmodifiedEventsByRecurrenceID(ctx context.Context, staffId, id uuid.UUID) *errLib.CommonError {
	// Lookup recurrence details for audit log
	var programName, locationName string
	var firstOccurrence, lastOccurrence time.Time

	row := s.db.QueryRowContext(ctx, `
		SELECT p.name, l.name, r.first_occurrence, r.last_occurrence
		FROM events.recurrences r
		LEFT JOIN program.programs p ON r.program_id = p.id
		LEFT JOIN facilities.locations l ON r.location_id = l.id
		WHERE r.id = $1
	`, id)
	var pName, lName sql.NullString
	_ = row.Scan(&pName, &lName, &firstOccurrence, &lastOccurrence)
	if pName.Valid {
		programName = pName.String
	}
	if lName.Valid {
		locationName = lName.String
	}

	return s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {
		if err := txRepo.DeleteUnmodifiedEventsByRecurrenceID(ctx, id); err != nil {
			return err
		}

		// Build description
		loc, _ := time.LoadLocation("America/Edmonton")
		if loc == nil {
			loc = time.FixedZone("MST", -7*60*60)
		}

		desc := "Deleted recurring events"
		if programName != "" {
			desc += fmt.Sprintf(" for '%s'", programName)
		}
		if locationName != "" {
			desc += fmt.Sprintf(" at %s", locationName)
		}
		if !firstOccurrence.IsZero() && !lastOccurrence.IsZero() {
			desc += fmt.Sprintf(" (%s to %s)",
				firstOccurrence.In(loc).Format("Jan 2, 2006"),
				lastOccurrence.In(loc).Format("Jan 2, 2006"))
		}

		return s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			staffId,
			desc,
		)
	})
}

func (s *Service) DeleteEvents(ctx context.Context, ids []uuid.UUID) *errLib.CommonError {
	var (
		err     *errLib.CommonError
		isStaff bool
		staffID uuid.UUID
	)

	// Fetch event details before deletion for audit log
	events, err := s.eventsRepository.GetEvents(ctx, values.GetEventsFilter{Ids: ids})
	if err != nil {
		return err
	}

	// Build description with event details
	var descriptions []string
	for _, event := range events {
		desc := ""
		if event.Program.Name != "" {
			desc = event.Program.Name
		}
		if event.Location.Name != "" {
			if desc != "" {
				desc += " at " + event.Location.Name
			} else {
				desc = event.Location.Name
			}
		}
		loc, _ := time.LoadLocation("America/Edmonton")
		if loc == nil {
			loc = time.FixedZone("MST", -7*60*60)
		}
		desc += " on " + event.StartAt.In(loc).Format("Jan 2, 2006 at 3:04 PM")
		descriptions = append(descriptions, desc)
	}

	return s.executeInTx(ctx, func(txRepo *repo.EventsRepository) *errLib.CommonError {
		if err = txRepo.DeleteEvents(ctx, ids); err != nil {
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

			activityDesc := "Deleted event(s): "
			if len(descriptions) > 0 {
				activityDesc += fmt.Sprintf("%v", descriptions)
			} else {
				activityDesc += fmt.Sprintf("IDs %v", ids)
			}

			return s.staffActivityLogsService.InsertStaffActivity(
				ctx,
				txRepo.GetTx(),
				staffID,
				activityDesc,
			)
		} else {
			return nil
		}
	})
}

func (s *Service) GetEventsRecurrences(ctx context.Context, filter values.GetEventsFilter) ([]values.ReadRecurrenceValues, *errLib.CommonError) {
	return s.recurrencesRepository.GetEventsRecurrences(ctx, filter.ProgramType, filter.ProgramID, filter.LocationID,
		filter.ParticipantID, filter.TeamID, filter.CreatedBy, filter.UpdatedBy, filter.Before, filter.After)
}

func generateEventsFromRecurrence(
	firstOccurrence, lastOccurrence time.Time,
	startTimeStr, endTimeStr string,
	mutater, programID, locationID, courtID, teamID uuid.UUID,
	membershipPlanIDs []uuid.UUID,
	priceID string,
	day time.Weekday,
	creditCost *int32,
	registrationRequired bool,
) ([]values.CreateEventValues, *errLib.CommonError) {
	const timeLayout = "15:04:05Z07:00"

	if firstOccurrence.After(lastOccurrence) {
		return nil, errLib.New("Recurrence start date must be before the end date", http.StatusBadRequest)
	}

	startTime, err := time.Parse(timeLayout, startTimeStr)
	if err != nil {
		return nil, errLib.New("Invalid start time format - must be HH:MM:SS±HH:MM (e.g. 09:00:00+00:00)", http.StatusBadRequest)
	}

	endTime, err := time.Parse(timeLayout, endTimeStr)
	if err != nil {
		return nil, errLib.New("Invalid end time format - must be HH:MM:SS±HH:MM (e.g. 17:00:00+00:00)", http.StatusBadRequest)
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
			CreatedBy: mutater,
			EventDetails: values.EventDetails{
				StartAt:                   start,
				EndAt:                     end,
				ProgramID:                 programID,
				LocationID:                locationID,
				CourtID:                   courtID,
				TeamID:                    teamID,
				RequiredMembershipPlanIDs: membershipPlanIDs,
				PriceID:                   priceID,
				CreditCost:                creditCost,
				RegistrationRequired:      registrationRequired,
			},
		})
	}

	for date := firstOccurrence; !date.After(lastOccurrence); date = date.AddDate(0, 0, 1) {
		if date.Weekday() == day {
			addEvent(date)
			date = date.AddDate(0, 0, 6) // Skip to next week
		}
	}

	return events, nil
}

// detectChanges compares an existing event with new update values to identify changes
func (s *Service) detectChanges(ctx context.Context, existing values.ReadEventValues, new values.UpdateEventValues) values.EventChanges {
	changes := values.DetectEventChanges(existing, new)

	// Fetch new location name if changed
	if changes.LocationChanged {
		_, newLocationName, _ := s.lookupNames(ctx, uuid.Nil, new.LocationID, uuid.Nil)
		changes.NewLocationName = newLocationName
	}

	return changes
}

// buildChangeMessage creates a human-readable message describing what changed
func (s *Service) buildChangeMessage(existing values.ReadEventValues, changes values.EventChanges) string {
	loc, _ := time.LoadLocation("America/Edmonton")
	if loc == nil {
		loc = time.FixedZone("MST", -7*60*60)
	}

	var parts []string

	if changes.StartTimeChanged || changes.EndTimeChanged {
		oldTime := changes.OldStartAt.In(loc).Format("Monday, January 2 at 3:04 PM")
		newTime := changes.NewStartAt.In(loc).Format("Monday, January 2 at 3:04 PM")
		parts = append(parts, fmt.Sprintf("Time changed from %s to %s", oldTime, newTime))
	}

	if changes.LocationChanged {
		parts = append(parts, fmt.Sprintf("Location changed from %s to %s",
			changes.OldLocationName, changes.NewLocationName))
	}

	return strings.Join(parts, "\n")
}

// notifyEnrolledUsersOfChanges sends notifications to all enrolled users about event changes
func (s *Service) notifyEnrolledUsersOfChanges(ctx context.Context, eventID uuid.UUID, existing values.ReadEventValues, changes values.EventChanges) {
	message := s.buildChangeMessage(existing, changes)
	if message == "" {
		return
	}

	// Build subject line
	programName := "Event"
	if existing.Program.Name != "" {
		programName = existing.Program.Name
	}
	subject := fmt.Sprintf("Event Update: %s", programName)

	// Use the updater's ID as sender (or system if not available)
	senderID := existing.UpdatedBy.ID

	request := dto.SendNotificationRequestDto{
		Channel:             string(dto.ChannelBoth),
		Subject:             subject,
		Message:             message,
		IncludeEventDetails: true,
	}

	result, err := s.notificationService.SendNotification(ctx, eventID, senderID, request)
	if err != nil {
		log.Printf("[EVENT-UPDATE-NOTIFY] Failed to send notifications for event %s: %v", eventID, err.Message)
		return
	}

	log.Printf("[EVENT-UPDATE-NOTIFY] Sent update notifications for event %s: email=%d/%d, push=%d/%d",
		eventID, result.EmailSent, result.EmailFailed, result.PushSent, result.PushFailed)
}
