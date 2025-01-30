package response_handlers

import (
	errLib "api/internal/libs/errors"
	"encoding/json"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, err *errLib.CommonError) {

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
