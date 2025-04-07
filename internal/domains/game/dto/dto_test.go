package game

import (
	"api/internal/libs/validators"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestDto_Validation(t *testing.T) {
	tests := []struct {
		name                 string
		dto                  *RequestDto
		expectErr            bool
		expectedErrorMessage string
	}{
		{
			name: "Valid DTO",
			dto: &RequestDto{
				Name:        "Go Programming Basics",
				Description: "Learn Go Programming",
			},
			expectErr: false,
		},
		{
			name: "Missing Name",
			dto: &RequestDto{
				Name:        "",
				Description: "Learn Go Programming",
			},
			expectErr:            true,
			expectedErrorMessage: "name: required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validators.ValidateDto(tc.dto)
			if tc.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
