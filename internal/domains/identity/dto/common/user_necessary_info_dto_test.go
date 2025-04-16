package identity

import (
	"api/internal/libs/validators"
	"strings"
	"testing"
)

func TestUserNecessaryInfoDto_Validation(t *testing.T) {
	tests := []struct {
		name           string
		input          UserBaseInfoRequestDto
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Valid Input",
			input: UserBaseInfoRequestDto{
				FirstName: "John",
				LastName:  "Doe",
				DOB:       "2000-01-01",
			},
			expectedError:  false,
			expectedErrMsg: "",
		},
		{
			name: "Missing First Name",
			input: UserBaseInfoRequestDto{
				FirstName: "",
				LastName:  "Doe",
				DOB:       "2000-01-01",
			},
			expectedError:  true,
			expectedErrMsg: "first_name: required",
		},
		{
			name: "Whitespace First Name",
			input: UserBaseInfoRequestDto{
				FirstName: "   ",
				LastName:  "Doe",
				DOB:       "2000-01-01",
			},
			expectedError:  true,
			expectedErrMsg: "first_name: cannot be empty or whitespace",
		},
		{
			name: "Missing Last Name",
			input: UserBaseInfoRequestDto{
				FirstName: "John",
				LastName:  "",
				DOB:       "2000-01-01",
			},
			expectedError:  true,
			expectedErrMsg: "last_name: required",
		},
		{
			name: "Whitespace Last Name",
			input: UserBaseInfoRequestDto{
				FirstName: "John",
				LastName:  "   ",
				DOB:       "2000-01-01",
			},
			expectedError:  true,
			expectedErrMsg: "last_name: cannot be empty or whitespace",
		},
		{
			name: "Valid Gender",
			input: UserBaseInfoRequestDto{
				FirstName: "John",
				LastName:  "Doe",
				DOB:       "2000-01-01",
				Gender:    "M",
			},
			expectedError:  false,
			expectedErrMsg: "",
		},
		{
			name: "Invalid Gender",
			input: UserBaseInfoRequestDto{
				FirstName: "John",
				LastName:  "Doe",
				DOB:       "2000-01-01",
				Gender:    "Male",
			},
			expectedError:  true,
			expectedErrMsg: "gender: must be one of the following values: M F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateDto(&tt.input)

			if tt.expectedError && err == nil {
				t.Errorf("Expected an error, but got none")
			} else if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			// Check specific error message
			if tt.expectedError {
				if err == nil || !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Expected error containing '%s', but got: %v", tt.expectedErrMsg, err)
				}
			}
		})
	}
}
