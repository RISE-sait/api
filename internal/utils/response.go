package utils

import (
	"api/internal/types"
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
		json.NewEncoder(w).Encode(response)
	}
}

func RespondWithError(w http.ResponseWriter, err *types.HTTPError) {

	w.WriteHeader(err.StatusCode)
	http.Error(w, err.Message, err.StatusCode)
}
