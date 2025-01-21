package tests

// import (
// 	"bytes"
// 	"testing"
// 	"time"

// 	dto "api/internal/dtos/membership"
// 	"api/internal/utils/validators"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// )

// func TestDecodeUpdateMembershipRequestBody(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		jsonBody       string
// 		expectError    bool
// 		expectedValues *dto.UpdateMembershipRequest
// 	}{
// 		{
// 			name: "Valid JSON",
// 			jsonBody: `{
// 				"name": "Updated Membership",
// 				"description": "Updated Description",
// 				"start_date": "2025-01-01T00:00:00Z",
// 				"end_date": "2025-12-31T23:59:59Z",
// 				"id": "7f7a7b7a-8a8a-9a9a-bb9b-cc9c9d9d9e9e"
// 			}`,
// 			expectError: false,
// 			expectedValues: &dto.UpdateMembershipRequest{
// 				Name:        "Updated Membership",
// 				Description: "Updated Description",
// 				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
// 				ID:          uuid.MustParse("7f7a7b7a-8a8a-9a9a-bb9b-cc9c9d9d9e9e"),
// 			},
// 		},
// 		{
// 			name: "Valid JSON: Missing ID",
// 			jsonBody: `{
// 				"name": "Updated Membership",
// 				"description": "Updated Description",
// 				"start_date": "2025-01-01T00:00:00Z",
// 				"end_date": "2025-12-31T23:59:59Z"
// 			}`,
// 			expectError: false,
// 		},
// 		{
// 			name: "Invalid JSON: Invalid UUID",
// 			jsonBody: `{
// 				"name": "Updated Membership",
// 				"description": "Updated Description",
// 				"start_date": "2025-01-01T00:00:00Z",
// 				"end_date": "2025-12-31T23:59:59Z",
// 				"id": "invalid-uuid"
// 			}`,
// 			expectError: true,
// 		},
// 		{
// 			name: "Valid JSON: Missing Required Fields (Name, StartDate, EndDate, ID)",
// 			jsonBody: `{
// 				"description": "Updated Description"
// 			}`,
// 			expectError: false,
// 		},
// 		{
// 			name: "Empty Name",
// 			jsonBody: `{
// 				"name": "",
// 				"description": "Updated Description",
// 				"start_date": "2025-01-01T00:00:00Z",
// 				"end_date": "2025-12-31T23:59:59Z",
// 				"id": "7f7a7b7a-8a8a-9a9a-bb9b-cc9c9d9d9e9e"
// 			}`,
// 			expectError: false,
// 		},
// 		{
// 			name: "Whitespace Name",
// 			jsonBody: `{
// 				"name": "   ",
// 				"description": "Updated Description",
// 				"start_date": "2025-01-01T00:00:00Z",
// 				"end_date": "2025-12-31T23:59:59Z",
// 				"id": "7f7a7b7a-8a8a-9a9a-bb9b-cc9c9d9d9e9e"
// 			}`,
// 			expectError: false,
// 		},
// 		{
// 			name: "Invalid EndDate Format",
// 			jsonBody: `{
// 				"name": "Updated Membership",
// 				"description": "Updated Description",
// 				"start_date": "2025-01-01T00:00:00Z",
// 				"end_date": "invalid-date",
// 				"id": "7f7a7b7a-8a8a-9a9a-bb9b-cc9c9d9d9e9e"
// 			}`,
// 			expectError: true,
// 		},
// 		{
// 			name: "Valid Empty Description",
// 			jsonBody: `{
// 				"name": "Updated Membership",
// 				"description": "",
// 				"start_date": "2025-01-01T00:00:00Z",
// 				"end_date": "2025-12-31T23:59:59Z",
// 				"id": "7f7a7b7a-8a8a-9a9a-bb9b-cc9c9d9d9e9e"
// 			}`,
// 			expectError: false,
// 			expectedValues: &dto.UpdateMembershipRequest{
// 				Name:        "Updated Membership",
// 				Description: "",
// 				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
// 				ID:          uuid.MustParse("7f7a7b7a-8a8a-9a9a-bb9b-cc9c9d9d9e9e"),
// 			},
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			reqBody := bytes.NewReader([]byte(tc.jsonBody))
// 			var target dto.UpdateMembershipRequest

// 			err := validators.DecodeRequestBody(reqBody, &target)
// 			if tc.expectError {
// 				assert.NotNil(t, err)
// 			} else {
// 				assert.Nil(t, err)

// 				expected := tc.expectedValues

// 				if expected != nil {
// 					if expected.Name != "" {
// 						assert.Equal(t, expected.Name, target.Name)
// 					}

// 					if expected.Description != "" {
// 						assert.Equal(t, expected.Description, target.Description)
// 					}

// 					if !expected.StartDate.IsZero() {
// 						assert.Equal(t, expected.StartDate, target.StartDate)
// 					}

// 					if !expected.EndDate.IsZero() {
// 						assert.Equal(t, expected.EndDate, target.EndDate)
// 					}

// 					if expected.ID != uuid.Nil {
// 						assert.Equal(t, expected.ID, target.ID)
// 					}
// 				}
// 			}
// 		})
// 	}
// }

// func TestUpdateMembershipRequestValidation(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		dto           dto.UpdateMembershipRequest
// 		expectError   bool
// 		expectedError string
// 	}{
// 		{
// 			name: "Valid UpdateMembershipRequest",
// 			dto: dto.UpdateMembershipRequest{
// 				Name:        "Valid Membership Name",
// 				Description: "Valid description",
// 				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
// 				ID:          uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "Invalid UpdateMembershipRequest (missing ID)",
// 			dto: dto.UpdateMembershipRequest{
// 				Name:        "Valid Membership Name",
// 				Description: "Valid description",
// 				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
// 			},
// 			expectError:   true,
// 			expectedError: "id: required",
// 		},
// 		{
// 			name: "Invalid UpdateMembershipRequest (whitespace name)",
// 			dto: dto.UpdateMembershipRequest{
// 				ID:          uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
// 				Name:        "   ",
// 				Description: "Valid description",
// 				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
// 			},
// 			expectError:   true,
// 			expectedError: "name: required and cannot be empty or whitespace",
// 		},
// 		{
// 			name: "Invalid UpdateMembershipRequest (missing name)",
// 			dto: dto.UpdateMembershipRequest{
// 				ID:          uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
// 				Description: "Valid description",
// 				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
// 			},
// 			expectError:   true,
// 			expectedError: "name: required and cannot be empty or whitespace",
// 		},
// 		{
// 			name: "Invalid UpdateMembershipRequest (empty ID)",
// 			dto: dto.UpdateMembershipRequest{
// 				ID:          uuid.Nil,
// 				Name:        "Valid Membership Name",
// 				Description: "Valid description",
// 				StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 				EndDate:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
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
