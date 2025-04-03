package event

//
//import (
//	"api/internal/domains/event/dto"
//	"api/internal/domains/event/entity"
//	"api/internal/domains/event/values"
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
//type MockEventsRepository struct {
//	mock.Mock
//}
//
//func (m *MockEventsRepository) GetGameById(c context.Context, id uuid.UUID) (*entity.Event, *errLib.CommonError) {
//	args := m.Called(c, id)
//
//	var event *entity.Event
//	var err *errLib.CommonError
//
//	if args.Get(0) != nil {
//		event = args.Get(0).(*entity.Event)
//	}
//	if args.Get(1) != nil {
//		err = args.Get(1).(*errLib.CommonError)
//	}
//
//	return event, err
//}
//
//func (m *MockEventsRepository) UpdateGame(c context.Context, inputCourse *entity.Event) (*entity.Event, *errLib.CommonError) {
//	args := m.Called(c, inputCourse)
//
//	var course *entity.Event
//	var err *errLib.CommonError
//
//	if args.Get(0) != nil {
//		course = args.Get(0).(*entity.Event)
//	}
//	if args.Get(1) != nil {
//		err = args.Get(1).(*errLib.CommonError)
//	}
//
//	return course, err
//}
//
//func (m *MockEventsRepository) GetBarberServices(c context.Context, name, description *string) ([]entity.Event, *errLib.CommonError) {
//	args := m.Called(c, name, description)
//
//	var courses []entity.Event
//	var err *errLib.CommonError
//
//	if args.Get(0) != nil {
//		courses = args.Get(0).([]entity.Event)
//	}
//	if args.Get(1) != nil {
//		err = args.Get(1).(*errLib.CommonError)
//	}
//
//	return courses, err
//}
//
//func (m *MockEventsRepository) DeleteSchedule(c context.Context, id uuid.UUID) *errLib.CommonError {
//	args := m.Called(c, id)
//
//	if args.Get(0) != nil {
//		return args.Get(0).(*errLib.CommonError)
//	}
//
//	return nil
//}
//
//func (m *MockEventsRepository) CreateSchedule(c context.Context, courseDetails *values.DateTimeDetails) (*entity.Event, *errLib.CommonError) {
//	args := m.Called(c, courseDetails)
//
//	var course *entity.Event
//	var err *errLib.CommonError
//
//	if args.Get(0) != nil {
//		course = args.Get(0).(*entity.Event)
//	}
//	if args.Get(1) != nil {
//		err = args.Get(1).(*errLib.CommonError)
//	}
//
//	return course, err
//}
//
//func SetupHandlersAndRepo() (*EventStaffsHandler, *MockEventsRepository) {
//	// Create a mock repository
//	mockRepo := new(MockEventsRepository)
//
//	// Return the handler with the mock repository
//	return NewEventStaffsHandler(mockRepo), mockRepo
//}
//
//func TestCreateCourse(t *testing.T) {
//	tests := []struct {
//		name           string
//		input          dto.EventRequestDto
//		mockSetup      func(*MockEventsRepository)
//		expectedStatus int
//	}{
//		{
//			name: "Success - Create Course",
//			input: dto.CourseRequestDto{
//				Name:        "Test Course",
//				Description: "A course for testing",
//			},
//			mockSetup: func(m *MockEventsRepository) {
//				m.On("CreateGame", mock.Anything, mock.Anything).Return(&entity.Event{
//					HubSpotId:          uuid.New(),
//					Name:        "Test Course",
//					Description: "A course for testing",
//				}, nil)
//			},
//			expectedStatus: http.StatusCreated,
//		},
//		{
//			name: "Failure - Invalid Payload",
//			input: dto.CourseRequestDto{
//				Name:        "Invalid Course",
//				Description: "Invalid payload",
//			},
//			mockSetup: func(m *MockEventsRepository) {
//				m.On("CreateGame", mock.Anything, mock.Anything).Return(
//					nil, &errLib.CommonError{Message: "", HTTPCode: http.StatusBadRequest})
//			}, expectedStatus: http.StatusBadRequest,
//		},
//		{
//			name: "Failure - Missing Required Fields",
//			input: dto.CourseRequestDto{
//				Description: "Missing name",
//			},
//			mockSetup: func(m *MockEventsRepository) {
//
//			},
//			expectedStatus: http.StatusBadRequest,
//		},
//		{
//			name: "Failure - Duplicate Course Name",
//			input: dto.CourseRequestDto{
//				Name:        "Duplicate Course",
//				Description: "First course with this name",
//			},
//			mockSetup: func(m *MockEventsRepository) {
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
//			r.Post("/courses", handlers.CreateGame)
//			r.ServeHTTP(rr, req)
//
//			require.Equal(t, tt.expectedStatus, rr.Code)
//
//			if rr.Code >= 400 { // Check for error response only when status code indicates an error
//				var response errLib.CommonError
//				err = json.Unmarshal(rr.Body.Bytes(), &response)
//				require.NoError(t, err)
//			} else { // Check for success response
//				var response dto.CourseResponse
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
//	courseDto := dto.CourseRequestDto{
//		Name:        "Updated Course",
//		Description: "Updated course description",
//	}
//
//	courseEntity := &entity.Event{
//		HubSpotId:          uuid.New(),
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
//	req, err := http.NewRequest("PUT", "/courses/"+courseEntity.HubSpotId.String(), bytes.NewBuffer(body))
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
//	var response dto.CourseResponse
//	err = json.Unmarshal(rr.Body.Bytes(), &response)
//	require.NoError(t, err)
//	require.Equal(t, courseDto.Name, response.Name)
//	require.Equal(t, courseDto.Description, response.Description)
//}
//
//func TestUpdateCourse_NotFound(t *testing.T) {
//	handlers, mockRepo := SetupHandlersAndRepo()
//
//	courseDto := dto.CourseRequestDto{
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
//	courseEntity := &entity.Event{
//		HubSpotId:          courseID,
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
//	var response dto.CourseResponse
//	err = json.Unmarshal(rr.Body.Bytes(), &response)
//	require.NoError(t, err)
//	require.Equal(t, courseEntity.Name, response.Name)
//}
//
//func TestGetCourses(t *testing.T) {
//	handlers, mockRepo := SetupHandlersAndRepo()
//
//	courses := []entity.Event{
//		{HubSpotId: uuid.New(), Name: "Course 1", Description: "First test course"},
//		{HubSpotId: uuid.New(), Name: "Course 2", Description: "Second test course"},
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
//	var response []dto.CourseResponse
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
