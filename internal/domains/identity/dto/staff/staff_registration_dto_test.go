package staff

import (
	"api/internal/domains/identity/values"
	staffValues "api/internal/domains/user/values/staff"
	"api/internal/libs/validators"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaffRegistrationRequestDto_Validation(t *testing.T) {
	tests := []struct {
		name                 string
		dto                  *RegistrationRequestDto
		expectErr            bool
		expectedErrorMessage string
	}{
		{
			name: "Valid DTO",
			dto: &RegistrationRequestDto{
				HubSpotID: "12345",
				RoleName:  "Manager",
				IsActive:  true,
			},
			expectErr: false,
		},
		{
			name: "Missing HubSpotID",
			dto: &RegistrationRequestDto{
				RoleName: "Manager",
				IsActive: true,
			},
			expectErr:            true,
			expectedErrorMessage: "hubspot_id: required",
		},
		{
			name: "Missing RoleName",
			dto: &RegistrationRequestDto{
				HubSpotID: "12345",
				IsActive:  true,
			},
			expectErr:            true,
			expectedErrorMessage: "role_name: required",
		},
		{
			name: "Missing IsActive",
			dto: &RegistrationRequestDto{
				HubSpotID: "12345",
				RoleName:  "Manager",
			},
			expectErr:            true,
			expectedErrorMessage: "is_active: required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validators.ValidateDto(tc.dto)
			if tc.expectErr {
				assert.NotNil(t, err)
				if tc.expectedErrorMessage != "" {
					assert.Contains(t, err.Message, tc.expectedErrorMessage)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestStaffRegistrationRequestDto_ToDetails(t *testing.T) {
	tests := []struct {
		name           string
		dto            *RegistrationRequestDto
		expectError    bool
		expectedValues *identity.StaffRegistrationRequestInfo
	}{
		{
			name: "Valid Input",
			dto: &RegistrationRequestDto{
				HubSpotID: "12345",
				RoleName:  "Manager",
				IsActive:  true,
			},
			expectError: false,
			expectedValues: &identity.StaffRegistrationRequestInfo{
				HubSpotID: "12345",
				Details: staffValues.Details{
					RoleName: "Manager",
					IsActive: true,
				},
			},
		},
		{
			name: "Invalid DTO - Missing RoleName",
			dto: &RegistrationRequestDto{
				HubSpotID: "12345",
				IsActive:  true,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vo, err := tc.dto.ToDetails()
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if tc.expectedValues != nil {
					assert.Equal(t, tc.expectedValues.HubSpotID, vo.HubSpotID)
					assert.Equal(t, tc.expectedValues.Details.RoleName, vo.Details.RoleName)
					assert.Equal(t, tc.expectedValues.Details.IsActive, vo.Details.IsActive)
				}
			}
		})
	}
}
