package customer

import (
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"net/http"
	"testing"
)

func TestCustomerWaiverSigningDto_ToValueObjects(t *testing.T) {
	tests := []struct {
		name          string
		input         WaiverSigningRequestDto
		expectedValue *identity.CustomerWaiverSigning
		expectedError *errLib.CommonError
	}{
		{
			name: "Valid Input",
			input: WaiverSigningRequestDto{
				WaiverUrl:      "https://example.com/waiver",
				IsWaiverSigned: true,
			},
			expectedValue: &identity.CustomerWaiverSigning{
				WaiverUrl:      "https://example.com/waiver",
				IsWaiverSigned: true,
			},
			expectedError: nil,
		},
		{
			name: "Invalid URL",
			input: WaiverSigningRequestDto{
				WaiverUrl:      "invalid-url",
				IsWaiverSigned: true,
			},
			expectedValue: nil,
			expectedError: errLib.New("waiver_url: must be a valid URL", http.StatusBadRequest),
		},
		{
			name: "Missing Waiver URL",
			input: WaiverSigningRequestDto{
				WaiverUrl:      "",
				IsWaiverSigned: true,
			},
			expectedValue: nil,
			expectedError: errLib.New("waiver_url: required", http.StatusBadRequest),
		},
		{
			name: "IsWaiverSigned Default Value",
			input: WaiverSigningRequestDto{
				WaiverUrl:      "https://example.com/waiver",
				IsWaiverSigned: false, // Default value
			},
			expectedValue: &identity.CustomerWaiverSigning{
				WaiverUrl:      "https://example.com/waiver",
				IsWaiverSigned: false,
			},
			expectedError: nil,
		},
		{
			name: "Missing waiver signed value",
			input: WaiverSigningRequestDto{
				WaiverUrl: "https://example.com/waiver",
			},
			expectedValue: &identity.CustomerWaiverSigning{
				WaiverUrl: "https://example.com/waiver",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.ToValueObjects()

			// Check if the error matches the expected error
			if err != nil && tt.expectedError == nil {
				t.Errorf("Expected no error, but got: %v", err)
			} else if err == nil && tt.expectedError != nil {
				t.Errorf("Expected error: %v, but got none", tt.expectedError)
			} else if err != nil && tt.expectedError != nil {
				if err.Message != tt.expectedError.Message || err.HTTPCode != tt.expectedError.HTTPCode {
					t.Errorf("Expected error: %v, but got: %v", tt.expectedError, err)
				}
			}

			// Check if the result matches the expected value
			if result != nil && tt.expectedValue == nil {
				t.Errorf("Expected no value, but got: %v", result)
			} else if result == nil && tt.expectedValue != nil {
				t.Errorf("Expected value: %v, but got none", tt.expectedValue)
			} else if result != nil && tt.expectedValue != nil {
				if result.WaiverUrl != tt.expectedValue.WaiverUrl || result.IsWaiverSigned != tt.expectedValue.IsWaiverSigned {
					t.Errorf("Expected value: %v, but got: %v", tt.expectedValue, result)
				}
			}
		})
	}
}
