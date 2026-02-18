package family

import (
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

// RequestLinkRequest is used when initiating a link request
// The caller's role (parent or child) determines the flow:
// - If caller has no parent_id: they're a parent adding a child
// - If caller has parent_id: they're a child adding/changing a parent
type RequestLinkRequest struct {
	TargetEmail string `json:"target_email" validate:"required,email"`
}

func (dto *RequestLinkRequest) Validate() *errLib.CommonError {
	return validators.ValidateDto(dto)
}

// ConfirmLinkRequest is used to confirm a link request with a verification code
type ConfirmLinkRequest struct {
	Code string `json:"code" validate:"required,len=6,numeric"`
}

func (dto *ConfirmLinkRequest) Validate() *errLib.CommonError {
	return validators.ValidateDto(dto)
}
