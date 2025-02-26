package response_handlers

import (
	errLib "api/internal/libs/errors"
	"encoding/json"
	"net/http"
)

// RespondWithSuccess sends a JSON response with the provided payload and HTTP status code.
//
// Parameters:
//   - w: The http.ResponseWriter to write the response to.
//   - payload: The data to be encoded as JSON and sent in the response body. Can be nil.
//   - status: The HTTP status code to set in the response.
//
// Behavior:
//   - Sets the HTTP status code in the response.
//   - If a payload is provided, it is encoded as JSON and sent in the response body.
//   - If JSON encoding fails, it returns an internal server error response using RespondWithError.
func RespondWithSuccess(w http.ResponseWriter, payload interface{}, status int) {

	w.WriteHeader(status)

	if payload != nil {

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {

			newErr := errLib.New("Failed to encode response", http.StatusInternalServerError)
			RespondWithError(w, newErr)
		}
	}
}
