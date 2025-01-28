package dto

type GetUserOptionalInfoRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	Role           string `json:"role"`
	IsActiveStaff  bool   `json:"is_active_staff"`
	WaiverUrl      string `json:"waiver_url"`
	IsSignedWaiver bool   `json:"is_signed_waiver"`
}

type UpdateUserNameRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateUserPasswordRequest struct {
	Email          string `json:"email" `
	HashedPassword string `json:"hashed_password"`
}
