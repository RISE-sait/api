package tests

import (
	"bytes"
	"testing"
	"time"

	dto "api/internal/dtos/course"
	"api/internal/utils/validators"

	"github.com/stretchr/testify/assert"
)

func TestDecodeRequestBody(t *testing.T) {
	t.Run("Valid JSON", func(t *testing.T) {
		validJSON := `{
			"name": "Go Programming Basics",
			"description": "Learn the basics of Go programming",
			"start_date": "2025-01-15T00:00:00Z",
			"end_date": "2025-02-15T00:00:00Z"
		}`
		reqBody := bytes.NewReader([]byte(validJSON))
		var target dto.CreateCourseRequestBody

		if err := validators.DecodeRequestBody(reqBody, &target); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		assert.Equal(t, "Go Programming Basics", target.Name)
		assert.Equal(t, "Learn the basics of Go programming", target.Description)
		assert.Equal(t, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), target.StartDate)
		assert.Equal(t, time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC), target.EndDate)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidJSON := `{
			"name": "Go Programming Basics",
			"start_date": "2025-01-15T00:00:00Z",
			"end_date": "2025-02-15T00:00:00Z"
		` // Missing closing brace
		reqBody := bytes.NewReader([]byte(invalidJSON))
		var target dto.CreateCourseRequestBody

		if err := validators.DecodeRequestBody(reqBody, &target); err == nil {
			t.Fatal("expected error for invalid JSON, but got nil")
		}
	})

	t.Run("Validation: Missing Name", func(t *testing.T) {
		validJSON := `{
			"description": "Learn the basics of Go programming",
			"start_date": "2025-01-15T00:00:00Z",
			"end_date": "2025-02-15T00:00:00Z"
		}`
		reqBody := bytes.NewReader([]byte(validJSON))
		var target dto.CreateCourseRequestBody

		if err := validators.DecodeRequestBody(reqBody, &target); err != nil {
			t.Fatalf("expected no decoding error, got: %v", err)
		}

		validationErr := validators.ValidateDto(&target)
		assert.NotNil(t, validationErr)
		assert.Contains(t, validationErr.Message, "name: required")
	})

	t.Run("Validation: Whitespace Name", func(t *testing.T) {
		validJSON := `{
			"name": "   ",
			"description": "Learn the basics of Go programming",
			"start_date": "2025-01-15T00:00:00Z",
			"end_date": "2025-02-15T00:00:00Z"
		}`
		reqBody := bytes.NewReader([]byte(validJSON))
		var target dto.CreateCourseRequestBody

		if err := validators.DecodeRequestBody(reqBody, &target); err != nil {
			t.Fatalf("expected no decoding error, got: %v", err)
		}

		validationErr := validators.ValidateDto(&target)
		assert.NotNil(t, validationErr)
		assert.Contains(t, validationErr.Message, "name: required and cannot be empty or whitespace")
	})

	t.Run("Validation: Missing StartDate", func(t *testing.T) {
		validJSON := `{
			"name": "Go Programming Basics",
			"description": "Learn the basics of Go programming",
			"end_date": "2025-02-15T00:00:00Z"
		}`
		reqBody := bytes.NewReader([]byte(validJSON))
		var target dto.CreateCourseRequestBody

		if err := validators.DecodeRequestBody(reqBody, &target); err != nil {
			t.Fatalf("expected no decoding error, got: %v", err)
		}

		validationErr := validators.ValidateDto(&target)
		assert.NotNil(t, validationErr)
		assert.Contains(t, validationErr.Message, "start_date: required")
	})

	t.Run("Validation: Missing EndDate", func(t *testing.T) {
		validJSON := `{
			"name": "Go Programming Basics",
			"description": "Learn the basics of Go programming",
			"start_date": "2025-01-15T00:00:00Z"
		}`
		reqBody := bytes.NewReader([]byte(validJSON))
		var target dto.CreateCourseRequestBody

		if err := validators.DecodeRequestBody(reqBody, &target); err != nil {
			t.Fatalf("expected no decoding error, got: %v", err)
		}

		validationErr := validators.ValidateDto(&target)
		assert.NotNil(t, validationErr)
		assert.Contains(t, validationErr.Message, "end_date: required")
	})
}

func TestValidateDto(t *testing.T) {
	// Test Case 1: Valid DTO
	validDTO := &dto.CreateCourseRequestBody{
		Name:        "Valid Course",
		Description: "A description of the dto",
		EndDate:     time.Now().Add(24 * time.Hour),
	}
	if err := validators.ValidateDto(validDTO); err != nil {
		t.Fatalf("expected no validation error, got: %v", err)
	}

	// Test Case 2: Invalid DTO (empty Name)
	invalidNameDTO := &dto.CreateCourseRequestBody{
		Name:        " ",
		Description: "A description of the dto",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(24 * time.Hour),
	}
	if err := validators.ValidateDto(invalidNameDTO); err == nil {
		t.Fatal("expected validation error for empty Name, but got nil")
	}

	// Test Case 3: Invalid DTO (missing StartDate)
	invalidStartDateDTO := &dto.CreateCourseRequestBody{
		Name:        "Valid Course",
		Description: "A description of the dto",
		StartDate:   time.Time{}, // Zero time to represent invalid value
		EndDate:     time.Now().Add(24 * time.Hour),
	}
	if err := validators.ValidateDto(invalidStartDateDTO); err == nil {
		t.Fatal("expected validation error for missing StartDate, but got nil")
	}

	// Test Case 4: Invalid DTO (missing EndDate)
	invalidEndDateDTO := &dto.CreateCourseRequestBody{
		Name:        "Valid Course",
		Description: "A description of the dto",
		StartDate:   time.Now(),
		EndDate:     time.Time{}, // Zero time to represent invalid value
	}
	if err := validators.ValidateDto(invalidEndDateDTO); err == nil {
		t.Fatal("expected validation error for missing EndDate, but got nil")
	}

	// Test Case 5: Invalid DTO (StartDate is after EndDate)
	invalidDateDTO := &dto.CreateCourseRequestBody{
		Name:        "Valid Course",
		Description: "A description of the dto",
		StartDate:   time.Now().Add(2 * time.Hour), // Start date later than End date
		EndDate:     time.Now(),
	}
	if err := validators.ValidateDto(invalidDateDTO); err == nil {
		t.Fatal("expected validation error for StartDate after EndDate, but got nil")
	}
}
