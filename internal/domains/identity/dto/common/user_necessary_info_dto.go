package identity

type UserNecessaryInfoDto struct {
	FirstName string `json:"first_name" validate:"required,notwhitespace"`
	LastName  string `json:"last_name" validate:"required,notwhitespace"`
	Age       int    `json:"age" validate:"required,gt=0"`
}
