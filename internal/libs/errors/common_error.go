package errLib

type CommonError struct {
	Message  string
	HTTPCode int
}

func New(message string, httpCode int) *CommonError {
	return &CommonError{
		Message:  message,
		HTTPCode: httpCode,
	}
}
