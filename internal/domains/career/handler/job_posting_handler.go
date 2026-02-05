package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"api/internal/di"
	"api/internal/domains/career/dto"
	db "api/internal/domains/career/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type JobPostingHandler struct {
	Queries *db.Queries
}

func NewJobPostingHandler(container *di.Container) *JobPostingHandler {
	return &JobPostingHandler{Queries: container.Queries.CareersDb}
}

func (h *JobPostingHandler) ListPublishedJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.Queries.ListPublishedJobPostings(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to list jobs", http.StatusInternalServerError))
		return
	}

	resp := make([]dto.JobPostingResponse, len(jobs))
	for i, j := range jobs {
		resp[i] = mapJobPostingToResponse(j)
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}

func (h *JobPostingHandler) GetJobPosting(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, parseErr := validators.ParseUUID(idStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	// Check if user is admin - if so, show any status
	role, _ := contextUtils.GetUserRole(r.Context())
	isAdmin := role == contextUtils.RoleAdmin || role == contextUtils.RoleSuperAdmin || role == contextUtils.RoleIT

	var job db.CareersJobPosting
	var err error

	if isAdmin {
		job, err = h.Queries.GetJobPostingById(r.Context(), id)
	} else {
		job, err = h.Queries.GetPublishedJobPostingById(r.Context(), id)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Job posting not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to get job posting", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapJobPostingToResponse(job), http.StatusOK)
}

func (h *JobPostingHandler) ListAllJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.Queries.ListAllJobPostings(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to list jobs", http.StatusInternalServerError))
		return
	}

	resp := make([]dto.JobPostingResponse, len(jobs))
	for i, j := range jobs {
		resp[i] = mapJobPostingToResponse(j)
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}

func (h *JobPostingHandler) CreateJobPosting(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateJobPostingRequest
	if parseErr := validators.ParseJSON(r.Body, &req); parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}
	if valErr := validators.ValidateDto(&req); valErr != nil {
		responseHandlers.RespondWithError(w, valErr)
		return
	}

	userID, userErr := contextUtils.GetUserID(r.Context())
	if userErr != nil {
		responseHandlers.RespondWithError(w, userErr)
		return
	}

	params := db.CreateJobPostingParams{
		Title:          req.Title,
		Position:       req.Position,
		EmploymentType: req.EmploymentType,
		LocationType:   req.LocationType,
		Description:    req.Description,
		ShowSalary:     req.ShowSalary,
		Status:         "draft",
		CreatedBy:      uuid.NullUUID{UUID: userID, Valid: true},
	}

	if req.Responsibilities != nil {
		params.Responsibilities = req.Responsibilities
	} else {
		params.Responsibilities = []string{}
	}
	if req.Requirements != nil {
		params.Requirements = req.Requirements
	} else {
		params.Requirements = []string{}
	}
	if req.NiceToHave != nil {
		params.NiceToHave = req.NiceToHave
	} else {
		params.NiceToHave = []string{}
	}

	if req.SalaryMin != nil {
		params.SalaryMin = sql.NullString{String: fmt.Sprintf("%.2f", *req.SalaryMin), Valid: true}
	}
	if req.SalaryMax != nil {
		params.SalaryMax = sql.NullString{String: fmt.Sprintf("%.2f", *req.SalaryMax), Valid: true}
	}
	if req.ClosingDate != nil {
		t, tErr := time.Parse(time.RFC3339, *req.ClosingDate)
		if tErr != nil {
			responseHandlers.RespondWithError(w, errLib.New("Invalid closing_date format, use RFC3339", http.StatusBadRequest))
			return
		}
		params.ClosingDate = sql.NullTime{Time: t, Valid: true}
	}

	job, err := h.Queries.CreateJobPosting(r.Context(), params)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to create job posting", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapJobPostingToResponse(job), http.StatusCreated)
}

func (h *JobPostingHandler) UpdateJobPosting(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, parseErr := validators.ParseUUID(idStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	var req dto.UpdateJobPostingRequest
	if jsonErr := validators.ParseJSON(r.Body, &req); jsonErr != nil {
		responseHandlers.RespondWithError(w, jsonErr)
		return
	}
	if valErr := validators.ValidateDto(&req); valErr != nil {
		responseHandlers.RespondWithError(w, valErr)
		return
	}

	params := db.UpdateJobPostingParams{
		ID:             id,
		Title:          req.Title,
		Position:       req.Position,
		EmploymentType: req.EmploymentType,
		LocationType:   req.LocationType,
		Description:    req.Description,
		ShowSalary:     req.ShowSalary,
	}

	if req.Responsibilities != nil {
		params.Responsibilities = req.Responsibilities
	} else {
		params.Responsibilities = []string{}
	}
	if req.Requirements != nil {
		params.Requirements = req.Requirements
	} else {
		params.Requirements = []string{}
	}
	if req.NiceToHave != nil {
		params.NiceToHave = req.NiceToHave
	} else {
		params.NiceToHave = []string{}
	}

	if req.SalaryMin != nil {
		params.SalaryMin = sql.NullString{String: fmt.Sprintf("%.2f", *req.SalaryMin), Valid: true}
	}
	if req.SalaryMax != nil {
		params.SalaryMax = sql.NullString{String: fmt.Sprintf("%.2f", *req.SalaryMax), Valid: true}
	}
	if req.ClosingDate != nil {
		t, tErr := time.Parse(time.RFC3339, *req.ClosingDate)
		if tErr != nil {
			responseHandlers.RespondWithError(w, errLib.New("Invalid closing_date format, use RFC3339", http.StatusBadRequest))
			return
		}
		params.ClosingDate = sql.NullTime{Time: t, Valid: true}
	}

	job, err := h.Queries.UpdateJobPosting(r.Context(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Job posting not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to update job posting", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapJobPostingToResponse(job), http.StatusOK)
}

func (h *JobPostingHandler) UpdateJobStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, parseErr := validators.ParseUUID(idStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	var req dto.UpdateJobStatusRequest
	if jsonErr := validators.ParseJSON(r.Body, &req); jsonErr != nil {
		responseHandlers.RespondWithError(w, jsonErr)
		return
	}
	if valErr := validators.ValidateDto(&req); valErr != nil {
		responseHandlers.RespondWithError(w, valErr)
		return
	}

	job, err := h.Queries.UpdateJobPostingStatus(r.Context(), db.UpdateJobPostingStatusParams{
		ID:     id,
		Status: req.Status,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Job posting not found", http.StatusNotFound))
			return
		}
		log.Printf("Failed to update job status: %v", err)
		responseHandlers.RespondWithError(w, errLib.New("Failed to update job status", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapJobPostingToResponse(job), http.StatusOK)
}

func (h *JobPostingHandler) DeleteJobPosting(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, parseErr := validators.ParseUUID(idStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	rowsAffected, err := h.Queries.DeleteJobPosting(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to delete job posting", http.StatusInternalServerError))
		return
	}
	if rowsAffected == 0 {
		responseHandlers.RespondWithError(w, errLib.New("Job posting not found", http.StatusNotFound))
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
