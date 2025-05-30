package dto

type NewsletterRequest struct {
	Email string `json:"email"`
	Tag   string `json:"tag"`
}
