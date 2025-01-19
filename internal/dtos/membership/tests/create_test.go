package tests

import (
	"bytes"
	"testing"
	"time"

	dto "api/internal/dtos/membership"
	"api/internal/utils/validators"

	"github.com/stretchr/testify/assert"
)

func TestDecodeCreateMembershipRequestRequestBody(t *testing.T) {
	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectedValues *dto.CreateMembershipRequest
	}{
		{
			name: "Valid JSON",
			jsonBody: `{
				"name": "Membership A",
				"description": "Description of Membership A",
				"start_date": "2025-01-01T00:00:00Z",
				"end_date": "2025-12-31T23:59:59Z"
			}`,
			expectError: false,
			expectedValues: &dto.CreateMembershipRequest{
				Name:        "Membership A",
				Description: "Description of Membership A",
				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			},
		},
		{
			name: "Invalid JSON: Missing EndDate",
			jsonBody: `{
				"name": "Membership A",
				"description": "Description of Membership A",
				"start_date": "2025-01-01T00:00:00Z"
			}`,
			expectError: false,
		},
		{
			name: "Invalid JSON: Malformed Date",
			jsonBody: `{
				"name": "Membership A",
				"description": "Description of Membership A",
				"start_date": "2025-01-01T00:00:00Z",
				"end_date": "invalid-date"
			}`,
			expectError: true,
		},
		{
			name: "Empty Name",
			jsonBody: `{
				"name": "",
				"description": "Description of Membership A",
				"start_date": "2025-01-01T00:00:00Z",
				"end_date": "2025-12-31T23:59:59Z"
			}`,
			expectError: false,
		},
		{
			name: "Whitespace Name",
			jsonBody: `{
				"name": "   ",
				"description": "Description of Membership A",
				"start_date": "2025-01-01T00:00:00Z",
				"end_date": "2025-12-31T23:59:59Z"
			}`,
			expectError: false,
		},
		{
			name: "Valid Empty Description",
			jsonBody: `{
				"name": "Membership A",
				"description": "",
				"start_date": "2025-01-01T00:00:00Z",
				"end_date": "2025-12-31T23:59:59Z"
			}`,
			expectError: false,
			expectedValues: &dto.CreateMembershipRequest{
				Name:        "Membership A",
				Description: "",
				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tc.jsonBody))
			var target dto.CreateMembershipRequest

			err := validators.DecodeRequestBody(reqBody, &target)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)

				expected := tc.expectedValues

				if expected != nil {
					if expected.Name != "" {
						assert.Equal(t, expected.Name, target.Name)
					}

					if expected.Description != "" {
						assert.Equal(t, expected.Description, target.Description)
					}

					if !expected.StartDate.IsZero() {
						assert.Equal(t, expected.StartDate, target.StartDate)
					}

					if !expected.EndDate.IsZero() {
						assert.Equal(t, expected.EndDate, target.EndDate)
					}
				}
			}
		})
	}
}

func TestValidateCreateMembershipRequestDto(t *testing.T) {
	tests := []struct {
		name          string
		dto           dto.CreateMembershipRequest
		expectError   bool
		expectedError string
	}{
		{
			name: "Valid Input",
			dto: dto.CreateMembershipRequest{
				Name:        "Membership A",
				Description: "Description of Membership A",
				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			},
			expectError: false,
		},
		{
			name: "Missing Name",
			dto: dto.CreateMembershipRequest{
				Description: "Description of Membership A",
				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			},
			expectError:   true,
			expectedError: "name: required and cannot be empty or whitespace",
		},
		{
			name: "Whitespace Name",
			dto: dto.CreateMembershipRequest{
				Name:        "   ",
				Description: "Description of Membership A",
				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			},
			expectError:   true,
			expectedError: "name: required and cannot be empty or whitespace",
		},
		{
			name: "Missing StartDate",
			dto: dto.CreateMembershipRequest{
				Name:        "Membership A",
				Description: "Description of Membership A",
				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			},
			expectError:   true,
			expectedError: "start_date: required",
		},
		{
			name: "Missing EndDate",
			dto: dto.CreateMembershipRequest{
				Name:        "Membership A",
				Description: "Description of Membership A",
				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectError:   true,
			expectedError: "end_date: required",
		},
		{
			name: "EndDate before StartDate",
			dto: dto.CreateMembershipRequest{
				Name:        "Membership A",
				Description: "Description of Membership A",
				StartDate:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
				EndDate:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectError:   true,
			expectedError: "end_date: EndDate must be after StartDate",
		},
		{
			name: "No Description",
			dto: dto.CreateMembershipRequest{
				Name:      "Membership A",
				StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			},
			expectError: false,
		},
		{
			name:          "Empty JSON Fields",
			dto:           dto.CreateMembershipRequest{},
			expectError:   true,
			expectedError: "name: required and cannot be empty or whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateDto(&tt.dto)
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
