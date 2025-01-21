package tests

// import (
// 	"bytes"
// 	"testing"
// 	"time"

// 	dto "api/internal/dtos/schedule"
// 	"api/internal/utils/validators"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// )

// func TestDecodeCreateScheduleRequestRequestBody(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		jsonBody       string
// 		expectError    bool
// 		expectedValues *dto.CreateScheduleRequest
// 	}{
// 		{
// 			name: "Valid JSON",
// 			jsonBody: `{
// 				"begin_datetime": "2025-01-01T00:00:00Z",
// 				"end_datetime": "2025-01-01T00:00:00Z",
// 				"course_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"facility_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"day": 1
// 			}`,
// 			expectError: false,
// 			expectedValues: &dto.CreateScheduleRequest{
// 				BeginDatetime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDatetime:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				CourseID:      uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				FacilityID:    uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Day:           1,
// 			},
// 		},
// 		{
// 			name: "Valid JSON: Nil UUID",
// 			jsonBody: `{
// 				"begin_datetime": "2025-01-01T00:00:00Z",
// 				"end_datetime": "2025-01-01T00:00:00Z",
// 				"course_id": "` + uuid.Nil.String() + `",
// 				"facility_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9"
// 			}`,
// 			expectError: false,
// 			expectedValues: &dto.CreateScheduleRequest{
// 				BeginDatetime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDatetime:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				CourseID:      uuid.Nil,
// 				FacilityID:    uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 			},
// 		},
// 		{
// 			name: "Invalid JSON: Malformed MembershipID",
// 			jsonBody: `{
// 				"begin_datetime": "2025-01-01T00:00:00Z",
// 				"end_datetime": "2025-01-01T00:00:00Z",
// 				"course_id": "invalid",
// 				"facility_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9"
// 			}`,
// 			expectError: true,
// 		},
// 		{
// 			name: "Invalid day",
// 			jsonBody: `{
// 				"begin_datetime": "2025-01-01T00:00:00Z",
// 				"end_datetime": "2025-01-01T00:00:00Z",
// 				"course_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"facility_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"day": "invalid"
// 			}`,
// 			expectError: true,
// 		},
// 		{
// 			name: "Invalid begin_datetime",
// 			jsonBody: `{
// 				"begin_datetime": "invalid",
// 				"end_datetime": "2025-01-01T00:00:00Z",
// 				"course_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"facility_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9"
// 			}`,
// 			expectError: true,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			reqBody := bytes.NewReader([]byte(tc.jsonBody))
// 			var target dto.CreateScheduleRequest

// 			err := validators.DecodeRequestBody(reqBody, &target)
// 			if tc.expectError {
// 				assert.NotNil(t, err)
// 			} else {
// 				assert.Nil(t, err)

// 				expected := tc.expectedValues

// 				if expected != nil {
// 					assert.Equal(t, expected.BeginDatetime, target.BeginDatetime)
// 					assert.Equal(t, expected.EndDatetime, target.EndDatetime)
// 					assert.Equal(t, expected.CourseID, target.CourseID)
// 					assert.Equal(t, expected.FacilityID, target.FacilityID)
// 					assert.Equal(t, expected.Day, target.Day)
// 				}
// 			}
// 		})
// 	}
// }

// func TestValidateDto(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		dto           dto.CreateScheduleRequest
// 		expectError   bool
// 		expectedError string
// 	}{
// 		{
// 			name: "Valid Input",
// 			dto: dto.CreateScheduleRequest{
// 				BeginDatetime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDatetime:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
// 				CourseID:      uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				FacilityID:    uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Day:           1,
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "Missing start datetime",
// 			dto: dto.CreateScheduleRequest{
// 				EndDatetime: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
// 				CourseID:    uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				FacilityID:  uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 			},
// 			expectError:   true,
// 			expectedError: "begin_datetime: required",
// 		},
// 		{
// 			name: "Missing course id",
// 			dto: dto.CreateScheduleRequest{
// 				BeginDatetime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDatetime:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
// 				FacilityID:    uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Day:           2,
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "Missing day",
// 			dto: dto.CreateScheduleRequest{
// 				BeginDatetime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDatetime:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
// 				FacilityID:    uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 			},
// 			expectError:   true,
// 			expectedError: "day: required",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := validators.ValidateDto(&tt.dto)
// 			if tt.expectError {
// 				assert.NotNil(t, err)
// 				if tt.expectedError != "" {
// 					assert.Contains(t, err.Message, tt.expectedError)
// 				}
// 			} else {
// 				assert.Nil(t, err)
// 			}
// 		})
// 	}
// }
