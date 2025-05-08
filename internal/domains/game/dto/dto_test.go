package game

import (
	"api/internal/libs/validators"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestRequestDto_Validation tests the validation logic for the RequestDto struct.
// It uses a table-driven approach to cover various edge cases and input scenarios.
func TestRequestDto_Validation(t *testing.T) {
	now := time.Now() // Capture current time for consistent test input

	tests := []struct {
		name      string      // Describes the test case
		dto       *RequestDto // The input DTO being validated
		expectErr bool        // Whether validation is expected to fail
	}{
		{
			name: "Valid DTO",
			dto: &RequestDto{
				HomeTeamID: uuid.New(),
				AwayTeamID: uuid.New(),
				HomeScore:  intPtr(3),
				AwayScore:  intPtr(2),
				StartTime:  now,
				EndTime:    timePtr(now.Add(90 * time.Minute)),
				LocationID: uuid.New(),
				Status:     "scheduled",
			},
			expectErr: false, // All fields are valid, so validation should pass
		},
		{
			name: "Missing required fields",
			dto: &RequestDto{
				AwayTeamID: uuid.New(),
				Status:     "invalid_status", // Invalid status, missing required fields
			},
			expectErr: true,
		},
		{
			name: "Home and Away team are the same",
			dto: &RequestDto{
				HomeTeamID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				AwayTeamID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), // Same as HomeTeamID
				StartTime:  now,
				LocationID: uuid.New(),
				Status:     "completed",
			},
			expectErr: true,
		},
	}

	// Run each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validators.ValidateDto(tc.dto)
			if tc.expectErr {
				assert.NotNil(t, err) // Expecting an error
			} else {
				assert.Nil(t, err) // Expecting no error
			}
		})
	}
}

// intPtr returns a pointer to the given int32 value.
// Useful for setting optional fields in DTOs.
func intPtr(i int32) *int32 {
	return &i
}

// timePtr returns a pointer to the given time.Time value.
// Used to set optional end times in tests.
func timePtr(t time.Time) *time.Time {
	return &t
}
