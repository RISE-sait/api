package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
)

type WaiverCreate struct {
	Email     string
	WaiverUrl string
	IsSigned  bool
}

func NewWaiverCreate(email string, waiverUrl string, isSigned bool) *WaiverCreate {
	return &WaiverCreate{
		Email:     email,
		WaiverUrl: waiverUrl,
		IsSigned:  isSigned,
	}
}

func (w *WaiverCreate) Validate() *errLib.CommonError {
	if w.WaiverUrl == "" {
		return errLib.New("Waiver URL cannot be empty", http.StatusBadRequest)
	}
	return nil
}
