package events_test

import (
	locationDb "api/internal/domains/location/persistence/sqlc/generated"

	identityDb "api/internal/domains/identity/persistence/sqlc/generated"

	"database/sql"

	"context"
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	eventDb "api/internal/domains/event/persistence/sqlc/generated"
	programDb "api/internal/domains/program/persistence/sqlc/generated"

	enrollmentDb "api/internal/domains/enrollment/persistence/sqlc/generated"
	dbTestUtils "api/utils/test_utils"
)

func TestEnrollCustomerInEvent(t *testing.T) {

	_, identityQueries, eventQueries, programQueries, enrollmentQueries, locationQueries, _, _, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	// Create a user to be the creator of the event
	createUserParams := identityDb.CreateUserParams{
		FirstName: "John",
		LastName:  "Doe",
	}

	createdUser, err := identityQueries.CreateUser(context.Background(), createUserParams)
	require.NoError(t, err)

	createProgramParams := programDb.CreateProgramParams{
		Name:  "Go Basics Practice",
		Level: "beginner",
		Type:  programDb.ProgramProgramTypeCourse,
	}

	createdProgram, err := programQueries.CreateProgram(context.Background(), createProgramParams)
	require.NoError(t, err)

	createLocationParams := locationDb.CreateLocationParams{
		Name:    "Main Conference Room",
		Address: "123 Main St",
	}

	createdLocation, err := locationQueries.CreateLocation(context.Background(), createLocationParams)
	require.NoError(t, err)

	now := time.Now().Truncate(time.Second)
	capacity := 20

	createEventParams := eventDb.CreateEventParams{
		StartAt:    now,
		EndAt:      now.Add(time.Hour * 24),
		LocationID: createdLocation.ID,
		ProgramID:  uuid.NullUUID{UUID: createdProgram.ID, Valid: true},
		Capacity: sql.NullInt32{
			Int32: int32(capacity),
			Valid: true,
		},
		CreatedBy: createdUser.ID,
	}

	createdEvent, err := eventQueries.CreateEvent(context.Background(), createEventParams)

	require.NoError(t, err)

	// Create a customer to enroll in the event
	createCustomerParams := identityDb.CreateUserParams{
		FirstName: "Jane",
		LastName:  "Smith",
	}

	createdCustomer, err := identityQueries.CreateUser(context.Background(), createCustomerParams)
	require.NoError(t, err)

	// Enroll the customer in the event
	enrollParams := enrollmentDb.EnrollCustomerInEventParams{
		CustomerID: createdCustomer.ID,
		EventID:    createdEvent.ID,
	}

	err = enrollmentQueries.EnrollCustomerInEvent(context.Background(), enrollParams)

	require.NoError(t, err)

	customers, err := eventQueries.GetEventCustomers(context.Background(), createdEvent.ID)

	require.NoError(t, err)
	require.Equal(t, len(customers), 1)

	customer := customers[0]

	require.Equal(t, customer.CustomerID, createdCustomer.ID)
	require.Equal(t, customer.CustomerFirstName, createdCustomer.FirstName)
	require.Equal(t, customer.CustomerLastName, createdCustomer.LastName)
	require.Equal(t, customer.CustomerEmail, createdCustomer.Email)
	require.Equal(t, customer.CustomerPhone, createdCustomer.Phone)
	require.Equal(t, customer.CustomerEnrollmentIsCancelled, false)
}

func TestEnrollCustomerInProgramEvents(t *testing.T) {

	_, identityQueries, eventQueries, programQueries, enrollmentQueries, locationQueries, _, _, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	// Create a user to be the creator of the event
	createUserParams := identityDb.CreateUserParams{
		FirstName: "John",
		LastName:  "Doe",
	}

	createdUser, err := identityQueries.CreateUser(context.Background(), createUserParams)
	require.NoError(t, err)

	createProgramParams := programDb.CreateProgramParams{
		Name:  "Go Basics Practice",
		Level: "beginner",
		Type:  programDb.ProgramProgramTypeCourse,
	}

	createdProgram, err := programQueries.CreateProgram(context.Background(), createProgramParams)
	require.NoError(t, err)

	createLocationParams := locationDb.CreateLocationParams{
		Name:    "Main Conference Room",
		Address: "123 Main St",
	}

	createdLocation, err := locationQueries.CreateLocation(context.Background(), createLocationParams)
	require.NoError(t, err)

	numEvents := 20
	capacity := int32(20)
	duration := time.Hour * 2 // Each event lasts 2 hours
	gap := time.Hour * 2

	locationIDs := make([]uuid.UUID, numEvents)
	programIDs := make([]uuid.UUID, numEvents)
	createdByIDs := make([]uuid.UUID, numEvents)
	updatedByIDs := make([]uuid.UUID, numEvents)
	startTimes := make([]time.Time, numEvents)
	endTimes := make([]time.Time, numEvents)
	capacities := make([]int32, numEvents)
	isCancelledArray := make([]bool, numEvents)
	cancellationReasons := make([]string, numEvents)

	// Set initial time (truncated to nearest hour)
	currentTime := time.Now().Truncate(time.Hour).Add(time.Hour * 1)

	for i := 0; i < numEvents; i++ {
		locationIDs[i] = createdLocation.ID
		programIDs[i] = createdProgram.ID
		createdByIDs[i] = createdUser.ID
		updatedByIDs[i] = createdUser.ID
		startTimes[i] = currentTime
		endTimes[i] = currentTime.Add(duration)
		capacities[i] = capacity
		isCancelledArray[i] = false
		cancellationReasons[i] = ""

		// Move to next time slot with gap
		currentTime = endTimes[i].Add(gap)
	}

	createEventsParams := eventDb.CreateEventsParams{
		LocationIds:         locationIDs,
		ProgramIds:          programIDs,
		CreatedByIds:        createdByIDs,
		UpdatedByIds:        updatedByIDs,
		StartAtArray:        startTimes,
		EndAtArray:          endTimes,
		Capacities:          capacities,
		IsCancelledArray:    isCancelledArray,
		CancellationReasons: cancellationReasons,
	}

	err = eventQueries.CreateEvents(context.Background(), createEventsParams)

	require.NoError(t, err)

	enrollParams := enrollmentDb.EnrollCustomerInProgramEventsParams{
		CustomerID: createdUser.ID,
		ProgramID:  createdProgram.ID,
	}

	err = enrollmentQueries.EnrollCustomerInProgramEvents(context.Background(), enrollParams)
	require.NoError(t, err)

	events, err := eventQueries.GetEvents(context.Background(), eventDb.GetEventsParams{
		UserID: uuid.NullUUID{
			UUID:  createdUser.ID,
			Valid: true,
		},
	})

	require.NoError(t, err)
	require.Equal(t, 20, len(events))
}
