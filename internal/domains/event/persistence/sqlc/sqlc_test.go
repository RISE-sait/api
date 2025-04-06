package sqlc_test

import (
	locationDb "api/internal/domains/location/persistence/sqlc/generated"
	dbTestUtils "api/utils/test_utils"

	identityDb "api/internal/domains/identity/persistence/sqlc/generated"

	"database/sql"

	"context"
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	eventDb "api/internal/domains/event/persistence/sqlc/generated"
	programDb "api/internal/domains/program/persistence/sqlc/generated"
)

func TestCreateEvent(t *testing.T) {

	identityQueries, eventQueries, programQueries, _, locationQueries, _, _, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

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

	require.Equal(t, createEventParams.StartAt.UTC(), createdEvent.StartAt.UTC())
	require.Equal(t, createEventParams.EndAt.UTC(), createdEvent.EndAt.UTC())

	require.Equal(t, createEventParams.LocationID, createdEvent.LocationID)
	require.Equal(t, createEventParams.ProgramID, createdEvent.ProgramID)
	require.Equal(t, createEventParams.Capacity, createdEvent.Capacity)
}

func TestUpdateEvent(t *testing.T) {

	identityQueries, eventQueries, programQueries, _, locationQueries, _, _, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

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

	// Now, update the createdEvent
	newBeginTime := now.Add(3 * time.Hour).UTC()
	newEndTime := now.Add(4 * time.Hour).UTC()

	updateEventParams := eventDb.UpdateEventParams{
		ID:         createdEvent.ID,
		StartAt:    newBeginTime,
		EndAt:      newEndTime,
		LocationID: createdEvent.LocationID,
		ProgramID:  createdEvent.ProgramID,
		Capacity:   createdEvent.Capacity,
		UpdatedBy:  createdEvent.CreatedBy,
	}

	updatedEvent, err := eventQueries.UpdateEvent(context.Background(), updateEventParams)
	require.NoError(t, err)

	// Assert updated createdEvent data (only comparing time)
	require.Equal(t, newBeginTime, updatedEvent.StartAt.UTC())
	require.Equal(t, newEndTime, updatedEvent.EndAt.UTC())
	require.Equal(t, createdEvent.LocationID, updatedEvent.LocationID)
	require.Equal(t, createdEvent.ProgramID, updatedEvent.ProgramID)
	require.Equal(t, createdEvent.Capacity, updatedEvent.Capacity)
}

func TestDeleteEvent(t *testing.T) {

	identityQueries, eventQueries, programQueries, _, locationQueries, _, _, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

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

	// Now, delete the createdEvent
	err = eventQueries.DeleteEvent(context.Background(), createdEvent.ID)

	require.NoError(t, err)

	// Try to fetch the deleted event

	_, err = eventQueries.GetEventById(context.Background(), createdEvent.ID)

	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}
