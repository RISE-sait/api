package program_test

import (
	databaseErrors "api/internal/constants"
	context "context"
	sql "database/sql"
	errors "errors"
	fmt "fmt"
	testing "testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	dbTestUtils "api/utils/test_utils"

	db "api/internal/domains/program/persistence/sqlc/generated"
)

func truncateProgramsTable(dbConn *sql.DB) {
	_, err := dbConn.ExecContext(context.Background(), `TRUNCATE TABLE program.programs RESTART IDENTITY CASCADE`)
	if err != nil {
		panic(fmt.Sprintf("failed to truncate programs table: %v", err))
	}
}

func TestCreateProgram(t *testing.T) {
	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()
	truncateProgramsTable(dbConn)

	queries := db.New(dbConn)

	params := db.CreateProgramParams{
		Name:        "Go Course",
		Description: "Learn Go programming",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}

	created, err := queries.CreateProgram(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, params.Name, created.Name)
	require.Equal(t, params.Description, created.Description)
	require.Equal(t, params.Level, created.Level)
	require.Equal(t, params.Type, created.Type)
}

func TestUpdateProgramValid(t *testing.T) {
	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()
	truncateProgramsTable(dbConn)

	queries := db.New(dbConn)

	original := db.CreateProgramParams{
		Name:        "Go Course",
		Description: "Learn Go programming",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}
	created, err := queries.CreateProgram(context.Background(), original)
	require.NoError(t, err)

	update := db.UpdateProgramParams{
		ID:          created.ID,
		Name:        "Advanced Go Course",
		Description: "Learn advanced Go programming",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}
	updated, err := queries.UpdateProgram(context.Background(), update)
	require.NoError(t, err)
	require.Equal(t, update.Name, updated.Name)
	require.Equal(t, update.Description, updated.Description)
}

func TestUpdatePracticeInvalidLevel(t *testing.T) {
	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()
	truncateProgramsTable(dbConn)

	queries := db.New(dbConn)

	created, err := queries.CreateProgram(context.Background(), db.CreateProgramParams{
		Name:        "Test",
		Description: "Desc",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	})
	require.NoError(t, err)

	_, err = dbConn.ExecContext(context.Background(), `UPDATE program.programs SET level = 'INVALID' WHERE id = $1`, created.ID)
	require.Error(t, err)
	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.InvalidTextRepresentation, string(pgErr.Code))
}

func TestCreateProgramUniqueNameConstraint(t *testing.T) {
	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()
	truncateProgramsTable(dbConn)

	queries := db.New(dbConn)

	params := db.CreateProgramParams{
		Name:        "Go Course",
		Description: "Learn Go",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}
	_, err := queries.CreateProgram(context.Background(), params)
	require.NoError(t, err)
	_, err = queries.CreateProgram(context.Background(), params)
	require.Error(t, err)
	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.UniqueViolation, string(pgErr.Code))
}

func TestGetAllPrograms(t *testing.T) {
	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()
	truncateProgramsTable(dbConn)

	queries := db.New(dbConn)
	enumTypes := []db.ProgramProgramType{
		db.ProgramProgramTypeCourse,
		db.ProgramProgramTypeGame,
		db.ProgramProgramTypePractice,
		
	}

	for i, tpe := range enumTypes {
		params := db.CreateProgramParams{
			Name:        fmt.Sprintf("Program %d", i),
			Description: fmt.Sprintf("Desc %d", i),
			Level:       db.ProgramProgramLevelAll,
			Type:        tpe,
		}
		_, err := queries.CreateProgram(context.Background(), params)
		require.NoError(t, err)
	}

	programs, err := queries.GetPrograms(context.Background(), db.NullProgramProgramType{Valid: false})
	require.NoError(t, err)
	require.EqualValues(t, len(enumTypes), len(programs))
}

func TestGetNotExistingProgram(t *testing.T) {
	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()
	queries := db.New(dbConn)

	_, err := queries.GetProgramById(context.Background(), uuid.New())
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}

func TestUpdateNonExistentProgram(t *testing.T) {
	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()
	queries := db.New(dbConn)

	update := db.UpdateProgramParams{
		ID:          uuid.New(),
		Name:        "Nonexistent",
		Description: "Update",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}
	_, err := queries.UpdateProgram(context.Background(), update)
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}

func TestCreateCourseWithWrongLevel(t *testing.T) {
	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()

	_, err := dbConn.ExecContext(context.Background(), "TRUNCATE program.programs CASCADE")
	require.NoError(t, err)

	queries := db.New(dbConn)

	// 👇 Intentionally use an invalid level
	params := db.CreateProgramParams{
		Name:        "Go Course",
		Description: "Invalid level test",
		Level:       "not_a_valid_enum", // Invalid value for enum
		Type:        db.ProgramProgramTypeCourse,
	}

	_, err = queries.CreateProgram(context.Background(), params)
	require.Error(t, err)

	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.InvalidTextRepresentation, string(pgErr.Code))
}

func TestDeleteProgram(t *testing.T) {
	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()
	truncateProgramsTable(dbConn)

	queries := db.New(dbConn)

	params := db.CreateProgramParams{
		Name:        "ToDelete",
		Description: "Desc",
		Level:       db.ProgramProgramLevelAll,
		Type:        db.ProgramProgramTypeCourse,
	}
	created, err := queries.CreateProgram(context.Background(), params)
	require.NoError(t, err)

	rows, err := queries.DeleteProgram(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), rows)

	_, err = queries.GetProgramById(context.Background(), created.ID)
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}
