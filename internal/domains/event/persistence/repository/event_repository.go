package event

import (
	db "api/internal/domains/event/persistence/sqlc/generated"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository struct {
	Queries *db.Queries
}

func NewEventsRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) CreateEvent(ctx context.Context, eventDetails values.CreateEventValues) *errLib.CommonError {

	userID, err := contextUtils.GetUserID(ctx)

	if err != nil {
		return err
	}

	if !db.DayEnum(eventDetails.Day).Valid() {

		validDaysDbValues := db.AllDayEnumValues()

		validDays := make([]string, len(validDaysDbValues))

		for i, value := range validDaysDbValues {
			validDays[i] = string(value)
		}

		return errLib.New("Invalid day provided. Valid days are: "+strings.Join(validDays, ", "), http.StatusBadRequest)
	}

	dbParams := db.CreateEventParams{
		ProgramStartAt: eventDetails.ProgramStartAt,
		EventStartTime: eventDetails.EventStartTime,
		EventEndTime:   eventDetails.EventEndTime,
		Day:            db.DayEnum(eventDetails.Day),
		LocationID:     eventDetails.LocationID,
		ProgramID: uuid.NullUUID{
			UUID:  eventDetails.ProgramID,
			Valid: eventDetails.ProgramID != uuid.Nil,
		},
		CreatedBy: userID,
	}

	if eventDetails.ProgramEndAt != nil {
		dbParams.ProgramEndAt = sql.NullTime{
			Time:  *eventDetails.ProgramEndAt,
			Valid: true,
		}
	}

	if eventDetails.Capacity != nil {
		dbParams.Capacity = sql.NullInt32{
			Int32: *eventDetails.Capacity,
			Valid: true,
		}
	}

	if dbErr := r.Queries.CreateEvent(ctx, dbParams); dbErr != nil {

		var pqErr *pq.Error
		if errors.As(dbErr, &pqErr) {

			constraintErrors := map[string]string{
				"fk_program":            "The referenced program doesn't exist",
				"fk_location":           "The referenced location doesn't exist",
				"unique_event_time":     "An event is already scheduled at this time",
				"check_event_times":     "End time/date must be after Begin time/date",
				"event_end_after_start": "End time/date must be after Begin time/date",
			}

			if msg, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(msg, http.StatusBadRequest)
			}
		}

		log.Printf("Failed to create eventDetails: %+v. Error: %v", eventDetails, dbErr.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) GetEventCreatedBy(ctx context.Context, eventID uuid.UUID) (uuid.UUID, *errLib.CommonError) {
	userID, err := r.Queries.GetEventCreatedBy(ctx, eventID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, errLib.New("Event not found", http.StatusNotFound)
		}
		log.Println("Failed to get event from db: ", err.Error())
		return uuid.Nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if !userID.Valid {
		return uuid.Nil, errLib.New("Unknown reason. Couldn't get the person who created this event.", http.StatusInternalServerError)
	}

	return userID.UUID, nil
}

func (r *Repository) GetEvent(ctx context.Context, id uuid.UUID) (values.ReadEventValues, *errLib.CommonError) {
	rows, err := r.Queries.GetEventById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadEventValues{}, errLib.New("Event not found", http.StatusNotFound)
		}
		log.Println("Failed to get event from db: ", err.Error())
		return values.ReadEventValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	var event values.ReadEventValues
	event.Staffs = []values.Staff{}
	event.Customers = []values.Customer{}

	for i, row := range rows {

		if i == 0 {

			event = values.ReadEventValues{
				ID:        row.ID,
				CreatedAt: row.CreatedAt,
				UpdatedAt: row.UpdatedAt,
				Details: values.Details{
					Day:            string(row.Day),
					ProgramStartAt: row.ProgramStartAt,
					EventStartTime: row.EventStartTime,
					EventEndTime:   row.EventEndTime,
					LocationID:     row.LocationID,
					ProgramID:      row.ProgramID.UUID,
				},
				CreatedBy:       row.CreatedBy.UUID,
				UpdatedBy:       row.UpdatedBy.UUID,
				ProgramName:     row.ProgramName.String,
				ProgramType:     string(row.ProgramType.ProgramProgramType),
				LocationName:    row.LocationName.String,
				LocationAddress: row.LocationAddress.String,
			}

			if row.ProgramEndAt.Valid {
				event.Details.ProgramEndAt = &row.ProgramEndAt.Time
			}

			if row.Capacity.Valid {
				event.Details.Capacity = &row.Capacity.Int32
			}
		}

		// Add staff member if exists in this row
		if row.StaffID.Valid {
			staff := values.Staff{
				ID:        row.StaffID.UUID,
				Email:     row.StaffEmail.String,
				FirstName: row.StaffFirstName.String,
				LastName:  row.StaffLastName.String,
				Phone:     row.StaffPhone.String,
				Gender:    stringToPtr(row.StaffGender),
				RoleName:  row.StaffRoleName.String,
			}
			// Check for duplicates
			exists := false
			for _, s := range event.Staffs {
				if s.ID == staff.ID {
					exists = true
					break
				}
			}
			if !exists {
				event.Staffs = append(event.Staffs, staff)
			}
		}

		// Add customer if exists in this row
		if row.CustomerID.Valid {
			customer := values.Customer{
				ID:                    row.CustomerID.UUID,
				FirstName:             row.CustomerFirstName.String,
				LastName:              row.CustomerLastName.String,
				Email:                 stringToPtr(row.CustomerEmail),
				Phone:                 stringToPtr(row.CustomerPhone),
				Gender:                stringToPtr(row.CustomerGender),
				IsEnrollmentCancelled: row.CustomerIsCancelled.Bool,
			}
			event.Customers = append(event.Customers, customer)
		}
	}

	// If no rows were returned at all
	if event.ID == uuid.Nil {
		return values.ReadEventValues{}, errLib.New("Event not found", http.StatusNotFound)
	}

	return event, nil
}

func (r *Repository) GetEvents(ctx context.Context, programTypeStr string, programID, locationID, userID, teamID, createdBy, updatedBy uuid.UUID, before, after time.Time) ([]values.ReadEventValues, *errLib.CommonError) {

	var programType db.NullProgramProgramType

	if programTypeStr == "" {
		programType.Valid = false
	} else {
		if !db.ProgramProgramType(programTypeStr).Valid() {

			validTypes := db.AllProgramProgramTypeValues()

			return nil, errLib.New(fmt.Sprintf("Invalid program type. Valid types are: %v", validTypes), http.StatusBadRequest)
		}
	}

	// Execute the query using SQLC generated function
	dbRows, err := r.Queries.GetEvents(ctx, db.GetEventsParams{
		ProgramID:  uuid.NullUUID{UUID: programID, Valid: programID != uuid.Nil},
		LocationID: uuid.NullUUID{UUID: locationID, Valid: locationID != uuid.Nil},
		Before:     sql.NullTime{Time: before, Valid: !before.IsZero()},
		After:      sql.NullTime{Time: after, Valid: !after.IsZero()},
		UserID:     uuid.NullUUID{UUID: userID, Valid: userID != uuid.Nil},
		TeamID:     uuid.NullUUID{UUID: teamID, Valid: teamID != uuid.Nil},
		CreatedBy:  uuid.NullUUID{UUID: createdBy, Valid: createdBy != uuid.Nil},
		UpdatedBy:  uuid.NullUUID{UUID: updatedBy, Valid: updatedBy != uuid.Nil},
		Type:       programType,
	})

	if err != nil {
		log.Println("Failed to get events from db: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	// Group the rows by event ID
	var events []values.ReadEventValues

	for _, row := range dbRows {

		event := values.ReadEventValues{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			Details: values.Details{
				Day:            string(row.Day),
				ProgramStartAt: row.ProgramStartAt,
				EventStartTime: row.EventStartTime,
				EventEndTime:   row.EventEndTime,
				ProgramID:      row.ProgramID.UUID,
				LocationID:     row.LocationID,
			},
			ProgramName:     row.ProgramName.String,
			ProgramType:     string(row.ProgramType.ProgramProgramType),
			LocationName:    row.LocationName,
			LocationAddress: row.LocationAddress,
			CreatedBy:       row.CreatedBy.UUID,
			UpdatedBy:       row.UpdatedBy.UUID,
		}

		if row.ProgramEndAt.Valid {
			event.Details.ProgramEndAt = &row.ProgramEndAt.Time
		}

		if row.TeamID.Valid && row.TeamName.Valid {
			event.TeamID = row.TeamID.UUID
			event.TeamName = row.TeamName.String
		}

		if row.Capacity.Valid {
			event.Details.Capacity = &row.Capacity.Int32
		}

		if row.TeamID.Valid && row.TeamName.Valid {
			event.TeamID = row.TeamID.UUID
			event.TeamName = row.TeamName.String
		}

		events = append(events, event)
	}
	return events, nil
}

// Helper function to convert sql.NullString to *string
func stringToPtr(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func (r *Repository) UpdateEvent(ctx context.Context, event values.UpdateEventValues) *errLib.CommonError {

	userID, err := contextUtils.GetUserID(ctx)

	if err != nil {
		return err
	}

	if !db.DayEnum(event.Day).Valid() {

		validDaysDbValues := db.AllDayEnumValues()

		validDays := make([]string, len(validDaysDbValues))

		for i, value := range validDaysDbValues {
			validDays[i] = string(value)
		}

		return errLib.New("Invalid day provided. Valid days are: "+strings.Join(validDays, ", "), http.StatusBadRequest)
	}

	dbEventParams := db.UpdateEventParams{
		ProgramStartAt: event.ProgramStartAt,
		LocationID:     event.LocationID,
		ProgramID:      uuid.NullUUID{UUID: event.ProgramID, Valid: event.ProgramID != uuid.Nil},
		EventStartTime: event.EventStartTime,
		EventEndTime:   event.EventEndTime,
		Day:            db.DayEnum(event.Day),
		ID:             event.ID,
		UpdatedBy:      userID,
	}

	if event.ProgramEndAt != nil {
		dbEventParams.ProgramEndAt = sql.NullTime{
			Time:  *event.ProgramEndAt,
			Valid: true,
		}
	}

	if dbErr := r.Queries.UpdateEvent(ctx, dbEventParams); dbErr != nil {

		var pqErr *pq.Error
		if errors.As(dbErr, &pqErr) {

			constraintErrors := map[string]string{
				"fk_program":            "The referenced program doesn't exist",
				"fk_location":           "The referenced location doesn't exist",
				"unique_event_time":     "An event is already scheduled at this time",
				"check_event_times":     "End time/date must be after Begin time/date",
				"event_end_after_start": "End time/date must be after Begin time/date",
			}

			if msg, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(msg, http.StatusBadRequest)
			}
		}

		log.Printf("Failed to update event: %+v. Error: %v", event, dbErr.Error())
		return errLib.New("Internal server error when updating event", http.StatusInternalServerError)
	}

	return nil

}

func (r *Repository) DeleteEvent(c context.Context, id uuid.UUID) *errLib.CommonError {
	err := r.Queries.DeleteEvent(c, id)

	if err != nil {
		log.Printf("Failed to delete event with HubSpotId: %s. Error: %s", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
