package validators

import (
	"testing"

	"github.com/google/uuid"
)

func TestParseUUID(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedUUID  uuid.UUID
		isExpectedErr bool
	}{
		{
			name:          "Valid UUID",
			input:         "f47ac10b-58cc-4372-a567-0e02b2c3d479", // example valid UUID
			expectedUUID:  uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
			isExpectedErr: false,
		},
		{
			name:          "Invalid UUID (empty string)",
			input:         "",
			expectedUUID:  uuid.Nil,
			isExpectedErr: true,
		},
		{
			name:          "Invalid UUID (invalid format)",
			input:         "invalid-uuid-format",
			expectedUUID:  uuid.Nil,
			isExpectedErr: true,
		},
		{
			name:          "Invalid UUID (parsing error)",
			input:         "f47ac10b-58cc-4372-a567-0e02b2c3d479zz", // invalid due to extra characters
			expectedUUID:  uuid.Nil,
			isExpectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUUID, err := ParseUUID(tt.input)

			if gotUUID != tt.expectedUUID {
				t.Errorf("ParseUUID() = %v, want %v", gotUUID, tt.expectedUUID)
			}

			if err != nil && !tt.isExpectedErr {
				t.Errorf("ParseUUID() test error = %v", err)
			}
		})
	}
}
