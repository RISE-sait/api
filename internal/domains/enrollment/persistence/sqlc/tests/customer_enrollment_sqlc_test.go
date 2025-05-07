package events_test

import (
	"context"
	"testing"
	"time"

	locationDb "api/internal/domains/location/persistence/sqlc/generated"

	eventDb "api/internal/domains/event/persistence/sqlc/generated"

	enrollmentDb "api/internal/domains/enrollment/persistence/sqlc/generated"
	identityDb "api/internal/domains/identity/persistence/sqlc/generated"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	dbTestUtils "api/utils/test_utils"
)

func TestEnrollCustomerInEvent(t *testing.T) {
	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	identityQueries := identityDb.New(db)
	enrollmentQueries := enrollmentDb.New(db)
	eventQueries := eventDb.New(db)
	locationQueries := locationDb.New(db)

	defer cleanup()
	// Seed the database with a default program of type 'course'
	// This is a workaround to avoid the need for a separate migration for the test database
	_, err := db.ExecContext(context.Background(), `
	INSERT INTO program.programs (id, name, type, level, description)
	VALUES (gen_random_uuid(), 'Course', 'course', 'all', 'Default test program')
	ON CONFLICT (type) DO NOTHING
`)
require.NoError(t, err)
	// Create a user to be the creator of the event
	eventCreator, err := identityQueries.CreateUser(context.Background(), identityDb.CreateUserParams{
		FirstName: "John",
		LastName:  "Doe",
	})
	require.NoError(t, err)

	// Use seeded program of type 'course'
	var createdProgram struct {
		ID          uuid.UUID
		Name        string
		Type        string
		Level       string
		Description string
	}

	err = db.QueryRowContext(context.Background(), `
		SELECT id, name, type, level, description
		FROM program.programs
		WHERE type = 'course'
	`).Scan(&createdProgram.ID, &createdProgram.Name, &createdProgram.Type, &createdProgram.Level, &createdProgram.Description)
	require.NoError(t, err)

	// Create a location
	createdLocation, err := locationQueries.CreateLocation(context.Background(), locationDb.CreateLocationParams{
		Name:    "Main Conference Room",
		Address: "123 Main St",
	})
	require.NoError(t, err)

	now := time.Now().Truncate(time.Second)

	_, err = eventQueries.CreateEvents(context.Background(), eventDb.CreateEventsParams{
		StartAtArray:            []time.Time{now},
		EndAtArray:              []time.Time{now.Add(time.Hour * 24)},
		LocationIds:             []uuid.UUID{createdLocation.ID},
		ProgramIds:              []uuid.UUID{createdProgram.ID},
		CreatedByIds:            []uuid.UUID{eventCreator.ID},
		IsCancelledArray:        []bool{false},
		IsDateTimeModifiedArray: []bool{false},
	})
	require.NoError(t, err)

	// Create a customer to enroll
	createdCustomer, err := identityQueries.CreateUser(context.Background(), identityDb.CreateUserParams{
		FirstName: "Jane",
		LastName:  "Smith",
	})
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

	err = enrollmentQueries.EnrollCustomerInEvent(context.Background(), enrollmentDb.EnrollCustomerInEventParams{
		CustomerID: createdCustomer.ID,
		EventID:    createdEvent.ID,
	})
	require.NoError(t, err)

	customers, err := eventQueries.GetEventCustomers(context.Background(), createdEvent.ID)
	require.NoError(t, err)
	require.Equal(t, 1, len(customers))

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
	enrollmentQueries := enrollmentDb.New(db)
	eventQueries := eventDb.New(db)
	locationQueries := locationDb.New(db)

	defer cleanup()

	// Seed the database with a default program of type 'course'
	// This is a workaround to avoid the need for a separate migration for the test database
	_, err := db.ExecContext(context.Background(), `
	INSERT INTO program.programs (id, name, type, level, description)
	VALUES (gen_random_uuid(), 'Course', 'course', 'all', 'Default test program')
	ON CONFLICT (type) DO NOTHING
	`)
	require.NoError(t, err)

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

	// Use seeded program of type 'course'
	var createdProgram struct {
		ID          uuid.UUID
		Name        string
		Type        string
		Level       string
		Description string
	}

	err = db.QueryRowContext(context.Background(), `
		SELECT id, name, type, level, description
		FROM program.programs
		WHERE type = 'course'
	`).Scan(&createdProgram.ID, &createdProgram.Name, &createdProgram.Type, &createdProgram.Level, &createdProgram.Description)
	require.NoError(t, err)

	createdLocation, err := locationQueries.CreateLocation(context.Background(), locationDb.CreateLocationParams{
		Name:    "Main Conference Room",
		Address: "123 Main St",
	})
	require.NoError(t, err)

	numEvents := 20
	capacity := int32(20)
	duration := time.Hour * 2
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
		currentTime = endTimes[i].Add(gap)
	}

	_, err = eventQueries.CreateEvents(context.Background(), eventDb.CreateEventsParams{
		LocationIds:             locationIDs,
		ProgramIds:              programIDs,
		CreatedByIds:            createdByIDs,
		StartAtArray:            startTimes,
		EndAtArray:              endTimes,
		IsCancelledArray:        isCancelledArray,
		CancellationReasons:     cancellationReasons,
		IsDateTimeModifiedArray: make([]bool, numEvents),
	})
	require.NoError(t, err)

	err = enrollmentQueries.EnrollCustomerInProgram(context.Background(), enrollmentDb.EnrollCustomerInProgramParams{
		CustomerID: customer.ID,
		ProgramID:  createdProgram.ID,
	})
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
