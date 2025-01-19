package tests

import (
	dto "api/internal/dtos/facilityType"
	"api/internal/utils/validators"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestDecodeUpdateFacilityTypeRequestBody(t *testing.T) {
	testUUID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"

	tests := []struct {
		name         string
		jsonInput    string
		expectError  bool
		expectedID   uuid.UUID
		expectedName string
	}{
		{
			name:         "Valid JSON",
			jsonInput:    fmt.Sprintf(`{"id": "%s", "name": "Test Facility"}`, testUUID),
			expectError:  false,
			expectedID:   uuid.MustParse(testUUID),
			expectedName: "Test Facility",
		},
		{
			name:        "Invalid JSON",
			jsonInput:   `{"id": "f47ac10b-58cc-4372-a567-0e02b2c3d479", "name": "Test Facility"`,
			expectError: true,
		},
		{
			name:        "Valid JSON with Empty ID",
			jsonInput:   `{"id": "", "name": "Test Facility"}`,
			expectError: true,
		},
		{
			name:         "Valid JSON with Whitespace Name",
			jsonInput:    fmt.Sprintf(`{"id": "%s", "name": "  "}`, testUUID),
			expectError:  false,
			expectedID:   uuid.MustParse(testUUID),
			expectedName: "  ",
		},
		{
			name:         "Missing ID Field in JSON",
			jsonInput:    `{"name": "Test Facility"}`,
			expectError:  false,
			expectedID:   uuid.Nil,
			expectedName: "Test Facility",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tt.jsonInput))
			var target dto.UpdateFacilityTypeRequest

			err := validators.DecodeRequestBody(reqBody, &target)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				if target.Id != tt.expectedID {
					t.Fatalf("expected Id to be '%v', got: %v", tt.expectedID, target.Id)
				}
				if target.Name != tt.expectedName {
					t.Fatalf("expected Name to be '%s', got: %s", tt.expectedName, target.Name)
				}
			}
		})
	}
}

func TestUpdateFacilityTypeRequestDto(t *testing.T) {
	testUUID := uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")

	tests := []struct {
		name         string
		dto          *dto.UpdateFacilityTypeRequest
		expectError  bool
		errorMessage string
	}{
		{
			name:        "Valid UpdateFacilityTypeRequest",
			dto:         &dto.UpdateFacilityTypeRequest{Id: testUUID, Name: "Valid Facility Name"},
			expectError: false,
		},
		{
			name:         "Invalid UpdateFacilityTypeRequest (missing ID)",
			dto:          &dto.UpdateFacilityTypeRequest{Name: "Valid Facility Name"},
			expectError:  true,
			errorMessage: "id: required",
		},
		{
			name:         "Invalid UpdateFacilityTypeRequest (whitespace name)",
			dto:          &dto.UpdateFacilityTypeRequest{Id: testUUID, Name: "   "},
			expectError:  true,
			errorMessage: "name: required and cannot be empty or whitespace",
		},
		{
			name:         "Invalid UpdateFacilityTypeRequest (missing name)",
			dto:          &dto.UpdateFacilityTypeRequest{Id: testUUID},
			expectError:  true,
			errorMessage: "name: required",
		},
		{
			name:         "Invalid UpdateFacilityTypeRequest (empty ID)",
			dto:          &dto.UpdateFacilityTypeRequest{Id: uuid.Nil, Name: "Valid Facility Name"},
			expectError:  true,
			errorMessage: "id: required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateDto(tt.dto)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected validation error, got nil")
				}
				if !strings.Contains(err.Message, tt.errorMessage) {
					t.Errorf("expected '%s' error, got: %v", tt.errorMessage, err.Message)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no validation error, got: %v", err)
				}
			}
		})
	}
}
