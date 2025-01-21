package tests

// import (
// 	dto "api/internal/dtos/membershipPlans"
// 	"api/internal/utils/validators"
// 	db "api/sqlc"
// 	"bytes"
// 	"testing"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// )

// func TestDecodeUpdateMembershipPlanRequest(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		jsonBody       string
// 		expectError    bool
// 		expectedValues *dto.UpdateMembershipPlanRequest
// 	}{
// 		{
// 			name: "Valid JSON",
// 			jsonBody: `{
// 				"membership_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"id": "5a9c1b38-49ba-4fe3-83cf-dfc9b4050ee9",
// 				"name": "Updated Premium Membership",
// 				"price": 2999,
// 				"payment_frequency": "week",
// 				"amt_periods": 24
// 			}`,
// 			expectError: false,
// 			expectedValues: &dto.UpdateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				ID:               uuid.MustParse("5a9c1b38-49ba-4fe3-83cf-dfc9b4050ee9"),
// 				Name:             "Updated Premium Membership",
// 				Price:            2999,
// 				PaymentFrequency: string(db.PaymentFrequencyWeek),
// 				AmtPeriods:       24,
// 			},
// 		},
// 		{
// 			name: "Invalid JSON: Malformed UUID",
// 			jsonBody: `{
// 				"membership_id": "invalid-uuid",
// 				"id": "5a9c1b38-49ba-4fe3-83cf-dfc9b4050ee9",
// 				"name": "Updated Premium Membership",
// 				"price": 2999,
// 				"payment_frequency": "week",
// 				"amt_periods": 24
// 			}`,
// 			expectError: true,
// 		},
// 		{
// 			name: "Validation: Missing Required Fields",
// 			jsonBody: `{
// 				"name": "Updated Premium Membership",
// 				"price": 2999
// 			}`,
// 			expectError: false,
// 		},
// 		{
// 			name: "Validation: Empty Name",
// 			jsonBody: `{
// 				"membership_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"id": "5a9c1b38-49ba-4fe3-83cf-dfc9b4050ee9",
// 				"name": "",
// 				"price": 2999,
// 				"payment_frequency": "week",
// 				"amt_periods": 24
// 			}`,
// 			expectError: false,
// 		},
// 		{
// 			name: "Invalid Payment Frequency",
// 			jsonBody: `{
// 				"membership_id": "4f5c063e-fb57-43b2-83a7-1b1068305ee9",
// 				"id": "5a9c1b38-49ba-4fe3-83cf-dfc9b4050ee9",
// 				"name": "Updated Premium Membership",
// 				"price": 2999,
// 				"payment_frequency": "invalid-freq",
// 				"amt_periods": 24
// 			}`,
// 			expectError: false,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			reqBody := bytes.NewReader([]byte(tc.jsonBody))
// 			var target dto.UpdateMembershipPlanRequest

// 			err := validators.DecodeRequestBody(reqBody, &target)
// 			if tc.expectError {
// 				assert.NotNil(t, err)
// 			} else {
// 				assert.Nil(t, err)

// 				expected := tc.expectedValues

// 				if expected != nil {
// 					assert.Equal(t, expected.MembershipID, target.MembershipID)
// 					assert.Equal(t, expected.ID, target.ID)
// 					assert.Equal(t, expected.Name, target.Name)
// 					assert.Equal(t, expected.Price, target.Price)
// 					assert.Equal(t, expected.PaymentFrequency, target.PaymentFrequency)
// 					assert.Equal(t, expected.AmtPeriods, target.AmtPeriods)
// 				}
// 			}
// 		})
// 	}
// }

// func TestUpdateMembershipPlanRequestValidation(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		dto           dto.UpdateMembershipPlanRequest
// 		expectError   bool
// 		expectedError string
// 	}{
// 		{
// 			name: "Valid UpdateMembershipPlanRequest",
// 			dto: dto.UpdateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				ID:               uuid.MustParse("5a9c1b38-49ba-4fe3-83cf-dfc9b4050ee9"),
// 				Name:             "Valid Membership Plan Name",
// 				Price:            2999,
// 				PaymentFrequency: string(db.PaymentFrequencyWeek),
// 				AmtPeriods:       24,
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "Invalid UpdateMembershipPlanRequest (missing ID)",
// 			dto: dto.UpdateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				Name:             "Valid Membership Plan Name",
// 				Price:            2999,
// 				PaymentFrequency: string(db.PaymentFrequencyWeek),
// 				AmtPeriods:       24,
// 			},
// 			expectError:   true,
// 			expectedError: "id: required",
// 		},
// 		{
// 			name: "Invalid UpdateMembershipPlanRequest (whitespace name)",
// 			dto: dto.UpdateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				ID:               uuid.MustParse("5a9c1b38-49ba-4fe3-83cf-dfc9b4050ee9"),
// 				Name:             "   ",
// 				Price:            2999,
// 				PaymentFrequency: string(db.PaymentFrequencyWeek),
// 				AmtPeriods:       24,
// 			},
// 			expectError:   true,
// 			expectedError: "name: required and cannot be empty or whitespace",
// 		},
// 		{
// 			name: "Invalid UpdateMembershipPlanRequest (missing name)",
// 			dto: dto.UpdateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				ID:               uuid.MustParse("5a9c1b38-49ba-4fe3-83cf-dfc9b4050ee9"),
// 				Price:            2999,
// 				PaymentFrequency: string(db.PaymentFrequencyWeek),
// 				AmtPeriods:       24,
// 			},
// 			expectError:   true,
// 			expectedError: "name: required and cannot be empty or whitespace",
// 		},
// 		{
// 			name: "Invalid UpdateMembershipPlanRequest (empty ID)",
// 			dto: dto.UpdateMembershipPlanRequest{
// 				MembershipID:     uuid.MustParse("4f5c063e-fb57-43b2-83a7-1b1068305ee9"),
// 				ID:               uuid.Nil,
// 				Name:             "Valid Membership Plan Name",
// 				Price:            2999,
// 				PaymentFrequency: string(db.PaymentFrequencyWeek),
// 				AmtPeriods:       24,
// 			},
// 			expectError:   true,
// 			expectedError: "id: required",
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
