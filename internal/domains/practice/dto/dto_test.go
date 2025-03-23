package practice

//
//import (
//	"api/internal/libs/validators"
//	"bytes"
//	"github.com/google/uuid"
//	"net/http"
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
//				"description": "Learn the basics of Go programming",
//"capacity":32
//			}`,
//			expectError: false,
//			expectedValues: &RequestDto{
//				Name:        "Go Programming Basics",
//				Description: "Learn the basics of Go programming",
//				Capacity:    int32(32),
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
//// Validate Dto
//
//func TestValidRequestDto(t *testing.T) {
//
//	requestDto := RequestDto{
//		Name:        "Go Programming Basics",
//		Description: "Learn Go Programming",
//		Capacity:    int32(50),
//		Level:       "all",
//	}
//
//	createRequestDto, err := requestDto.ToCreateValueObjects()
//
//	assert.Nil(t, err)
//
//	assert.Equal(t, createRequestDto.Name, "Go Programming Basics")
//	assert.Equal(t, createRequestDto.Description, "Learn Go Programming")
//	assert.Equal(t, createRequestDto.Capacity, int32(50))
//}
//
//func TestMissingNameRequestDto(t *testing.T) {
//
//	requestDto := RequestDto{
//		Description: "Learn Go Programming",
//		Capacity:    int32(50),
//		Level:       "all",
//	}
//
//	createRequestDto, err := requestDto.ToCreateValueObjects()
//
//	assert.NotNil(t, err)
//
//	assert.Equal(t, err.Message, "name: required")
//	assert.Equal(t, err.HTTPCode, http.StatusBadRequest)
//	assert.Equal(t, createRequestDto.Name, "")
//}
//
//func TestBlankNameRequestDto(t *testing.T) {
//
//	requestDto := RequestDto{
//		Name:        "          ",
//		Description: "Learn Go Programming",
//		Capacity:    int32(50),
//	}
//
//	createRequestDto, err := requestDto.ToCreateValueObjects()
//
//	assert.NotNil(t, err)
//
//	assert.Contains(t, err.Message, "name: cannot be empty or whitespace")
//	assert.Equal(t, err.HTTPCode, http.StatusBadRequest)
//	assert.Equal(t, createRequestDto.Name, "")
//}
//
//func TestUpdateRequestDtoValidUUID(t *testing.T) {
//
//	requestDto := RequestDto{
//		Name:        "Learn Go Programming Name",
//		Description: "Learn Go Programming Description",
//		Capacity:    int32(50),
//		Level:       "all",
//	}
//
//	id := uuid.New()
//
//	updateRequestDto, err := requestDto.ToUpdateValueObjects(id.String())
//
//	assert.Nil(t, err)
//
//	assert.Equal(t, updateRequestDto.Details.Name, "Learn Go Programming Name")
//	assert.Equal(t, updateRequestDto.Details.Description, "Learn Go Programming Description")
//
//	assert.Equal(t, updateRequestDto.ID.String(), id.String())
//}
//
//func TestUpdateRequestDtoInvalidUUID(t *testing.T) {
//
//	requestDto := RequestDto{
//		Name:        "Learn Go Programming Name",
//		Description: "Learn Go Programming Description",
//	}
//
//	updateRequestDto, err := requestDto.ToUpdateValueObjects("wefwfwefew")
//
//	assert.NotNil(t, err)
//
//	assert.Contains(t, err.Message, "invalid UUID: wefwfwefew")
//
//	assert.Equal(t, updateRequestDto.Details.Name, "")
//	assert.Equal(t, updateRequestDto.Details.Description, "")
//}
//
//func TestUpdateRequestDtoMissingCapacity(t *testing.T) {
//
//	requestDto := RequestDto{
//		Name:        "Learn Go Programming Name",
//		Description: "Learn Go Programming Description",
//	}
//
//	id := uuid.New()
//
//	updateRequestDto, err := requestDto.ToUpdateValueObjects(id.String())
//
//	assert.NotNil(t, err)
//
//	assert.Contains(t, err.Message, "capacity: required")
//
//	assert.Equal(t, updateRequestDto.Details.Name, "")
//	assert.Equal(t, updateRequestDto.Details.Description, "")
//}
//
//func TestUpdateRequestDtoCapacity0(t *testing.T) {
//
//	requestDto := RequestDto{
//		Name:        "Learn Go Programming Name",
//		Description: "Learn Go Programming Description",
//		Capacity:    int32(0),
//	}
//
//	id := uuid.New()
//
//	updateRequestDto, err := requestDto.ToUpdateValueObjects(id.String())
//
//	assert.NotNil(t, err)
//
//	assert.Contains(t, err.Message, "capacity: required")
//
//	assert.Equal(t, updateRequestDto.Details.Name, "")
//	assert.Equal(t, updateRequestDto.Details.Description, "")
//}
