package identity

//
//import (
//	dto "api/internal/domains/identity/dto/common"
//	"api/internal/domains/identity/dto/customer"
//	values "api/internal/domains/identity/values"
//	"api/internal/libs/validators"
//	"bytes"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestDecodeRequestBody(t *testing.T) {
//	tests := []struct {
//		name           string
//		jsonBody       string
//		expectError    bool
//		expectedValues *RegularCustomerRegistrationRequestDto
//	}{
//		{
//			name: "Valid Input",
//			jsonBody: `{
//"age": 25,
//  "first_name": "John",
//  "has_consent_to_email_marketing": true,
//  "has_consent_to_sms": true,
//  "last_name": "Doe",
//  "phone_number": "+14034661009",
//  "waivers": [
//    {
//      "is_waiver_signed": true,
//      "waiver_url": "https://example.com/waiver1"
//    }
//  ]
//}`,
//			expectError: false,
//			expectedValues: &RegularCustomerRegistrationRequestDto{
//				RegistrationBaseRequestDto: RegistrationBaseRequestDto{
//					CustomerWaiversSigningDto: []customer.WaiverSigningRequestDto{
//						{
//							IsWaiverSigned: true,
//							WaiverURL:      "https://example.com/waiver1",
//						},
//					},
//					UserNecessaryInfoRequestDto: dto.UserNecessaryInfoRequestDto{
//						Age:       25,
//						FirstName: "John",
//						LastName:  "Doe",
//					},
//				},
//				PhoneNumber: func() string {
//					phone := "+14034661009"
//					return phone
//				}(),
//				HasConsentToSmS:            true,
//				HasConsentToEmailMarketing: true,
//			},
//		},
//		{
//			name: "Invalid JSON - Missing closing brace",
//			jsonBody: `{
//				"waivers": [
//					{
//						"is_waiver_signed": true,
//						"waiver_url": "https://example.com/waiver1"
//					}
//				],
//				"age": 25,
//				"first_name": "John"
//			`,
//			expectError: true,
//		},
//		{
//			name: "Missing Waivers",
//			jsonBody: `{
//				"age": 25,
//				"first_name": "John",
//				"last_name": "Doe"
//			}`,
//			expectError: false, // Expecting validation error for missing waivers
//			expectedValues: &RegularCustomerRegistrationRequestDto{
//				RegistrationBaseRequestDto: RegistrationBaseRequestDto{
//					CustomerWaiversSigningDto: nil,
//					UserNecessaryInfoRequestDto: dto.UserNecessaryInfoRequestDto{
//						Age:       25,
//						FirstName: "John",
//						LastName:  "Doe",
//					},
//				},
//				PhoneNumber: func() string {
//					phone := ""
//					return phone
//				}(),
//				HasConsentToSmS:            false,
//				HasConsentToEmailMarketing: false,
//			},
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			reqBody := bytes.NewReader([]byte(tc.jsonBody))
//			var target RegularCustomerRegistrationRequestDto
//
//			err := validators.ParseJSON(reqBody, &target)
//			if tc.expectError {
//				assert.NotNil(t, err)
//			} else {
//				assert.Nil(t, err)
//				if tc.expectedValues != nil {
//					assert.Equal(t, tc.expectedValues.CustomerWaiversSigningDto, target.CustomerWaiversSigningDto)
//					assert.Equal(t, tc.expectedValues.Age, target.Age)
//					assert.Equal(t, tc.expectedValues.FirstName, target.FirstName)
//					assert.Equal(t, tc.expectedValues.LastName, target.LastName)
//					assert.Equal(t, tc.expectedValues.PhoneNumber, target.PhoneNumber)
//					assert.Equal(t, tc.expectedValues.HasConsentToSmS, target.HasConsentToSmS)
//					assert.Equal(t, tc.expectedValues.HasConsentToEmailMarketing, target.HasConsentToEmailMarketing)
//				}
//			}
//		})
//	}
//}
//
//func TestCustomerRegistrationDto_Validation(t *testing.T) {
//	tests := []struct {
//		name                 string
//		dto                  *RegularCustomerRegistrationRequestDto
//		expectErr            bool
//		expectedErrorMessage string
//	}{
//		{
//			name: "Valid DTO",
//			dto: &RegularCustomerRegistrationRequestDto{
//				HasConsentToEmailMarketing: true,
//				HasConsentToSmS:            true,
//				PhoneNumber:                "+14034661009",
//				RegistrationBaseRequestDto: RegistrationBaseRequestDto{
//					CustomerWaiversSigningDto: []customer.WaiverSigningRequestDto{
//						{
//							IsWaiverSigned: true,
//							WaiverURL:      "https://example.com/waiver1",
//						},
//					},
//					UserNecessaryInfoRequestDto: dto.UserNecessaryInfoRequestDto{
//						Age:       25,
//						FirstName: "John",
//						LastName:  "Doe",
//					},
//				},
//			},
//			expectErr: false,
//		},
//		{
//			name: "Invalid number",
//			dto: &RegularCustomerRegistrationRequestDto{
//				HasConsentToEmailMarketing: true,
//				HasConsentToSmS:            true,
//				PhoneNumber:                "+1ewre4034661009",
//				RegistrationBaseRequestDto: RegistrationBaseRequestDto{
//					CustomerWaiversSigningDto: []customer.WaiverSigningRequestDto{
//						{
//							IsWaiverSigned: true,
//							WaiverURL:      "https://example.com/waiver1",
//						},
//					},
//					UserNecessaryInfoRequestDto: dto.UserNecessaryInfoRequestDto{
//						Age:       25,
//						FirstName: "John",
//						LastName:  "Doe",
//					},
//				},
//			},
//			expectErr:            true,
//			expectedErrorMessage: "phone_number: must be a valid phone number",
//		},
//		{
//			name: "Missing Waivers",
//			dto: &RegularCustomerRegistrationRequestDto{
//				RegistrationBaseRequestDto: RegistrationBaseRequestDto{
//					UserNecessaryInfoRequestDto: dto.UserNecessaryInfoRequestDto{
//						Age:       25,
//						FirstName: "John",
//						LastName:  "Doe",
//					},
//				},
//			},
//			expectErr:            true,
//			expectedErrorMessage: "waivers: required",
//		},
//		{
//			name: "Missing First Name",
//			dto: &RegularCustomerRegistrationRequestDto{
//				RegistrationBaseRequestDto: RegistrationBaseRequestDto{
//					CustomerWaiversSigningDto: []customer.WaiverSigningRequestDto{
//						{
//							IsWaiverSigned: true,
//							WaiverURL:      "https://example.com/waiver1",
//						},
//					},
//					UserNecessaryInfoRequestDto: dto.UserNecessaryInfoRequestDto{
//						Age:      25,
//						LastName: "Doe",
//					},
//				},
//			},
//			expectErr:            true,
//			expectedErrorMessage: "first_name: required",
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			err := validators.ValidateDto(tc.dto)
//			if tc.expectErr {
//				assert.NotNil(t, err)
//				if tc.expectedErrorMessage != "" {
//					assert.Contains(t, err.Message, tc.expectedErrorMessage)
//				}
//			} else {
//				assert.Nil(t, err)
//			}
//		})
//	}
//}
//
//func TestCustomerRegistrationDto_ToCreateRegularCustomerValueObject(t *testing.T) {
//	tests := []struct {
//		name           string
//		dto            *RegularCustomerRegistrationRequestDto
//		email          string
//		expectError    bool
//		expectedValues *values.RegularCustomerRegistrationRequestInfo
//	}{
//		{
//			name: "Valid Input",
//			dto: &RegularCustomerRegistrationRequestDto{
//				PhoneNumber:                "+41034661009",
//				HasConsentToSmS:            true,
//				HasConsentToEmailMarketing: true,
//				RegistrationBaseRequestDto: RegistrationBaseRequestDto{
//					CustomerWaiversSigningDto: []customer.WaiverSigningRequestDto{
//						{
//							IsWaiverSigned: true,
//							WaiverURL:      "https://example.com/waiver1",
//						},
//					},
//					UserNecessaryInfoRequestDto: dto.UserNecessaryInfoRequestDto{
//						Age:       25,
//						FirstName: "John",
//						LastName:  "Doe",
//					},
//				},
//			},
//			email:       "john.doe@example.com",
//			expectError: false,
//			expectedValues: &values.RegularCustomerRegistrationRequestInfo{
//				Phone: "+41034661009",
//				UserRegistrationRequestNecessaryInfo: values.UserRegistrationRequestNecessaryInfo{
//					Age:       25,
//					FirstName: "John",
//					LastName:  "Doe",
//				},
//				Email: "john.doe@example.com",
//				Waivers: []values.CustomerWaiverSigning{
//					{
//						IsWaiverSigned: true,
//						WaiverUrl:      "https://example.com/waiver1",
//					},
//				},
//			},
//		},
//		{
//			name: "Invalid DTO - Missing Waivers",
//			dto: &RegularCustomerRegistrationRequestDto{
//				RegistrationBaseRequestDto: RegistrationBaseRequestDto{
//					UserNecessaryInfoRequestDto: dto.UserNecessaryInfoRequestDto{
//						Age:       25,
//						FirstName: "John",
//						LastName:  "Doe",
//					},
//				},
//			},
//			email:       "john.doe@example.com",
//			expectError: true,
//		},
//		{
//			name: "Invalid DTO - Missing phone number",
//			dto: &RegularCustomerRegistrationRequestDto{
//				HasConsentToSmS:            true,
//				HasConsentToEmailMarketing: true,
//				RegistrationBaseRequestDto: RegistrationBaseRequestDto{
//					CustomerWaiversSigningDto: []customer.WaiverSigningRequestDto{
//						{
//							IsWaiverSigned: true,
//							WaiverURL:      "https://example.com/waiver1",
//						},
//					},
//					UserNecessaryInfoRequestDto: dto.UserNecessaryInfoRequestDto{
//						Age:       25,
//						FirstName: "John",
//						LastName:  "Doe",
//					},
//				},
//			},
//			email:       "john.doe@example.com",
//			expectError: true,
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			vo, err := tc.dto.ToCreateRegularCustomerValueObject(tc.email)
//			if tc.expectError {
//				assert.NotNil(t, err)
//			} else {
//				assert.Nil(t, err)
//				if tc.expectedValues != nil {
//					assert.Equal(t, tc.expectedValues.UserRegistrationRequestNecessaryInfo, vo.UserRegistrationRequestNecessaryInfo)
//					assert.Equal(t, tc.expectedValues.Email, vo.Email)
//					assert.Equal(t, tc.expectedValues.Waivers, vo.Waivers)
//				}
//			}
//		})
//	}
//}
//
//func TestCustomerRegistrationDto_ToCreateChildValueObject(t *testing.T) {
//	tests := []struct {
//		name           string
//		dto            ChildRegistrationRequestDto
//		parentEmail    string
//		expectError    bool
//		expectedValues *values.ChildRegistrationRequestInfo
//	}{
//		{
//			name: "Valid Input",
//			dto: ChildRegistrationRequestDto{
//				RegistrationBaseRequestDto: RegistrationBaseRequestDto{
//					CustomerWaiversSigningDto: []customer.WaiverSigningRequestDto{
//						{
//							IsWaiverSigned: true,
//							WaiverURL:      "https://example.com/waiver1",
//						},
//					},
//					UserNecessaryInfoRequestDto: dto.UserNecessaryInfoRequestDto{
//						Age:       10,
//						FirstName: "Alice",
//						LastName:  "Doe",
//					},
//				},
//			},
//			parentEmail: "john.doe@example.com",
//			expectError: false,
//			expectedValues: &values.ChildRegistrationRequestInfo{
//				UserRegistrationRequestNecessaryInfo: values.UserRegistrationRequestNecessaryInfo{
//					Age:       10,
//					FirstName: "Alice",
//					LastName:  "Doe",
//				},
//				ParentEmail: "john.doe@example.com",
//				Waivers: []values.CustomerWaiverSigning{
//					{
//						IsWaiverSigned: true,
//						WaiverUrl:      "https://example.com/waiver1",
//					},
//				},
//			},
//		},
//		{
//			name: "Invalid DTO - Missing Waivers",
//			dto: ChildRegistrationRequestDto{
//				RegistrationBaseRequestDto: RegistrationBaseRequestDto{
//					UserNecessaryInfoRequestDto: dto.UserNecessaryInfoRequestDto{
//						Age:       10,
//						FirstName: "Alice",
//						LastName:  "Doe",
//					},
//				},
//			},
//			parentEmail: "john.doe@example.com",
//			expectError: true,
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			vo, err := tc.dto.ToCreateChildValueObject(tc.parentEmail)
//			if tc.expectError {
//				assert.NotNil(t, err)
//			} else {
//				assert.Nil(t, err)
//				if tc.expectedValues != nil {
//					assert.Equal(t, tc.expectedValues.UserRegistrationRequestNecessaryInfo, vo.UserRegistrationRequestNecessaryInfo)
//					assert.Equal(t, tc.expectedValues.ParentEmail, vo.ParentEmail)
//					assert.Equal(t, tc.expectedValues.Waivers, vo.Waivers)
//				}
//			}
//		})
//	}
//}
