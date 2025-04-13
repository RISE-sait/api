package program_test

import (
	databaseErrors "api/internal/constants"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/stretchr/testify/require"

	dbTestUtils "api/utils/test_utils"

	db "api/internal/domains/program/persistence/sqlc/generated"
)

func TestCreateProgram(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	queries := db.New(dbConn)

	defer cleanup()

	CreateProgramParams := db.CreateProgramParams{
		Name:        "Go Course",
		Description: "Learn Go programming",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}

	createdProgram, err := queries.CreateProgram(context.Background(), CreateProgramParams)

	require.NoError(t, err)

	// Assert course data
	require.Equal(t, CreateProgramParams.Name, createdProgram.Name)
	require.Equal(t, CreateProgramParams.Description, createdProgram.Description)
	require.Equal(t, CreateProgramParams.Level, createdProgram.Level)
	require.Equal(t, CreateProgramParams.Type, createdProgram.Type)
}

func TestUpdateProgramValid(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	queries := db.New(dbConn)

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

	createdProgram, err := queries.CreateProgram(context.Background(), CreateProgramParams)
	require.NoError(t, err)

	// Now, update the course
	newName := "Advanced Go Course"
	updateParams := db.UpdateProgramParams{
		ID:          createdProgram.ID,
		Name:        newName,
		Description: "Learn advanced Go programming",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}

	updatedProgram, err := queries.UpdateProgram(context.Background(), updateParams)

	require.NoError(t, err)
	require.Equal(t, newName, updatedProgram.Name)
	require.Equal(t, "Learn advanced Go programming", updatedProgram.Description)
	require.Equal(t, db.ProgramProgramLevelAll, updatedProgram.Level)
	require.Equal(t, db.ProgramProgramTypeCourse, updatedProgram.Type)
}

func TestUpdatePracticeInvalidLevel(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	queries := db.New(dbConn)

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

	createdProgram, err := queries.CreateProgram(context.Background(), CreateProgramParams)
	require.NoError(t, err)

	// Now, update the course
	newName := "Advanced Go Course"
	updateParams := db.UpdateProgramParams{
		ID:          createdProgram.ID,
		Name:        newName,
		Description: "Learn advanced Go programming",
	}

	_, err = queries.UpdateProgram(context.Background(), updateParams)
	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.InvalidTextRepresentation, string(pgErr.Code))

}

func TestCreateProgramUniqueNameConstraint(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	queries := db.New(dbConn)

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

	_, err := queries.CreateProgram(context.Background(), createCourseParams)
	require.NoError(t, err)

	// Attempt to create another course with the same name
	_, err = queries.CreateProgram(context.Background(), createCourseParams)
	require.Error(t, err)

	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.UniqueViolation, string(pgErr.Code))
}

func TestGetAllPrograms(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	queries := db.New(dbConn)

	defer cleanup()

	// Create some courses
	for i := 1; i <= 5; i++ {
		createCourseParams := db.CreateProgramParams{
			Name:        fmt.Sprintf("Course %d", i),
			Description: fmt.Sprintf("Description %d", i),
			Level:       db.ProgramProgramLevelAll,
			Type:        db.ProgramProgramTypeCourse,
		}
		createdProgram, err := queries.CreateProgram(context.Background(), createCourseParams)
		require.NoError(t, err)

		require.Equal(t, createCourseParams.Name, createdProgram.Name)
		require.Equal(t, createCourseParams.Description, createdProgram.Description)
		require.Equal(t, createCourseParams.Level, createdProgram.Level)
		require.Equal(t, createCourseParams.Type, createdProgram.Type)
	}

	// Fetch all courses
	courses, err := queries.GetPrograms(context.Background(), db.NullProgramProgramType{
		ProgramProgramType: db.ProgramProgramTypeCourse,
		Valid:              true,
	})
	require.NoError(t, err)
	require.EqualValues(t, 5, len(courses))
}

func TestGetNotExistingProgram(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	queries := db.New(dbConn)

	defer cleanup()

	// Fetch all courses
	_, err := queries.GetProgramById(context.Background(), uuid.New())
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}

func TestUpdateNonExistentProgram(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	queries := db.New(dbConn)

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

	_, err := queries.UpdateProgram(context.Background(), updateParams)
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}

func TestCreateCourseWithWrongLevel(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	queries := db.New(dbConn)

	defer cleanup()

	// Create a course with a null description
	CreateProgramParams := db.CreateProgramParams{
		Name:        "Go Course",
		Description: "wefwefew",
		Level:       "jhwwf",
		Type:        db.ProgramProgramTypeCourse,
	}

	_, err := queries.CreateProgram(context.Background(), CreateProgramParams)

	require.Error(t, err)

	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.InvalidTextRepresentation, string(pgErr.Code))
}

func TestDeleteProgram(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	queries := db.New(dbConn)

	defer cleanup()

	// Create a course to delete
	name := "Go Course"
	CreateProgramParams := db.CreateProgramParams{
		Name:        name,
		Description: "Learn Go programming",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}

	createdProgram, err := queries.CreateProgram(context.Background(), CreateProgramParams)
	require.NoError(t, err)

	// Delete the course
	impactedRows, err := queries.DeleteProgram(context.Background(), createdProgram.ID)
	require.NoError(t, err)

	require.Equal(t, impactedRows, int64(1))

	// Attempt to fetch the deleted course (expecting error)
	_, err = queries.GetProgramById(context.Background(), createdProgram.ID)

	require.Error(t, err)

	require.Equal(t, sql.ErrNoRows, err)
}
