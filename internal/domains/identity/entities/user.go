package entities

type UserInfo struct {
	Name      string
	Email     string
	StaffInfo StaffInfo
}

type StaffInfo struct {
	Role     string
	IsActive bool
}

type JWTClaims struct {
	Name          string
	Email         string
	Role          string
	IsActiveStaff bool
	IssuedAt      int64
	ExpiresAt     int64
	Issuer        string
}
