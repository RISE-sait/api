package tests_test

//
//import (
//	"api/internal/custom_types"
//	courseTestUtils "api/internal/domains/course/persistence/test_utils"
//	eventTestUtils "api/internal/domains/event/persistence/test_utils"
//	gameTestUtils "api/internal/domains/game/persistence/test_utils"
//	facilityDb "api/internal/domains/location/persistence/sqlc/generated"
//	facilityTestUtils "api/internal/domains/location/persistence/test_utils"
//	practiceTestUtils "api/internal/domains/practice/persistence/test_utils"
//
//	"api/utils/test_utils"
//	"context"
//	"github.com/google/uuid"
//	"log"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/require"
//
//	eventDb "api/internal/domains/event/persistence/sqlc/generated"
//	practiceDb "api/internal/domains/practice/persistence/sqlc/generated"
//)
//
//func TestCreateEvent(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	practiceQueries, _ := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//	facilityQueries, _ := facilityTestUtils.SetupFacilityTestDbQueries(t, dbConn)
//	_, _ = courseTestUtils.SetupCourseTestDb(t, dbConn)
//	_, _ = gameTestUtils.SetupGameTestDb(t, dbConn)
//	eventQueries, _ := eventTestUtils.SetupEventTestDbQueries(t, dbConn)
//
//	practiceName := "Go Basics Practice"
//	createPracticeParams := practiceDb.CreatePracticeParams{
//		Name:  practiceName,
//		Level: "beginner",
//	}
//
//	practice, err := practiceQueries.CreateTeam(context.Background(), createPracticeParams)
//	require.NoError(t, err)
//
//	facilityCategory, err := facilityQueries.CreateFacilityCategory(context.Background(), "Basketball")
//
//	facilityName := "Main Conference Room"
//	createFacilityParams := facilityDb.CreateFacilityParams{
//		Name:               facilityName,
//		FacilityCategoryID: facilityCategory.ID,
//	}
//
//	facility, err := facilityQueries.CreateFacility(context.Background(), createFacilityParams)
//	require.NoError(t, err)
//
//	now := time.Now()
//
//	createEventParams := eventDb.CreateEventParams{
//		EventStartAt: now,
//		EventEndAt:   now.Add(time.Hour * 24),
//		EventStartTime: custom_types.TimeWithTimeZone{
//			Time: now.Add(1 * time.Hour).UTC().Format("15:04:05+00:00"),
//		},
//		EventEndTime: custom_types.TimeWithTimeZone{
//			Time: now.Add(2 * time.Hour).UTC().Format("15:04:05+00:00"),
//		},
//		LocationID: facility.ID,
//		PracticeID: uuid.NullUUID{UUID: practice.ID, Valid: true},
//		Day:        eventDb.DayEnum("MONDAY"), // Sample day
//	}
//
//	event, err := eventQueries.CreateEvent(context.Background(), createEventParams)
//
//	require.NoError(t, err)
//
//	require.Equal(t, createEventParams.EventStartTime.Time, event.EventStartTime.Time)
//	require.Equal(t, createEventParams.EventEndTime.Time, event.EventEndTime.Time)
//
//	require.Equal(t, createEventParams.LocationID, event.LocationID)
//	require.Equal(t, createEventParams.PracticeID, event.PracticeID)
//	require.Equal(t, createEventParams.Day, event.Day)
//}
//
//func TestUpdateEvent(t *testing.T) {
//	dbConn, _ := test_utils.SetupTestDB(t)
//	_, _ = gameTestUtils.SetupGameTestDb(t, dbConn)
//
//	practiceQueries, _ := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//	facilityQueries, _ := facilityTestUtils.SetupFacilityTestDbQueries(t, dbConn)
//	eventQueries, _ := eventTestUtils.SetupEventTestDbQueries(t, dbConn)
//
//	// Create practice and facility
//	practice, _ := practiceQueries.CreateTeam(context.Background(), practiceDb.CreatePracticeParams{
//		Name:  "Advanced Go",
//		Level: "intermediate",
//	})
//
//	facilityType, _ := facilityQueries.CreateFacilityCategory(context.Background(), "Soccer")
//	facility, _ := facilityQueries.CreateFacility(context.Background(), facilityDb.CreateFacilityParams{
//		Name:               "Indoor Stadium",
//		FacilityCategoryID: facilityType.ID,
//	})
//
//	now := time.Now()
//
//	// Create an event to update
//	createEventParams := eventDb.CreateEventParams{
//		EventStartTime: custom_types.TimeWithTimeZone{
//			Time: now.Add(1 * time.Hour).UTC().Format("15:04:05+00:00"),
//		},
//		EventEndTime: custom_types.TimeWithTimeZone{
//			Time: now.Add(2 * time.Hour).UTC().Format("15:04:05+00:00"),
//		},
//		LocationID: facility.ID,
//		PracticeID: uuid.NullUUID{UUID: practice.ID, Valid: true},
//		Day:        eventDb.DayEnum("TUESDAY"),
//	}
//
//	event, err := eventQueries.CreateEvent(context.Background(), createEventParams)
//	require.NoError(t, err)
//
//	// Now, update the event
//	newBeginTime := now.Add(3 * time.Hour).UTC().Format("15:04:05+00:00")
//	newEndTime := now.Add(4 * time.Hour).UTC().Format("15:04:05+00:00")
//
//	updateEventParams := eventDb.UpdateEventParams{
//		ID:               event.ID,
//		EventStartTime: custom_types.TimeWithTimeZone{Time: newBeginTime},
//		EventEndTime:   custom_types.TimeWithTimeZone{Time: newEndTime},
//		LocationID:       event.LocationID,
//		PracticeID:       event.PracticeID,
//		Day:              event.Day,
//	}
//
//	updatedEvent, err := eventQueries.UpdateEvent(context.Background(), updateEventParams)
//	require.NoError(t, err)
//
//	// Assert updated event data (only comparing time)
//	require.Equal(t, newBeginTime, updatedEvent.EventStartTime.Time)
//	require.Equal(t, newEndTime, updatedEvent.EventEndTime.Time)
//}
//
//func TestDeleteEvent(t *testing.T) {
//	dbConn, _ := test_utils.SetupTestDB(t)
//	_, _ = gameTestUtils.SetupGameTestDb(t, dbConn)
//
//	practiceQueries, _ := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//	facilityQueries, _ := facilityTestUtils.SetupFacilityTestDbQueries(t, dbConn)
//	eventQueries, _ := eventTestUtils.SetupEventTestDbQueries(t, dbConn)
//
//	// Create practice and facility first
//	practiceName := "Basic Go"
//	createPracticeParams := practiceDb.CreatePracticeParams{
//		Name:  practiceName,
//		Level: "beginner",
//	}
//
//	practice, err := practiceQueries.CreateTeam(context.Background(), createPracticeParams)
//	require.NoError(t, err)
//
//	facilityType, err := facilityQueries.CreateFacilityCategory(context.Background(), "Basketball")
//	require.NoError(t, err)
//
//	facilityName := "Main Court"
//	createFacilityParams := facilityDb.CreateFacilityParams{
//		Name:               facilityName,
//		FacilityCategoryID: facilityType.ID,
//	}
//
//	facility, err := facilityQueries.CreateFacility(context.Background(), createFacilityParams)
//	require.NoError(t, err)
//
//	now := time.Now()
//
//	// Create an event to delete
//	createEventParams := eventDb.CreateEventParams{
//		EventStartTime: custom_types.TimeWithTimeZone{
//			Time: now.Add(1 * time.Hour).UTC().Format("15:04:05+00:00"),
//		},
//		EventEndTime: custom_types.TimeWithTimeZone{
//			Time: now.Add(2 * time.Hour).UTC().Format("15:04:05+00:00"),
//		},
//		LocationID: facility.ID,
//		PracticeID: uuid.NullUUID{UUID: practice.ID, Valid: true},
//		Day:        eventDb.DayEnum("WEDNESDAY"),
//	}
//
//	event, err := eventQueries.CreateEvent(context.Background(), createEventParams)
//	require.NoError(t, err)
//
//	// Delete the event
//	affectedRows, err := eventQueries.DeleteEvent(context.Background(), event.ID)
//	require.NoError(t, err)
//	require.Equal(t, int64(1), affectedRows)
//
//	// Attempt to fetch the deleted event (expecting an error)
//	_, err = eventQueries.GetEventById(context.Background(), event.ID)
//	require.Error(t, err)
//}
//
//func TestGetEventById(t *testing.T) {
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	// Set up practice, facility, and event queries
//	_, _ = gameTestUtils.SetupGameTestDb(t, dbConn)
//	practiceQueries, _ := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//	facilityQueries, _ := facilityTestUtils.SetupFacilityTestDbQueries(t, dbConn)
//	eventQueries, _ := eventTestUtils.SetupEventTestDbQueries(t, dbConn)
//
//	// Create practice and facility first
//	practiceName := "Go Advanced"
//	createPracticeParams := practiceDb.CreatePracticeParams{
//		Name:  practiceName,
//		Level: "advanced",
//	}
//
//	practice, err := practiceQueries.CreateTeam(context.Background(), createPracticeParams)
//	require.NoError(t, err)
//
//	facilityType, err := facilityQueries.CreateFacilityCategory(context.Background(), "Indoor")
//	require.NoError(t, err)
//
//	facilityName := "Main Stadium"
//	createFacilityParams := facilityDb.CreateFacilityParams{
//		Name:               facilityName,
//		FacilityCategoryID: facilityType.ID,
//	}
//
//	facility, err := facilityQueries.CreateFacility(context.Background(), createFacilityParams)
//	require.NoError(t, err)
//
//	now := time.Now()
//
//	// Create an event
//	createEventParams := eventDb.CreateEventParams{
//		EventStartTime: custom_types.TimeWithTimeZone{
//			Time: now.Add(1 * time.Hour).UTC().Format("15:04:05+00:00"),
//		},
//		EventEndTime: custom_types.TimeWithTimeZone{
//			Time: now.Add(2 * time.Hour).UTC().Format("15:04:05+00:00"),
//		},
//		LocationID: facility.ID,
//		PracticeID: uuid.NullUUID{UUID: practice.ID, Valid: true},
//		Day:        eventDb.DayEnum("THURSDAY"),
//	}
//
//	event, err := eventQueries.CreateEvent(context.Background(), createEventParams)
//	require.NoError(t, err)
//
//	// Get event by ID
//	fetchedEvent, err := eventQueries.GetEventById(context.Background(), event.ID)
//	require.NoError(t, err)
//
//	// Assert fetched event matches created event
//	require.Equal(t, event.ID, fetchedEvent.ID)
//	require.Equal(t, event.EventStartTime.Time, fetchedEvent.EventStartTime.Time)
//	require.Equal(t, event.EventEndTime.Time, fetchedEvent.EventEndTime.Time)
//}
//
//func TestGetEvents(t *testing.T) {
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	// Set up practice, facility, and event queries
//	_, _ = gameTestUtils.SetupGameTestDb(t, dbConn)
//	practiceQueries, _ := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//	facilityQueries, _ := facilityTestUtils.SetupFacilityTestDbQueries(t, dbConn)
//	eventQueries, _ := eventTestUtils.SetupEventTestDbQueries(t, dbConn)
//
//	// Create practice and facility first
//	practiceName := "Go Basics Practice"
//	createPracticeParams := practiceDb.CreatePracticeParams{
//		Name:  practiceName,
//		Level: "beginner",
//	}
//
//	practice, err := practiceQueries.CreateTeam(context.Background(), createPracticeParams)
//	require.NoError(t, err)
//
//	facilityType, err := facilityQueries.CreateFacilityCategory(context.Background(), "Basketball")
//	require.NoError(t, err)
//
//	facilityName := "Main Conference Room"
//	createFacilityParams := facilityDb.CreateFacilityParams{
//		Name:               facilityName,
//		FacilityCategoryID: facilityType.ID,
//	}
//
//	facility, err := facilityQueries.CreateFacility(context.Background(), createFacilityParams)
//	require.NoError(t, err)
//
//	baseTime := time.Date(2025, time.February, 17, 1, 0, 0, 0, time.UTC)
//
//	// Create some events
//	for i := 1; i <= 5; i++ {
//
//		beginTime := baseTime.Add(time.Duration(i*2) * time.Hour).UTC().Format("15:04:05+00:00")
//		endTime := baseTime.Add(time.Duration((i*2)+1) * time.Hour).UTC().Format("15:04:05+00:00")
//
//		log.Println("begin ", beginTime)
//		log.Println("end ", endTime)
//
//		createEventParams := eventDb.CreateEventParams{
//			EventStartTime: custom_types.TimeWithTimeZone{
//				Time: beginTime,
//			},
//			EventEndTime: custom_types.TimeWithTimeZone{
//				Time: endTime,
//			},
//			LocationID: facility.ID,
//			PracticeID: uuid.NullUUID{UUID: practice.ID, Valid: true},
//			Day:        eventDb.DayEnum("FRIDAY"),
//		}
//		_, err := eventQueries.CreateEvent(context.Background(), createEventParams)
//		require.NoError(t, err)
//	}
//
//	// Fetch events
//	events, err := eventQueries.GetBarberServices(context.Background())
//	require.NoError(t, err)
//
//	// Assert that at least 5 events exist
//	require.GreaterOrEqual(t, len(events), 5)
//}
