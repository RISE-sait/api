package identity

type UserAuthenticationResponseDto struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     *string `json:"email,omitempty"`
	Role      string  `json:"role"`
	Age       int     `json:"age"`
}
