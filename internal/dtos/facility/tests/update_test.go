package tests

import (
	dto "api/internal/dtos/facility"
	"api/internal/utils/validators"
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDecodeUpdateFacilityRequestRequestBody(t *testing.T) {
	tests := []struct {
		name          string
		json          string
		expectError   bool
		expectedError string
		expected      *dto.UpdateFacilityRequest
	}{
		{
			name: "Valid JSON",
			json: `{
				"id": "d3f6e7a8-cbe3-4a9f-8f21-7a27a54a6de9",
				"name": "Facility B",
				"location": "Location B",
				"facility_type_id": "b2f6ae19-62ff-4e64-aecc-08b432a8b593"
			}`,
			expectError: false,
			expected: &dto.UpdateFacilityRequest{
				ID:             uuid.MustParse("d3f6e7a8-cbe3-4a9f-8f21-7a27a54a6de9"),
				Name:           "Facility B",
				Location:       "Location B",
				FacilityTypeID: uuid.MustParse("b2f6ae19-62ff-4e64-aecc-08b432a8b593"),
			},
		},
		{
			name:        "Invalid JSON",
			json:        `{"id": "d3f6e7a8-cbe3-4a9f-8f21-7a27a54a6de9", "name": "Facility B", "location": "Location B", "facility_type_id": "b2f6ae19-62ff-4e64-aecc-08b432a8b593"`,
			expectError: true,
		},
		{
			name: "Validation: Missing ID",
			json: `{
				"name": "Facility B",
				"location": "Location B",
				"facility_type_id": "b2f6ae19-62ff-4e64-aecc-08b432a8b593"
			}`,
			expected: &dto.UpdateFacilityRequest{
				Name:           "Facility B",
				Location:       "Location B",
				FacilityTypeID: uuid.MustParse("b2f6ae19-62ff-4e64-aecc-08b432a8b593"),
			},
			expectError: false,
		},
		{
			name: "Validation: Missing Name",
			json: `{
				"id": "d3f6e7a8-cbe3-4a9f-8f21-7a27a54a6de9",
				"location": "Location B",
				"facility_type_id": "b2f6ae19-62ff-4e64-aecc-08b432a8b593"
			}`,
			expected: &dto.UpdateFacilityRequest{
				ID:             uuid.MustParse("d3f6e7a8-cbe3-4a9f-8f21-7a27a54a6de9"),
				Location:       "Location B",
				FacilityTypeID: uuid.MustParse("b2f6ae19-62ff-4e64-aecc-08b432a8b593"),
			},
			expectError: false,
		},
		{
			name: "Validation: Missing Location",
			json: `{
				"id": "d3f6e7a8-cbe3-4a9f-8f21-7a27a54a6de9",
				"name": "Facility B",
				"facility_type_id": "b2f6ae19-62ff-4e64-aecc-08b432a8b593"
			}`,
			expected: &dto.UpdateFacilityRequest{
				ID:             uuid.MustParse("d3f6e7a8-cbe3-4a9f-8f21-7a27a54a6de9"),
				Name:           "Facility B",
				FacilityTypeID: uuid.MustParse("b2f6ae19-62ff-4e64-aecc-08b432a8b593"),
			},
			expectError: false,
		},
		{
			name: "Validation: Missing FacilityTypeID",
			json: `{
				"id": "d3f6e7a8-cbe3-4a9f-8f21-7a27a54a6de9",
				"name": "Facility B",
				"location": "Location B"
			}`,
			expected: &dto.UpdateFacilityRequest{
				ID:       uuid.MustParse("d3f6e7a8-cbe3-4a9f-8f21-7a27a54a6de9"),
				Name:     "Facility B",
				Location: "Location B",
			},
			expectError: false,
		},
		{
			name: "Validation: Invalid FacilityTypeID",
			json: `{
				"id": "d3f6e7a8-cbe3-4a9f-8f21-7a27a54a6de9",
				"name": "Facility B",
				"location": "Location B",
				"facility_type_id": "invalid-uuid"
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tt.json))
			var target dto.UpdateFacilityRequest

			err := validators.DecodeRequestBody(reqBody, &target)
			if tt.expectError {
				assert.NotNil(t, err)

				if tt.expectedError != "" {
					assert.Contains(t, err.Message, tt.expectedError)
				}
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, &target)
			}
		})
	}
}

func TestValidateUpdateFacilityDto(t *testing.T) {
	tests := []struct {
		name          string
		dto           *dto.UpdateFacilityRequest
		expectError   bool
		expectedError string
	}{
		{
			name:        "Valid JSON",
			expectError: false,
			dto: &dto.UpdateFacilityRequest{
				ID:             uuid.MustParse("d2fa7ac4-736e-40ac-8e1d-01c3ea4a7ad3"),
				Name:           "Updated Facility",
				Location:       "Updated Location",
				FacilityTypeID: uuid.MustParse("fcbfc5a7-bb5f-4132-bd7b-b97b6871f8b0"),
			},
		},
		{
			name: "Missing Name",
			dto: &dto.UpdateFacilityRequest{
				ID:             uuid.MustParse("d2fa7ac4-736e-40ac-8e1d-01c3ea4a7ad3"),
				Location:       "Updated Location",
				FacilityTypeID: uuid.MustParse("fcbfc5a7-bb5f-4132-bd7b-b97b6871f8b0"),
			},
			expectError:   true,
			expectedError: "name: required and cannot be empty or whitespace",
		},
		{
			name: "Whitespace Name",
			dto: &dto.UpdateFacilityRequest{
				ID:             uuid.MustParse("d2fa7ac4-736e-40ac-8e1d-01c3ea4a7ad3"),
				Name:           "   ",
				Location:       "Updated Location",
				FacilityTypeID: uuid.MustParse("fcbfc5a7-bb5f-4132-bd7b-b97b6871f8b0"),
			},
			expectError:   true,
			expectedError: "name: required and cannot be empty or whitespace",
		},
		{
			name: "Missing location",
			dto: &dto.UpdateFacilityRequest{
				ID:             uuid.MustParse("d2fa7ac4-736e-40ac-8e1d-01c3ea4a7ad3"),
				Name:           "eeeegegeg",
				FacilityTypeID: uuid.MustParse("fcbfc5a7-bb5f-4132-bd7b-b97b6871f8b0"),
			},
			expectError:   true,
			expectedError: "location: required and cannot be empty or whitespace",
		},
		{
			name: "Missing facility type id",
			dto: &dto.UpdateFacilityRequest{
				ID:       uuid.MustParse("d2fa7ac4-736e-40ac-8e1d-01c3ea4a7ad3"),
				Name:     "eeeegegeg",
				Location: "Updated Location",
			},
			expectError:   true,
			expectedError: "facility_type_id: required",
		},
		{
			name: "Empty Fields",
			dto: &dto.UpdateFacilityRequest{
				ID:       uuid.UUID{},
				Name:     " ",
				Location: " ",
			},
			expectError:   true,
			expectedError: "id: required",
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
