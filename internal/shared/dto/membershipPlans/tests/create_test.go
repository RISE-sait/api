package tests

// import (
// 	"bytes"
// 	"testing"

// 	dto "api/internal/dtos/membershipPlans"
// 	"api/internal/utils/validators"
// 	db "api/sqlc"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// )

// func TestDecodeCreateScheduleRequestRequestBody(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		jsonBody       string
// 		expectError    bool
// 		expectedValues *dto.CreateMembershipPlanRequest
// 	}{
// 		{
// 			name: "Valid JSON",
// 			jsonBody: `{
// 				"membership_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"name": "Premium Membership",
// 				"price": 1999,
// 				"payment_frequency": "month",
// 				"amt_periods": 12
// 			}`,
// 			expectError: false,
// 			expectedValues: &dto.CreateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Name:             "Premium Membership",
// 				Price:            1999,
// 				PaymentFrequency: string(db.PaymentFrequencyMonth),
// 				AmtPeriods:       12,
// 			},
// 		},
// 		{
// 			name: "Valid JSON: Missing MembershipID",
// 			jsonBody: `{
// 				"name": "Premium Membership",
// 				"price": 1999,
// 				"payment_frequency": "monthly",
// 				"amt_periods": 12
// 			}`,
// 			expectError: false,
// 		},
// 		{
// 			name: "Invalid JSON: Malformed MembershipID",
// 			jsonBody: `{
// 				"membership_id": "invalid-uuid",
// 				"name": "Premium Membership",
// 				"price": 1999,
// 				"payment_frequency": "monthly",
// 				"amt_periods": 12
// 			}`,
// 			expectError: true,
// 		},
// 		{
// 			name: "Empty Name",
// 			jsonBody: `{
// 				"membership_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"name": "",
// 				"price": 1999,
// 				"payment_frequency": "monthly",
// 				"amt_periods": 12
// 			}`,
// 			expectError: false,
// 		},
// 		{
// 			name: "Missing Fields: Name and Price",
// 			jsonBody: `{
// 				"membership_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"payment_frequency": "monthly",
// 				"amt_periods": 12
// 			}`,
// 			expectError: false,
// 		},
// 		{
// 			name: "Invalid Price",
// 			jsonBody: `{
// 				"membership_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"name": "Premium Membership",
// 				"price": "invalid-price",
// 				"payment_frequency": "monthly",
// 				"amt_periods": 12
// 			}`,
// 			expectError: true,
// 		},
// 		{
// 			name: "Valid Empty PaymentFrequency",
// 			jsonBody: `{
// 				"membership_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"name": "Premium Membership",
// 				"price": 1999,
// 				"payment_frequency": "",
// 				"amt_periods": 12
// 			}`,
// 			expectError: false,
// 			expectedValues: &dto.CreateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Name:             "Premium Membership",
// 				Price:            1999,
// 				PaymentFrequency: "",
// 				AmtPeriods:       12,
// 			},
// 		},
// 		{
// 			name: "Valid Empty AmtPeriods",
// 			jsonBody: `{
// 				"membership_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"name": "Premium Membership",
// 				"price": 1999,
// 				"payment_frequency": "monthly",
// 				"amt_periods": null
// 			}`,
// 			expectError: false,
// 			expectedValues: &dto.CreateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Name:             "Premium Membership",
// 				Price:            1999,
// 				PaymentFrequency: "monthly",
// 			},
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			reqBody := bytes.NewReader([]byte(tc.jsonBody))
// 			var target dto.CreateMembershipPlanRequest

// 			err := validators.DecodeRequestBody(reqBody, &target)
// 			if tc.expectError {
// 				assert.NotNil(t, err)
// 			} else {
// 				assert.Nil(t, err)

// 				expected := tc.expectedValues

// 				if expected != nil {
// 					if expected.MembershipID != uuid.Nil {
// 						assert.Equal(t, expected.MembershipID, target.MembershipID)
// 					}

// 					if expected.Name != "" {
// 						assert.Equal(t, expected.Name, target.Name)
// 					}

// 					if expected.Price != 0 {
// 						assert.Equal(t, expected.Price, target.Price)
// 					}

// 					if expected.PaymentFrequency != "" {
// 						assert.Equal(t, expected.PaymentFrequency, target.PaymentFrequency)
// 					}

// 					if expected.AmtPeriods != 0 {
// 						assert.Equal(t, expected.AmtPeriods, target.AmtPeriods)
// 					}
// 				}
// 			}
// 		})
// 	}
// }

// func TestValidateCreateMembershipPlanRequestDto(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		dto           dto.CreateMembershipPlanRequest
// 		expectError   bool
// 		expectedError string
// 	}{
// 		{
// 			name: "Valid Input",
// 			dto: dto.CreateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Name:             "Premium Membership",
// 				Price:            1999,
// 				PaymentFrequency: string(db.PaymentFrequencyMonth),
// 				AmtPeriods:       12,
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "Invalid payment frequency",
// 			dto: dto.CreateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Name:             "Premium Membership",
// 				Price:            1999,
// 				PaymentFrequency: "yearly",
// 				AmtPeriods:       12,
// 			},
// 			expectError:   true,
// 			expectedError: "payment_frequency: must be one of 'week', 'month', or 'day'",
// 		},
// 		{
// 			name: "Validation: Whitespace Name",
// 			dto: dto.CreateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Name:             "",
// 				Price:            1999,
// 				PaymentFrequency: "yearly",
// 				AmtPeriods:       12,
// 			},
// 			expectError:   true,
// 			expectedError: "name: required",
// 		},
// 		{
// 			name: "Validation: wrong frequency and empty name",
// 			dto: dto.CreateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Name:             " ",
// 				Price:            1999,
// 				PaymentFrequency: "yearly",
// 				AmtPeriods:       12,
// 			},
// 			expectError:   true,
// 			expectedError: "name: required and cannot be empty or whitespace",
// 		},
// 		{
// 			name: "Validation: Missing Price",
// 			dto: dto.CreateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Name:             "Premium Membership",
// 				PaymentFrequency: "monthly",
// 				AmtPeriods:       12,
// 			},
// 			expectError:   true,
// 			expectedError: "price: required",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := validators.ValidateDto(&tt.dto)
// 			if tt.expectError {
// 				assert.NotNil(t, err)
// 				if tt.expectedError != "" {
// 					assert.Contains(t, err.Message, tt.expectedError)
// 				}
// 			} else {
// 				assert.Nil(t, err)
// 			}
// 		})
// 	}
// }
