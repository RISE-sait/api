package auth

type UserInfo struct {
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	StaffInfo *StaffInfo `json:"staff_info,omitempty"`
}

type StaffInfo struct {
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

type JWTClaims struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	IsActiveStaff bool   `json:"is_active_staff"`
	IssuedAt      int64  `json:"iat"`
	ExpiresAt     int64  `json:"exp"`
	Issuer        string `json:"iss"`
}
