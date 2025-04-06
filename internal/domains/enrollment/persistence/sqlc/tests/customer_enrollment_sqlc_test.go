package events_test

import (
	eventTestUtils "api/internal/domains/event/persistence/test_utils"

	enrollmentTestUtils "api/internal/domains/enrollment/persistence/test_utils"
	locationDb "api/internal/domains/location/persistence/sqlc/generated"
	locationTestUtils "api/internal/domains/location/persistence/test_utils"
	programTestUtils "api/internal/domains/program/persistence/test_utils"
	teamTestUtils "api/internal/domains/team/persistence/test_utils"

	identityDb "api/internal/domains/identity/persistence/sqlc/generated"
	identityTestUtils "api/internal/domains/identity/persistence/test_utils"
	userTestUtils "api/internal/domains/user/persistence/test_utils"

	"database/sql"

	"api/utils/test_utils"
	"context"
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	eventDb "api/internal/domains/event/persistence/sqlc/generated"
	programDb "api/internal/domains/program/persistence/sqlc/generated"

	enrollmentDb "api/internal/domains/enrollment/persistence/sqlc/generated"
)

func dbSetup(t *testing.T) (identityQ *identityDb.Queries, eventQ *eventDb.Queries, programQ *programDb.Queries, enrollmentQ *enrollmentDb.Queries, locationQ *locationDb.Queries, cleanup func()) {
	dbConn, _ := test_utils.SetupTestDB(t)

	identityQueries, identityCleanup := identityTestUtils.SetupIdentityTestDb(t, dbConn)

	_, userCleanup := userTestUtils.SetupUsersTestDb(t, dbConn)
	_, staffCleanup := userTestUtils.SetupStaffsTestDb(t, dbConn)
	_, teamCleanup := teamTestUtils.SetupTeamTestDbQueries(t, dbConn)
	programQueries, programCleanup := programTestUtils.SetupProgramTestDbQueries(t, dbConn)
	locationQueries, locationCleanup := locationTestUtils.SetupLocationTestDbQueries(t, dbConn)
	eventQueries, eventCleanup := eventTestUtils.SetupEventTestDbQueries(t, dbConn)
	enrollmentQueries, enrollmentCleanup := enrollmentTestUtils.SetupEnrollmentTestDbQueries(t, dbConn)

	cleanup = func() {
		enrollmentCleanup()
		eventCleanup()
		locationCleanup()
		programCleanup()
		teamCleanup()
		staffCleanup()
		userCleanup()
		identityCleanup()
	}

	return identityQueries, eventQueries, programQueries, enrollmentQueries, locationQueries, cleanup
}

func TestEnrollCustomerInEvent(t *testing.T) {

	identityQueries, eventQueries, programQueries, enrollmentQueries, locationQueries, cleanup := dbSetup(t)

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

	identityQueries, eventQueries, programQueries, postPaymentQueries, locationQueries, cleanup := dbSetup(t)

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

	err = postPaymentQueries.EnrollCustomerInProgramEvents(context.Background(), enrollParams)
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
