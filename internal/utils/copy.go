package utils

import (
	"net/http"

	"github.com/jinzhu/copier"
)

func CopyStruct(src interface{}, dst interface{}) (*HTTPError, bool) {
	if err := copier.Copy(dst, src); err != nil {
		return CreateHTTPError("Failed to copy data", http.StatusInternalServerError), false
	}
	return nil, true
}
