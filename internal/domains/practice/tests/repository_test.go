package course

// import (
// 	test_utils "api/internal/domains/course/test_utils"
// 	"api/internal/domains/course/values"
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/require"

// 	_ "github.com/lib/pq"
// )

// // Example test functions

// func TestCreateCourse(t *testing.T) {
// 	repo, cleanup := test_utils.SetupTestRepository(t)
// 	defer cleanup()

// 	courseDetails := &values.CourseDetails{
// 		Name:        "Go Course",
// 		Description: "Learn Go programming",
// 		StartDate:   time.Now().Truncate(time.Second).UTC(),
// 		EndDate:     time.Now().AddDate(0, 0, 7).Truncate(time.Second).UTC(),
// 		Capacity:    30,
// 	}

// 	course, err := repo.CreateCourse(context.Background(), courseDetails)

// 	var errToCheck error

// 	if err != nil {

// 		errToCheck = fmt.Errorf("%v", err.Message)
// 	}

// 	actualStartDate := course.StartDate.Truncate(time.Second).UTC()
// 	actualEndDate := course.EndDate.Truncate(time.Second).UTC()

// 	require.NoError(t, errToCheck)
// 	require.NotEqual(t, course.HubSpotId, uuid.Nil)
// 	require.Equal(t, courseDetails.Name, course.Name)
// 	require.Equal(t, courseDetails.Description, course.Description)
// 	require.Equal(t, courseDetails.Capacity, course.Capacity)
// 	require.Equal(t, courseDetails.StartDate, actualStartDate)
// 	require.Equal(t, courseDetails.EndDate, actualEndDate)
// }

// func TestUpdateCourse(t *testing.T) {
// 	repo, cleanup := test_utils.SetupTestRepository(t)
// 	defer cleanup()

// 	// Create a mock course first
// 	courseDetails := &values.CourseDetails{
// 		Name:        "Go Course",
// 		Description: "Learn Go programming",
// 		StartDate:   time.Now().Truncate(time.Second).UTC(),
// 		EndDate:     time.Now().AddDate(0, 0, 7).Truncate(time.Second).UTC(),
// 		Capacity:    30,
// 	}

// 	course, err := repo.CreateCourse(context.Background(), courseDetails)

// 	var errToCheck error

// 	if err != nil {

// 		errToCheck = fmt.Errorf("%v", err.Message)
// 	}

// 	require.NoError(t, errToCheck)

// 	// Update course details
// 	updatedCourseDetails := &values.CourseAllFields{
// 		HubSpotId: course.HubSpotId,
// 		CourseDetails: values.CourseDetails{
// 			Name:        "Updated Go Course",
// 			Description: "Learn advanced Go programming",
// 			StartDate:   time.Now().Truncate(time.Second).UTC(),
// 			EndDate:     time.Now().AddDate(0, 0, 7).Truncate(time.Second).UTC(),
// 			Capacity:    30,
// 		},
// 	}
// 	err = repo.UpdateCourse(context.Background(), updatedCourseDetails)

// 	if err != nil {

// 		errToCheck = fmt.Errorf("%v", err.Message)
// 	}

// 	require.NoError(t, errToCheck)

// 	course, err = repo.GetCourseById(context.Background(), course.HubSpotId)

// 	if err != nil {

// 		errToCheck = fmt.Errorf("%v", err.Message)
// 	}

// 	require.NoError(t, errToCheck)

// 	actualStartDate := course.StartDate.Truncate(time.Second).UTC()
// 	actualEndDate := course.EndDate.Truncate(time.Second).UTC()

// 	require.EqualValues(t, (*course).CourseDetails.Name, "Updated Go Course")
// 	require.EqualValues(t, (*course).CourseDetails.Description, "Learn advanced Go programming")
// 	require.EqualValues(t, actualStartDate, updatedCourseDetails.CourseDetails.StartDate)
// 	require.EqualValues(t, actualEndDate, updatedCourseDetails.CourseDetails.EndDate)

// 	require.EqualValues(t, updatedCourseDetails.CourseDetails.Capacity, (*course).CourseDetails.Capacity)
// }

// func TestGetCoursesWithFilters(t *testing.T) {
// 	repo, cleanup := test_utils.SetupTestRepository(t)
// 	defer cleanup()

// 	now := time.Now()

// 	// Create some courses
// 	for i := 1; i <= 5; i++ {
// 		courseDetails := &values.CourseDetails{
// 			Name:        fmt.Sprintf("Course %d", i),
// 			Description: fmt.Sprintf("Description %d", i),
// 			StartDate:   time.Now(),
// 			EndDate:     time.Now().AddDate(0, 0, 7),
// 			Capacity:    100,
// 		}
// 		_, err := repo.CreateCourse(context.Background(), courseDetails)

// 		var errToCheck error

// 		if err != nil {

// 			errToCheck = fmt.Errorf("%v", err.Message)
// 		}

// 		require.NoError(t, errToCheck)
// 	}

// 	name := "Course 1"
// 	description := "Description 1"

// 	// Fetch courses based on the filter
// 	courses, err := repo.GetCourses(context.Background(), &name, &description)

// 	var errToCheck error

// 	if err != nil {

// 		errToCheck = fmt.Errorf("%v", err.Message)
// 	}

// 	require.NoError(t, errToCheck)

// 	// Ensure that only the filtered courses are returned
// 	require.Len(t, courses, 1)
// 	require.Equal(t, "Course 1", courses[0].Name)

// 	// Check for other filters as well
// 	require.Equal(t, "Description 1", courses[0].Description)
// 	require.Equal(t, int32(100), courses[0].Capacity)
// 	require.True(t, courses[0].StartDate.After(now))
// 	require.True(t, courses[0].EndDate.After(courses[0].StartDate))
// }

// func TestCreateCourseWithDuplicateName(t *testing.T) {
// 	repo, cleanup := test_utils.SetupTestRepository(t)
// 	defer cleanup()

// 	// Create the first course
// 	courseDetails1 := &values.CourseDetails{
// 		Name:        "Go Course",
// 		Description: "Learn Go programming",
// 		StartDate:   time.Now(),
// 		EndDate:     time.Now().AddDate(0, 0, 7),
// 		Capacity:    30,
// 	}
// 	_, err := repo.CreateCourse(context.Background(), courseDetails1)

// 	var errToCheck error

// 	if err != nil {

// 		errToCheck = fmt.Errorf("%v", err.Message)
// 	}

// 	require.NoError(t, errToCheck)

// 	// Attempt to create a second course with the same name
// 	courseDetails2 := &values.CourseDetails{
// 		Name:        "Go Course", // Duplicate name
// 		Description: "Another Go programming course",
// 		StartDate:   time.Now(),
// 		EndDate:     time.Now().AddDate(0, 0, 7),
// 		Capacity:    30,
// 	}
// 	_, err = repo.CreateCourse(context.Background(), courseDetails2)

// 	errToCheck = fmt.Errorf("%v", err.Message)

// 	// Check that an error is returned due to the duplicate name
// 	require.Error(t, errToCheck)
// 	require.Equal(t, "Course name already exists", errToCheck.Error())
// 	require.Equal(t, http.StatusConflict, err.HTTPCode)
// }

// func TestUpdateCourseWithDuplicateName(t *testing.T) {
// 	repo, cleanup := test_utils.SetupTestRepository(t)
// 	defer cleanup()

// 	// Create the first course
// 	courseDetails1 := &values.CourseDetails{
// 		Name:        "Go Course",
// 		Description: "Learn Go programming",
// 		StartDate:   time.Now(),
// 		EndDate:     time.Now().AddDate(0, 0, 7),
// 		Capacity:    30,
// 	}
// 	_, err := repo.CreateCourse(context.Background(), courseDetails1)

// 	var errToCheck error
// 	if err != nil {
// 		errToCheck = fmt.Errorf("%v", err.Message)
// 	}
// 	require.NoError(t, errToCheck)

// 	// Create a second course with the same name
// 	courseDetails2 := &values.CourseDetails{
// 		Name:        "Another Go Course",
// 		Description: "Another Go programming course",
// 		StartDate:   time.Now(),
// 		EndDate:     time.Now().AddDate(0, 0, 7),
// 		Capacity:    30,
// 	}

// 	course2, err := repo.CreateCourse(context.Background(), courseDetails2)

// 	if err != nil {
// 		errToCheck = fmt.Errorf("%v", err.Message)
// 	}

// 	require.NoError(t, errToCheck)

// 	// Attempt to update the second course with the same name as the first one
// 	updatedCourse := &values.CourseAllFields{
// 		HubSpotId: course2.HubSpotId, // Same HubSpotId as the first course
// 		CourseDetails: values.CourseDetails{
// 			Name:        "Go Course", // Duplicate name
// 			Description: "Updated description",
// 			StartDate:   time.Now(),
// 			EndDate:     time.Now().AddDate(0, 0, 7),
// 			Capacity:    30,
// 		},
// 	}
// 	err = repo.UpdateCourse(context.Background(), updatedCourse)

// 	if err != nil {
// 		errToCheck = fmt.Errorf("%v", err.Message)
// 	}
// 	// Check for duplicate name error
// 	require.Error(t, errToCheck)
// 	require.Equal(t, "Course name already exists", err.Message)
// 	require.Equal(t, http.StatusConflict, err.HTTPCode)
// }
