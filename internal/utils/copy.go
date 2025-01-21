package utils

import (
	"api/internal/types"
	"net/http"

	"github.com/jinzhu/copier"
)

func CopyStruct(src interface{}, dst interface{}) (*types.HTTPError, bool) {
	if err := copier.Copy(dst, src); err != nil {
		return CreateHTTPError("Failed to copy data", http.StatusInternalServerError), false
	}
	return nil, true
}
