package entity

type UserInfo struct {
	FirstName *string
	LastName  *string
	Email     string
	StaffInfo *StaffInfo
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
