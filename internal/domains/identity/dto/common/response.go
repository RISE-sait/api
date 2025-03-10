package identity

type UserAuthenticationResponseDto struct {
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Email       *string `json:"email,omitempty"`
	Role        string  `json:"role"`
	Phone       *string `json:"phone,omitempty"`
	Age         int     `json:"age"`
	CountryCode *string `json:"country_code,omitempty"`
}
