package program

import (
	databaseErrors "api/internal/constants"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"api/utils/test_utils"

	"github.com/stretchr/testify/require"

	programTestUtils "api/internal/domains/program/persistence/test_utils"

	db "api/internal/domains/program/persistence/sqlc/generated"
)

func TestCreateProgram(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := programTestUtils.SetupProgramTestDbQueries(t, dbConn)

	defer cleanup()

	name := "Go Course"
	description := "Learn Go programming"

	CreateProgramParams := db.CreateProgramParams{
		Name:        name,
		Description: description,
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}

	err := queries.CreateProgram(context.Background(), CreateProgramParams)

	require.NoError(t, err)

	programs, err := queries.GetPrograms(context.Background())

	require.NoError(t, err)

	practice := programs[0]

	// Assert course data
	require.Equal(t, name, practice.Name)
	require.Equal(t, description, practice.Description)
	require.Equal(t, CreateProgramParams.Name, practice.Name)
}

func TestUpdateProgramValid(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := programTestUtils.SetupProgramTestDbQueries(t, dbConn)

	defer cleanup()

	// Create a course to update
	name := "Go Course"
	description := "Learn Go programming"

	CreateProgramParams := db.CreateProgramParams{
		Name:        name,
		Description: description,
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}

	err := queries.CreateProgram(context.Background(), CreateProgramParams)
	require.NoError(t, err)

	programs, err := queries.GetPrograms(context.Background())

	require.NoError(t, err)

	practice := programs[0]

	// Now, update the course
	newName := "Advanced Go Course"
	updateParams := db.UpdateProgramParams{
		ID:          practice.ID,
		Name:        newName,
		Description: "Learn advanced Go programming",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}

	err = queries.UpdateProgram(context.Background(), updateParams)

	// Get the updated course and verify
	updatedCourse, err := queries.GetProgramById(context.Background(), practice.ID)
	require.NoError(t, err)
	require.Equal(t, newName, updatedCourse.Name)
	require.Equal(t, "Learn advanced Go programming", updatedCourse.Description)
}

func TestUpdatePracticeInvalidLevel(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := programTestUtils.SetupProgramTestDbQueries(t, dbConn)

	defer cleanup()

	// Create a course to update
	name := "Go Course"
	description := "Learn Go programming"

	CreateProgramParams := db.CreateProgramParams{
		Name:        name,
		Description: description,
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}

	err := queries.CreateProgram(context.Background(), CreateProgramParams)
	require.NoError(t, err)

	programs, err := queries.GetPrograms(context.Background())

	require.NoError(t, err)

	practice := programs[0]

	// Now, update the course
	newName := "Advanced Go Course"
	updateParams := db.UpdateProgramParams{
		ID:          practice.ID,
		Name:        newName,
		Description: "Learn advanced Go programming",
	}

	err = queries.UpdateProgram(context.Background(), updateParams)
	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.InvalidTextRepresentation, string(pgErr.Code))

}

func TestCreateProgramUniqueNameConstraint(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := programTestUtils.SetupProgramTestDbQueries(t, dbConn)

	defer cleanup()

	// Create a course
	name := "Go Course"
	description := "Learn Go programming"
	createCourseParams := db.CreateProgramParams{
		Name:        name,
		Description: description,
		Level:       db.ProgramProgramLevelAdvanced,
		Type:        db.ProgramProgramTypeCourse,
	}

	err := queries.CreateProgram(context.Background(), createCourseParams)
	require.NoError(t, err)

	// Attempt to create another course with the same name
	err = queries.CreateProgram(context.Background(), createCourseParams)
	require.Error(t, err)

	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.UniqueViolation, string(pgErr.Code))
}

func TestGetAllPrograms(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := programTestUtils.SetupProgramTestDbQueries(t, dbConn)

	defer cleanup()

	// Create some courses
	for i := 1; i <= 5; i++ {
		createCourseParams := db.CreateProgramParams{
			Name:        fmt.Sprintf("Course %d", i),
			Description: fmt.Sprintf("Description %d", i),
			Level:       db.ProgramProgramLevelAll,
			Type:        db.ProgramProgramTypeCourse,
		}
		err := queries.CreateProgram(context.Background(), createCourseParams)
		require.NoError(t, err)
	}

	// Fetch all courses
	courses, err := queries.GetPrograms(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 5, len(courses))
}

func TestUpdateNonExistentProgram(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := programTestUtils.SetupProgramTestDbQueries(t, dbConn)

	defer cleanup()

	// Attempt to update a practice that doesn't exist
	nonExistentId := uuid.New() // Random UUID

	updateParams := db.UpdateProgramParams{
		ID:          nonExistentId,
		Name:        "Updated Practice",
		Description: "Updated practice description",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeGame,
	}

	err := queries.UpdateProgram(context.Background(), updateParams)
	require.Nil(t, err)
}

func TestCreateCourseWithWrongLevel(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := programTestUtils.SetupProgramTestDbQueries(t, dbConn)

	defer cleanup()

	// Create a course with a null description
	CreateProgramParams := db.CreateProgramParams{
		Name:        "Go Course",
		Description: "wefwefew",
		Level:       "jhwwf",
		Type:        db.ProgramProgramTypeCourse,
	}

	err := queries.CreateProgram(context.Background(), CreateProgramParams)

	require.Error(t, err)

	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.InvalidTextRepresentation, string(pgErr.Code))
}

func TestDeleteProgram(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := programTestUtils.SetupProgramTestDbQueries(t, dbConn)

	defer cleanup()

	// Create a course to delete
	name := "Go Course"
	CreateProgramParams := db.CreateProgramParams{
		Name:        name,
		Description: "Learn Go programming",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}

	err := queries.CreateProgram(context.Background(), CreateProgramParams)
	require.NoError(t, err)

	programs, err := queries.GetPrograms(context.Background())

	require.NoError(t, err)

	createdPractice := programs[0]

	// Delete the course
	impactedRows, err := queries.DeleteProgram(context.Background(), createdPractice.ID)
	require.NoError(t, err)

	require.Equal(t, impactedRows, int64(1))

	// Attempt to fetch the deleted course (expecting error)
	_, err = queries.GetProgramById(context.Background(), createdPractice.ID)

	require.Error(t, err)

	require.Equal(t, sql.ErrNoRows, err)
}
