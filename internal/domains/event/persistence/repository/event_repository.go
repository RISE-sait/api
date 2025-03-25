package event

import (
	db "api/internal/domains/event/persistence/sqlc/generated"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
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

func (r *Repository) CreateEvent(c context.Context, eventDetails values.CreateEventValues) *errLib.CommonError {

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
		ProgramEndAt:   eventDetails.ProgramEndAt,
		EventStartTime: eventDetails.EventStartTime,
		EventEndTime:   eventDetails.EventEndTime,
		Day:            db.DayEnum(eventDetails.Day),
		LocationID: uuid.NullUUID{
			UUID:  eventDetails.LocationID,
			Valid: eventDetails.LocationID != uuid.Nil,
		},
		ProgramID: uuid.NullUUID{
			UUID:  eventDetails.ProgramID,
			Valid: eventDetails.ProgramID != uuid.Nil,
		},
	}

	if eventDetails.Capacity != nil {
		dbParams.Capacity = sql.NullInt32{
			Int32: *eventDetails.Capacity,
			Valid: true,
		}
	}

	if err := r.Queries.CreateEvent(c, dbParams); err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

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

		log.Printf("Failed to create eventDetails: %+v. Error: %v", eventDetails, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) GetEvent(ctx context.Context, id uuid.UUID) (values.ReadEventValues, *errLib.CommonError) {
	rows, err := r.Queries.GetEventStuffById(ctx, id)
	if err != nil {
		log.Println("Failed to get event from db: ", err.Error())
		return values.ReadEventValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	var event values.ReadEventValues
	event.Staffs = []values.Staff{}
	event.Customers = []values.Customer{}

	for _, row := range rows {

		event = values.ReadEventValues{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			Details: values.Details{
				Day:            string(row.Day),
				ProgramStartAt: row.ProgramStartAt,
				ProgramEndAt:   row.ProgramEndAt,
				EventStartTime: row.EventStartTime,
				EventEndTime:   row.EventEndTime,
				LocationID:     row.LocationID.UUID,
				ProgramID:      row.ProgramID.UUID,
			},
			ProgramName:     row.ProgramName.String,
			ProgramType:     string(row.ProgramType.ProgramProgramType),
			LocationName:    row.LocationName.String,
			LocationAddress: row.LocationAddress.String,
			Staffs:          []values.Staff{},
			Customers:       []values.Customer{},
		}
		if row.Capacity.Valid {
			event.Details.Capacity = &row.Capacity.Int32
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
			// Check for duplicates
			exists := false
			for _, c := range event.Customers {
				if c.ID == customer.ID {
					exists = true
					break
				}
			}
			if !exists {
				event.Customers = append(event.Customers, customer)
			}
		}
	}

	// If no rows were returned at all
	if event.ID == uuid.Nil {
		return values.ReadEventValues{}, errLib.New("Event not found", http.StatusNotFound)
	}

	return event, nil
}

func (r *Repository) GetEvents(ctx context.Context, programTypeStr string, programID, locationID uuid.UUID, before, after time.Time, userID uuid.UUID) ([]values.ReadEventValues, *errLib.CommonError) {

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
		Type:       programType,
	})

	if err != nil {
		log.Println("Failed to get events from db: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	// Group the rows by event ID
	eventMap := make(map[uuid.UUID]*values.ReadEventValues)

	for _, row := range dbRows {
		// Get or create the event
		if _, exists := eventMap[row.ID]; !exists {
			eventMap[row.ID] = &values.ReadEventValues{
				ID:        row.ID,
				CreatedAt: row.CreatedAt,
				UpdatedAt: row.UpdatedAt,
				Details: values.Details{
					Day:            string(row.Day),
					ProgramStartAt: row.ProgramStartAt,
					ProgramEndAt:   row.ProgramEndAt,
					EventStartTime: row.EventStartTime,
					EventEndTime:   row.EventEndTime,
					ProgramID:      row.ProgramID.UUID,
					LocationID:     row.LocationID.UUID,
				},
				ProgramName:     row.ProgramName.String,
				ProgramType:     string(row.ProgramType.ProgramProgramType),
				LocationName:    row.LocationName.String,
				LocationAddress: row.LocationAddress.String,
				Staffs:          []values.Staff{},
				Customers:       []values.Customer{},
			}
			if row.Capacity.Valid {
				eventMap[row.ID].Details.Capacity = &row.Capacity.Int32
			}
		}
		event := eventMap[row.ID]

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
			// Check if this staff member already exists for this event
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
			// Check if this customer already exists for this event
			exists := false
			for _, c := range event.Customers {
				if c.ID == customer.ID {
					exists = true
					break
				}
			}
			if !exists {
				event.Customers = append(event.Customers, customer)
			}
		}
	}

	// Convert the map to a slice
	events := make([]values.ReadEventValues, 0, len(eventMap))
	for _, event := range eventMap {
		events = append(events, *event)
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
func (r *Repository) UpdateEvent(c context.Context, event values.UpdateEventValues) *errLib.CommonError {

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
		ProgramEndAt:   event.ProgramEndAt,
		LocationID:     uuid.NullUUID{UUID: event.LocationID, Valid: event.LocationID != uuid.Nil},
		ProgramID:      uuid.NullUUID{UUID: event.ProgramID, Valid: event.ProgramID != uuid.Nil},
		EventStartTime: event.EventStartTime,
		EventEndTime:   event.EventEndTime,
		Day:            db.DayEnum(event.Day),
		ID:             event.ID,
	}

	err := r.Queries.UpdateEvent(c, dbEventParams)

	if err != nil {
		log.Printf("Failed to update event: %+v. Error: %v", event, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
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
