package response_handlers

import (
	"api/internal/libs/errors"
	"encoding/json"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, err *errLib.CommonError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPCode)
	response := map[string]interface{}{
		"error": map[string]string{
			"message": err.Message,
		},
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}
