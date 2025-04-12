package event

import (
	"api/internal/di"
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

func NewEventsRepository(container *di.Container) *EventsRepository {
	return &EventsRepository{
		Queries: container.Queries.EventDb,
	}
}

func (r *EventsRepository) CreateEvents(ctx context.Context, eventDetails values.CreateEventsValues) *errLib.CommonError {

	var (
		locationIDs, programIDs, teamIDs, createdByIds []uuid.UUID
		startAtArray, endAtArray                       []time.Time
		capacities                                     []int32
		isCancelledArray                               []bool
	)

	for _, event := range eventDetails.Events {
		locationIDs = append(locationIDs, event.LocationID)
		programIDs = append(programIDs, event.ProgramID)
		teamIDs = append(teamIDs, event.TeamID)
		startAtArray = append(startAtArray, event.StartAt)
		endAtArray = append(endAtArray, event.EndAt)
		createdByIds = append(createdByIds, eventDetails.CreatedBy)
		capacities = append(capacities, event.Capacity)
		isCancelledArray = append(isCancelledArray, false)
	}

	dbParams := db.CreateEventsParams{
		LocationIds:         locationIDs,
		ProgramIds:          programIDs,
		TeamIds:             teamIDs,
		StartAtArray:        startAtArray,
		EndAtArray:          endAtArray,
		CreatedByIds:        createdByIds,
		Capacities:          capacities,
		IsCancelledArray:    isCancelledArray,
		CancellationReasons: nil,
	}

	dbErr := r.Queries.CreateEvents(ctx, dbParams)

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

func (r *EventsRepository) GetEvent(ctx context.Context, id uuid.UUID) (values.ReadEventValues, *errLib.CommonError) {
	dbEvent, err := r.Queries.GetEventById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadEventValues{}, errLib.New("Event not found", http.StatusNotFound)
		}
		log.Println("Failed to get event from db: ", err.Error())
		return values.ReadEventValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	eventValue := values.ReadEventValues{
		ID:        dbEvent.ID,
		CreatedAt: dbEvent.CreatedAt,
		UpdatedAt: dbEvent.UpdatedAt,
		StartAt:   dbEvent.StartAt,
		EndAt:     dbEvent.EndAt,
		Capacity:  dbEvent.Capacity.Int32,
		Location: struct {
			ID      uuid.UUID
			Name    string
			Address string
		}{
			ID:      dbEvent.LocationID,
			Name:    dbEvent.LocationName,
			Address: dbEvent.LocationAddress,
		},
		CreatedBy: values.ReadPersonValues{
			ID:        dbEvent.CreatedBy,
			FirstName: dbEvent.CreatorFirstName,
			LastName:  dbEvent.CreatorLastName,
		},
		UpdatedBy: values.ReadPersonValues{
			ID:        dbEvent.UpdatedBy,
			FirstName: dbEvent.UpdaterFirstName,
			LastName:  dbEvent.UpdaterLastName,
		},
	}

	eventCustomers := make([]values.Customer, 0)
	eventStaffs := make([]values.Staff, 0)

	dbStaffs, err := r.Queries.GetEventStaffs(ctx, id)

	if err != nil {
		log.Println("Failed to get event staffs from db: ", err.Error())
		return values.ReadEventValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	for _, dbStaff := range dbStaffs {

		staff := values.Staff{
			ReadPersonValues: values.ReadPersonValues{

				ID:        dbStaff.StaffID,
				FirstName: dbStaff.StaffFirstName,
				LastName:  dbStaff.StaffLastName,
			},
			Phone:    dbStaff.StaffPhone.String,
			Gender:   stringToPtr(dbStaff.StaffGender),
			RoleName: dbStaff.StaffRoleName,
		}

		eventStaffs = append(eventStaffs, staff)
	}

	dbCustomers, err := r.Queries.GetEventCustomers(ctx, id)

	if err != nil {
		log.Println("Failed to get event customers from db: ", err.Error())
		return values.ReadEventValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	for _, dbCustomer := range dbCustomers {

		customer := values.Customer{
			ReadPersonValues: values.ReadPersonValues{

				ID:        dbCustomer.CustomerID,
				FirstName: dbCustomer.CustomerFirstName,
				LastName:  dbCustomer.CustomerLastName,
			},
			Phone:  stringToPtr(dbCustomer.CustomerPhone),
			Gender: stringToPtr(dbCustomer.CustomerGender),
		}
		eventCustomers = append(eventCustomers, customer)
	}

	return eventValue, nil
}

func (r *EventsRepository) GetEvents(ctx context.Context, programTypeStr string, programID, locationID, userID, teamID, createdBy, updatedBy uuid.UUID, before, after time.Time) ([]values.ReadEventValues, *errLib.CommonError) {

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
	}

	programType := db.ProgramProgramType(programTypeStr)

	if programTypeStr != "" {
		if programType.Valid() {
			param.Type = db.NullProgramProgramType{
				ProgramProgramType: programType,
				Valid:              true,
			}
		} else {
			validTypes := db.AllProgramProgramTypeValues()

			return nil, errLib.New(fmt.Sprintf("Invalid program type. Valid types are: %v", validTypes), http.StatusBadRequest)
		}

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
