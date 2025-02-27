package game

//
//import (
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
//		expectedValues *RequestDto
//	}{
//		{
//			name: "Valid Input",
//			jsonBody: `{
//				"name": "Go Programming Basics",
//				"description": "Learn the basics of Go programming"
//			}`,
//			expectError: false,
//			expectedValues: &RequestDto{
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
//			var target RequestDto
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
//
//func TestRequestDto_Validation(t *testing.T) {
//	tests := []struct {
//		name                 string
//		dto                  *RequestDto
//		expectErr            bool
//		expectedErrorMessage string
//	}{
//		{
//			name: "Valid DTO",
//			dto: &RequestDto{
//				Name:        "Go Programming Basics",
//				Description: "Learn Go Programming",
//			},
//			expectErr: false,
//		},
//		{
//			name: "Missing Name",
//			dto: &RequestDto{
//				Name:        "",
//				Description: "Learn Go Programming",
//			},
//			expectErr:            true,
//			expectedErrorMessage: "name: required",
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			err := validators.ValidateDto(tc.dto)
//			if tc.expectErr {
//				assert.NotNil(t, err)
//			} else {
//				assert.Nil(t, err)
//			}
//		})
//	}
//}
