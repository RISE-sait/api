package handlers

import (
	"api/internal/libs/errors"
	"encoding/json"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, err *errors.CommonError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPCode)
	response := map[string]interface{}{
		"error": map[string]string{
			"message": err.Message,
		},
	}
	json.NewEncoder(w).Encode(response)
}
