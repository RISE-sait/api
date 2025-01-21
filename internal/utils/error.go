package utils

import "api/internal/types"

func CreateHTTPError(message string, statusCode int) *types.HTTPError {
	return &types.HTTPError{Message: message, StatusCode: statusCode}
}
