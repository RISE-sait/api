package customer

import (
	identity "api/internal/domains/identity/dto/common"
	"api/internal/libs/validators"
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeRequestBody(t *testing.T) {
	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectedValues *ParentRegistrationRequestDto
	}{
		{
			name: "Valid Input",
			jsonBody: `{
				"first_name": "John",
				"last_name": "Doe",
				"age": 30,
				"country_code": "US",
				"phone_number": "+15141234567",
				"has_consent_to_sms": true,
				"has_consent_to_email_marketing": true
			}`,
			expectError: false,
			expectedValues: &ParentRegistrationRequestDto{
				UserBaseInfoRequestDto: identity.UserBaseInfoRequestDto{
					FirstName:   "John",
					LastName:    "Doe",
					Age:         30,
					CountryCode: "US",
				},
				PhoneNumber:                "+15141234567",
				HasConsentToSmS:            true,
				HasConsentToEmailMarketing: true,
			},
		},
		{
			name: "Invalid JSON - Missing closing brace",
			jsonBody: `{
				"first_name": "John",
				"last_name": "Doe",
				"age": 30,
				"country_code": "US",
				"phone_number": "+15141234567",
				"has_consent_to_sms": true,
				"has_consent_to_email_marketing": true
			`,
			expectError: true,
		},
		{
			name: "Missing First Name",
			jsonBody: `{
				"last_name": "Doe",
				"age": 30,
				"country_code": "US",
				"phone_number": "+15141234567",
				"has_consent_to_sms": true,
				"has_consent_to_email_marketing": true
			}`,
			expectError: false,
		},
		{
			name: "Wrong age type",
			jsonBody: `{
				"first_name": "John",
				"last_name": "Doe",
				"age": "30",
				"country_code": "US",
				"phone_number": "+15141234567"
			}`,
			expectError: true,
		},
		{
			name: "Wrong consent type",
			jsonBody: `{
				"first_name": "John",
				"last_name": "Doe",
				"age": 30,
				"country_code": "US",
				"phone_number": "+15141234567",
				"has_consent_to_sms": "true",
				"has_consent_to_email_marketing": "false"
			}`,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tc.jsonBody))
			var target ParentRegistrationRequestDto

			err := validators.ParseJSON(reqBody, &target)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if tc.expectedValues != nil {
					assert.Equal(t, tc.expectedValues.FirstName, target.FirstName)
					assert.Equal(t, tc.expectedValues.LastName, target.LastName)
					assert.Equal(t, tc.expectedValues.Age, target.Age)
					assert.Equal(t, tc.expectedValues.CountryCode, target.CountryCode)
					assert.Equal(t, tc.expectedValues.PhoneNumber, target.PhoneNumber)
					assert.Equal(t, tc.expectedValues.HasConsentToSmS, target.HasConsentToSmS)
					assert.Equal(t, tc.expectedValues.HasConsentToEmailMarketing, target.HasConsentToEmailMarketing)
				}
			}
		})
	}
}

// Validate Dto

func TestValidRequestDto(t *testing.T) {
	dto := ParentRegistrationRequestDto{
		UserBaseInfoRequestDto: identity.UserBaseInfoRequestDto{
			FirstName:   "John",
			LastName:    "Doe",
			Age:         30,
			CountryCode: "US",
		},
		PhoneNumber:                "+15141234567",
		HasConsentToSmS:            true,
		HasConsentToEmailMarketing: true,
	}

	email := "john.doe@example.com"
	createRequestDto, err := dto.ToParent(email)

	assert.Nil(t, err)

	assert.Equal(t, createRequestDto.FirstName, "John")
	assert.Equal(t, createRequestDto.LastName, "Doe")
	assert.Equal(t, createRequestDto.Age, int32(30))
	assert.Equal(t, createRequestDto.CountryCode, "US")
	assert.Equal(t, createRequestDto.Phone, "+15141234567")
	assert.Equal(t, createRequestDto.HasConsentToSms, true)
	assert.Equal(t, createRequestDto.HasConsentToEmailMarketing, true)
	assert.Equal(t, createRequestDto.Email, email)
}

func TestMissingFirstNameRequestDto(t *testing.T) {
	dto := ParentRegistrationRequestDto{
		UserBaseInfoRequestDto: identity.UserBaseInfoRequestDto{
			LastName:    "Doe",
			Age:         30,
			CountryCode: "US",
		},
		PhoneNumber:                "+15141234567",
		HasConsentToSmS:            true,
		HasConsentToEmailMarketing: true,
	}

	email := "john.doe@example.com"
	_, err := dto.ToParent(email)

	assert.NotNil(t, err)
	assert.Equal(t, err.Message, "first_name: required")
	assert.Equal(t, err.HTTPCode, http.StatusBadRequest)
}

func TestInvalidPhoneRequestDto(t *testing.T) {
	dto := ParentRegistrationRequestDto{
		UserBaseInfoRequestDto: identity.UserBaseInfoRequestDto{
			FirstName:   "John",
			LastName:    "Doe",
			Age:         30,
			CountryCode: "US",
		},
		PhoneNumber:                "+x x 15141234567",
		HasConsentToSmS:            true,
		HasConsentToEmailMarketing: true,
	}

	email := "john.doe@example.com"
	_, err := dto.ToParent(email)

	assert.NotNil(t, err)

	assert.Contains(t, err.Message, "must be a valid phone number")
}

func TestInvalidEmailRequestDto(t *testing.T) {
	dto := ParentRegistrationRequestDto{
		UserBaseInfoRequestDto: identity.UserBaseInfoRequestDto{
			FirstName:   "John",
			LastName:    "Doe",
			Age:         30,
			CountryCode: "US",
		},
		PhoneNumber:                "+15141234567",
		HasConsentToSmS:            true,
		HasConsentToEmailMarketing: true,
	}

	email := "john.doeismymum.com"
	_, err := dto.ToParent(email)

	assert.NotNil(t, err)

	assert.Equal(t, err.Message, "Invalid email format")
}
