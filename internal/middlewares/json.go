package middlewares

import "net/http"

// SetJSONContentType is a middleware that sets the "Content-Type" header to
// "application/json" for all HTTP responses.
//
// This middleware wraps an existing http.Handler and ensures that any response
// will have the appropriate content type header for JSON data, regardless of
// whether the handler explicitly sets it.
//
// Parameters:
//   - handler: The http.Handler to wrap
//
// Returns:
//   - http.Handler: A new handler that sets the JSON content type
func SetJSONContentType(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler.ServeHTTP(w, r)
	})
}
