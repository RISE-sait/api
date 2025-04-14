package staff

import (
	values "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"github.com/google/uuid"
	"net/http"
	"testing"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestRequestDto_ToUpdateRequestValues(t *testing.T) {
	validUUID := uuid.New()
	invalidUUID := "not-a-uuid"

	tests := []struct {
		name          string
		idStr         string
		dto           RequestDto
		expected      values.UpdateValues
		expectError   bool
		expectedError *errLib.CommonError
	}{
		{
			name:  "Valid Input",
			idStr: validUUID.String(),
			dto:   RequestDto{IsActive: boolPtr(true), RoleName: "Coach"},
			expected: values.UpdateValues{
				ID:       validUUID,
				IsActive: true,
				RoleName: "Coach",
			},
			expectError: false,
		},
		{
			name:  "Valid Input",
			idStr: validUUID.String(),
			dto:   RequestDto{IsActive: boolPtr(false), RoleName: "Coach"},
			expected: values.UpdateValues{
				ID:       validUUID,
				IsActive: false,
				RoleName: "Coach",
			},
			expectError: false,
		},
		{
			name:        "Invalid UUID",
			idStr:       invalidUUID,
			dto:         RequestDto{IsActive: boolPtr(true), RoleName: "Coach"},
			expectError: true,
			expectedError: &errLib.CommonError{
				Message:  "invalid UUID: not-a-uuid, error: invalid UUID length: 10",
				HTTPCode: http.StatusBadRequest,
			},
		},
		{
			name:        "Missing IsActive",
			idStr:       validUUID.String(),
			dto:         RequestDto{RoleName: "Coach"}, // IsActive missing
			expectError: true,
			expectedError: &errLib.CommonError{
				Message:  "is_active: required",
				HTTPCode: http.StatusBadRequest,
			},
		},
		{
			name:        "Missing RoleName",
			idStr:       validUUID.String(),
			dto:         RequestDto{IsActive: boolPtr(true)}, // RoleName missing
			expectError: true,
			expectedError: &errLib.CommonError{
				Message:  "role_name: required",
				HTTPCode: http.StatusBadRequest,
			},
		},
		{
			name:        "Empty RoleName",
			idStr:       validUUID.String(),
			dto:         RequestDto{IsActive: boolPtr(true), RoleName: ""},
			expectError: true,
			expectedError: &errLib.CommonError{
				Message:  "role_name: required",
				HTTPCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.dto.ToUpdateRequestValues(tt.idStr)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
					return
				}
				if tt.expectedError != nil {
					if err.Message != tt.expectedError.Message {
						t.Errorf("Expected error message '%s', got '%s'", tt.expectedError.Message, err.Message)
					}
					if err.HTTPCode != tt.expectedError.HTTPCode {
						t.Errorf("Expected status code %d, got %d", tt.expectedError.HTTPCode, err.HTTPCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result.ID != tt.expected.ID {
					t.Errorf("Expected ID %s, got %s", tt.expected.ID, result.ID)
				}
				if result.IsActive != tt.expected.IsActive {
					t.Errorf("Expected IsActive %t, got %t", tt.expected.IsActive, result.IsActive)
				}
				if result.RoleName != tt.expected.RoleName {
					t.Errorf("Expected RoleName %s, got %s", tt.expected.RoleName, result.RoleName)
				}
			}
		})
	}
}
