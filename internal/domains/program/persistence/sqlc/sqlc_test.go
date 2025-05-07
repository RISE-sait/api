package program_test

// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"fmt"
// 	"testing"

// 	"github.com/google/uuid"
// 	"github.com/lib/pq"
// 	"github.com/stretchr/testify/require"

// 	databaseErrors "api/internal/constants"
// 	db "api/internal/domains/program/persistence/sqlc/generated"
// 	dbTestUtils "api/utils/test_utils"
// )

// func TestCreateProgram(t *testing.T) {
// 	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
// 	defer cleanup()

// 	queries := db.New(dbConn)

// 	params := db.CreateProgramParams{
// 		Name:        "Go Course",
// 		Description: "Learn Go programming",
// 		Level:       db.ProgramProgramLevelAll,
// 		Type:        db.ProgramProgramTypeCourse,
// 	}

// 	created, err := queries.CreateProgram(context.Background(), params)
// 	require.NoError(t, err)

// 	require.Equal(t, params.Name, created.Name)
// 	require.Equal(t, params.Description, created.Description)
// 	require.Equal(t, params.Level, created.Level)
// 	require.Equal(t, params.Type, created.Type)
// }

// func TestUpdateProgramValid(t *testing.T) {
// 	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
// 	defer cleanup()

// 	queries := db.New(dbConn)

// 	original := db.CreateProgramParams{
// 		Name:        "Go Course",
// 		Description: "Learn Go programming",
// 		Level:       db.ProgramProgramLevelAll,
// 		Type:        db.ProgramProgramTypeCourse,
// 	}
// 	created, err := queries.CreateProgram(context.Background(), original)
// 	require.NoError(t, err)

// 	updateParams := db.UpdateProgramParams{
// 		ID:          created.ID,
// 		Name:        "Advanced Go Course",
// 		Description: "Learn advanced Go programming",
// 		Level:       db.ProgramProgramLevelAll,
// 		Type:        db.ProgramProgramTypeCourse,
// 	}

// 	updated, err := queries.UpdateProgram(context.Background(), updateParams)
// 	require.NoError(t, err)

// 	require.Equal(t, updateParams.Name, updated.Name)
// 	require.Equal(t, updateParams.Description, updated.Description)
// 	require.Equal(t, updateParams.Level, updated.Level)
// 	require.Equal(t, updateParams.Type, updated.Type)
// }

// func TestUpdatePracticeInvalidLevel(t *testing.T) {
// 	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
// 	defer cleanup()

// 	queries := db.New(dbConn)

// 	created, err := queries.CreateProgram(context.Background(), db.CreateProgramParams{
// 		Name:        "Test",
// 		Description: "Desc",
// 		Level:       db.ProgramProgramLevelAll,
// 		Type:        db.ProgramProgramTypeCourse,
// 	})
// 	require.NoError(t, err)

// 	_, err = dbConn.ExecContext(context.Background(), `UPDATE program.programs SET level = 'INVALID' WHERE id = $1`, created.ID)
// 	require.Error(t, err)

// 	var pgErr *pq.Error
// 	require.True(t, errors.As(err, &pgErr))
// 	require.Equal(t, databaseErrors.InvalidTextRepresentation, string(pgErr.Code))
// }

// func TestCreateProgramUniqueNameConstraint(t *testing.T) {
// 	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
// 	defer cleanup()

// 	queries := db.New(dbConn)

// 	params := db.CreateProgramParams{
// 		Name:        "Go Course",
// 		Description: "Learn Go",
// 		Level:       db.ProgramProgramLevelAll,
// 		Type:        db.ProgramProgramTypeCourse,
// 	}

// 	_, err := queries.CreateProgram(context.Background(), params)
// 	require.NoError(t, err)

// 	_, err = queries.CreateProgram(context.Background(), params)
// 	require.Error(t, err)

// 	var pgErr *pq.Error
// 	require.True(t, errors.As(err, &pgErr))
// 	require.Equal(t, databaseErrors.UniqueViolation, string(pgErr.Code))
// }

// func TestGetAllPrograms(t *testing.T) {
// 	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
// 	defer cleanup()

// 	queries := db.New(dbConn)

// 	for i := 1; i <= 5; i++ {
// 		params := db.CreateProgramParams{
// 			Name:        fmt.Sprintf("Course %d", i),
// 			Description: fmt.Sprintf("Description %d", i),
// 			Level:       db.ProgramProgramLevelAll,
// 			Type:        db.ProgramProgramTypeCourse,
// 		}
// 		created, err := queries.CreateProgram(context.Background(), params)
// 		require.NoError(t, err)

// 		require.Equal(t, params.Name, created.Name)
// 		require.Equal(t, params.Description, created.Description)
// 		require.Equal(t, params.Level, created.Level)
// 		require.Equal(t, params.Type, created.Type)
// 	}

// 	courses, err := queries.GetPrograms(context.Background(), db.NullProgramProgramType{
// 		ProgramProgramType: db.ProgramProgramTypeCourse,
// 		Valid:              true,
// 	})
// 	require.NoError(t, err)
// 	require.EqualValues(t, 5, len(courses))
// }

// func TestGetNotExistingProgram(t *testing.T) {
// 	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
// 	defer cleanup()

// 	queries := db.New(dbConn)

// 	_, err := queries.GetProgramById(context.Background(), uuid.New())
// 	require.Error(t, err)
// 	require.Equal(t, sql.ErrNoRows, err)
// }

// func TestUpdateNonExistentProgram(t *testing.T) {
// 	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
// 	defer cleanup()

// 	queries := db.New(dbConn)

// 	update := db.UpdateProgramParams{
// 		ID:          uuid.New(),
// 		Name:        "Nonexistent",
// 		Description: "Update",
// 		Level:       db.ProgramProgramLevelAll,
// 		Type:        db.ProgramProgramTypeGame,
// 	}
// 	_, err := queries.UpdateProgram(context.Background(), update)
// 	require.Error(t, err)
// 	require.Equal(t, sql.ErrNoRows, err)
// }

// func TestCreateCourseWithWrongLevel(t *testing.T) {
// 	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
// 	defer cleanup()

// 	queries := db.New(dbConn)

// 	params := db.CreateProgramParams{
// 		Name:        "Go Course",
// 		Description: "Invalid",
// 		Level:       "not_a_valid_enum",
// 		Type:        db.ProgramProgramTypeCourse,
// 	}

// 	_, err := queries.CreateProgram(context.Background(), params)
// 	require.Error(t, err)

// 	var pgErr *pq.Error
// 	require.True(t, errors.As(err, &pgErr))
// 	require.Equal(t, databaseErrors.InvalidTextRepresentation, string(pgErr.Code))
// }

// func TestDeleteProgram(t *testing.T) {
// 	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
// 	defer cleanup()

// 	queries := db.New(dbConn)

// 	params := db.CreateProgramParams{
// 		Name:        "ToDelete",
// 		Description: "Desc",
// 		Level:       db.ProgramProgramLevelAll,
// 		Type:        db.ProgramProgramTypeCourse,
// 	}
// 	created, err := queries.CreateProgram(context.Background(), params)
// 	require.NoError(t, err)

// 	rows, err := queries.DeleteProgram(context.Background(), created.ID)
// 	require.NoError(t, err)
// 	require.Equal(t, int64(1), rows)

// 	_, err = queries.GetProgramById(context.Background(), created.ID)
// 	require.Error(t, err)
// 	require.Equal(t, sql.ErrNoRows, err)
// }
