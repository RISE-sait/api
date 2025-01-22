package response_handlers

import (
	errLib "api/internal/libs/errors"
	"encoding/json"
	"net/http"
)

type Response struct {
	Data interface{} `json:"data,omitempty"`
}

func RespondWithSuccess(w http.ResponseWriter, payload interface{}, status int) {

	w.WriteHeader(status)

	if payload != nil {
		response := Response{
			Data: payload,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {

			newErr := errLib.New("Failed to encode response", http.StatusInternalServerError)
			RespondWithError(w, newErr)
		}
	}
}
