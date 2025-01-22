package validators

import (
	errLib "api/internal/libs/errors"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"reflect"
	"strings"
)

var validate *validator.Validate

func notWhiteSpace(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

func init() {
	validate = validator.New()
	validate.RegisterValidation("notwhitespace", notWhiteSpace)
}

func ParseRequestBodyToJSON(body io.Reader, target interface{}) *errLib.CommonError {
	if err := json.NewDecoder(body).Decode(target); err != nil {
		return errLib.New(err.Error(), http.StatusBadRequest)
	}
	return nil
}

func ValidateDto(dto interface{}) *errLib.CommonError {

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
			case "notwhitespace":
				customMessage = fmt.Sprintf("%s: cannot be empty or whitespace", fieldName)
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
