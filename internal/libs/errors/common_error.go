package errLib

type CommonError struct {
	Message  string
	HTTPCode int
}

// New creates a new CommonError
func New(message string, httpCode int) *CommonError {
	return &CommonError{
		Message:  message,
		HTTPCode: httpCode,
	}
}

// Error implements the error interface
func (e *CommonError) Error() string {
	return e.Message
}
