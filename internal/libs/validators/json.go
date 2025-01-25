package validators

import (
	errLib "api/internal/libs/errors"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

func ParseJSON(body io.Reader, target interface{}) *errLib.CommonError {
	if err := validatePointerToStruct(target); err != nil {
		return err
	}

	if err := json.NewDecoder(body).Decode(target); err != nil {
		return errLib.New("Invalid JSON format", http.StatusBadRequest)
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
