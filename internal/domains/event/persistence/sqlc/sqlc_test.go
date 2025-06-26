package sqlc_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	identityDb "api/internal/domains/identity/persistence/sqlc/generated"
	locationDb "api/internal/domains/location/persistence/sqlc/generated"
	dbTestUtils "api/utils/test_utils"

	"github.com/google/uuid"

	courtDb "api/internal/domains/court/persistence/sqlc/generated"
	eventDb "api/internal/domains/event/persistence/sqlc/generated"
	programDb "api/internal/domains/program/persistence/sqlc/generated"

	"github.com/stretchr/testify/require"
)

func TestGetNonExistingEvent(t *testing.T) {
	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	eventQueries := eventDb.New(db)

	defer cleanup()

	_, err := eventQueries.GetEventById(context.Background(), uuid.New())

	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}

func TestCreateEvents(t *testing.T) {
	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	identityQueries := identityDb.New(db)
	eventQueries := eventDb.New(db)
	programQueries := programDb.New(db)
	locationQueries := locationDb.New(db)

	defer cleanup()

	creator, err := identityQueries.CreateUser(context.Background(), identityDb.CreateUserParams{
		FirstName: "John",
		LastName:  "Doe",
	})

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

	now := time.Now().Truncate(time.Second)
	testTimes := []struct {
		startAt  time.Time
		endAt    time.Time
		capacity int32
	}{
		{now, now.Add(time.Hour), 20},
		{now.Add(2 * time.Hour), now.Add(3 * time.Hour), 15},
		{now.AddDate(0, 0, 1), now.AddDate(0, 0, 1).Add(2 * time.Hour), 30},
	}

	createEventsParams := eventDb.CreateEventsParams{
		StartAtArray:            make([]time.Time, len(testTimes)),
		EndAtArray:              make([]time.Time, len(testTimes)),
		LocationIds:             make([]uuid.UUID, len(testTimes)),
		ProgramIds:              make([]uuid.UUID, len(testTimes)),
		CreatedByIds:            make([]uuid.UUID, len(testTimes)),
		IsCancelledArray:        make([]bool, len(testTimes)),
		IsDateTimeModifiedArray: make([]bool, len(testTimes)),
	}

	for i, tm := range testTimes {
		createEventsParams.StartAtArray[i] = tm.startAt
		createEventsParams.EndAtArray[i] = tm.endAt
		createEventsParams.LocationIds[i] = createdLocation.ID
		createEventsParams.ProgramIds[i] = createdProgram.ID
		createEventsParams.CreatedByIds[i] = creator.ID
		createEventsParams.IsDateTimeModifiedArray[i] = false
		createEventsParams.IsCancelledArray[i] = false
	}

	_, err = eventQueries.CreateEvents(context.Background(), createEventsParams)

	require.NoError(t, err)

	events, err := eventQueries.GetEvents(context.Background(), eventDb.GetEventsParams{
		CreatedBy: uuid.NullUUID{
			UUID:  creator.ID,
			Valid: true,
		},
	})

	require.NoError(t, err)
	require.Len(t, events, 3)

	for _, testTime := range testTimes {
		found := false
		for _, event := range events {
			if event.StartAt.Equal(testTime.startAt) {
				require.Equal(t, testTime.startAt.UTC(), event.StartAt.UTC())
				require.Equal(t, testTime.endAt.UTC(), event.EndAt.UTC())
				require.Equal(t, createdLocation.ID, event.LocationID)
				require.Equal(t, createdProgram.ID, event.ProgramID)
				found = true
				break
			}
		}
		require.True(t, found, "expected event not found: %v", testTime)
	}
}

func TestUpdateEvent(t *testing.T) {
	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	identityQueries := identityDb.New(db)
	eventQueries := eventDb.New(db)
	programQueries := programDb.New(db)
	locationQueries := locationDb.New(db)

	defer cleanup()

	creator, err := identityQueries.CreateUser(context.Background(), identityDb.CreateUserParams{
		FirstName: "John",
		LastName:  "Doe",
	})

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

	now := time.Now().Truncate(time.Second)

	createEventsParams := eventDb.CreateEventsParams{
		StartAtArray:            []time.Time{now},
		EndAtArray:              []time.Time{now.Add(24 * time.Hour)},
		LocationIds:             []uuid.UUID{createdLocation.ID},
		ProgramIds:              []uuid.UUID{createdProgram.ID},
		CreatedByIds:            []uuid.UUID{creator.ID},
		IsCancelledArray:        []bool{false},
		IsDateTimeModifiedArray: []bool{false},
	}

	_, err = eventQueries.CreateEvents(context.Background(), createEventsParams)
	require.NoError(t, err)

	createdEvents, err := eventQueries.GetEvents(context.Background(), eventDb.GetEventsParams{
		CreatedBy: uuid.NullUUID{
			UUID:  creator.ID,
			Valid: true,
		},
	})

	require.Len(t, createdEvents, 1)
	originalEvent := createdEvents[0]

	require.NoError(t, err)

	// Now, update the createdEvent
	newStart := now.Add(3 * time.Hour).UTC()
	newEnd := now.Add(4 * time.Hour).UTC()

	updateEventParams := eventDb.UpdateEventParams{
		ID:         originalEvent.ID,
		StartAt:    newStart,
		EndAt:      newEnd,
		LocationID: originalEvent.LocationID,
		ProgramID:  originalEvent.ProgramID,
		UpdatedBy:  originalEvent.CreatedBy,
	}

	updatedEvent, err := eventQueries.UpdateEvent(context.Background(), updateEventParams)
	require.NoError(t, err)

	// Assert updated createdEvent data (only comparing time)
	require.Equal(t, newStart, updatedEvent.StartAt.UTC())
	require.Equal(t, newEnd, updatedEvent.EndAt.UTC())
	require.Equal(t, originalEvent.LocationID, updatedEvent.LocationID)
	require.Equal(t, originalEvent.ProgramID, updatedEvent.ProgramID)
}

func TestDeleteEvent(t *testing.T) {
	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	identityQueries := identityDb.New(db)
	eventQueries := eventDb.New(db)
	programQueries := programDb.New(db)
	locationQueries := locationDb.New(db)

	defer cleanup()

	// Create a user to be the creator of the event
	createUserParams := identityDb.CreateUserParams{
		FirstName: "John",
		LastName:  "Doe",
	}

	creator, err := identityQueries.CreateUser(context.Background(), createUserParams)
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

	createEventsParams := eventDb.CreateEventsParams{
		StartAtArray:            []time.Time{now, now.Add(48 * time.Hour)},
		EndAtArray:              []time.Time{now.Add(24 * time.Hour), now.Add(72 * time.Hour)},
		LocationIds:             []uuid.UUID{createdLocation.ID, createdLocation.ID},
		ProgramIds:              []uuid.UUID{createdProgram.ID, createdProgram.ID},
		CreatedByIds:            []uuid.UUID{creator.ID, creator.ID},
		IsCancelledArray:        []bool{false, false},
		IsDateTimeModifiedArray: []bool{false, false},
	}

	_, err = eventQueries.CreateEvents(context.Background(), createEventsParams)

	require.NoError(t, err)

	createdEvents, err := eventQueries.GetEvents(context.Background(), eventDb.GetEventsParams{
		CreatedBy: uuid.NullUUID{
			UUID:  creator.ID,
			Valid: true,
		},
	})

	require.Len(t, createdEvents, 2)
	createdEvent1 := createdEvents[0]
	createdEvent2 := createdEvents[1]

	// Now, delete the createdEvent
	err = eventQueries.DeleteEventsByIds(context.Background(), []uuid.UUID{
		createdEvent1.ID,
		createdEvent2.ID,
	})

	require.NoError(t, err)

	// Try to fetch the deleted event

	var filter eventDb.GetEventsParams

	retrievedEvents, err := eventQueries.GetEvents(context.Background(), filter)

	require.Nil(t, err)
	require.Equal(t, 0, len(retrievedEvents))
}
func TestGetEventsByCourt(t *testing.T) {
	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	identityQueries := identityDb.New(db)
	eventQueries := eventDb.New(db)
	programQueries := programDb.New(db)
	locationQueries := locationDb.New(db)
	courtQueries := courtDb.New(db)

	defer cleanup()

	creator, err := identityQueries.CreateUser(context.Background(), identityDb.CreateUserParams{
		FirstName: "John",
		LastName:  "Doe",
	})
	require.NoError(t, err)

	createdProgram, err := programQueries.CreateProgram(context.Background(), programDb.CreateProgramParams{
		Name:  "Test Program",
		Level: "beginner",
		Type:  programDb.ProgramProgramTypeCourse,
	})
	require.NoError(t, err)

	createdLocation, err := locationQueries.CreateLocation(context.Background(), locationDb.CreateLocationParams{
		Name:    "Main Court",
		Address: "123 Street",
	})
	require.NoError(t, err)

	court1, err := courtQueries.CreateCourt(context.Background(), courtDb.CreateCourtParams{LocationID: createdLocation.ID, Name: "Court A"})
	require.NoError(t, err)
	court2, err := courtQueries.CreateCourt(context.Background(), courtDb.CreateCourtParams{LocationID: createdLocation.ID, Name: "Court B"})
	require.NoError(t, err)

	now := time.Now().Truncate(time.Second)

	createEventsParams := eventDb.CreateEventsParams{
		StartAtArray:            []time.Time{now, now.Add(time.Hour)},
		EndAtArray:              []time.Time{now.Add(time.Hour), now.Add(2 * time.Hour)},
		LocationIds:             []uuid.UUID{createdLocation.ID, createdLocation.ID},
		ProgramIds:              []uuid.UUID{createdProgram.ID, createdProgram.ID},
		CourtIds:                []uuid.UUID{court1.ID, court2.ID},
		TeamIds:                 []uuid.UUID{uuid.Nil, uuid.Nil},
		CreatedByIds:            []uuid.UUID{creator.ID, creator.ID},
		RecurrenceIds:           []uuid.UUID{uuid.Nil, uuid.Nil},
		IsCancelledArray:        []bool{false, false},
		IsDateTimeModifiedArray: []bool{false, false},
		CancellationReasons:     []string{"", ""},
	}

	_, err = eventQueries.CreateEvents(context.Background(), createEventsParams)
	require.NoError(t, err)

	events, err := eventQueries.GetEvents(context.Background(), eventDb.GetEventsParams{
		CourtID: uuid.NullUUID{UUID: court1.ID, Valid: true},
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, court1.ID, events[0].CourtID.UUID)
}
