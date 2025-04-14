package membership_plan

import (
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"testing"

	"github.com/stretchr/testify/assert"
)

//func TestDecodeRequestBody(t *testing.T) {
//	tests := []struct {
//		name           string
//		jsonBody       string
//		expectError    bool
//		expectedValues *UpdateRequestDto
//	}{
//		{
//			name: "Valid Input",
//			jsonBody: `{
//				"name": "Go Programming Basics",
//				"description": "Learn the basics of Go programming"
//			}`,
//			expectError: false,
//			expectedValues: &UpdateRequestDto{
//				Name:        "Go Programming Basics",
//				Description: "Learn the basics of Go programming",
//			},
//		},
//		{
//			name: "Invalid JSON - Missing closing brace",
//			jsonBody: `{
//				"name": "Go Programming Basics"
//			`,
//			expectError: true,
//		},
//		{
//			name: "Missing Name",
//			jsonBody: `{
//				"description": "Learn the basics of Go programming"
//			}`,
//			expectError: false, // Expecting validation error for missing name
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			reqBody := bytes.NewReader([]byte(tc.jsonBody))
//			var target UpdateRequestDto
//
//			err := validators.ParseJSON(reqBody, &target)
//			if tc.expectError {
//				assert.NotNil(t, err)
//			} else {
//				assert.Nil(t, err)
//				if tc.expectedValues != nil {
//					assert.Equal(t, tc.expectedValues.Name, target.Name)
//					assert.Equal(t, tc.expectedValues.Description, target.Description)
//				}
//			}
//		})
//	}
//}

func TestRequestDto_Validation(t *testing.T) {
	tests := []struct {
		name                 string
		dto                  *PlanRequestDto
		expectErr            bool
		expectedErrorMessage string
	}{
		{
			name: "Valid DTO",
			dto: &PlanRequestDto{
				MembershipID:        uuid.New(),
				Name:                "Go Programming Basics",
				AmtPeriods:          int32Ptr(5),
				StripePriceID:       "price_123",
				StripeJoiningFeesID: "fee_123",
			},
			expectErr: false,
		},
		{
			name: "Valid DTO with omited amt periods",
			dto: &PlanRequestDto{
				MembershipID:        uuid.New(),
				Name:                "Go Programming Basics",
				StripePriceID:       "price_123",
				StripeJoiningFeesID: "fee_123",
				AmtPeriods:          nil,
			},
			expectErr: false,
		},
		{
			name: "invalid DTO with 0 amt periods",
			dto: &PlanRequestDto{
				MembershipID:        uuid.New(),
				Name:                "Go Programming Basics",
				StripePriceID:       "price_123",
				StripeJoiningFeesID: "fee_123",
				AmtPeriods:          int32Ptr(0),
			},
			expectErr:            true,
			expectedErrorMessage: `amt_periods: must be greater than 0`,
		},
		{
			name: "Invalid DTO with empty stripe price ID",
			dto: &PlanRequestDto{
				MembershipID:        uuid.New(),
				Name:                "Go Programming Basics",
				AmtPeriods:          int32Ptr(5),
				StripePriceID:       "",
				StripeJoiningFeesID: "fee_123",
			},
			expectErr:            true,
			expectedErrorMessage: `stripe_price_id: required`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validators.ValidateDto(tc.dto)
			if tc.expectErr {
				assert.NotNil(t, err)
				assert.Contains(t, tc.expectedErrorMessage, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}
