package identity

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"regexp"
)

type CustomerWaiverCreateDto struct {
	WaiverUrl      string `json:"waiver_url"`
	IsWaiverSigned bool   `json:"is_waiver_signed"`
}

func NewCustomerWaiverCreateDto(waiverUrl string, isSigned bool) *CustomerWaiverCreateDto {
	return &CustomerWaiverCreateDto{
		WaiverUrl:      waiverUrl,
		IsWaiverSigned: isSigned,
	}
}

func (vu *CustomerWaiverCreateDto) Validate() *errLib.CommonError {

	if vu.WaiverUrl == "" {
		return errLib.New("'waiver_url' cannot be empty or whitespace", http.StatusBadRequest)
	}

	// validate waiver url
	urlRegex := regexp.MustCompile(`^(http|https):\/\/[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(\/\S*)?$`)
	if !urlRegex.MatchString(vu.WaiverUrl) {
		return errLib.New("Invalid URL format", http.StatusBadRequest)
	}

	return nil
}
