package game

//
//import (
//	dto "api/internal/domains/course/dto"
//	entity "api/internal/domains/course/entity"
//	values "api/internal/domains/course/values"
//	errLib "api/internal/libs/errors"
//	"context"
//	"github.com/google/uuid"
//	"github.com/stretchr/testify/mock"
//
//	"bytes"
//	"encoding/json"
//	"github.com/go-chi/chi"
//	"github.com/stretchr/testify/require"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//)
//
//type MockCourseRepository struct {
//	mock.Mock
//}
//
//func (m *MockCourseRepository) GetCourseById(c context.Context, id uuid.UUID) (*entity.Course, *errLib.CommonError) {
//	args := m.Called(c, id)
//
//	var course *entity.Course
//	var err *errLib.CommonError
//
//	if args.Get(0) != nil {
//		course = args.Get(0).(*entity.Course)
//	}
//	if args.Get(1) != nil {
//		err = args.Get(1).(*errLib.CommonError)
//	}
//
//	return course, err
//}
//
//func (m *MockCourseRepository) UpdateCourse(c context.Context, inputCourse *entity.Course) (*entity.Course, *errLib.CommonError) {
//	args := m.Called(c, inputCourse)
//
//	var course *entity.Course
//	var err *errLib.CommonError
//
//	if args.Get(0) != nil {
//		course = args.Get(0).(*entity.Course)
//	}
//	if args.Get(1) != nil {
//		err = args.Get(1).(*errLib.CommonError)
//	}
//
//	return course, err
//}
//
//func (m *MockCourseRepository) GetCourses(c context.Context, name, description *string) ([]entity.Course, *errLib.CommonError) {
//	args := m.Called(c, name, description)
//
//	var courses []entity.Course
//	var err *errLib.CommonError
//
//	if args.Get(0) != nil {
//		courses = args.Get(0).([]entity.Course)
//	}
//	if args.Get(1) != nil {
//		err = args.Get(1).(*errLib.CommonError)
//	}
//
//	return courses, err
//}
//
//func (m *MockCourseRepository) DeleteCourse(c context.Context, id uuid.UUID) *errLib.CommonError {
//	args := m.Called(c, id)
//
//	if args.Get(0) != nil {
//		return args.Get(0).(*errLib.CommonError)
//	}
//
//	return nil
//}
//
//func (m *MockCourseRepository) CreateCourse(c context.Context, courseDetails *values.Details) (*entity.Course, *errLib.CommonError) {
//	args := m.Called(c, courseDetails)
//
//	var course *entity.Course
//	var err *errLib.CommonError
//
//	if args.Get(0) != nil {
//		course = args.Get(0).(*entity.Course)
//	}
//	if args.Get(1) != nil {
//		err = args.Get(1).(*errLib.CommonError)
//	}
//
//	return course, err
//}
//
//func SetupHandlersAndRepo() (*Handler, *MockCourseRepository) {
//	// Create a mock repository
//	mockRepo := new(MockCourseRepository)
//
//	// Return the handler with the mock repository
//	return NewHandler(mockRepo), mockRepo
//}
//
//func TestCreateCourse(t *testing.T) {
//	tests := []struct {
//		name           string
//		input          dto.RequestDto
//		mockSetup      func(*MockCourseRepository)
//		expectedStatus int
//	}{
//		{
//			name: "Success - Create Course",
//			input: dto.RequestDto{
//				Name:        "Test Course",
//				Description: "A course for testing",
//			},
//			mockSetup: func(m *MockCourseRepository) {
//				m.On("CreateGame", mock.Anything, mock.Anything).Return(&entity.Course{
//					ID:          uuid.New(),
//					Name:        "Test Course",
//					Description: "A course for testing",
//				}, nil)
//			},
//			expectedStatus: http.StatusCreated,
//		},
//		{
//			name: "Failure - Invalid Payload",
//			input: dto.RequestDto{
//				Name:        "Invalid Course",
//				Description: "Invalid payload",
//			},
//			mockSetup: func(m *MockCourseRepository) {
//				m.On("CreateGame", mock.Anything, mock.Anything).Return(
//					nil, &errLib.CommonError{Message: "", HTTPCode: http.StatusBadRequest})
//			}, expectedStatus: http.StatusBadRequest,
//		},
//		{
//			name: "Failure - Missing Required Fields",
//			input: dto.RequestDto{
//				Description: "Missing name",
//			},
//			mockSetup: func(m *MockCourseRepository) {
//
//			},
//			expectedStatus: http.StatusBadRequest,
//		},
//		{
//			name: "Failure - Duplicate Course Name",
//			input: dto.RequestDto{
//				Name:        "Duplicate Course",
//				Description: "First course with this name",
//			},
//			mockSetup: func(m *MockCourseRepository) {
//				m.On("CreateGame", mock.Anything, mock.Anything).Return(
//					nil, &errLib.CommonError{Message: "", HTTPCode: http.StatusConflict},
//				)
//			},
//			expectedStatus: http.StatusConflict,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			handlers, mockRepo := SetupHandlersAndRepo()
//			tt.mockSetup(mockRepo)
//
//			body, err := json.Marshal(tt.input)
//			require.NoError(t, err)
//
//			req, err := http.NewRequest("POST", "/courses", bytes.NewBuffer(body))
//			require.NoError(t, err)
//
//			rr := httptest.NewRecorder()
//			r := chi.NewRouter()
//			r.Post("/courses", handlers.CreateCourse)
//			r.ServeHTTP(rr, req)
//
//			require.Equal(t, tt.expectedStatus, rr.Code)
//
//			if rr.Code >= 400 { // Check for error response only when status code indicates an error
//				var response errLib.CommonError
//				err = json.Unmarshal(rr.Body.Bytes(), &response)
//				require.NoError(t, err)
//			} else { // Check for success response
//				var response dto.ResponseDto
//				err = json.Unmarshal(rr.Body.Bytes(), &response)
//				require.NoError(t, err)
//				require.Equal(t, tt.input.Name, response.Name)
//				require.Equal(t, tt.input.Description, response.Description)
//			}
//			mockRepo.AssertExpectations(t)
//		})
//	}
//}
//
//func TestUpdateCourse(t *testing.T) {
//	handlers, mockRepo := SetupHandlersAndRepo()
//
//	// Prepare the course to update
//	courseDto := dto.RequestDto{
//		Name:        "Updated Course",
//		Description: "Updated course description",
//	}
//
//	courseEntity := &entity.Course{
//		ID:          uuid.New(),
//		Name:        courseDto.Name,
//		Description: courseDto.Description,
//	}
//
//	mockRepo.On("UpdateGame", mock.Anything, mock.Anything).Return(courseEntity, nil).Once()
//
//	body, err := json.Marshal(courseDto)
//	require.NoError(t, err)
//
//	// Simulate a PUT request
//	req, err := http.NewRequest("PUT", "/courses/"+courseEntity.ID.String(), bytes.NewBuffer(body))
//	require.NoError(t, err)
//
//	rr := httptest.NewRecorder()
//	r := chi.NewRouter()
//	r.Put("/courses/{id}", handlers.UpdateGame)
//
//	r.ServeHTTP(rr, req)
//
//	// Assert the response code
//	require.Equal(t, http.StatusNoContent, rr.Code)
//
//	// Additional assertions can be added here
//	var response dto.ResponseDto
//	err = json.Unmarshal(rr.Body.Bytes(), &response)
//	require.NoError(t, err)
//	require.Equal(t, courseDto.Name, response.Name)
//	require.Equal(t, courseDto.Description, response.Description)
//}
//
//func TestUpdateCourse_NotFound(t *testing.T) {
//	handlers, mockRepo := SetupHandlersAndRepo()
//
//	courseDto := dto.RequestDto{
//		Name:        "Updated Course",
//		Description: "Updated course description",
//	}
//
//	mockRepo.On("UpdateGame", mock.Anything, mock.Anything).Return(&errLib.CommonError{Message: "Course not found", HTTPCode: http.StatusNotFound})
//
//	body, err := json.Marshal(courseDto)
//	require.NoError(t, err)
//
//	// Simulate a PUT request
//	req, err := http.NewRequest("PUT", "/courses/invalid-id", bytes.NewBuffer(body))
//	require.NoError(t, err)
//
//	rr := httptest.NewRecorder()
//	r := chi.NewRouter()
//	r.Put("/courses/{id}", handlers.UpdateGame)
//
//	r.ServeHTTP(rr, req)
//
//	require.Equal(t, http.StatusBadRequest, rr.Code)
//}
//
//func TestGetCourseById(t *testing.T) {
//	handlers, mockRepo := SetupHandlersAndRepo()
//
//	courseID := uuid.New()
//	courseEntity := &entity.Course{
//		ID:          courseID,
//		Name:        "Test Course",
//		Description: "A test course",
//	}
//
//	mockRepo.On("GetGameById", mock.Anything, courseID).Return(courseEntity, nil)
//
//	req, err := http.NewRequest("GET", "/courses/"+courseID.String(), nil)
//	require.NoError(t, err)
//
//	rr := httptest.NewRecorder()
//	r := chi.NewRouter()
//	r.Get("/courses/{id}", handlers.GetGameById)
//
//	r.ServeHTTP(rr, req)
//
//	require.Equal(t, http.StatusOK, rr.Code)
//
//	var response dto.ResponseDto
//	err = json.Unmarshal(rr.Body.Bytes(), &response)
//	require.NoError(t, err)
//	require.Equal(t, courseEntity.Name, response.Name)
//}
//
//func TestGetCourses(t *testing.T) {
//	handlers, mockRepo := SetupHandlersAndRepo()
//
//	courses := []entity.Course{
//		{ID: uuid.New(), Name: "Course 1", Description: "First test course"},
//		{ID: uuid.New(), Name: "Course 2", Description: "Second test course"},
//	}
//
//	mockRepo.On("GetGames", mock.Anything, mock.Anything, mock.Anything).Return(courses, nil)
//
//	req, err := http.NewRequest("GET", "/courses", nil)
//	require.NoError(t, err)
//
//	rr := httptest.NewRecorder()
//	r := chi.NewRouter()
//	r.Get("/courses", handlers.GetGames)
//
//	r.ServeHTTP(rr, req)
//
//	require.Equal(t, http.StatusOK, rr.Code)
//
//	var response []dto.ResponseDto
//	err = json.Unmarshal(rr.Body.Bytes(), &response)
//	require.NoError(t, err)
//	require.Len(t, response, 2)
//}
//
//func TestDeleteCourse(t *testing.T) {
//	handlers, mockRepo := SetupHandlersAndRepo()
//
//	courseID := uuid.New()
//
//	mockRepo.On("DeleteGame", mock.Anything, courseID).Return(nil)
//
//	req, err := http.NewRequest("DELETE", "/courses/"+courseID.String(), nil)
//	require.NoError(t, err)
//
//	rr := httptest.NewRecorder()
//	r := chi.NewRouter()
//	r.Delete("/courses/{id}", handlers.DeleteGame)
//
//	r.ServeHTTP(rr, req)
//
//	require.Equal(t, http.StatusNoContent, rr.Code)
//}
//
//func TestDeleteCourse_NotFound(t *testing.T) {
//	handlers, mockRepo := SetupHandlersAndRepo()
//
//	courseID := uuid.New()
//
//	mockRepo.On("DeleteGame", mock.Anything, courseID).Return(&errLib.CommonError{Message: "Course not found", HTTPCode: http.StatusNotFound})
//
//	req, err := http.NewRequest("DELETE", "/courses/"+courseID.String(), nil)
//	require.NoError(t, err)
//
//	rr := httptest.NewRecorder()
//	r := chi.NewRouter()
//	r.Delete("/courses/{id}", handlers.DeleteGame)
//
//	r.ServeHTTP(rr, req)
//
//	require.Equal(t, http.StatusNotFound, rr.Code)
//}
