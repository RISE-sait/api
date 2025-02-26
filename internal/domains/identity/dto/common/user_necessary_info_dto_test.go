package identity

import (
	"api/internal/libs/validators"
	"testing"
)

func TestUserNecessaryInfoDto_Validation(t *testing.T) {
	tests := []struct {
		name          string
		input         UserNecessaryInfoDto
		expectedError bool
	}{
		{
			name: "Valid Input",
			input: UserNecessaryInfoDto{
				FirstName: "John",
				LastName:  "Doe",
				Age:       25,
			},
			expectedError: false,
		},
		{
			name: "Missing First Name",
			input: UserNecessaryInfoDto{
				FirstName: "", // Missing first name
				LastName:  "Doe",
				Age:       25,
			},
			expectedError: true,
		},
		{
			name: "Whitespace First Name",
			input: UserNecessaryInfoDto{
				FirstName: "   ", // Whitespace first name
				LastName:  "Doe",
				Age:       25,
			},
			expectedError: true,
		},
		{
			name: "Missing Last Name",
			input: UserNecessaryInfoDto{
				FirstName: "John",
				LastName:  "", // Missing last name
				Age:       25,
			},
			expectedError: true,
		},
		{
			name: "Whitespace Last Name",
			input: UserNecessaryInfoDto{
				FirstName: "John",
				LastName:  "   ", // Whitespace last name
				Age:       25,
			},
			expectedError: true,
		},
		{
			name: "Invalid Age (Zero)",
			input: UserNecessaryInfoDto{
				FirstName: "John",
				LastName:  "Doe",
				Age:       0, // Invalid age (zero)
			},
			expectedError: true,
		},
		{
			name: "Invalid Age (Negative)",
			input: UserNecessaryInfoDto{
				FirstName: "John",
				LastName:  "Doe",
				Age:       -5, // Invalid age (negative)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateDto(&tt.input)

			// Check if the error matches the expected result
			if tt.expectedError && err == nil {
				t.Errorf("Expected an error, but got none")
			} else if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
		})
	}
}
