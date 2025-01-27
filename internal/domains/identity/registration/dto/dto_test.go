package dto

import (
	"api/internal/libs/validators"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeReadUserOptionalInfoRequestBody(t *testing.T) {
	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectedValues *CreateUserRequest
	}{
		{
			name: "Valid Input",
			jsonBody: `{
				"email": "test@example.com",
				"password": "hashedpass123"
			}`,
			expectError: false,
			expectedValues: &CreateUserRequest{
				Email:    "test@example.com",
				Password: "hashedpass123",
			},
		},
		{
			name: "Invalid Email",
			jsonBody: `{
				"name": "John Doe",
				"email": "invalid-email",
				"password": "hashedpass123"
			}`,
			expectError: false,
		},
		{
			name: "Missing Password",
			jsonBody: `{
				"email": "test@example.com"
			}`,
			expectError: false,
			expectedValues: &CreateUserRequest{
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
			var target CreateUserRequest

			err := validators.ParseJSON(reqBody, &target)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if tc.expectedValues != nil {
					assert.Equal(t, tc.expectedValues.Email, target.Email)

					assert.Equal(t, tc.expectedValues.Password, target.Password)
				}
			}
		})
	}
}
