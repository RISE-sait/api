package course

// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"fmt"
// 	"os"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/lib/pq"

// 	"github.com/stretchr/testify/require"

// 	"github.com/testcontainers/testcontainers-go"
// 	"github.com/testcontainers/testcontainers-go/wait"

// 	db "api/internal/domains/course/persistence/sqlc/generated"
// )

// var dbInstance *sql.DB

// func setupTestDB(t *testing.T) (*db.Queries, func()) {
// 	ctx := context.Background()

// 	// Start a PostgreSQL container
// 	req := testcontainers.ContainerRequest{
// 		Image:        "postgres:13",
// 		ExposedPorts: []string{"5432/tcp"},
// 		Env: map[string]string{
// 			"POSTGRES_USER":     "postgres",
// 			"POSTGRES_PASSWORD": "root",
// 			"POSTGRES_DB":       "testdb",
// 		},
// 		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
// 	}

// 	_ = os.Setenv("TESTCONTAINERS_DEBUG", "true")

// 	// Create the container
// 	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
// 		ContainerRequest: req,
// 		Started:          true,
// 	})
// 	require.NoError(t, err)

// 	host, err := postgresC.Host(ctx)
// 	require.NoError(t, err)

// 	port, err := postgresC.MappedPort(ctx, "5432")
// 	require.NoError(t, err)

// 	dsn := fmt.Sprintf("postgresql://postgres:root@%s:%s/testdb?sslmode=disable", host, port.Port())

// 	// Open DB connection
// 	sqlDb, err := sql.Open("postgres", dsn)
// 	require.NoError(t, err)

// 	dbInstance = sqlDb

// 	require.NoError(t, sqlDb.Ping())

// 	migrationSQL := `
// 	CREATE TABLE IF NOT EXISTS courses (
// 	    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
// 	    name VARCHAR(50) NOT NULL UNIQUE,
// 	    description TEXT,
// 	    capacity INT NOT NULL,
// 	    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
// 	    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
// 	    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
// 	    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
// 	    CONSTRAINT end_after_start CHECK (end_date > start_date)
// 	);`

// 	_, err = sqlDb.Exec(migrationSQL)
// 	require.NoError(t, err)

// 	queries := db.New(sqlDb)

// 	// Return cleanup function to stop and remove the container after tests
// 	cleanup := func() {

// 		_, err := dbInstance.Exec("DELETE FROM courses")
// 		require.NoError(t, err)
// 	}

// 	return queries, cleanup
// }

// func TestCreateCourse(t *testing.T) {

// 	queries, cleanup := setupTestDB(t)
// 	defer cleanup()

// 	name := "Go Course"
// 	description := "Learn Go programming"

// 	createCourseParams := db.CreateCourseParams{
// 		Name:        name,
// 		Description: sql.NullString{String: description, Valid: description != ""},
// 		StartDate:   time.Now().Truncate(time.Second).UTC(),
// 		EndDate:     time.Now().AddDate(0, 0, 7).Truncate(time.Second).UTC(),
// 		Capacity:    100,
// 	}

// 	course, err := queries.CreateCourse(context.Background(), createCourseParams)

// 	require.NoError(t, err)

// 	expectedStartDate := createCourseParams.StartDate.Truncate(time.Second).UTC()
// 	actualStartDate := course.StartDate.Truncate(time.Second).UTC()

// 	expectedEndDate := createCourseParams.EndDate.Truncate(time.Second).UTC()
// 	actualEndDate := course.EndDate.Truncate(time.Second).UTC()

// 	// Assert course data
// 	require.Equal(t, name, course.Name)
// 	require.Equal(t, description, course.Description.String)
// 	require.Equal(t, createCourseParams.Capacity, course.Capacity)
// 	require.Equal(t, expectedStartDate, actualStartDate)
// 	require.Equal(t, expectedEndDate, actualEndDate)
// }

// func TestUpdateCourse(t *testing.T) {
// 	queries, cleanup := setupTestDB(t)
// 	defer cleanup()

// 	// Create a course to update
// 	name := "Go Course"
// 	description := "Learn Go programming"
// 	createCourseParams := db.CreateCourseParams{
// 		Name:        name,
// 		Description: sql.NullString{String: description, Valid: description != ""},
// 		StartDate:   time.Now().Truncate(time.Second).UTC(),
// 		EndDate:     time.Now().Truncate(time.Second).AddDate(0, 0, 7).UTC(),
// 	}

// 	course, err := queries.CreateCourse(context.Background(), createCourseParams)
// 	require.NoError(t, err)

// 	// Now, update the course
// 	newName := "Advanced Go Course"
// 	updateParams := db.UpdateCourseParams{
// 		ID:          course.ID,
// 		Name:        newName,
// 		StartDate:   time.Now().Truncate(time.Second).UTC(),
// 		EndDate:     time.Now().Truncate(time.Second).AddDate(0, 0, 7).UTC(),
// 		Description: sql.NullString{String: "Learn advanced Go programming", Valid: true},
// 		Capacity:    200,
// 	}

// 	_, err = queries.UpdateCourse(context.Background(), updateParams)
// 	require.NoError(t, err)

// 	// Get the updated course and verify
// 	updatedCourse, err := queries.GetCourseById(context.Background(), course.ID)
// 	require.NoError(t, err)
// 	require.Equal(t, newName, updatedCourse.Name)
// 	require.Equal(t, "Learn advanced Go programming", updatedCourse.Description.String)
// 	require.Equal(t, updateParams.Capacity, updatedCourse.Capacity)

// 	expectedStartDate := updateParams.StartDate.Truncate(time.Second).UTC()
// 	actualStartDate := updatedCourse.StartDate.Truncate(time.Second).UTC()
// 	require.Equal(t, expectedStartDate, actualStartDate)

// 	expectedEndDate := updateParams.EndDate.Truncate(time.Second).UTC()
// 	actualEndDate := updatedCourse.EndDate.Truncate(time.Second).UTC()
// 	require.Equal(t, expectedEndDate, actualEndDate)
// }

// func TestCreateCourseUniqueNameConstraint(t *testing.T) {
// 	queries, cleanup := setupTestDB(t)
// 	defer cleanup()

// 	// Create a course
// 	name := "Go Course"
// 	description := "Learn Go programming"
// 	createCourseParams := db.CreateCourseParams{
// 		Name:        name,
// 		Description: sql.NullString{String: description, Valid: description != ""},
// 		StartDate:   time.Now().Truncate(time.Second).UTC(),
// 		EndDate:     time.Now().Truncate(time.Second).AddDate(0, 0, 7).UTC(),
// 		Capacity:    100,
// 	}

// 	_, err := queries.CreateCourse(context.Background(), createCourseParams)
// 	require.NoError(t, err)

// 	// Attempt to create another course with the same name
// 	_, err = queries.CreateCourse(context.Background(), createCourseParams)
// 	require.Error(t, err)

// 	var pgErr *pq.Error
// 	require.True(t, errors.As(err, &pgErr))
// 	require.Equal(t, "23505", string(pgErr.Code)) // 23505 is the error code for unique violation
// }

// func TestGetAllCourses(t *testing.T) {
// 	queries, cleanup := setupTestDB(t)
// 	defer cleanup()

// 	// Create some courses
// 	for i := 1; i <= 5; i++ {
// 		createCourseParams := db.CreateCourseParams{
// 			Name:        fmt.Sprintf("Course %d", i),
// 			Description: sql.NullString{String: fmt.Sprintf("Description %d", i), Valid: true},
// 			StartDate:   time.Now().Truncate(time.Second).UTC(),
// 			EndDate:     time.Now().Truncate(time.Second).AddDate(0, 0, 7).UTC(),
// 			Capacity:    100,
// 		}
// 		_, err := queries.CreateCourse(context.Background(), createCourseParams)
// 		require.NoError(t, err)
// 	}

// 	params := db.GetCoursesParams{
// 		Name:        sql.NullString{String: "", Valid: false},
// 		Description: sql.NullString{String: "", Valid: false},
// 	}

// 	// Fetch all courses
// 	courses, err := queries.GetCourses(context.Background(), params)
// 	require.NoError(t, err)
// 	require.EqualValues(t, 5, len(courses))
// }

// func TestGetCoursesWithFilter(t *testing.T) {
// 	queries, cleanup := setupTestDB(t)
// 	defer cleanup()

// 	// Create some courses
// 	for i := 1; i <= 5; i++ {
// 		createCourseParams := db.CreateCourseParams{
// 			Name:        fmt.Sprintf("Course %d", i),
// 			Description: sql.NullString{String: fmt.Sprintf("Description %d", i), Valid: true},
// 			StartDate:   time.Now().Truncate(time.Second).UTC(),
// 			EndDate:     time.Now().Truncate(time.Second).AddDate(0, 0, 7).UTC(),
// 			Capacity:    100,
// 		}
// 		_, err := queries.CreateCourse(context.Background(), createCourseParams)
// 		require.NoError(t, err)
// 	}

// 	// Set a filter (e.g., filter courses by name)
// 	params := db.GetCoursesParams{
// 		Name:        sql.NullString{String: "Course 1", Valid: true}, // Filter for "Course 1"
// 		Description: sql.NullString{String: "", Valid: false},        // No filter on description
// 	}

// 	// Fetch courses with filter
// 	courses, err := queries.GetCourses(context.Background(), params)
// 	require.NoError(t, err)

// 	// Ensure that only the filtered course(s) are returned
// 	require.EqualValues(t, 1, len(courses))       // Only 1 course should match the filter ("Course 1")
// 	require.Equal(t, "Course 1", courses[0].Name) // Ensure the filtered course is "Course 1"
// }

// func TestUpdateNonExistentCourse(t *testing.T) {
// 	queries, cleanup := setupTestDB(t)
// 	defer cleanup()

// 	// Attempt to update a course that doesn't exist
// 	nonExistentId := uuid.New() // Random UUID

// 	updateParams := db.UpdateCourseParams{
// 		ID:          nonExistentId,
// 		Name:        "Updated Course",
// 		StartDate:   time.Now().Truncate(time.Second).Truncate(time.Second).UTC(),
// 		EndDate:     time.Now().AddDate(0, 0, 7).Truncate(time.Second).UTC(),
// 		Description: sql.NullString{String: "Updated course description", Valid: true},
// 	}

// 	affectedRows, err := queries.UpdateCourse(context.Background(), updateParams)
// 	require.NoError(t, err)

// 	require.Equal(t, affectedRows, int64(0))
// }

// func TestCreateCourseInvalidData(t *testing.T) {
// 	queries, cleanup := setupTestDB(t)
// 	defer cleanup()

// 	// Create a course with invalid start and end dates (end date is before start date)
// 	createCourseParams := db.CreateCourseParams{
// 		Name:        "Go Course",
// 		Description: sql.NullString{String: "Learn Go programming", Valid: true},
// 		StartDate:   time.Now().AddDate(0, 0, 7).Truncate(time.Second),
// 		EndDate:     time.Now().Truncate(time.Second).Truncate(time.Second), // Invalid end date (before start date)
// 	}

// 	_, err := queries.CreateCourse(context.Background(), createCourseParams)
// 	require.Error(t, err)
// }

// func TestCreateCourseWithNullDescription(t *testing.T) {
// 	queries, cleanup := setupTestDB(t)
// 	defer cleanup()

// 	// Create a course with a null description
// 	createCourseParams := db.CreateCourseParams{
// 		Name:        "Go Course",
// 		Description: sql.NullString{String: "", Valid: false},
// 		StartDate:   time.Now().Truncate(time.Second),
// 		EndDate:     time.Now().Truncate(time.Second).AddDate(0, 0, 7),
// 		Capacity:    100,
// 	}

// 	course, err := queries.CreateCourse(context.Background(), createCourseParams)
// 	require.NoError(t, err)

// 	// Fetch the course and check if description is null
// 	require.NoError(t, err)
// 	require.False(t, course.Description.Valid) // Should be null
// }

// func TestDeleteCourse(t *testing.T) {
// 	queries, cleanup := setupTestDB(t)
// 	defer cleanup()

// 	// Create a course to delete
// 	name := "Go Course"
// 	createCourseParams := db.CreateCourseParams{
// 		Name:        name,
// 		Description: sql.NullString{String: "Learn Go programming", Valid: true},
// 		StartDate:   time.Now().Truncate(time.Second),
// 		EndDate:     time.Now().Truncate(time.Second).AddDate(0, 0, 7),
// 	}

// 	course, err := queries.CreateCourse(context.Background(), createCourseParams)
// 	require.NoError(t, err)

// 	// Delete the course
// 	impactedRows, err := queries.DeleteCourse(context.Background(), course.ID)
// 	require.NoError(t, err)

// 	require.Equal(t, impactedRows, int64(1))

// 	// Attempt to fetch the deleted course (expecting error)
// 	_, err = queries.GetCourseById(context.Background(), course.ID)
// 	require.Error(t, err)
// }
