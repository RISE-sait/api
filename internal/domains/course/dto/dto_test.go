package dto

import (
	"bytes"
	"testing"
	"time"

	"api/internal/libs/validators"

	"github.com/stretchr/testify/assert"
)

func TestDecodeRequestBody(t *testing.T) {
	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectedValues *CourseRequestDto
	}{
		{
			name: "Valid Input",
			jsonBody: `{
				"name": "Go Programming Basics",
				"description": "Learn the basics of Go programming",
				"start_date": "2025-01-15T00:00:00Z",
				"end_date": "2025-02-15T00:00:00Z"
			}`,
			expectError: false,
			expectedValues: &CourseRequestDto{
				Name:        "Go Programming Basics",
				Description: "Learn the basics of Go programming",
				StartDate:   time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2025, time.February, 15, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Invalid JSON - Missing closing brace",
			jsonBody: `{
				"name": "Go Programming Basics",
				"start_date": "2025-01-15T00:00:00Z",
				"end_date": "2025-02-15T00:00:00Z"
			`,
			expectError: true,
		},
		{
			name: "Missing Name",
			jsonBody: `{
				"description": "Learn the basics of Go programming",
				"start_date": "2025-01-15T00:00:00Z",
				"end_date": "2025-02-15T00:00:00Z"
			}`,
			expectError: false, // Expecting validation error for missing name
		},
		{
			name: "Invalid Date Format",
			jsonBody: `{
				"name": "Go Programming Basics",
				"description": "Learn the basics of Go programming",
				"start_date": "invalid-date-format",
				"end_date": "2025-02-15T00:00:00Z"
			}`,
			expectError: true, // Invalid date format
		},
		{
			name: "End Date Before Start Date",
			jsonBody: `{
				"name": "Go Programming Basics",
				"description": "Learn the basics of Go programming",
				"start_date": "2025-02-15T00:00:00Z",
				"end_date": "2025-01-15T00:00:00Z"
			}`,
			expectError: false, // End date should be after start date
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tc.jsonBody))
			var target CourseRequestDto

			err := validators.ParseJSON(reqBody, &target)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if tc.expectedValues != nil {
					assert.Equal(t, tc.expectedValues.StartDate, target.StartDate)
					assert.Equal(t, tc.expectedValues.EndDate, target.EndDate)
					assert.Equal(t, tc.expectedValues.Name, target.Name)
					assert.Equal(t, tc.expectedValues.Description, target.Description)
				}
			}
		})
	}
}

func TestCourseRequestDto_Validation(t *testing.T) {
	tests := []struct {
		name                 string
		dto                  *CourseRequestDto
		expectErr            bool
		expectedErrorMessage string
	}{
		{
			name: "Valid DTO",
			dto: &CourseRequestDto{
				Name:        "Go Programming Basics",
				Description: "Learn Go Programming",
				StartDate:   time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2025, time.February, 15, 0, 0, 0, 0, time.UTC),
			},
			expectErr: false,
		},
		{
			name: "Missing Name",
			dto: &CourseRequestDto{
				Name:        "",
				Description: "Learn Go Programming",
				StartDate:   time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2025, time.February, 15, 0, 0, 0, 0, time.UTC),
			},
			expectErr:            true,
			expectedErrorMessage: "name: required",
		},
		{
			name: "Invalid End Date",
			dto: &CourseRequestDto{
				Name:        "Go Programming Basics",
				Description: "Learn Go Programming",
				StartDate:   time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC), // End date before Start date
			},
			expectErr:            true,
			expectedErrorMessage: "end_date: must be greater than start_date", // The expected error message for invalid date range
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.dto.validate()
			if tc.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
