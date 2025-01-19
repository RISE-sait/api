package tests

import (
	userOptionalInfo "api/internal/dtos/user/optionalInfo"
	"api/internal/utils/validators"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeReadUserOptionalInfoRequestBody(t *testing.T) {
	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectedValues *userOptionalInfo.GetUserOptionalInfoRequest
	}{
		{
			name: "Valid Input",
			jsonBody: `{
				"email": "test@example.com",
				"hashed_password": "hashedpass123"
			}`,
			expectError: false,
			expectedValues: &userOptionalInfo.GetUserOptionalInfoRequest{
				Email:    "test@example.com",
				Password: "hashedpass123",
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
			name: "Missing Password",
			jsonBody: `{
				"email": "test@example.com"
			}`,
			expectError: false,
			expectedValues: &userOptionalInfo.GetUserOptionalInfoRequest{
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
					assert.Equal(t, tc.expectedValues.Email, target.Email)

					assert.Equal(t, tc.expectedValues.Password, target.HashedPassword)
				}
			}
		})
	}
}

func TestValidateGetUserOptionalInfoRequest(t *testing.T) {
	tests := []struct {
		name          string
		dto           *userOptionalInfo.GetUserOptionalInfoRequest
		expectError   bool
		expectedError string
	}{
		{
			name: "Valid Input",
			dto: &userOptionalInfo.GetUserOptionalInfoRequest{
				Email:    "test@example.com",
				Password: "hashedpass123",
			},
			expectError: false,
		},
		{
			name: "Invalid Email Format",
			dto: &userOptionalInfo.GetUserOptionalInfoRequest{
				Email:    "invalid-email",
				Password: "hashedpass123",
			},
			expectError:   true,
			expectedError: "email: must be a valid email address",
		},
		{
			name: "Empty password",
			dto: &userOptionalInfo.GetUserOptionalInfoRequest{
				Email: "test@example.com",
			},
			expectError:   true,
			expectedError: "password: cannot be empty or whitespace",
		},
		{
			name: "Empty Email",
			dto: &userOptionalInfo.GetUserOptionalInfoRequest{
				Email:    "",
				Password: "hashedpass123",
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
