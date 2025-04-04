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
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type EventsRepository struct {
	Queries *db.Queries
}

func NewEventsRepository(dbQueries *db.Queries) *EventsRepository {
	return &EventsRepository{
		Queries: dbQueries,
	}
}

func (r *EventsRepository) CreateEvent(ctx context.Context, eventDetails values.CreateEventValues) *errLib.CommonError {

	dbParams := db.CreateEventParams{
		StartAt:   eventDetails.StartAt,
		EndAt:     eventDetails.EndAt,
		CreatedBy: eventDetails.CreatedBy,
		Capacity: sql.NullInt32{
			Int32: eventDetails.Capacity,
			Valid: true,
		},
		LocationID: eventDetails.LocationID,
		TeamID: uuid.NullUUID{
			UUID:  eventDetails.TeamID,
			Valid: eventDetails.TeamID != uuid.Nil,
		},
		ProgramID: uuid.NullUUID{
			UUID:  eventDetails.ProgramID,
			Valid: eventDetails.ProgramID != uuid.Nil,
		},
	}

	_, dbErr := r.Queries.CreateEvent(ctx, dbParams)

	if dbErr != nil {

		var pqErr *pq.Error
		if errors.As(dbErr, &pqErr) {

			constraintErrors := map[string]struct {
				Message string
				Status  int
			}{
				"fk_created_by": {
					Message: "The referenced user doesn't exist",
					Status:  http.StatusBadRequest,
				},
				"fk_updated_by": {
					Message: "The referenced user doesn't exist",
					Status:  http.StatusBadRequest,
				},
				"fk_program": {
					Message: "The referenced program doesn't exist",
					Status:  http.StatusBadRequest,
				},
				"fk_team": {
					Message: "The referenced team doesn't exist",
					Status:  http.StatusBadRequest,
				},
				"events_location_id_fkey": {
					Message: "The referenced location doesn't exist",
					Status:  http.StatusBadRequest,
				},
				"check_start_end": {
					Message: "Event end time must be after start time",
					Status:  http.StatusBadRequest,
				},
				"no_overlapping_events": {
					Message: "An event is already scheduled at this time and location",
					Status:  http.StatusConflict,
				},
			}

			if errInfo, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(errInfo.Message, errInfo.Status)
			}
		}

		log.Printf("Failed to create eventDetails: %+v. Error: %v", eventDetails, dbErr.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *EventsRepository) GetEventCreatedBy(ctx context.Context, eventID uuid.UUID) (uuid.UUID, *errLib.CommonError) {
	userID, err := r.Queries.GetEventCreatedBy(ctx, eventID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, errLib.New("Event not found", http.StatusNotFound)
		}
		log.Println("Failed to get event from db: ", err.Error())
		return uuid.Nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return userID, nil
}

func (r *EventsRepository) GetEvent(ctx context.Context, id uuid.UUID) (values.ReadEventValues, *errLib.CommonError) {
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
				StartAt:   row.StartAt,
				EndAt:     row.EndAt,
				Capacity:  row.Capacity.Int32,
				Location: struct {
					ID      uuid.UUID
					Name    string
					Address string
				}{
					ID:      row.LocationID,
					Name:    row.LocationName,
					Address: row.LocationAddress,
				},
				CreatedBy: values.ReadPersonValues{
					ID:        row.CreatedBy,
					FirstName: row.CreatorFirstName,
					LastName:  row.CreatorLastName,
				},
				UpdatedBy: values.ReadPersonValues{
					ID:        row.UpdatedBy,
					FirstName: row.UpdaterFirstName,
					LastName:  row.UpdaterLastName,
				},
			}

			if row.TeamID.Valid && row.TeamName.Valid {
				event.Team = &struct {
					ID   uuid.UUID
					Name string
				}{
					ID:   row.TeamID.UUID,
					Name: row.TeamName.String,
				}
			}

			if row.ProgramID.Valid && row.ProgramName.Valid {
				event.Program = &struct {
					ID          uuid.UUID
					Name        string
					Description string
					Type        string
				}{
					ID:          row.ProgramID.UUID,
					Name:        row.ProgramName.String,
					Description: row.ProgramDescription.String,
					Type:        string(row.ProgramType.ProgramProgramType),
				}
			}
		}

		// Add staff member if exists in this row
		if row.StaffID.Valid {
			staff := values.Staff{
				ReadPersonValues: values.ReadPersonValues{

					ID:        row.StaffID.UUID,
					FirstName: row.StaffFirstName.String,
					LastName:  row.StaffLastName.String,
				},
				Phone:    row.StaffPhone.String,
				Gender:   stringToPtr(row.StaffGender),
				RoleName: row.StaffRoleName.String,
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
				ReadPersonValues: values.ReadPersonValues{
					ID:        row.CustomerID.UUID,
					FirstName: row.CustomerFirstName.String,
					LastName:  row.CustomerLastName.String,
				},
				Phone:                  stringToPtr(row.CustomerPhone),
				Gender:                 stringToPtr(row.CustomerGender),
				HasCancelledEnrollment: row.CustomerEnrollmentIsCancelled.Bool,
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

func (r *EventsRepository) GetEvents(ctx context.Context, programTypeStr string, programID, locationID, userID, teamID, createdBy, updatedBy uuid.UUID, before, after time.Time) ([]values.ReadEventValues, *errLib.CommonError) {

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
	param := db.GetEventsParams{
		ProgramID:  uuid.NullUUID{UUID: programID, Valid: programID != uuid.Nil},
		LocationID: uuid.NullUUID{UUID: locationID, Valid: locationID != uuid.Nil},
		Before:     sql.NullTime{Time: before, Valid: !before.IsZero()},
		After:      sql.NullTime{Time: after, Valid: !after.IsZero()},
		UserID:     uuid.NullUUID{UUID: userID, Valid: userID != uuid.Nil},
		TeamID:     uuid.NullUUID{UUID: teamID, Valid: teamID != uuid.Nil},
		CreatedBy:  uuid.NullUUID{UUID: createdBy, Valid: createdBy != uuid.Nil},
		UpdatedBy:  uuid.NullUUID{UUID: updatedBy, Valid: updatedBy != uuid.Nil},
		Type:       programType,
	}

	dbRows, err := r.Queries.GetEvents(ctx, param)

	if err != nil {
		log.Println("Failed to get events from db: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	// Group the rows by event ID
	var events []values.ReadEventValues

	for _, row := range dbRows {

		event := values.ReadEventValues{
			ID:        row.ID,
			Capacity:  row.Capacity.Int32,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			StartAt:   row.StartAt,
			EndAt:     row.EndAt,
			Location: struct {
				ID      uuid.UUID
				Name    string
				Address string
			}{
				ID:      row.LocationID,
				Name:    row.LocationName,
				Address: row.LocationAddress,
			},
			CreatedBy: values.ReadPersonValues{
				ID:        row.CreatedBy,
				FirstName: row.CreatorFirstName,
				LastName:  row.CreatorLastName,
			},
			UpdatedBy: values.ReadPersonValues{
				ID:        row.UpdatedBy,
				FirstName: row.UpdaterFirstName,
				LastName:  row.UpdaterLastName,
			},
		}

		if row.ProgramID.Valid && row.ProgramName.Valid && row.ProgramType.Valid && row.ProgramDescription.Valid {
			event.Program = &struct {
				ID          uuid.UUID
				Name        string
				Description string
				Type        string
			}{
				ID:          row.ProgramID.UUID,
				Name:        row.ProgramName.String,
				Description: row.ProgramDescription.String,
				Type:        string(row.ProgramType.ProgramProgramType),
			}
		}

		if row.TeamID.Valid && row.TeamName.Valid {
			event.Team = &struct {
				ID   uuid.UUID
				Name string
			}{
				ID:   row.TeamID.UUID,
				Name: row.TeamName.String,
			}
		}

		if row.Capacity.Valid {
			event.Capacity = row.Capacity.Int32
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

func (r *EventsRepository) UpdateEvent(ctx context.Context, event values.UpdateEventValues) *errLib.CommonError {

	userID, err := contextUtils.GetUserID(ctx)

	if err != nil {
		return err
	}

	dbEventParams := db.UpdateEventParams{
		StartAt:    event.StartAt,
		EndAt:      event.EndAt,
		UpdatedBy:  userID,
		ID:         event.ID,
		LocationID: event.LocationID,
		TeamID: uuid.NullUUID{
			UUID:  event.TeamID,
			Valid: event.TeamID != uuid.Nil,
		},
		ProgramID: uuid.NullUUID{
			UUID:  event.ProgramID,
			Valid: event.ProgramID != uuid.Nil,
		},
		Capacity: sql.NullInt32{
			Int32: event.Capacity,
			Valid: event.Capacity != 0,
		},
	}

	_, dbErr := r.Queries.UpdateEvent(ctx, dbEventParams)

	if dbErr != nil {

		var pqErr *pq.Error
		if errors.As(dbErr, &pqErr) {

			constraintErrors := map[string]struct {
				Message string
				Status  int
			}{
				"fk_created_by": {
					Message: "The referenced user doesn't exist",
					Status:  http.StatusBadRequest,
				},
				"fk_updated_by": {
					Message: "The referenced user doesn't exist",
					Status:  http.StatusBadRequest,
				},
				"fk_program": {
					Message: "The referenced program doesn't exist",
					Status:  http.StatusBadRequest,
				},
				"fk_team": {
					Message: "The referenced team doesn't exist",
					Status:  http.StatusBadRequest,
				},
				"events_location_id_fkey": {
					Message: "The referenced location doesn't exist",
					Status:  http.StatusBadRequest,
				},
				"check_start_end": {
					Message: "Event end time must be after start time",
					Status:  http.StatusBadRequest,
				},
				"no_overlapping_events": {
					Message: "An event is already scheduled at this time and location",
					Status:  http.StatusConflict,
				},
			}

			if errInfo, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(errInfo.Message, errInfo.Status)
			}
		}

		log.Printf("Failed to update event: %+v. Error: %v", event, dbErr.Error())
		return errLib.New("Internal server error when updating event", http.StatusInternalServerError)
	}

	return nil

}

func (r *EventsRepository) DeleteEvent(c context.Context, id uuid.UUID) *errLib.CommonError {
	err := r.Queries.DeleteEvent(c, id)

	if err != nil {
		log.Printf("Failed to delete event with HubSpotId: %s. Error: %s", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
