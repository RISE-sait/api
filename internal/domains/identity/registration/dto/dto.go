package dto

type CreateUserRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	Role           string `json:"role"`
	IsActiveStaff  bool   `json:"is_active_staff"`
	WaiverUrl      string `json:"waiver_url"`
	IsSignedWaiver bool   `json:"is_signed_waiver"`
}
