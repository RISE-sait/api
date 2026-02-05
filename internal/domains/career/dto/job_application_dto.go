package dto

import "time"

type JobApplicationResponse struct {
	ID            string     `json:"id"`
	JobID         string     `json:"job_id"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Email         string     `json:"email"`
	Phone         *string    `json:"phone,omitempty"`
	ResumeURL     string     `json:"resume_url"`
	CoverLetter   *string    `json:"cover_letter,omitempty"`
	LinkedinURL   *string    `json:"linkedin_url,omitempty"`
	PortfolioURL  *string    `json:"portfolio_url,omitempty"`
	Status        string     `json:"status"`
	InternalNotes *string    `json:"internal_notes,omitempty"`
	Rating        *int32     `json:"rating,omitempty"`
	ReviewedBy    *string    `json:"reviewed_by,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type UpdateApplicationStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=received reviewing interview offer hired rejected withdrawn"`
}

type UpdateApplicationNotesRequest struct {
	Notes string `json:"notes"`
}

type UpdateApplicationRatingRequest struct {
	Rating int32 `json:"rating" validate:"required,min=1,max=5"`
}
