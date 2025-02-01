package dto

import (
	"api/internal/libs/validators"
	"bytes"
	"testing"
	"time"

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
			}},
		{
			name: "Invalid JSON",
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
			expectError: false,
			expectedValues: &CourseRequestDto{
				Description: "Learn the basics of Go programming",
				StartDate:   time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2025, time.February, 15, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Invalid JSON",
			jsonBody: `{
				"name": "John Doe",
				email: test@example.com
			}`,
			expectError: true,
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
