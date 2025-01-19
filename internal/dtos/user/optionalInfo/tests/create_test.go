package tests

import (
	userOptionalInfo "api/internal/dtos/user/optionalInfo"
	"api/internal/utils/validators"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeCreateUserOptionalInfoRequestBody(t *testing.T) {
	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectedValues *userOptionalInfo.CreateUserOptionalInfoRequest
	}{
		{
			name: "Valid Input",
			jsonBody: `{
				"name": "John Doe",
				"email": "test@example.com",
				"hashed_password": "hashedpass123"
			}`,
			expectError: false,
			expectedValues: &userOptionalInfo.CreateUserOptionalInfoRequest{
				Name:           "John Doe",
				Email:          "test@example.com",
				HashedPassword: "hashedpass123",
			},
		},
		{
			name: "Invalid Email",
			jsonBody: `{
				"name": "John Doe",
				"email": "invalid-email",
				"hashed_password": "hashedpass123"
			}`,
			expectError: false,
		},
		{
			name: "Missing Optional Fields",
			jsonBody: `{
				"email": "test@example.com"
			}`,
			expectError: false,
			expectedValues: &userOptionalInfo.CreateUserOptionalInfoRequest{
				Email: "test@example.com",
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
			var target userOptionalInfo.CreateUserOptionalInfoRequest

			err := validators.DecodeRequestBody(reqBody, &target)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if tc.expectedValues != nil {
					assert.Equal(t, tc.expectedValues.Name, target.Name)
					assert.Equal(t, tc.expectedValues.Email, target.Email)
					assert.Equal(t, tc.expectedValues.HashedPassword, target.HashedPassword)
				}
			}
		})
	}
}

func TestValidateCreateUserOptionalInfoRequest(t *testing.T) {
	tests := []struct {
		name        string
		dto         *userOptionalInfo.CreateUserOptionalInfoRequest
		expectError bool
	}{
		{
			name: "Valid Input",
			dto: &userOptionalInfo.CreateUserOptionalInfoRequest{
				Name:           "John Doe",
				Email:          "test@example.com",
				HashedPassword: "hashedpass123",
			},
			expectError: false,
		},
		{
			name: "Invalid Email Format",
			dto: &userOptionalInfo.CreateUserOptionalInfoRequest{
				Name:           "John Doe",
				Email:          "invalid-email",
				HashedPassword: "hashedpass123",
			},
			expectError: true,
		},
		{
			name: "Empty Optional Fields",
			dto: &userOptionalInfo.CreateUserOptionalInfoRequest{
				Email: "test@example.com",
			},
			expectError: false,
		},
		{
			name: "Empty Email",
			dto: &userOptionalInfo.CreateUserOptionalInfoRequest{
				Name:           "John Doe",
				HashedPassword: "hashedpass123",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateDto(tt.dto)
			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
