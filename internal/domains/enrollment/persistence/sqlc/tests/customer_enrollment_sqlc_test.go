package events_test

import (
	locationDb "api/internal/domains/location/persistence/sqlc/generated"

	eventDb "api/internal/domains/event/persistence/sqlc/generated"
	programDb "api/internal/domains/program/persistence/sqlc/generated"

	enrollmentDb "api/internal/domains/enrollment/persistence/sqlc/generated"
	identityDb "api/internal/domains/identity/persistence/sqlc/generated"

	"context"
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbTestUtils "api/utils/test_utils"
)

func TestEnrollCustomerInEvent(t *testing.T) {

	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	identityQueries := identityDb.New(db)
	programQueries := programDb.New(db)
	enrollmentQueries := enrollmentDb.New(db)
	eventQueries := eventDb.New(db)
	locationQueries := locationDb.New(db)

	defer cleanup()

	// Create a user to be the creator of the event
	createUserParams := identityDb.CreateUserParams{
		FirstName: "John",
		LastName:  "Doe",
	}

	eventCreator, err := identityQueries.CreateUser(context.Background(), createUserParams)
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

	createEventsParams := eventDb.CreateEventsParams{
		StartAtArray:            []time.Time{now},
		EndAtArray:              []time.Time{now.Add(time.Hour * 24)},
		LocationIds:             []uuid.UUID{createdLocation.ID},
		ProgramIds:              []uuid.UUID{createdProgram.ID},
		Capacities:              []int32{int32(capacity)},
		CreatedByIds:            []uuid.UUID{eventCreator.ID},
		IsCancelledArray:        []bool{false},
		IsDateTimeModifiedArray: []bool{false},
	}

	_, err = eventQueries.CreateEvents(context.Background(), createEventsParams)

	require.NoError(t, err)

	// Create a customer to enroll in the event
	createCustomerParams := identityDb.CreateUserParams{
		FirstName: "Jane",
		LastName:  "Smith",
	}

	createdCustomer, err := identityQueries.CreateUser(context.Background(), createCustomerParams)
	require.NoError(t, err)

	events, err := eventQueries.GetEvents(context.Background(), eventDb.GetEventsParams{
		CreatedBy: uuid.NullUUID{
			UUID:  eventCreator.ID,
			Valid: true,
		},
	})

	require.NoError(t, err)
	require.Equal(t, 1, len(events))
	createdEvent := events[0]

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

	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	identityQueries := identityDb.New(db)
	programQueries := programDb.New(db)
	enrollmentQueries := enrollmentDb.New(db)
	eventQueries := eventDb.New(db)
	locationQueries := locationDb.New(db)

	defer cleanup()

	creator, err := identityQueries.CreateUser(context.Background(), identityDb.CreateUserParams{
		FirstName: "John",
		LastName:  "Doe",
	})
	require.NoError(t, err)

	customer, err := identityQueries.CreateUser(context.Background(), identityDb.CreateUserParams{
		FirstName: "Klint",
		LastName:  "Doe",
	})

	require.NoError(t, err)

	err = identityQueries.CreateAthlete(context.Background(), customer.ID)

	require.NoError(t, err)

	createdProgram, err := programQueries.CreateProgram(context.Background(), programDb.CreateProgramParams{
		Name:  "Go Basics Practice",
		Level: "beginner",
		Type:  programDb.ProgramProgramTypeCourse,
	})
	require.NoError(t, err)

	createdLocation, err := locationQueries.CreateLocation(context.Background(), locationDb.CreateLocationParams{
		Name:    "Main Conference Room",
		Address: "123 Main St",
	})
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
		createdByIDs[i] = creator.ID
		updatedByIDs[i] = creator.ID
		startTimes[i] = currentTime
		endTimes[i] = currentTime.Add(duration)
		capacities[i] = capacity
		isCancelledArray[i] = false
		cancellationReasons[i] = ""

		// Move to next time slot with gap
		currentTime = endTimes[i].Add(gap)
	}

	createEventsParams := eventDb.CreateEventsParams{
		LocationIds:             locationIDs,
		ProgramIds:              programIDs,
		CreatedByIds:            createdByIDs,
		StartAtArray:            startTimes,
		EndAtArray:              endTimes,
		Capacities:              capacities,
		IsCancelledArray:        isCancelledArray,
		CancellationReasons:     cancellationReasons,
		IsDateTimeModifiedArray: make([]bool, numEvents),
	}

	_, err = eventQueries.CreateEvents(context.Background(), createEventsParams)

	require.NoError(t, err)

	enrollParams := enrollmentDb.EnrollCustomerInProgramParams{
		CustomerID: customer.ID,
		ProgramID:  createdProgram.ID,
	}

	err = enrollmentQueries.EnrollCustomerInProgram(context.Background(), enrollParams)
	require.NoError(t, err)

	events, err := eventQueries.GetEvents(context.Background(), eventDb.GetEventsParams{
		ParticipantID: uuid.NullUUID{
			UUID:  customer.ID,
			Valid: true,
		},
	})

	require.NoError(t, err)
	require.Equal(t, 20, len(events))
}
