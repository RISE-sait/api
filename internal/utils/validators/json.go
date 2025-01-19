package validators

import (
	"api/internal/utils"
	db "api/sqlc"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func validateEndDate(fl validator.FieldLevel) bool {
	startDate := fl.Parent().FieldByName("StartDate").Interface().(time.Time)
	endDate := fl.Field().Interface().(time.Time)

	// Return true if EndDate is after StartDate
	return endDate.After(startDate)
}

func notWhiteSpace(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

func requiredAndNotWhiteSpace(fl validator.FieldLevel) bool {
	// First, check if the field is empty (required check)
	if fl.Field().String() == "" {
		return false
	}

	// Then, check if it's not just whitespace
	return strings.TrimSpace(fl.Field().String()) != ""
}

func DayValidator(fl validator.FieldLevel) bool {
	// Extract the value from the field
	val := fl.Field().Int()

	// Check if the value is between 0 and 6
	return val >= 0 && val <= 6
}

func PaymentFrequencyValidator(fl validator.FieldLevel) bool {
	// Extract the value from the field
	val := fl.Field().String()

	// Check if the value is one of the valid PaymentFrequency values
	switch strings.ToLower(val) {
	case string(db.PaymentFrequencyWeek), string(db.PaymentFrequencyMonth), string(db.PaymentFrequencyDay):
		return true
	}
	return false
}

func init() {
	validate = validator.New()
	validate.RegisterValidation("notwhitespace", notWhiteSpace)
	validate.RegisterValidation("enddate", validateEndDate)
	validate.RegisterValidation("payment_frequency", PaymentFrequencyValidator)
	validate.RegisterValidation("required_and_notwhitespace", requiredAndNotWhiteSpace)
	validate.RegisterValidation("day", DayValidator)
}

func DecodeRequestBody(body io.Reader, target interface{}) *utils.HTTPError {
	if err := json.NewDecoder(body).Decode(target); err != nil {
		return utils.CreateHTTPError(fmt.Sprintf("Bad request: %v", err), http.StatusUnprocessableEntity)
	}
	return nil
}

func ValidateDto(dto interface{}) *utils.HTTPError {

	dtoType := reflect.TypeOf(dto).Elem()

	if err := validate.Struct(dto); err != nil {
		return parseValidationErrors(err, dtoType)
	}

	return nil

}

func DecodeAndValidateRequestBody(body io.Reader, target interface{}) *utils.HTTPError {
	if err := DecodeRequestBody(body, target); err != nil {
		return err
	}

	if err := ValidateDto(target); err != nil {
		return err
	}

	return nil
}

func parseValidationErrors(err error, structType reflect.Type) *utils.HTTPError {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {

		var errorMessages []string

		for _, e := range validationErrors {

			fieldName := getJSONFieldName(e, structType)
			var customMessage string

			switch e.Tag() {

			case "email":
				customMessage = fmt.Sprintf("%s: must be a valid email address", fieldName)
			case "enddate":
				customMessage = fmt.Sprintf("%s: EndDate must be after StartDate", fieldName)
			case "required":
				customMessage = fmt.Sprintf("%s: required", fieldName)
			case "notwhitespace":
				customMessage = fmt.Sprintf("%s: cannot be empty or whitespace", fieldName)
			case "payment_frequency":
				customMessage = fmt.Sprintf("%s: must be one of 'week', 'month', or 'day'", fieldName)
			case "day":
				customMessage = fmt.Sprintf("%s: must be between 0 and 6 (inclusive)", fieldName)
			default:
				customMessage = fmt.Sprintf("%s: validation failed on '%s'", fieldName, e.Tag())
			}

			errorMessages = append(errorMessages, customMessage)

		}

		return utils.CreateHTTPError(strings.Join(errorMessages, ", "), http.StatusBadRequest)
	}

	// Handle other validation errors
	fmt.Printf("Unhandled validation error: %v\n", err)
	return utils.CreateHTTPError("Internal server error", http.StatusInternalServerError)
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
