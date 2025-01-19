package utils

// HTTPError is a custom error type that includes an error message and an HTTP status code
type HTTPError struct {
	Message    string
	StatusCode int
}

func CreateHTTPError(message string, statusCode int) *HTTPError {
	return &HTTPError{Message: message, StatusCode: statusCode}
}
