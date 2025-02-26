package handler

import (
	"api/internal/domains/course/dto"
	"api/internal/domains/course/entity"
	"api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"bytes"
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockCourseRepository struct {
	mock.Mock
}

func (m *MockCourseRepository) GetCourseById(c context.Context, id uuid.UUID) (*entity.Course, *errLib.CommonError) {
	args := m.Called(c, id)

	var course *entity.Course
	var err *errLib.CommonError

	if args.Get(0) != nil {
		course = args.Get(0).(*entity.Course)
	}
	if args.Get(1) != nil {
		err = args.Get(1).(*errLib.CommonError)
	}

	return course, err
}

func (m *MockCourseRepository) UpdateCourse(c context.Context, inputCourse *entity.Course) (*entity.Course, *errLib.CommonError) {
	args := m.Called(c, inputCourse)

	var course *entity.Course
	var err *errLib.CommonError

	if args.Get(0) != nil {
		course = args.Get(0).(*entity.Course)
	}
	if args.Get(1) != nil {
		err = args.Get(1).(*errLib.CommonError)
	}

	return course, err
}

func (m *MockCourseRepository) GetCourses(c context.Context, name, description *string) ([]entity.Course, *errLib.CommonError) {
	args := m.Called(c, name, description)

	var courses []entity.Course
	var err *errLib.CommonError

	if args.Get(0) != nil {
		courses = args.Get(0).([]entity.Course)
	}
	if args.Get(1) != nil {
		err = args.Get(1).(*errLib.CommonError)
	}

	return courses, err
}

func (m *MockCourseRepository) DeleteCourse(c context.Context, id uuid.UUID) *errLib.CommonError {
	args := m.Called(c, id)

	if args.Get(0) != nil {
		return args.Get(0).(*errLib.CommonError)
	}

	return nil
}

func (m *MockCourseRepository) CreateCourse(c context.Context, courseDetails *values.CourseDetails) (*entity.Course, *errLib.CommonError) {
	args := m.Called(c, courseDetails)

	var course *entity.Course
	var err *errLib.CommonError

	if args.Get(0) != nil {
		course = args.Get(0).(*entity.Course)
	}
	if args.Get(1) != nil {
		err = args.Get(1).(*errLib.CommonError)
	}

	return course, err
}

func SetupHandlersAndRepo() (*Handler, *MockCourseRepository) {
	// Create a mock repository
	mockRepo := new(MockCourseRepository)

	// Return the handler with the mock repository
	return NewHandler(mockRepo), mockRepo
}

func TestCreateCourse(t *testing.T) {
	tests := []struct {
		name           string
		input          dto.CourseRequestDto
		mockSetup      func(*MockCourseRepository)
		expectedStatus int
	}{
		{
			name: "Success - Create Course",
			input: dto.CourseRequestDto{
				Name:        "Test Course",
				Description: "A course for testing",
			},
			mockSetup: func(m *MockCourseRepository) {
				m.On("CreateCourse", mock.Anything, mock.Anything).Return(&entity.Course{
					ID:          uuid.New(),
					Name:        "Test Course",
					Description: "A course for testing",
				}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Failure - Invalid Payload",
			input: dto.CourseRequestDto{
				Name:        "Invalid Course",
				Description: "Invalid payload",
			},
			mockSetup: func(m *MockCourseRepository) {
				m.On("CreateCourse", mock.Anything, mock.Anything).Return(
					nil, &errLib.CommonError{Message: "", HTTPCode: http.StatusBadRequest})
			}, expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Failure - Missing Required Fields",
			input: dto.CourseRequestDto{
				Description: "Missing name",
			},
			mockSetup: func(m *MockCourseRepository) {

			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Failure - Duplicate Course Name",
			input: dto.CourseRequestDto{
				Name:        "Duplicate Course",
				Description: "First course with this name",
			},
			mockSetup: func(m *MockCourseRepository) {
				m.On("CreateCourse", mock.Anything, mock.Anything).Return(
					nil, &errLib.CommonError{Message: "", HTTPCode: http.StatusConflict},
				)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockRepo := SetupHandlersAndRepo()
			tt.mockSetup(mockRepo)

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/courses", bytes.NewBuffer(body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r := chi.NewRouter()
			r.Post("/courses", handlers.CreateCourse)
			r.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)

			if rr.Code >= 400 { // Check for error response only when status code indicates an error
				var response errLib.CommonError
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
			} else { // Check for success response
				var response dto.CourseResponse
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.input.Name, response.Name)
				require.Equal(t, tt.input.Description, response.Description)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetCourseById(t *testing.T) {
	handlers, mockRepo := SetupHandlersAndRepo()

	courseID := uuid.New()
	courseEntity := &entity.Course{
		ID:          courseID,
		Name:        "Test Course",
		Description: "A test course",
	}

	mockRepo.On("GetCourseById", mock.Anything, courseID).Return(courseEntity, nil)

	req, err := http.NewRequest("GET", "/courses/"+courseID.String(), nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/courses/{id}", handlers.GetCourseById)

	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var response dto.CourseResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, courseEntity.Name, response.Name)
}

func TestGetCourses(t *testing.T) {
	handlers, mockRepo := SetupHandlersAndRepo()

	courses := []entity.Course{
		{ID: uuid.New(), Name: "Course 1", Description: "First test course"},
		{ID: uuid.New(), Name: "Course 2", Description: "Second test course"},
	}

	mockRepo.On("GetCourses", mock.Anything, mock.Anything, mock.Anything).Return(courses, nil)

	req, err := http.NewRequest("GET", "/courses", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/courses", handlers.GetCourses)

	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var response []dto.CourseResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response, 2)
}

func TestDeleteCourse(t *testing.T) {
	handlers, mockRepo := SetupHandlersAndRepo()

	courseID := uuid.New()

	mockRepo.On("DeleteCourse", mock.Anything, courseID).Return(nil)

	req, err := http.NewRequest("DELETE", "/courses/"+courseID.String(), nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Delete("/courses/{id}", handlers.DeleteCourse)

	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
}

func TestDeleteCourse_NotFound(t *testing.T) {
	handlers, mockRepo := SetupHandlersAndRepo()

	courseID := uuid.New()

	mockRepo.On("DeleteCourse", mock.Anything, courseID).Return(&errLib.CommonError{Message: "Course not found", HTTPCode: http.StatusNotFound})

	req, err := http.NewRequest("DELETE", "/courses/"+courseID.String(), nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Delete("/courses/{id}", handlers.DeleteCourse)

	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}
