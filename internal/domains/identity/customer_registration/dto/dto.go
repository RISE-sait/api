package dto

type CreateUserRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	WaiverUrl      string `json:"waiver_url"`
	IsSignedWaiver bool   `json:"is_signed_waiver"`
}
