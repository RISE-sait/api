package game

//
//import (
//	courseTestUtils "api/internal/domains/course/persistence/test_utils"
//	"api/utils/test_utils"
//	"context"
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
//func TestCreateCourse(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, _ := courseTestUtils.SetupCourseTestDb(t, dbConn)
//
//	name := "Go Course"
//	description := "Learn Go programming"
//
//	createCourseParams := db.CreateCourseParams{
//		Name:        name,
//		Description: description,
//		Capacity:    50,
//	}
//
//	course, err := queries.CreateCourse(context.Background(), createCourseParams)
//
//	require.NoError(t, err)
//
//	// Assert course data
//	require.Equal(t, name, course.Name)
//	require.Equal(t, description, course.Description)
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
//		Description: description,
//	}
//
//	course, err := queries.CreateCourse(context.Background(), createCourseParams)
//	require.NoError(t, err)
//
//	// Now, update the course
//	newName := "Advanced Go Course"
//	updateParams := db.UpdateCourseParams{
//		ID:          course.ID,
//		Name:        newName,
//		Description: "Learn advanced Go programming",
//	}
//
//	_, err = queries.UpdateCourse(context.Background(), updateParams)
//	require.NoError(t, err)
//
//	// Get the updated course and verify
//	updatedCourse, err := queries.GetCourseById(context.Background(), course.ID)
//	require.NoError(t, err)
//	require.Equal(t, newName, updatedCourse.Name)
//	require.Equal(t, "Learn advanced Go programming", updatedCourse.Description)
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
//		Description: description,
//	}
//
//	_, err := queries.CreateCourse(context.Background(), createCourseParams)
//	require.NoError(t, err)
//
//	// Attempt to create another course with the same name
//	_, err = queries.CreateCourse(context.Background(), createCourseParams)
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
//			Description: fmt.Sprintf("Description %d", i),
//		}
//		_, err := queries.CreateCourse(context.Background(), createCourseParams)
//		require.NoError(t, err)
//	}
//
//	// Fetch all courses
//	courses, err := queries.GetCourses(context.Background())
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
//			Description: fmt.Sprintf("Description %d", i),
//		}
//		_, err := queries.CreateCourse(context.Background(), createCourseParams)
//		require.NoError(t, err)
//	}
//
//	// Fetch courses with filter
//	courses, err := queries.GetCourses(context.Background())
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
//		ID:          nonExistentId,
//		Name:        "Updated Course",
//		Description: "Updated course description",
//	}
//
//	row, err := queries.UpdateCourse(context.Background(), updateParams)
//
//	require.Equal(t, int64(0), row)
//
//	require.Nil(t, err)
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
//		Description: "",
//	}
//
//	_, err := queries.CreateCourse(context.Background(), createCourseParams)
//	require.NoError(t, err)
//
//	// Fetch the course and check if description is null
//	require.NoError(t, err)
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
//		Description: "Learn Go programming",
//	}
//
//	course, err := queries.CreateCourse(context.Background(), createCourseParams)
//	require.NoError(t, err)
//
//	// Delete the course
//	impactedRows, err := queries.DeleteCourse(context.Background(), course.ID)
//	require.NoError(t, err)
//
//	require.Equal(t, impactedRows, int64(1))
//
//	// Attempt to fetch the deleted course (expecting error)
//	_, err = queries.GetCourseById(context.Background(), course.ID)
//	require.Error(t, err)
//}
