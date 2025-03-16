package course

import (
	databaseErrors "api/internal/constants"
	courseTestUtils "api/internal/domains/course/persistence/test_utils"
	"api/utils/test_utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"testing"

	"github.com/stretchr/testify/require"

	db "api/internal/domains/course/persistence/sqlc/generated"
)

func TestCreateCourse(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := courseTestUtils.SetupCourseTestDb(t, dbConn)

	defer cleanup()

	name := "Go Course"
	description := "Learn Go programming"

	createCourseParams := db.CreateCourseParams{
		Name:        name,
		Description: description,
		Capacity:    50,
	}

	course, err := queries.CreateCourse(context.Background(), createCourseParams)

	require.NoError(t, err)

	// Assert course data
	require.Equal(t, name, course.Name)
	require.Equal(t, description, course.Description)
}

func TestUpdateCourse(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := courseTestUtils.SetupCourseTestDb(t, dbConn)

	defer cleanup()

	// Create a course to update
	name := "Go Course"
	description := "Learn Go programming"
	createCourseParams := db.CreateCourseParams{
		Name:        name,
		Description: description,
	}

	course, err := queries.CreateCourse(context.Background(), createCourseParams)
	require.NoError(t, err)

	// Now, update the course
	newName := "Advanced Go Course"
	updateParams := db.UpdateCourseParams{
		ID:          course.ID,
		Name:        newName,
		Description: "Learn advanced Go programming",
	}

	_, err = queries.UpdateCourse(context.Background(), updateParams)
	require.NoError(t, err)

	// Get the updated course and verify
	updatedCourse, err := queries.GetCourseById(context.Background(), course.ID)
	require.NoError(t, err)
	require.Equal(t, newName, updatedCourse.Name)
	require.Equal(t, "Learn advanced Go programming", updatedCourse.Description)
}

func TestCreateCourseUniqueNameConstraint(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := courseTestUtils.SetupCourseTestDb(t, dbConn)

	defer cleanup()

	// Create a course
	name := "Go Course"
	description := "Learn Go programming"
	createCourseParams := db.CreateCourseParams{
		Name:        name,
		Description: description,
	}

	_, err := queries.CreateCourse(context.Background(), createCourseParams)
	require.NoError(t, err)

	// Attempt to create another course with the same name
	_, err = queries.CreateCourse(context.Background(), createCourseParams)
	require.Error(t, err)

	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.UniqueViolation, string(pgErr.Code)) // 23505 is the error code for unique violation
}

func TestGetAllCourses(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := courseTestUtils.SetupCourseTestDb(t, dbConn)

	defer cleanup()

	// Create some courses
	for i := 1; i <= 5; i++ {
		createCourseParams := db.CreateCourseParams{
			Name:        fmt.Sprintf("Course %d", i),
			Description: fmt.Sprintf("Description %d", i),
		}
		_, err := queries.CreateCourse(context.Background(), createCourseParams)
		require.NoError(t, err)
	}

	// Fetch all courses
	courses, err := queries.GetCourses(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 5, len(courses))
}

func TestGetNonExistingCourse(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := courseTestUtils.SetupCourseTestDb(t, dbConn)

	defer cleanup()

	_, err := queries.GetCourseById(context.Background(), uuid.Nil)

	require.Equal(t, sql.ErrNoRows, err)
}

func TestUpdateNonExistentCourse(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := courseTestUtils.SetupCourseTestDb(t, dbConn)

	defer cleanup()

	// Attempt to update a course that doesn't exist
	nonExistentId := uuid.New() // Random UUID

	updateParams := db.UpdateCourseParams{
		ID:          nonExistentId,
		Name:        "Updated Course",
		Description: "Updated course description",
	}

	rows, err := queries.UpdateCourse(context.Background(), updateParams)

	require.Equal(t, int64(0), rows)

	require.Nil(t, err)
}

func TestUpdateCourseWithSameValues(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := courseTestUtils.SetupCourseTestDb(t, dbConn)

	defer cleanup()

	name := "Go Course"
	description := "Learn Go programming"

	createCourseParams := db.CreateCourseParams{
		Name:        name,
		Description: description,
	}

	createdCourse, err := queries.CreateCourse(context.Background(), createCourseParams)
	require.NoError(t, err)

	updateParams := db.UpdateCourseParams{
		ID:          createdCourse.ID,
		Name:        name,
		Description: description,
	}

	impactedRows, err := queries.UpdateCourse(context.Background(), updateParams)

	require.Equal(t, int64(1), impactedRows)

	require.Nil(t, err)
}

func TestUpdateCourseWithDuplicateUniqueValues(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := courseTestUtils.SetupCourseTestDb(t, dbConn)

	defer cleanup()

	name := "Go Course"
	description := "Learn Go programming"

	createCourseParams := db.CreateCourseParams{
		Name:        name,
		Description: description,
	}

	createdCourse, err := queries.CreateCourse(context.Background(), createCourseParams)

	require.NoError(t, err)

	require.Equal(t, name, createdCourse.Name)
	require.Equal(t, description, createdCourse.Description)

	createCourseParams2 := db.CreateCourseParams{
		Name:        "Go Course 2",
		Description: description,
	}

	createdCourse2, err := queries.CreateCourse(context.Background(), createCourseParams2)
	require.NoError(t, err)

	require.Equal(t, "Go Course 2", createdCourse2.Name)
	require.Equal(t, description, createdCourse.Description)

	updateCourseParams := db.UpdateCourseParams{
		ID:          createdCourse2.ID,
		Name:        name,
		Description: description,
	}

	impactedRows, err := queries.UpdateCourse(context.Background(), updateCourseParams)

	require.Equal(t, 0, int(impactedRows))
	require.Equal(t, string(err.(*pq.Error).Code), databaseErrors.UniqueViolation)

}

func TestCreateCourseWithEmptyDescription(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := courseTestUtils.SetupCourseTestDb(t, dbConn)

	defer cleanup()

	// Create a course with an empty description
	createCourseParams := db.CreateCourseParams{
		Name:        "Go Course",
		Description: "",
	}

	createdCourse, err := queries.CreateCourse(context.Background(), createCourseParams)
	require.NoError(t, err)

	require.Equal(t, "Go Course", createdCourse.Name)
	require.Equal(t, "", createdCourse.Description)
	require.Equal(t, 0, int(createdCourse.Capacity))
}

func TestDeleteCourse(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := courseTestUtils.SetupCourseTestDb(t, dbConn)

	defer cleanup()

	// Create a course to delete
	createCourseParams := db.CreateCourseParams{
		Name:        "Go Course",
		Description: "Learn Go programming",
	}

	course, err := queries.CreateCourse(context.Background(), createCourseParams)
	require.NoError(t, err)

	// Delete the course
	impactedRows, err := queries.DeleteCourse(context.Background(), course.ID)
	require.NoError(t, err)

	require.Equal(t, impactedRows, int64(1))

	// Attempt to fetch the deleted course (expecting error)
	_, err = queries.GetCourseById(context.Background(), course.ID)
	require.Error(t, err)
}
