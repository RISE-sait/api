package dto

import (
	"api/internal/libs/validators"
	"bytes"
	"testing"

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
				"description": "Learn the basics of Go programming"
			}`,
			expectError: false,
			expectedValues: &CourseRequestDto{
				Name:        "Go Programming Basics",
				Description: "Learn the basics of Go programming",
			},
		},
		{
			name: "Invalid JSON - Missing closing brace",
			jsonBody: `{
				"name": "Go Programming Basics"
			`,
			expectError: true,
		},
		{
			name: "Missing Name",
			jsonBody: `{
				"description": "Learn the basics of Go programming"
			}`,
			expectError: false, // Expecting validation error for missing name
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
			},
			expectErr: false,
		},
		{
			name: "Missing Name",
			dto: &CourseRequestDto{
				Name:        "",
				Description: "Learn Go Programming",
			},
			expectErr:            true,
			expectedErrorMessage: "name: required",
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
