package users

import (
	"api/internal/utils/validators"
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDecodeUpdateUserEmailRequestBody(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectedValues *UpdateUserEmailRequest
	}{
		{
			name: "Valid Input",
			jsonBody: `{
				"id": "` + validID.String() + `",
				"email": "test@example.com"
			}`,
			expectError: false,
			expectedValues: &UpdateUserEmailRequest{
				Id:    validID,
				Email: "test@example.com",
			},
		},
		{
			name: "Invalid UUID",
			jsonBody: `{
				"id": "invalid-uuid",
				"email": "test@example.com"
			}`,
			expectError: true,
		},
		{
			name: "Missing Email",
			jsonBody: `{
				"id": "` + validID.String() + `",
				"email": ""
			}`,
			expectError: false,
		},
		{
			name: "Invalid JSON",
			jsonBody: `{
				"id": "` + validID.String() + `",
				email: test@example.com
			}`,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tc.jsonBody))
			var target UpdateUserEmailRequest

			err := validators.DecodeRequestBody(reqBody, &target)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if tc.expectedValues != nil {
					assert.Equal(t, tc.expectedValues.Id, target.Id)
					assert.Equal(t, tc.expectedValues.Email, target.Email)
				}
			}
		})
	}
}

func TestValidateUpdateUserEmailRequest(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name          string
		dto           *UpdateUserEmailRequest
		expectError   bool
		expectedError string
	}{
		{
			name: "Valid Input",
			dto: &UpdateUserEmailRequest{
				Id:    validID,
				Email: "test@example.com",
			},
			expectError: false,
		},
		{
			name: "Invalid Email Format",
			dto: &UpdateUserEmailRequest{
				Id:    validID,
				Email: "invalid-email",
			},
			expectError: true,
		},
		{
			name: "Empty Email",
			dto: &UpdateUserEmailRequest{
				Id:    validID,
				Email: "",
			},
			expectError: true,
		},
		{
			name: "Empty UUID",
			dto: &UpdateUserEmailRequest{
				Id:    uuid.Nil,
				Email: "test@example.com",
			},
			expectError: true,
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
