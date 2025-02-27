package enrollment

//
//import (
//	courseTestUtils "api/internal/domains/course/persistence/test_utils"
//	eventTestUtils "api/internal/domains/event/persistence/test_utils"
//	"api/utils/test_utils"
//	"context"
//	"database/sql"
//	"errors"
//	"fmt"
//	"github.com/google/uuid"
//	"github.com/lib/pq"
//	"testing"
//
//	"github.com/stretchr/testify/require"
//
//	db "api/internal/domains/course/persistence/sqlc/generated"
//)
//
//func TestCreateEnrollment(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	eventQueries, _ := eventTestUtils.SetupEventTestDbQueries(t, dbConn)
//	queries, _ := courseTestUtils.SetupCourseTestDb(t, dbConn)
//
//	name := "Go Course"
//	description := "Learn Go programming"
//
//	createCourseParams := db.CreateCourseParams{
//		Name:        name,
//		Description: sql.NullString{String: description, Valid: description != ""},
//	}
//
//	course, err := queries.CreateGame(context.Background(), createCourseParams)
//
//	require.NoError(t, err)
//
//	// Assert course data
//	require.Equal(t, name, course.Name)
//	require.Equal(t, description, course.Description.String)
//}
//
//func TestUpdateCourse(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, _ := courseTestUtils.SetupCourseTestDb(t, dbConn)
//
//	// Create a course to update
//	name := "Go Course"
//	description := "Learn Go programming"
//	createCourseParams := db.CreateCourseParams{
//		Name:        name,
//		Description: sql.NullString{String: description, Valid: description != ""},
//	}
//
//	course, err := queries.CreateGame(context.Background(), createCourseParams)
//	require.NoError(t, err)
//
//	// Now, update the course
//	newName := "Advanced Go Course"
//	updateParams := db.UpdateCourseParams{
//		HubSpotId:          course.HubSpotId,
//		Name:        newName,
//		Description: sql.NullString{String: "Learn advanced Go programming", Valid: true},
//	}
//
//	_, err = queries.UpdateGame(context.Background(), updateParams)
//	require.NoError(t, err)
//
//	// Get the updated course and verify
//	updatedCourse, err := queries.GetGameById(context.Background(), course.HubSpotId)
//	require.NoError(t, err)
//	require.Equal(t, newName, updatedCourse.Name)
//	require.Equal(t, "Learn advanced Go programming", updatedCourse.Description.String)
//}
//
//func TestCreateCourseUniqueNameConstraint(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, _ := courseTestUtils.SetupCourseTestDb(t, dbConn)
//
//	// Create a course
//	name := "Go Course"
//	description := "Learn Go programming"
//	createCourseParams := db.CreateCourseParams{
//		Name:        name,
//		Description: sql.NullString{String: description, Valid: description != ""},
//	}
//
//	_, err := queries.CreateGame(context.Background(), createCourseParams)
//	require.NoError(t, err)
//
//	// Attempt to create another course with the same name
//	_, err = queries.CreateGame(context.Background(), createCourseParams)
//	require.Error(t, err)
//
//	var pgErr *pq.Error
//	require.True(t, errors.As(err, &pgErr))
//	require.Equal(t, "23505", string(pgErr.Code)) // 23505 is the error code for unique violation
//}
//
//func TestGetAllCourses(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, _ := courseTestUtils.SetupCourseTestDb(t, dbConn)
//
//	// Create some courses
//	for i := 1; i <= 5; i++ {
//		createCourseParams := db.CreateCourseParams{
//			Name:        fmt.Sprintf("Course %d", i),
//			Description: sql.NullString{String: fmt.Sprintf("Description %d", i), Valid: true},
//		}
//		_, err := queries.CreateGame(context.Background(), createCourseParams)
//		require.NoError(t, err)
//	}
//
//	params := db.GetCoursesParams{
//		Name:        sql.NullString{String: "", Valid: false},
//		Description: sql.NullString{String: "", Valid: false},
//	}
//
//	// Fetch all courses
//	courses, err := queries.GetGames(context.Background(), params)
//	require.NoError(t, err)
//	require.EqualValues(t, 5, len(courses))
//}
//
//func TestGetCoursesWithFilter(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, _ := courseTestUtils.SetupCourseTestDb(t, dbConn)
//
//	// Create some courses
//	for i := 1; i <= 5; i++ {
//		createCourseParams := db.CreateCourseParams{
//			Name:        fmt.Sprintf("Course %d", i),
//			Description: sql.NullString{String: fmt.Sprintf("Description %d", i), Valid: true},
//		}
//		_, err := queries.CreateGame(context.Background(), createCourseParams)
//		require.NoError(t, err)
//	}
//
//	// Set a filter (e.g., filter courses by name)
//	params := db.GetCoursesParams{
//		Name:        sql.NullString{String: "Course 1", Valid: true}, // Filter for "Course 1"
//		Description: sql.NullString{String: "", Valid: false},        // No filter on description
//	}
//
//	// Fetch courses with filter
//	courses, err := queries.GetGames(context.Background(), params)
//	require.NoError(t, err)
//
//	// Ensure that only the filtered course(s) are returned
//	require.EqualValues(t, 1, len(courses))       // Only 1 course should match the filter ("Course 1")
//	require.Equal(t, "Course 1", courses[0].Name) // Ensure the filtered course is "Course 1"
//}
//
//func TestUpdateNonExistentCourse(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, _ := courseTestUtils.SetupCourseTestDb(t, dbConn)
//
//	// Attempt to update a course that doesn't exist
//	nonExistentId := uuid.New() // Random UUID
//
//	updateParams := db.UpdateCourseParams{
//		HubSpotId:          nonExistentId,
//		Name:        "Updated Course",
//		Description: sql.NullString{String: "Updated course description", Valid: true},
//	}
//
//	affectedRows, err := queries.UpdateGame(context.Background(), updateParams)
//	require.NoError(t, err)
//
//	require.Equal(t, affectedRows, int64(0))
//}
//
//func TestCreateCourseWithNullDescription(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, _ := courseTestUtils.SetupCourseTestDb(t, dbConn)
//
//	// Create a course with a null description
//	createCourseParams := db.CreateCourseParams{
//		Name:        "Go Course",
//		Description: sql.NullString{String: "", Valid: false},
//	}
//
//	course, err := queries.CreateGame(context.Background(), createCourseParams)
//	require.NoError(t, err)
//
//	// Fetch the course and check if description is null
//	require.NoError(t, err)
//	require.False(t, course.Description.Valid) // Should be null
//}
//
//func TestDeleteCourse(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, _ := courseTestUtils.SetupCourseTestDb(t, dbConn)
//
//	// Create a course to delete
//	name := "Go Course"
//	createCourseParams := db.CreateCourseParams{
//		Name:        name,
//		Description: sql.NullString{String: "Learn Go programming", Valid: true},
//	}
//
//	course, err := queries.CreateGame(context.Background(), createCourseParams)
//	require.NoError(t, err)
//
//	// Delete the course
//	impactedRows, err := queries.DeleteGame(context.Background(), course.HubSpotId)
//	require.NoError(t, err)
//
//	require.Equal(t, impactedRows, int64(1))
//
//	// Attempt to fetch the deleted course (expecting error)
//	_, err = queries.GetGameById(context.Background(), course.HubSpotId)
//	require.Error(t, err)
//}
