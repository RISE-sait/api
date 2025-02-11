package course

import (
	"api/internal/domains/course/dto"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"

	test_utils "api/internal/domains/course/test_utils"
)

func TestCreateCourse(t *testing.T) {
	// Initialize the repository and controller
	repo, cleanup := test_utils.SetupTestRepository(t)
	defer cleanup()

	// Initialize the service and controller
	service := NewCourseService(repo)
	controller := NewCourseController(service)

	// Create a course request DTO
	courseDto := dto.CourseRequestDto{
		Name:        "Test Course",
		Description: "A course for testing",
		Capacity:    30,
		StartDate:   time.Now().Add(24 * time.Hour).Truncate(time.Second).UTC(), // Start tomorrow
		EndDate:     time.Now().Add(48 * time.Hour).Truncate(time.Second).UTC(), // End in two days
	}

	// Marshal the DTO into JSON
	body, err := json.Marshal(courseDto)

	require.NoError(t, err)

	// Create a new HTTP POST request to the /courses endpoint
	req, err := http.NewRequest("POST", "/courses", bytes.NewBuffer(body))
	require.NoError(t, err)

	// Set up the HTTP response recorder
	rr := httptest.NewRecorder()

	// Set up the router with the controller
	r := chi.NewRouter()
	r.Post("/courses", controller.CreateCourse)

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Assert the response code and any other expected behaviors
	require.Equal(t, http.StatusCreated, rr.Code)

	// Additional assertions to verify the response can be added here
	var response dto.CourseResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)

	require.NoError(t, err)
	require.Equal(t, courseDto.Name, response.Name)
	require.Equal(t, courseDto.Description, response.Description)
	require.Equal(t, courseDto.Capacity, response.Capacity)
	require.Equal(t, courseDto.StartDate, response.StartDate)
	require.Equal(t, courseDto.EndDate, response.EndDate)
}

func TestCreateCourse_InvalidPayload(t *testing.T) {

	repo, cleanup := test_utils.SetupTestRepository(t)
	defer cleanup()

	// Initialize the service and controller
	service := NewCourseService(repo)
	controller := NewCourseController(service)

	invalidJSON := `{"name": "Invalid Course", "capacity": "invalid_int"}`

	req, err := http.NewRequest("POST", "/courses", bytes.NewBuffer([]byte(invalidJSON)))
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/courses", controller.CreateCourse)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateCourse_MissingRequiredFields(t *testing.T) {
	repo, cleanup := test_utils.SetupTestRepository(t)
	defer cleanup()

	// Initialize the service and controller
	service := NewCourseService(repo)
	controller := NewCourseController(service)

	missingFieldsDto := dto.CourseRequestDto{
		Description: "Missing name",
		Capacity:    30,
		StartDate:   time.Now().Add(24 * time.Hour).Truncate(time.Second).UTC(),
		EndDate:     time.Now().Add(48 * time.Hour).Truncate(time.Second).UTC(),
	}

	body, err := json.Marshal(missingFieldsDto)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/courses", bytes.NewBuffer(body))
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/courses", controller.CreateCourse)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateCourse_InvalidDateRange(t *testing.T) {
	repo, cleanup := test_utils.SetupTestRepository(t)
	defer cleanup()

	// Initialize the service and controller
	service := NewCourseService(repo)
	controller := NewCourseController(service)
	invalidDatesDto := dto.CourseRequestDto{
		Name:        "Invalid Date Course",
		Description: "End date before start date",
		Capacity:    30,
		StartDate:   time.Now().Add(48 * time.Hour).Truncate(time.Second).UTC(),
		EndDate:     time.Now().Add(24 * time.Hour).Truncate(time.Second).UTC(),
	}

	body, err := json.Marshal(invalidDatesDto)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/courses", bytes.NewBuffer(body))
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/courses", controller.CreateCourse)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateCourse_DuplicateCourseName(t *testing.T) {
	repo, cleanup := test_utils.SetupTestRepository(t)
	defer cleanup()

	// Initialize the service and controller
	service := NewCourseService(repo)
	controller := NewCourseController(service)
	courseDto := dto.CourseRequestDto{
		Name:        "Duplicate Course",
		Description: "First course with this name",
		Capacity:    30,
		StartDate:   time.Now().Add(24 * time.Hour).Truncate(time.Second).UTC(),
		EndDate:     time.Now().Add(48 * time.Hour).Truncate(time.Second).UTC(),
	}

	body, err := json.Marshal(courseDto)
	require.NoError(t, err)

	rr1 := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/courses", controller.CreateCourse)

	req1, err := http.NewRequest("POST", "/courses", bytes.NewBuffer(body))
	require.NoError(t, err)
	r.ServeHTTP(rr1, req1)

	require.Equal(t, http.StatusCreated, rr1.Code)

	var response dto.CourseResponse
	err = json.Unmarshal(rr1.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, courseDto.Name, response.Name)

	// Second request (should fail due to duplicate name)
	rr2 := httptest.NewRecorder()
	req2, err := http.NewRequest("POST", "/courses", bytes.NewBuffer(body))
	require.NoError(t, err)
	r.ServeHTTP(rr2, req2)

	require.Equal(t, http.StatusConflict, rr2.Code)
}
