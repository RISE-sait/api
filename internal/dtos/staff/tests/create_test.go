package tests

import (
	"bytes"
	"testing"

	dto "api/internal/dtos/staff"
	"api/internal/utils/validators"

	"github.com/stretchr/testify/assert"
)

func TestDecodeCreateStaffRequestBody(t *testing.T) {
	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectedValues *dto.CreateStaffRequest
	}{
		{
			name: "Valid Input",
			jsonBody: `{
				"email": "klint",
				"role": "wdwd",
				"is_active": true
		}`,
			expectError: false,
			expectedValues: &dto.CreateStaffRequest{
				Email:    "klint",
				Role:     "wdwd",
				IsActive: true,
			},
		},
		{
			name: "empty email",
			jsonBody: `{
				"email": "",
				"role": "wdwd",
				"is_active": true
		}`,
			expectError: false,
		},
		{
			name: "empty role",
			jsonBody: `{
				"email": "klint",
				"role": "",
				"is_active": true
		}`,
			expectError: false,
		},
		{
			name: "invalid isActive",
			jsonBody: `{
				"email": "klint",
				"role": "",
				"is_active": false
		}`,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tc.jsonBody))
			var target dto.CreateStaffRequest

			err := validators.DecodeRequestBody(reqBody, &target)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)

				expected := tc.expectedValues

				if expected != nil {
					assert.Equal(t, expected.Email, target.Email)
					assert.Equal(t, expected.Role, target.Role)

					assert.Equal(t, expected.IsActive, target.IsActive)
					assert.Equal(t, expected.IsActive, target.IsActive)
				}
			}
		})
	}
}

func TestValidateCreateStaffRequestDto(t *testing.T) {
	tests := []struct {
		name          string
		dto           dto.CreateStaffRequest
		expectError   bool
		expectedError string
	}{
		{
			name: "Valid Input",
			dto: dto.CreateStaffRequest{
				Email:    "klintlee1@gmail.com",
				Role:     "wdwd",
				IsActive: true,
			},
			expectError: false,
		},
		{
			name: "Invalid email",
			dto: dto.CreateStaffRequest{
				Email:    "klintlee1",
				Role:     "wdwd",
				IsActive: true,
			},
			expectError: true,
		},
		{
			name: "Missing role",
			dto: dto.CreateStaffRequest{
				Email:    "klintlee1",
				IsActive: true,
			},
			expectError: true,
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
