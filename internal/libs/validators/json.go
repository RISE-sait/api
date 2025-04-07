package validators

import (
	errLib "api/internal/libs/errors"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator"
)

var validate *validator.Validate

func ParseJSON(body io.Reader, target interface{}) *errLib.CommonError {
	if err := validatePointerToStruct(target); err != nil {
		return err
	}

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			return errLib.New(fmt.Sprintf("Invalid JSON syntax at position %d", syntaxErr.Offset), http.StatusBadRequest)
		case errors.As(err, &unmarshalTypeErr):
			return errLib.New(fmt.Sprintf("Invalid type for field '%s' at position %d", unmarshalTypeErr.Field, unmarshalTypeErr.Offset), http.StatusBadRequest)
		case strings.Contains(err.Error(), "unexpected end of JSON"):
			return errLib.New("Unexpected end of JSON. Make sure all brackets and commas are properly placed.", http.StatusBadRequest)
		case strings.Contains(err.Error(), "cannot unmarshal"):
			return errLib.New("Invalid data type provided. Check if all fields match the expected type.", http.StatusBadRequest)
		default:
			log.Println(err)
			return errLib.New("Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		}
	}
	return nil
}

func validatePointerToStruct(v interface{}) *errLib.CommonError {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return errLib.New("Internal error: invalid target type", http.StatusInternalServerError)
	}
	return nil
}

func notWhiteSpace(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

func init() {
	validate = validator.New()
	validate.RegisterValidation("notwhitespace", notWhiteSpace)
}

// ValidateDto validates the given DTO using the go-playground/validator library.
func ValidateDto(dto interface{}) *errLib.CommonError {

	if err := validatePointerToStruct(dto); err != nil {
		return err
	}

	dtoType := reflect.TypeOf(dto).Elem()

	if err := validate.Struct(dto); err != nil {
		return parseValidationErrors(err, dtoType)
	}

	return nil

}

func parseValidationErrors(err error, structType reflect.Type) *errLib.CommonError {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {

		var errorMessages []string

		for _, e := range validationErrors {

			fieldName := getJSONFieldName(e, structType)
			var customMessage string

			switch e.Tag() {

			case "email":
				customMessage = fmt.Sprintf("%s: must be a valid email address", fieldName)
			case "e164":
				customMessage = fmt.Sprintf("%s: must be a valid phone number", fieldName)
			case "min":
				customMessage = fmt.Sprintf("%s: must be at least %s characters long", fieldName, e.Param())
			case "required":
				customMessage = fmt.Sprintf("%s: required", fieldName)
			case "notwhitespace":
				customMessage = fmt.Sprintf("%s: cannot be empty or whitespace", fieldName)
			case "url":
				customMessage = fmt.Sprintf("%s: must be a valid URL", fieldName)
			case "gt":
				customMessage = fmt.Sprintf("%s: must be greater than %s", fieldName, getJSONFieldNameForParam(e.Param(), structType))
			case "gtcsfield":
				customMessage = fmt.Sprintf("%s: must be greater than %s", fieldName, getJSONFieldNameForParam(e.Param(), structType))
			case "oneof":
				customMessage = fmt.Sprintf("%s: must be one of the following values: %s", fieldName, strings.Join(strings.Split(e.Param(), ","), ", "))
			default:
				customMessage = fmt.Sprintf("%s: validation failed on '%s'", fieldName, e.Tag())
			}

			errorMessages = append(errorMessages, customMessage)

		}

		return errLib.New(strings.Join(errorMessages, ", "), http.StatusBadRequest)
	}

	// Handle other validation errors
	fmt.Printf("Unhandled validation error: %v\n", err)
	return errLib.New("Internal server error", http.StatusInternalServerError)
}

func getJSONFieldName(e validator.FieldError, structType reflect.Type) string {
	fieldName := e.Field() // Default to the Go field name

	// Recursively search for the field in the struct and its embedded fields
	var findField func(reflect.Type, string) string
	findField = func(t reflect.Type, fieldName string) string {
		// Iterate over the struct's fields
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// If the field is embedded, recursively search its type
			if field.Anonymous {
				if result := findField(field.Type, fieldName); result != "" {
					return result
				}
				continue
			}

			// If the field name matches, return the JSON tag
			if field.Name == fieldName {
				jsonTag := field.Tag.Get("json")
				if jsonTag != "" {
					return strings.Split(jsonTag, ",")[0] // Take the first JSON tag
				}
				return field.Name // Fallback to Go field name
			}
		}
		return "" // Field not found
	}

	// Start the recursive search
	if result := findField(structType, fieldName); result != "" {
		return result
	}

	log.Println("Field not found:", fieldName)
	return fieldName // If no match is found, return the Go field name by default
}

func getJSONFieldNameForParam(param string, structType reflect.Type) string {
	// Iterate over the struct's fields to find the one that matches the field name
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Compare the field name with the provided param
		if field.Name == param {
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				return strings.Split(jsonTag, ",")[0] // Take the first JSON tag
			}
			return field.Name // Fallback to Go field name
		}
	}
	return param // If no match is found, return the param itself
}
