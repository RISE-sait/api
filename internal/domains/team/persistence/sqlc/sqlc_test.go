package practice

//
//import (
//	databaseErrors "api/internal/constants"
//	"context"
//	"database/sql"
//	"errors"
//	"fmt"
//	"testing"
//
//	"github.com/google/uuid"
//	"github.com/lib/pq"
//
//	"api/utils/test_utils"
//	"github.com/stretchr/testify/require"
//
//	practiceTestUtils "api/internal/domains/practice/persistence/test_utils"
//
//	db "api/internal/domains/practice/persistence/sqlc/generated"
//)
//
//func TestCreatePractice(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, cleanup := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//
//	defer cleanup()
//
//	name := "Go Course"
//	description := "Learn Go programming"
//
//	createPracticeParams := db.CreatePracticeParams{
//		Name:        name,
//		Description: description,
//		Level:       db.PracticeLevelAll,
//		Capacity:    100,
//	}
//
//	err := queries.CreatePractice(context.Background(), createPracticeParams)
//
//	require.NoError(t, err)
//
//	practices, err := queries.GetPractices(context.Background())
//
//	require.NoError(t, err)
//
//	practice := practices[0]
//
//	// Assert course data
//	require.Equal(t, name, practice.Name)
//	require.Equal(t, description, practice.Description)
//	require.Equal(t, createPracticeParams.Capacity, practice.Capacity)
//	require.Equal(t, createPracticeParams.Name, practice.Name)
//}
//
//func TestUpdatePracticeValid(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, cleanup := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//
//	defer cleanup()
//
//	// Create a course to update
//	name := "Go Course"
//	description := "Learn Go programming"
//
//	createPracticeParams := db.CreatePracticeParams{
//		Name:        name,
//		Description: description,
//		Level:       db.PracticeLevelAll,
//		Capacity:    100,
//	}
//
//	err := queries.CreatePractice(context.Background(), createPracticeParams)
//	require.NoError(t, err)
//
//	practices, err := queries.GetPractices(context.Background())
//
//	require.NoError(t, err)
//
//	practice := practices[0]
//
//	// Now, update the course
//	newName := "Advanced Go Course"
//	updateParams := db.UpdatePracticeParams{
//		ID:          practice.ID,
//		Name:        newName,
//		Description: "Learn advanced Go programming",
//		Capacity:    200,
//		Level:       db.PracticeLevelAll,
//	}
//
//	err = queries.UpdatePractice(context.Background(), updateParams)
//
//	// Get the updated course and verify
//	updatedCourse, err := queries.GetPracticeById(context.Background(), practice.ID)
//	require.NoError(t, err)
//	require.Equal(t, newName, updatedCourse.Name)
//	require.Equal(t, "Learn advanced Go programming", updatedCourse.Description)
//	require.Equal(t, updateParams.Capacity, updatedCourse.Capacity)
//}
//
//func TestUpdatePracticeInvalidLevel(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, cleanup := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//
//	defer cleanup()
//
//	// Create a course to update
//	name := "Go Course"
//	description := "Learn Go programming"
//
//	createPracticeParams := db.CreatePracticeParams{
//		Name:        name,
//		Description: description,
//		Level:       db.PracticeLevelAll,
//		Capacity:    100,
//	}
//
//	err := queries.CreatePractice(context.Background(), createPracticeParams)
//	require.NoError(t, err)
//
//	practices, err := queries.GetPractices(context.Background())
//
//	require.NoError(t, err)
//
//	practice := practices[0]
//
//	// Now, update the course
//	newName := "Advanced Go Course"
//	updateParams := db.UpdatePracticeParams{
//		ID:          practice.ID,
//		Name:        newName,
//		Description: "Learn advanced Go programming",
//		Capacity:    200,
//	}
//
//	err = queries.UpdatePractice(context.Background(), updateParams)
//	var pgErr *pq.Error
//	require.True(t, errors.As(err, &pgErr))
//	require.Equal(t, databaseErrors.InvalidTextRepresentation, string(pgErr.Code))
//
//}
//
//func TestCreatePracticeUniqueNameConstraint(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, cleanup := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//
//	defer cleanup()
//
//	// Create a course
//	name := "Go Course"
//	description := "Learn Go programming"
//	createCourseParams := db.CreatePracticeParams{
//		Name:        name,
//		Description: description,
//		Capacity:    100,
//		Level:       db.PracticeLevelAdvanced,
//	}
//
//	err := queries.CreatePractice(context.Background(), createCourseParams)
//	require.NoError(t, err)
//
//	// Attempt to create another course with the same name
//	err = queries.CreatePractice(context.Background(), createCourseParams)
//	require.Error(t, err)
//
//	var pgErr *pq.Error
//	require.True(t, errors.As(err, &pgErr))
//	require.Equal(t, databaseErrors.UniqueViolation, string(pgErr.Code))
//}
//
//func TestGetAllPractices(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, cleanup := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//
//	defer cleanup()
//
//	// Create some courses
//	for i := 1; i <= 5; i++ {
//		createCourseParams := db.CreatePracticeParams{
//			Name:        fmt.Sprintf("Course %d", i),
//			Description: fmt.Sprintf("Description %d", i),
//			Capacity:    100,
//			Level:       db.PracticeLevelAll,
//		}
//		err := queries.CreatePractice(context.Background(), createCourseParams)
//		require.NoError(t, err)
//	}
//
//	// Fetch all courses
//	courses, err := queries.GetPractices(context.Background())
//	require.NoError(t, err)
//	require.EqualValues(t, 5, len(courses))
//}
//
//func TestUpdateNonExistentPractice(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, cleanup := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//
//	defer cleanup()
//
//	// Attempt to update a practice that doesn't exist
//	nonExistentId := uuid.New() // Random UUID
//
//	updateParams := db.UpdatePracticeParams{
//		ID:          nonExistentId,
//		Name:        "Updated Practice",
//		Description: "Updated practice description",
//		Capacity:    150,
//		Level:       db.PracticeLevelAll,
//	}
//
//	err := queries.UpdatePractice(context.Background(), updateParams)
//	require.Nil(t, err)
//}
//
//func TestCreateCourseWithWrongLevel(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, cleanup := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//
//	defer cleanup()
//
//	// Create a course with a null description
//	createPracticeParams := db.CreatePracticeParams{
//		Name:        "Go Course",
//		Description: "wefwefew",
//		Capacity:    100,
//		Level:       "jhwwf",
//	}
//
//	err := queries.CreatePractice(context.Background(), createPracticeParams)
//
//	require.Error(t, err)
//
//	var pgErr *pq.Error
//	require.True(t, errors.As(err, &pgErr))
//	require.Equal(t, databaseErrors.InvalidTextRepresentation, string(pgErr.Code))
//}
//
//func TestDeleteCourse(t *testing.T) {
//
//	dbConn, _ := test_utils.SetupTestDB(t)
//
//	queries, cleanup := practiceTestUtils.SetupPracticeTestDbQueries(t, dbConn)
//
//	defer cleanup()
//
//	// Create a course to delete
//	name := "Go Course"
//	createPracticeParams := db.CreatePracticeParams{
//		Name:        name,
//		Description: "Learn Go programming",
//		Level:       db.PracticeLevelAll,
//		Capacity:    0,
//	}
//
//	err := queries.CreatePractice(context.Background(), createPracticeParams)
//	require.NoError(t, err)
//
//	practices, err := queries.GetPractices(context.Background())
//
//	require.NoError(t, err)
//
//	createdPractice := practices[0]
//
//	// Delete the course
//	impactedRows, err := queries.DeletePractice(context.Background(), createdPractice.ID)
//	require.NoError(t, err)
//
//	require.Equal(t, impactedRows, int64(1))
//
//	// Attempt to fetch the deleted course (expecting error)
//	_, err = queries.GetPracticeById(context.Background(), createdPractice.ID)
//
//	require.Error(t, err)
//
//	require.Equal(t, sql.ErrNoRows, err)
//}
