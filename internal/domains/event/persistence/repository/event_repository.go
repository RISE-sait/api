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
	Tx      *sql.Tx
}

func NewEventsRepository(container *di.Container) *EventsRepository {
	return &EventsRepository{
		Queries: container.Queries.EventDb,
	}
}

func (r *EventsRepository) GetTx() *sql.Tx {
	return r.Tx
}

func (r *EventsRepository) WithTx(tx *sql.Tx) *EventsRepository {
	return &EventsRepository{
		Queries: r.Queries.WithTx(tx),
		Tx:      tx,
	}
}

var constraintErrors = map[string]struct {
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

func (r *EventsRepository) CreateEvents(ctx context.Context, eventDetails []values.CreateEventValues) *errLib.CommonError {

	var (
		locationIDs, programIDs, teamIDs, createdByIds, recurrenceIds []uuid.UUID
		startAtArray, endAtArray                                      []time.Time
		capacities                                                    []int32
		isCancelledArray, isDateTimeModifiedArray                     []bool
	)

	recurrenceId := uuid.Nil

	if len(eventDetails) > 1 {
		recurrenceId = uuid.New()
	}

	for _, event := range eventDetails {
		locationIDs = append(locationIDs, event.LocationID)
		programIDs = append(programIDs, event.ProgramID)
		recurrenceIds = append(recurrenceIds, recurrenceId)
		teamIDs = append(teamIDs, event.TeamID)
		startAtArray = append(startAtArray, event.StartAt)
		endAtArray = append(endAtArray, event.EndAt)
		createdByIds = append(createdByIds, event.CreatedBy)
		capacities = append(capacities, event.Capacity)
		isCancelledArray = append(isCancelledArray, false)
		isDateTimeModifiedArray = append(isDateTimeModifiedArray, false)
	}

	dbParams := db.CreateEventsParams{
		LocationIds:             locationIDs,
		ProgramIds:              programIDs,
		TeamIds:                 teamIDs,
		StartAtArray:            startAtArray,
		EndAtArray:              endAtArray,
		CreatedByIds:            createdByIds,
		Capacities:              capacities,
		RecurrenceIds:           recurrenceIds,
		IsCancelledArray:        isCancelledArray,
		IsDateTimeModifiedArray: isDateTimeModifiedArray,
		CancellationReasons:     nil,
	}

	impactedRows, dbErr := r.Queries.CreateEvents(ctx, dbParams)

	if dbErr != nil {

		var pqErr *pq.Error
		if errors.As(dbErr, &pqErr) {

			if errInfo, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(errInfo.Message, errInfo.Status)
			}
		}

		log.Printf("Failed to create eventDetails: %+v. Error: %v", eventDetails, dbErr.Error())
		return errLib.New("Internal server error when creating events", http.StatusInternalServerError)
	}

	if impactedRows < int64(len(eventDetails)) {

		return errLib.New("Not all events were created successfully, likely due to overlapping events", http.StatusBadRequest)
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

func (r *EventsRepository) GetEvents(ctx context.Context, filter values.GetEventsFilter) ([]values.ReadEventValues, *errLib.CommonError) {

	param := db.GetEventsParams{
		ProgramID:     uuid.NullUUID{UUID: filter.ProgramID, Valid: filter.ProgramID != uuid.Nil},
		LocationID:    uuid.NullUUID{UUID: filter.LocationID, Valid: filter.LocationID != uuid.Nil},
		Before:        sql.NullTime{Time: filter.Before, Valid: !filter.Before.IsZero()},
		After:         sql.NullTime{Time: filter.After, Valid: !filter.After.IsZero()},
		ParticipantID: uuid.NullUUID{UUID: filter.ParticipantID, Valid: filter.ParticipantID != uuid.Nil},
		TeamID:        uuid.NullUUID{UUID: filter.TeamID, Valid: filter.TeamID != uuid.Nil},
		CreatedBy:     uuid.NullUUID{UUID: filter.CreatedBy, Valid: filter.CreatedBy != uuid.Nil},
		UpdatedBy:     uuid.NullUUID{UUID: filter.UpdatedBy, Valid: filter.UpdatedBy != uuid.Nil},
	}

	if filter.ProgramType != "" {
		programType := db.ProgramProgramType(filter.ProgramType)
		if !programType.Valid() {
			return nil, errLib.New(fmt.Sprintf("Invalid program type. Valid types are: %v", db.AllProgramProgramTypeValues()), http.StatusBadRequest)
		}
		param.Type = db.NullProgramProgramType{ProgramProgramType: programType, Valid: true}
	}

	dbRows, err := r.Queries.GetEvents(ctx, param)

	if err != nil {
		log.Println("Failed to get events from db: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	events := make([]values.ReadEventValues, 0, len(dbRows))

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
			Program: struct {
				ID          uuid.UUID
				Name        string
				Description string
				Type        string
			}{
				ID:          row.ProgramID,
				Name:        row.ProgramName,
				Description: row.ProgramDescription,
				Type:        string(row.ProgramType),
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
		ProgramID: event.ProgramID,
		Capacity: sql.NullInt32{
			Int32: event.Capacity,
			Valid: event.Capacity != 0,
		},
	}

	_, dbErr := r.Queries.UpdateEvent(ctx, dbEventParams)

	if dbErr != nil {

		var pqErr *pq.Error
		if errors.As(dbErr, &pqErr) {

			if errInfo, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(errInfo.Message, errInfo.Status)
			}
		}

		log.Printf("Failed to update event: %+v. Error: %v", event, dbErr.Error())
		return errLib.New("Internal server error when updating event", http.StatusInternalServerError)
	}

	return nil

}

func (r *EventsRepository) DeleteEvents(c context.Context, ids []uuid.UUID) *errLib.CommonError {
	err := r.Queries.DeleteEventsByIds(c, ids)

	if err != nil {
		log.Printf("Failed to delete event with Ids: %s. Error: %s", ids, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *EventsRepository) DeleteUnmodifiedEventsByRecurrenceID(c context.Context, id uuid.UUID) *errLib.CommonError {
	err := r.Queries.DeleteUnmodifiedEventsByRecurrenceID(c, uuid.NullUUID{
		UUID:  id,
		Valid: true,
	})

	if err != nil {
		log.Printf("Failed to delete event with recurrence Id: %s. Error: %s", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
