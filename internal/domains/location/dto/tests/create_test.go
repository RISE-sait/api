package tests

import (
	"bytes"
	"strings"
	"testing"

	dto "api/internal/domains/location/dto"
	"api/internal/libs/validators"

	"github.com/stretchr/testify/assert"
)

func TestDecodeCreateFacilityRequestBody(t *testing.T) {
	tests := []struct {
		name          string
		jsonInput     string
		expectError   bool
		expectedError string
	}{
		{
			name: "Valid JSON",
			jsonInput: `{
				"name": "Facility A",
				"location": "Address A"
			}`,
			expectError: false,
		},
		{
			name: "Invalid JSON",
			jsonInput: `{
				"name": "Facility A",
				"location": "Address A"
			`,
			expectError: true,
		},
		{
			name: "Validation: Missing Name",
			jsonInput: `{
				"location": "Address A"
			}`,
			expectError: false,
		},
		{
			name: "Validation: Whitespace Name",
			jsonInput: `{
				"name": "   ",
				"location": "Address A"
			}`,
			expectError: false,
		},
		{
			name: "Validation: Missing Address",
			jsonInput: `{
				"name": "Facility A"
			}`,
			expectError: false,
		},
		{
			name: "Validation: Missing FacilityCategoryID",
			jsonInput: `{
				"name": "Facility A",
				"location": "Address A"
			}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tt.jsonInput))
			var target dto.RequestDto

			err := validators.ParseJSON(reqBody, &target)
			if tt.expectError {
				assert.NotNil(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Message, tt.expectedError)
				}
			} else {
				assert.Nil(t, err)

				if strings.TrimSpace(target.Name) != "" {
					assert.Equal(t, "Facility A", target.Name)
				}
				if strings.TrimSpace(target.Location) != "" {
					assert.Equal(t, "Address A", target.Location)
				}
			}
		})
	}
}
