package tests

import (
	userOptionalInfo "api/internal/dtos/user/optionalInfo"
	"api/internal/utils/validators"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeUpdateUsernameRequestBody(t *testing.T) {
	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectedValues *userOptionalInfo.UpdateUserNameRequest
	}{
		{
			name: "Valid Input",
			jsonBody: `{
				"name": "John Doe",
				"email": "test@example.com"
			}`,
			expectError: false,
			expectedValues: &userOptionalInfo.UpdateUserNameRequest{
				Name:  "John Doe",
				Email: "test@example.com",
			},
		},
		{
			name: "Invalid name",
			jsonBody: `{
				"name": "",
				"email": "invalid-email"
			}`,
			expectError: false,
			expectedValues: &userOptionalInfo.UpdateUserNameRequest{
				Name:  "",
				Email: "invalid-email",
			},
		},
		{
			name: "Invalid Email",
			jsonBody: `{
				"name": "John Doe",
				"email": "invalid-email"
			}`,
			expectError: false,
			expectedValues: &userOptionalInfo.UpdateUserNameRequest{
				Name:  "John Doe",
				Email: "invalid-email",
			},
		},
		{
			name: "Missing Optional Fields",
			jsonBody: `{
				"email": "test@example.com"
			}`,
			expectError: false,
			expectedValues: &userOptionalInfo.UpdateUserNameRequest{
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

					if tc.expectedValues.Name != "" {
						assert.Equal(t, tc.expectedValues.Name, target.Name)
					}

					if tc.expectedValues.Email != "" {
						assert.Equal(t, tc.expectedValues.Email, target.Email)
					}
				}
			}
		})
	}
}

func TestValidateUpdateUsernameRequest(t *testing.T) {
	tests := []struct {
		name          string
		dto           *userOptionalInfo.UpdateUserNameRequest
		expectError   bool
		expectedError string
	}{
		{
			name: "Valid Input",
			dto: &userOptionalInfo.UpdateUserNameRequest{
				Name:  "John Doe",
				Email: "test@example.com",
			},
			expectError: false,
		},
		{
			name: "Invalid Email Format",
			dto: &userOptionalInfo.UpdateUserNameRequest{
				Name:  "John Doe",
				Email: "invalid-email",
			},
			expectError:   true,
			expectedError: "email: must be a valid email address",
		},
		{
			name: "Empty name",
			dto: &userOptionalInfo.UpdateUserNameRequest{
				Name:  "          ",
				Email: "test@example.com",
			},
			expectError:   true,
			expectedError: "name: cannot be empty or whitespace",
		},
		{
			name: "Empty Email",
			dto: &userOptionalInfo.UpdateUserNameRequest{
				Name: "John Doe",
			},
			expectError:   true,
			expectedError: "email: must be a valid email address",
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
