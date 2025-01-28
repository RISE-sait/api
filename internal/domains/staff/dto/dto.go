package dto

type CreateStaffRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	Role          string `json:"role"`
	IsActiveStaff bool   `json:"is_active_staff"`
}
