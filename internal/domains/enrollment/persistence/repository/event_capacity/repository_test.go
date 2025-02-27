package event_capacity

//
//import (
//	"api/internal/domains/course/entity"
//	courseTestUtils "api/internal/domains/course/persistence/test_utils"
//	"api/internal/domains/course/values"
//	"api/utils/test_utils"
//	"context"
//	"fmt"
//	"net/http"
//	"testing"
//
//	"github.com/google/uuid"
//	"github.com/stretchr/testify/require"
//
//	_ "github.com/lib/pq"
//)
//
//func SetupCourseRepo(t *testing.T) *Repository {
//	testDb, _ := test_utils.SetupTestDB(t)
//
//	queries, _ := courseTestUtils.SetupCourseTestDb(t, testDb)
//
//	return NewEnrollmentRepository(queries)
//}
//
//func TestCreateCourse(t *testing.T) {
//
//	repo := SetupCourseRepo(t)
//
//	courseDetails := &values.CourseDetails{
//		Name:        "Go Course",
//		Description: "Learn Go programming",
//	}
//
//	course, err := repo.CreateGame(context.Background(), courseDetails)
//
//	var errToCheck error
//
//	if err != nil {
//
//		errToCheck = fmt.Errorf("%v", err.Message)
//	}
//
//	require.NoError(t, errToCheck)
//	require.NotEqual(t, course.HubSpotId, uuid.Nil)
//	require.Equal(t, courseDetails.Name, course.Name)
//	require.Equal(t, courseDetails.Description, course.Description)
//}
//
//func TestUpdateCourse(t *testing.T) {
//
//	repo := SetupCourseRepo(t)
//
//	// Create a mock createdCourse first
//	courseDetails := &values.CourseDetails{
//		Name:        "Go Course",
//		Description: "Learn Go programming",
//	}
//
//	createdCourse, err := repo.CreateGame(context.Background(), courseDetails)
//
//	var errToCheck error
//
//	if err != nil {
//
//		errToCheck = fmt.Errorf("%v", err.Message)
//	}
//
//	require.NoError(t, errToCheck)
//
//	// Update createdCourse details
//	updatedCourseDetails := &entity.Course{
//		HubSpotId:          createdCourse.HubSpotId,
//		Name:        "Updated Go Course",
//		Description: "Learn advanced Go programming",
//	}
//
//	_, err = repo.UpdateGame(context.Background(), updatedCourseDetails)
//
//	if err != nil {
//
//		errToCheck = fmt.Errorf("%v", err.Message)
//	}
//
//	require.NoError(t, errToCheck)
//
//	courseToBeChecked, err := repo.GetGameById(context.Background(), createdCourse.HubSpotId)
//
//	if err != nil {
//
//		errToCheck = fmt.Errorf("%v", err.Message)
//	}
//
//	require.NoError(t, errToCheck)
//
//	require.EqualValues(t, (*courseToBeChecked).Name, "Updated Go Course")
//	require.EqualValues(t, (*courseToBeChecked).Description, "Learn advanced Go programming")
//
//}
//
//func TestGetCoursesWithFilters(t *testing.T) {
//
//	repo := SetupCourseRepo(t)
//
//	// Create some courses
//	for i := 1; i <= 5; i++ {
//		courseDetails := &values.CourseDetails{
//			Name:        fmt.Sprintf("Course %d", i),
//			Description: fmt.Sprintf("Description %d", i),
//		}
//		_, err := repo.CreateGame(context.Background(), courseDetails)
//
//		var errToCheck error
//
//		if err != nil {
//
//			errToCheck = fmt.Errorf("%v", err.Message)
//		}
//
//		require.NoError(t, errToCheck)
//	}
//
//	name := "Course 1"
//	description := "Description 1"
//
//	// Fetch courses based on the filter
//	courses, err := repo.GetGames(context.Background(), &name, &description)
//
//	var errToCheck error
//
//	if err != nil {
//
//		errToCheck = fmt.Errorf("%v", err.Message)
//	}
//
//	require.NoError(t, errToCheck)
//
//	// Ensure that only the filtered courses are returned
//	require.Len(t, courses, 1)
//	require.Equal(t, "Course 1", courses[0].Name)
//
//	// Check for other filters as well
//	require.Equal(t, "Description 1", courses[0].Description)
//}
//
//func TestCreateCourseWithDuplicateName(t *testing.T) {
//
//	repo := SetupCourseRepo(t)
//
//	// Create the first course
//	courseDetails1 := &values.CourseDetails{
//		Name:        "Go Course",
//		Description: "Learn Go programming",
//	}
//	_, err := repo.CreateGame(context.Background(), courseDetails1)
//
//	var errToCheck error
//
//	if err != nil {
//
//		errToCheck = fmt.Errorf("%v", err.Message)
//	}
//
//	require.NoError(t, errToCheck)
//
//	// Attempt to create a second course with the same name
//	courseDetails2 := &values.CourseDetails{
//		Name:        "Go Course", // Duplicate name
//		Description: "Another Go programming course",
//	}
//	_, err = repo.CreateGame(context.Background(), courseDetails2)
//
//	errToCheck = fmt.Errorf("%v", err.Message)
//
//	// Check that an error is returned due to the duplicate name
//	require.Error(t, errToCheck)
//	require.Equal(t, "Course name already exists", errToCheck.Error())
//	require.Equal(t, http.StatusConflict, err.HTTPCode)
//}
//
//func TestUpdateCourseWithDuplicateName(t *testing.T) {
//
//	repo := SetupCourseRepo(t)
//
//	// Create the first course
//	courseDetails1 := &values.CourseDetails{
//		Name:        "Go Course",
//		Description: "Learn Go programming",
//	}
//	_, err := repo.CreateGame(context.Background(), courseDetails1)
//
//	var errToCheck error
//	if err != nil {
//		errToCheck = fmt.Errorf("%v", err.Message)
//	}
//	require.NoError(t, errToCheck)
//
//	// Create a second course with the same name
//	courseDetails2 := &values.CourseDetails{
//		Name:        "Another Go Course",
//		Description: "Another Go programming course",
//	}
//
//	course2, err := repo.CreateGame(context.Background(), courseDetails2)
//
//	if err != nil {
//		errToCheck = fmt.Errorf("%v", err.Message)
//	}
//
//	require.NoError(t, errToCheck)
//
//	// Attempt to update the second course with the same name as the first one
//	updatedCourse := &entity.Course{
//		HubSpotId:          course2.HubSpotId,  // Same HubSpotId as the first course
//		Name:        "Go Course", // Duplicate name
//		Description: "Updated description",
//	}
//
//	_, err = repo.UpdateGame(context.Background(), updatedCourse)
//
//	if err != nil {
//		errToCheck = fmt.Errorf("%v", err.Message)
//	}
//	// Check for duplicate name error
//	require.Error(t, errToCheck)
//	require.Equal(t, "Course name already exists", err.Message)
//	require.Equal(t, http.StatusConflict, err.HTTPCode)
//}
