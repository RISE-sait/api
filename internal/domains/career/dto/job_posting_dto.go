package dto

import "time"

type CreateJobPostingRequest struct {
	Title            string   `json:"title" validate:"required,notwhitespace"`
	Position         string   `json:"position" validate:"required,notwhitespace"`
	EmploymentType   string   `json:"employment_type" validate:"required,oneof=full_time part_time contract internship volunteer"`
	LocationType     string   `json:"location_type" validate:"required,oneof=on_site remote hybrid"`
	Description      string   `json:"description" validate:"required,notwhitespace"`
	Responsibilities []string `json:"responsibilities"`
	Requirements     []string `json:"requirements"`
	NiceToHave       []string `json:"nice_to_have"`
	SalaryMin        *float64 `json:"salary_min"`
	SalaryMax        *float64 `json:"salary_max"`
	ShowSalary       bool     `json:"show_salary"`
	ClosingDate      *string  `json:"closing_date"`
}

type UpdateJobPostingRequest struct {
	Title            string   `json:"title" validate:"required,notwhitespace"`
	Position         string   `json:"position" validate:"required,notwhitespace"`
	EmploymentType   string   `json:"employment_type" validate:"required,oneof=full_time part_time contract internship volunteer"`
	LocationType     string   `json:"location_type" validate:"required,oneof=on_site remote hybrid"`
	Description      string   `json:"description" validate:"required,notwhitespace"`
	Responsibilities []string `json:"responsibilities"`
	Requirements     []string `json:"requirements"`
	NiceToHave       []string `json:"nice_to_have"`
	SalaryMin        *float64 `json:"salary_min"`
	SalaryMax        *float64 `json:"salary_max"`
	ShowSalary       bool     `json:"show_salary"`
	ClosingDate      *string  `json:"closing_date"`
}

type UpdateJobStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft published closed archived"`
}

type JobPostingResponse struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Position         string   `json:"position"`
	EmploymentType   string   `json:"employment_type"`
	LocationType     string   `json:"location_type"`
	Description      string   `json:"description"`
	Responsibilities []string `json:"responsibilities"`
	Requirements     []string `json:"requirements"`
	NiceToHave       []string `json:"nice_to_have"`
	SalaryMin        *string  `json:"salary_min,omitempty"`
	SalaryMax        *string  `json:"salary_max,omitempty"`
	ShowSalary       bool     `json:"show_salary"`
	Status           string   `json:"status"`
	ClosingDate      *string  `json:"closing_date,omitempty"`
	CreatedBy        *string  `json:"created_by,omitempty"`
	PublishedAt      *string  `json:"published_at,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}
