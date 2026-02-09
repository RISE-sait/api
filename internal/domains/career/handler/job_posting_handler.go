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

// ListPublishedJobs returns all published job postings.
// @Summary List published job postings
// @Description Returns all job postings with status "published" for public viewing
// @Tags careers
// @Produce json
// @Success 200 {array} dto.JobPostingResponse "List of published job postings"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /jobs [get]
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

// GetJobPosting retrieves a single job posting by ID.
// @Summary Get job posting by ID
// @Description Returns a job posting. Public users only see published postings; admins see all statuses.
// @Tags careers
// @Produce json
// @Param id path string true "Job Posting ID" format(uuid)
// @Success 200 {object} dto.JobPostingResponse "Job posting details"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Job posting not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /jobs/{id} [get]
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

// ListAllJobs returns all job postings regardless of status (admin only).
// @Summary List all job postings
// @Description Returns all job postings including drafts, closed, etc. Admin only.
// @Tags careers
// @Produce json
// @Security Bearer
// @Success 200 {array} dto.JobPostingResponse "List of all job postings"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /jobs/all [get]
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

// CreateJobPosting creates a new job posting (admin only).
// @Summary Create job posting
// @Description Creates a new job posting with draft status. Admin only.
// @Tags careers
// @Accept json
// @Produce json
// @Param job body dto.CreateJobPostingRequest true "Job posting details"
// @Security Bearer
// @Success 201 {object} dto.JobPostingResponse "Created job posting"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /jobs [post]
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

// UpdateJobPosting updates an existing job posting (admin only).
// @Summary Update job posting
// @Description Updates job posting details. Admin only.
// @Tags careers
// @Accept json
// @Produce json
// @Param id path string true "Job Posting ID" format(uuid)
// @Param job body dto.UpdateJobPostingRequest true "Updated job posting details"
// @Security Bearer
// @Success 200 {object} dto.JobPostingResponse "Updated job posting"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found: Job posting not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /jobs/{id} [put]
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

// UpdateJobStatus updates the status of a job posting (admin only).
// @Summary Update job posting status
// @Description Updates job posting status (draft, published, closed). Admin only.
// @Tags careers
// @Accept json
// @Produce json
// @Param id path string true "Job Posting ID" format(uuid)
// @Param status body dto.UpdateJobStatusRequest true "New status"
// @Security Bearer
// @Success 200 {object} dto.JobPostingResponse "Updated job posting"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found: Job posting not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /jobs/{id}/status [patch]
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

// DeleteJobPosting deletes a job posting (admin only).
// @Summary Delete job posting
// @Description Permanently deletes a job posting. Admin only.
// @Tags careers
// @Param id path string true "Job Posting ID" format(uuid)
// @Security Bearer
// @Success 204 "Job posting deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found: Job posting not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /jobs/{id} [delete]
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
