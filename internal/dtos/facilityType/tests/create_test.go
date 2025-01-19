package tests

import (
	"api/internal/utils/validators"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CreateFacilityTypeRequest struct {
	Name string `json:"name" validate:"required_and_notwhitespace"`
}

func TestDecodeFacilityTypeRequestBody(t *testing.T) {
	tests := []struct {
		name        string
		jsonInput   string
		expectError bool
		expected    *CreateFacilityTypeRequest
	}{
		{
			name:        "Valid JSON",
			jsonInput:   `{"name": "Test Facility"}`,
			expectError: false,
			expected: &CreateFacilityTypeRequest{
				Name: "Test Facility",
			},
		},
		{
			name:        "Invalid JSON",
			jsonInput:   `{"name": "Test Facility"`,
			expectError: true,
		},
		{
			name:        "Empty Name",
			jsonInput:   `{"name": ""}`,
			expectError: false,
		},
		{
			name:        "Whitespace Name",
			jsonInput:   `{"name": "   "}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tt.jsonInput))
			var target CreateFacilityTypeRequest

			err := validators.DecodeRequestBody(reqBody, &target)
			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)

				if tt.expected != nil {
					assert.Equal(t, tt.expected, &target)
				}
			}
		})
	}
}

func TestValidateFacilityTypeDto(t *testing.T) {
	tests := []struct {
		name          string
		dto           *CreateFacilityTypeRequest
		expectError   bool
		expectedError string
	}{
		{
			name:        "Valid DTO",
			dto:         &CreateFacilityTypeRequest{Name: "Valid Facility"},
			expectError: false,
		},
		{
			name:          "Empty Name",
			dto:           &CreateFacilityTypeRequest{Name: ""},
			expectError:   true,
			expectedError: "name: required",
		},
		{
			name:          "Whitespace Name",
			dto:           &CreateFacilityTypeRequest{Name: "   "},
			expectError:   true,
			expectedError: "name: required, cannot be empty or whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateDto(tt.dto)
			if tt.expectError {
				assert.NotNil(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Message, tt.expectedError)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
