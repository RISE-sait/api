package enrollment

import (
	"api/internal/libs/validators"
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDecodeRequestBody(t *testing.T) {
	validCustomerID := uuid.New()
	validEventID := uuid.New()

	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectedValues *RequestDto
	}{
		{
			name: "Valid Input",
			jsonBody: `{
				"customer_id": "` + validCustomerID.String() + `",
				"event_id": "` + validEventID.String() + `"
			}`,
			expectError: false,
			expectedValues: &RequestDto{
				CustomerId: validCustomerID,
				EventId:    validEventID,
			},
		},
		{
			name: "Invalid JSON - Missing closing brace",
			jsonBody: `{
				"customer_id": "` + validCustomerID.String() + `"
			`,
			expectError: true,
		},
		{
			name: "Missing Customer ID",
			jsonBody: `{
				"event_id": "` + validEventID.String() + `"
			}`,
			expectError: false,
		},
		{
			name: "Invalid UUID format",
			jsonBody: `{
				"customer_id": "invalid-uuid",
				"event_id": "` + validEventID.String() + `"
			}`,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tc.jsonBody))
			var target RequestDto

			err := validators.ParseJSON(reqBody, &target)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if tc.expectedValues != nil {
					assert.Equal(t, tc.expectedValues.CustomerId, target.CustomerId)
					assert.Equal(t, tc.expectedValues.EventId, target.EventId)
				}
			}
		})
	}
}

func TestEnrollmentRequestDto_Validation(t *testing.T) {
	validCustomerID := uuid.New()
	validEventID := uuid.New()

	tests := []struct {
		name                 string
		dto                  CreateRequestDto
		expectErr            bool
		expectedErrorMessage string
	}{
		{
			name: "Valid DTO",
			dto: CreateRequestDto{
				RequestDto: RequestDto{
					CustomerId: validCustomerID,
					EventId:    validEventID,
				},
			},
			expectErr: false,
		},
		{
			name: "Missing CustomerId",
			dto: CreateRequestDto{
				RequestDto: RequestDto{
					EventId: validEventID,
				},
			},
			expectErr:            true,
			expectedErrorMessage: "customer_id: required",
		},
		{
			name: "Missing EventId",
			dto: CreateRequestDto{
				RequestDto: RequestDto{
					CustomerId: validCustomerID,
				},
			},
			expectErr:            true,
			expectedErrorMessage: "event_id: required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			err := validate(&tc.dto.RequestDto)
			if tc.expectErr {
				assert.NotNil(t, err)

				assert.Contains(t, err.Message, tc.expectedErrorMessage)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestEnrollmentRequestDto_ToCreateValueObjects(t *testing.T) {
	validCustomerID := uuid.New()
	validEventID := uuid.New()

	tests := []struct {
		name        string
		dto         CreateRequestDto
		expectError bool
	}{
		{
			name: "Valid Conversion",
			dto: CreateRequestDto{
				RequestDto: RequestDto{
					CustomerId: validCustomerID,
					EventId:    validEventID,
				},
			},
			expectError: false,
		},
		{
			name: "Missing CustomerId",
			dto: CreateRequestDto{
				RequestDto: RequestDto{
					EventId: validEventID,
				},
			},
			expectError: true,
		},
		{
			name: "Missing EventId",
			dto: CreateRequestDto{
				RequestDto: RequestDto{
					CustomerId: validCustomerID,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vo, err := tc.dto.ToCreateValueObjects()
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.dto.CustomerId, vo.CustomerId)
				assert.Equal(t, tc.dto.EventId, vo.EventId)
			}
		})
	}
}

func TestEnrollmentRequestDto_ToUpdateValueObjects(t *testing.T) {
	validCustomerID := uuid.New()
	validEventID := uuid.New()
	validEnrollmentID := uuid.New()

	tests := []struct {
		name        string
		dto         UpdateRequestDto
		idStr       string
		expectError bool
	}{
		{
			name: "Valid Update Conversion",
			dto: UpdateRequestDto{
				RequestDto: RequestDto{
					CustomerId: validCustomerID,
					EventId:    validEventID,
				},
				IsCancelled: false,
				ID:          validEnrollmentID,
			},
			idStr:       validEnrollmentID.String(),
			expectError: false,
		},
		{
			name: "Invalid UUID for Enrollment ID",
			dto: UpdateRequestDto{
				RequestDto: RequestDto{
					CustomerId: validCustomerID,
					EventId:    validEventID,
				},
			},
			idStr:       "invalid-uuid",
			expectError: true,
		},
		{
			name: "Missing CustomerId",
			dto: UpdateRequestDto{
				RequestDto: RequestDto{

					EventId: validEventID,
				}},
			idStr:       validEnrollmentID.String(),
			expectError: true,
		},
		{
			name: "Missing EventId",
			dto: UpdateRequestDto{
				RequestDto: RequestDto{CustomerId: validCustomerID},
			},
			idStr:       validEnrollmentID.String(),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vo, err := tc.dto.ToUpdateValueObjects(tc.idStr)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.dto.CustomerId, vo.EnrollmentDetails.CustomerId)
				assert.Equal(t, tc.dto.EventId, vo.EnrollmentDetails.EventId)
				assert.Equal(t, tc.idStr, vo.ID.String())
			}
		})
	}
}
