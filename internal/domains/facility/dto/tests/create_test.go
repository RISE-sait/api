package tests

import (
	"bytes"
	"strings"
	"testing"

	"api/internal/domains/facility/dto"
	"api/internal/libs/validators"

	"github.com/google/uuid"
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
				"location": "Location A",
				"facility_type_id": "b2f6ae19-62ff-4e64-aecc-08b432a8b593"
			}`,
			expectError: false,
		},
		{
			name: "Invalid JSON",
			jsonInput: `{
				"name": "Facility A",
				"location": "Location A",
				"facility_type_id": "b2f6ae19-62ff-4e64-aecc-08b432a8b593"
			`, // Missing closing brace
			expectError: true,
		},
		{
			name: "Validation: Missing Name",
			jsonInput: `{
				"location": "Location A",
				"facility_type_id": "b2f6ae19-62ff-4e64-aecc-08b432a8b593"
			}`,
			expectError: false,
		},
		{
			name: "Validation: Whitespace Name",
			jsonInput: `{
				"name": "   ",
				"location": "Location A",
				"facility_type_id": "b2f6ae19-62ff-4e64-aecc-08b432a8b593"
			}`,
			expectError: false,
		},
		{
			name: "Validation: Missing Location",
			jsonInput: `{
				"name": "Facility A",
				"facility_type_id": "b2f6ae19-62ff-4e64-aecc-08b432a8b593"
			}`,
			expectError: false,
		},
		{
			name: "Validation: Missing FacilityTypeID",
			jsonInput: `{
				"name": "Facility A",
				"location": "Location A"
			}`,
			expectError: false,
		},
		{
			name: "Validation: Invalid FacilityTypeID",
			jsonInput: `{
				"name": "Facility A",
				"location": "Location A",
				"facility_type_id": "invalid-uuid"
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tt.jsonInput))
			var target dto.FacilityRequestDto

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
					assert.Equal(t, "Location A", target.Location)
				}
				if target.FacilityTypeID != uuid.Nil {
					assert.Equal(t, uuid.MustParse("b2f6ae19-62ff-4e64-aecc-08b432a8b593"), target.FacilityTypeID)
				}
			}
		})
	}
}
