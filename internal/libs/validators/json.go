package validators

import (
	errLib "api/internal/libs/errors"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type validationError struct {
	Field   string
	Message string
}

var validate *validator.Validate

func notWhiteSpace(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

func init() {
	validate = validator.New()
	validate.RegisterValidation("notwhitespace", notWhiteSpace)
}

func ParseAndValidateJSON(body io.Reader, target interface{}) *errLib.CommonError {
	if err := validatePointerToStruct(target); err != nil {
		return err
	}

	if err := json.NewDecoder(body).Decode(target); err != nil {
		return errLib.New("Invalid JSON format", http.StatusBadRequest)
	}

	if err := validate.Struct(target); err != nil {
		return handleValidationErrors(err, reflect.TypeOf(target).Elem())
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

func handleValidationErrors(err error, structType reflect.Type) *errLib.CommonError {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return errLib.New("Internal validation error", http.StatusInternalServerError)
	}

	messages := make([]validationError, 0, len(validationErrors))
	for _, e := range validationErrors {
		messages = append(messages, createValidationError(e, structType))
	}

	return formatValidationError(messages)
}

func createValidationError(e validator.FieldError, structType reflect.Type) validationError {
	field := getJSONFieldName(e, structType)
	message := getErrorMessage(e.Tag(), field)
	return validationError{Field: field, Message: message}
}

func getErrorMessage(tag, field string) string {
	messages := map[string]string{
		"required":      "is required",
		"email":         "must be a valid email address",
		"notwhitespace": "cannot be empty or whitespace",
	}

	if msg, ok := messages[tag]; ok {
		return fmt.Sprintf("%s %s", field, msg)
	}
	return fmt.Sprintf("%s: validation failed on '%s'", field, tag)
}

func formatValidationError(errors []validationError) *errLib.CommonError {
	if len(errors) == 0 {
		return errLib.New("Validation failed", http.StatusBadRequest)
	}

	messages := make([]string, 0, len(errors))
	for _, e := range errors {
		messages = append(messages, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}

	return errLib.New(strings.Join(messages, ", "), http.StatusBadRequest)
}

func getJSONFieldName(e validator.FieldError, structType reflect.Type) string {
	// Iterate over the struct's fields to find the one that matches the field name
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// If the field name matches, get its JSON tag
		if field.Name == e.Field() {
			// Get the JSON tag and return the first part (in case of multiple tags)
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				return strings.Split(jsonTag, ",")[0] // Get the first part of the JSON tag
			}
			return field.Name // Fallback to the Go field name if no JSON tag exists
		}
	}

	return e.Field() // If no match is found, return the Go field name by default
}
