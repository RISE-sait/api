package waiver

import (
	"bytes"
	"testing"

	"api/internal/utils/validators"

	"github.com/stretchr/testify/assert"
)

func TestDecodeUpdateWaiverRequest(t *testing.T) {
	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectedValues *UpdateWaiverRequest
	}{
		{
			name: "Valid JSON",
			jsonBody: `{
				"email":"klintlee1@gmail.com",
				"signed_status":true
			}`,
			expectError: false,
			expectedValues: &UpdateWaiverRequest{
				Email:    "klintlee1@gmail.com",
				IsSigned: true,
			},
		},
		{
			name: "Invalid JSON: Invalid signed status",
			jsonBody: `{
				"email":"klintlee1@gmail.com",
				"signed_status":"invalid"
			}`,
			expectError: true,
		},
		{
			name: "Invalid JSON: Malformed JSON",
			jsonBody: `{
				"begin_datetime": "2025-01-01T00:00:00Z",
				"end_datetime": "2025-01-01T00:00:00Z",
				"course_id": "invalid",
				"facility_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9
			}`,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tc.jsonBody))
			var target UpdateWaiverRequest

			err := validators.DecodeRequestBody(reqBody, &target)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)

				expected := tc.expectedValues

				if expected != nil {
					assert.Equal(t, expected.Email, target.Email)
					assert.Equal(t, expected.IsSigned, target.IsSigned)
				}
			}
		})
	}
}

func TestValidateUpdateWaiverRequest(t *testing.T) {
	tests := []struct {
		name          string
		dto           UpdateWaiverRequest
		expectError   bool
		expectedError string
	}{
		{
			name: "Valid UpdateWaiverRequest",
			dto: UpdateWaiverRequest{
				Email:    "klintlee1@gmail.com",
				IsSigned: true,
			},
			expectError: false,
		},
		{
			name: "Invalid email",
			dto: UpdateWaiverRequest{
				Email:    "wefewfewfwef",
				IsSigned: true,
			},
			expectError:   true,
			expectedError: "email: must be a valid email address",
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
