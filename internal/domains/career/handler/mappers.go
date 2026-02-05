package handler

import (
	db "api/internal/domains/career/persistence/sqlc/generated"
	"api/internal/domains/career/dto"
)

func mapJobPostingToResponse(p db.CareersJobPosting) dto.JobPostingResponse {
	resp := dto.JobPostingResponse{
		ID:               p.ID.String(),
		Title:            p.Title,
		Position:         p.Position,
		EmploymentType:   p.EmploymentType,
		LocationType:     p.LocationType,
		Description:      p.Description,
		Responsibilities: p.Responsibilities,
		Requirements:     p.Requirements,
		NiceToHave:       p.NiceToHave,
		ShowSalary:       p.ShowSalary,
		Status:           p.Status,
	}

	if resp.Responsibilities == nil {
		resp.Responsibilities = []string{}
	}
	if resp.Requirements == nil {
		resp.Requirements = []string{}
	}
	if resp.NiceToHave == nil {
		resp.NiceToHave = []string{}
	}

	if p.SalaryMin.Valid {
		resp.SalaryMin = &p.SalaryMin.String
	}
	if p.SalaryMax.Valid {
		resp.SalaryMax = &p.SalaryMax.String
	}
	if p.ClosingDate.Valid {
		s := p.ClosingDate.Time.Format("2006-01-02T15:04:05Z07:00")
		resp.ClosingDate = &s
	}
	if p.CreatedBy.Valid {
		s := p.CreatedBy.UUID.String()
		resp.CreatedBy = &s
	}
	if p.PublishedAt.Valid {
		s := p.PublishedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		resp.PublishedAt = &s
	}
	if p.CreatedAt.Valid {
		t := p.CreatedAt.Time
		resp.CreatedAt = &t
	}
	if p.UpdatedAt.Valid {
		t := p.UpdatedAt.Time
		resp.UpdatedAt = &t
	}

	return resp
}

func mapJobApplicationToResponse(a db.CareersJobApplication) dto.JobApplicationResponse {
	resp := dto.JobApplicationResponse{
		ID:        a.ID.String(),
		JobID:     a.JobID.String(),
		FirstName: a.FirstName,
		LastName:  a.LastName,
		Email:     a.Email,
		ResumeURL: a.ResumeUrl,
		Status:    a.Status,
	}

	if a.Phone.Valid {
		resp.Phone = &a.Phone.String
	}
	if a.CoverLetter.Valid {
		resp.CoverLetter = &a.CoverLetter.String
	}
	if a.LinkedinUrl.Valid {
		resp.LinkedinURL = &a.LinkedinUrl.String
	}
	if a.PortfolioUrl.Valid {
		resp.PortfolioURL = &a.PortfolioUrl.String
	}
	if a.InternalNotes.Valid {
		resp.InternalNotes = &a.InternalNotes.String
	}
	if a.Rating.Valid {
		resp.Rating = &a.Rating.Int32
	}
	if a.ReviewedBy.Valid {
		s := a.ReviewedBy.UUID.String()
		resp.ReviewedBy = &s
	}
	if a.CreatedAt.Valid {
		t := a.CreatedAt.Time
		resp.CreatedAt = &t
	}
	if a.UpdatedAt.Valid {
		t := a.UpdatedAt.Time
		resp.UpdatedAt = &t
	}

	return resp
}
