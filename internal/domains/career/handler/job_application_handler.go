package handler

import (
	"database/sql"
	"log"
	"net/http"

	"api/internal/di"
	"api/internal/domains/career/dto"
	db "api/internal/domains/career/persistence/sqlc/generated"
	"api/internal/domains/career/service"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type JobApplicationHandler struct {
	Queries *db.Queries
}

func NewJobApplicationHandler(container *di.Container) *JobApplicationHandler {
	return &JobApplicationHandler{Queries: container.Queries.CareersDb}
}

// SubmitApplication submits a job application with resume upload.
// @Summary Submit job application
// @Description Submit an application for a published job posting with resume file upload
// @Tags careers
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Job Posting ID" format(uuid)
// @Param first_name formData string true "Applicant first name"
// @Param last_name formData string true "Applicant last name"
// @Param email formData string true "Applicant email"
// @Param resume formData file true "Resume file (PDF, DOC, DOCX)"
// @Param phone formData string false "Phone number"
// @Param cover_letter formData string false "Cover letter text"
// @Param linkedin_url formData string false "LinkedIn profile URL"
// @Param portfolio_url formData string false "Portfolio URL"
// @Success 201 {object} dto.JobApplicationResponse "Application submitted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input or missing required fields"
// @Failure 404 {object} map[string]interface{} "Not Found: Job posting not found or not accepting applications"
// @Failure 429 {object} map[string]interface{} "Too Many Requests: Rate limit exceeded"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /jobs/{id}/apply [post]
func (h *JobApplicationHandler) SubmitApplication(w http.ResponseWriter, r *http.Request) {
	jobIdStr := chi.URLParam(r, "id")
	jobID, parseErr := validators.ParseUUID(jobIdStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	// Verify job is published
	job, jobErr := h.Queries.GetPublishedJobPostingById(r.Context(), jobID)
	if jobErr != nil {
		if jobErr == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Job posting not found or not accepting applications", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to verify job posting", http.StatusInternalServerError))
		return
	}

	// Parse multipart form (max 10MB total)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to parse form data", http.StatusBadRequest))
		return
	}

	// Required fields
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")

	if firstName == "" || lastName == "" || email == "" {
		responseHandlers.RespondWithError(w, errLib.New("first_name, last_name, and email are required", http.StatusBadRequest))
		return
	}

	// Resume file (required)
	file, header, fileErr := r.FormFile("resume")
	if fileErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Resume file is required", http.StatusBadRequest))
		return
	}
	defer file.Close()

	resumeURL, uploadErr := service.UploadResume(file, header.Filename, header.Size)
	if uploadErr != nil {
		responseHandlers.RespondWithError(w, uploadErr)
		return
	}

	// Build params
	params := db.CreateJobApplicationParams{
		JobID:     jobID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		ResumeUrl: resumeURL,
	}

	if phone := r.FormValue("phone"); phone != "" {
		params.Phone = sql.NullString{String: phone, Valid: true}
	}
	if coverLetter := r.FormValue("cover_letter"); coverLetter != "" {
		params.CoverLetter = sql.NullString{String: coverLetter, Valid: true}
	}
	if linkedinURL := r.FormValue("linkedin_url"); linkedinURL != "" {
		params.LinkedinUrl = sql.NullString{String: linkedinURL, Valid: true}
	}
	if portfolioURL := r.FormValue("portfolio_url"); portfolioURL != "" {
		params.PortfolioUrl = sql.NullString{String: portfolioURL, Valid: true}
	}

	application, createErr := h.Queries.CreateJobApplication(r.Context(), params)
	if createErr != nil {
		log.Printf("Failed to create job application: %v", createErr)
		responseHandlers.RespondWithError(w, errLib.New("Failed to submit application", http.StatusInternalServerError))
		return
	}

	// Send emails (fire-and-forget)
	service.SendApplicationEmails(email, firstName, lastName, job.Title)

	responseHandlers.RespondWithSuccess(w, mapJobApplicationToResponse(application), http.StatusCreated)
}

// ListApplicationsByJob lists all applications for a specific job posting (admin only).
// @Summary List applications by job
// @Description Returns all applications for a specific job posting. Admin only.
// @Tags careers
// @Produce json
// @Param job_id path string true "Job Posting ID" format(uuid)
// @Security Bearer
// @Success 200 {array} dto.JobApplicationResponse "List of applications"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /jobs/{job_id}/applications [get]
func (h *JobApplicationHandler) ListApplicationsByJob(w http.ResponseWriter, r *http.Request) {
	jobIdStr := chi.URLParam(r, "job_id")
	jobID, parseErr := validators.ParseUUID(jobIdStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	apps, err := h.Queries.ListJobApplicationsByJobId(r.Context(), jobID)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to list applications", http.StatusInternalServerError))
		return
	}

	resp := make([]dto.JobApplicationResponse, len(apps))
	for i, a := range apps {
		resp[i] = mapJobApplicationToResponse(a)
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}

// ListAllApplications lists all job applications (admin only).
// @Summary List all applications
// @Description Returns all job applications across all job postings. Admin only.
// @Tags careers
// @Produce json
// @Security Bearer
// @Success 200 {array} dto.JobApplicationResponse "List of all applications"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /applications [get]
func (h *JobApplicationHandler) ListAllApplications(w http.ResponseWriter, r *http.Request) {
	apps, err := h.Queries.ListAllJobApplications(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to list applications", http.StatusInternalServerError))
		return
	}

	resp := make([]dto.JobApplicationResponse, len(apps))
	for i, a := range apps {
		resp[i] = mapJobApplicationToResponse(a)
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}

// GetApplication retrieves a single application by ID (admin only).
// @Summary Get application by ID
// @Description Returns a single job application with full details. Admin only.
// @Tags careers
// @Produce json
// @Param id path string true "Application ID" format(uuid)
// @Security Bearer
// @Success 200 {object} dto.JobApplicationResponse "Application details"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found: Application not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /applications/{id} [get]
func (h *JobApplicationHandler) GetApplication(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, parseErr := validators.ParseUUID(idStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	app, err := h.Queries.GetJobApplicationById(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Application not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to get application", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapJobApplicationToResponse(app), http.StatusOK)
}

// UpdateApplicationStatus updates the status of an application (admin only).
// @Summary Update application status
// @Description Updates application status and sends notification email to applicant. Admin only.
// @Tags careers
// @Accept json
// @Produce json
// @Param id path string true "Application ID" format(uuid)
// @Param status body dto.UpdateApplicationStatusRequest true "New status"
// @Security Bearer
// @Success 200 {object} dto.JobApplicationResponse "Updated application"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found: Application not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /applications/{id}/status [patch]
func (h *JobApplicationHandler) UpdateApplicationStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, parseErr := validators.ParseUUID(idStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	var req dto.UpdateApplicationStatusRequest
	if jsonErr := validators.ParseJSON(r.Body, &req); jsonErr != nil {
		responseHandlers.RespondWithError(w, jsonErr)
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

	// Get application first to get email info for notifications
	existingApp, getErr := h.Queries.GetJobApplicationById(r.Context(), id)
	if getErr != nil {
		if getErr == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Application not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to get application", http.StatusInternalServerError))
		return
	}

	app, err := h.Queries.UpdateJobApplicationStatus(r.Context(), db.UpdateJobApplicationStatusParams{
		ID:         id,
		Status:     req.Status,
		ReviewedBy: uuid.NullUUID{UUID: userID, Valid: true},
	})
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to update application status", http.StatusInternalServerError))
		return
	}

	// Get job title for email
	job, jobErr := h.Queries.GetJobPostingById(r.Context(), existingApp.JobID)
	if jobErr == nil {
		service.SendStatusChangeEmail(existingApp.Email, existingApp.FirstName, job.Title, req.Status)
	}

	responseHandlers.RespondWithSuccess(w, mapJobApplicationToResponse(app), http.StatusOK)
}

// UpdateApplicationNotes updates internal notes on an application (admin only).
// @Summary Update application notes
// @Description Updates internal notes for an application. Admin only.
// @Tags careers
// @Accept json
// @Produce json
// @Param id path string true "Application ID" format(uuid)
// @Param notes body dto.UpdateApplicationNotesRequest true "Notes content"
// @Security Bearer
// @Success 200 {object} dto.JobApplicationResponse "Updated application"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found: Application not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /applications/{id}/notes [patch]
func (h *JobApplicationHandler) UpdateApplicationNotes(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, parseErr := validators.ParseUUID(idStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	var req dto.UpdateApplicationNotesRequest
	if jsonErr := validators.ParseJSON(r.Body, &req); jsonErr != nil {
		responseHandlers.RespondWithError(w, jsonErr)
		return
	}

	app, err := h.Queries.UpdateJobApplicationNotes(r.Context(), db.UpdateJobApplicationNotesParams{
		ID:            id,
		InternalNotes: sql.NullString{String: req.Notes, Valid: req.Notes != ""},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Application not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to update application notes", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapJobApplicationToResponse(app), http.StatusOK)
}

// UpdateApplicationRating updates the rating of an application (admin only).
// @Summary Update application rating
// @Description Updates the rating (1-5) for an application. Admin only.
// @Tags careers
// @Accept json
// @Produce json
// @Param id path string true "Application ID" format(uuid)
// @Param rating body dto.UpdateApplicationRatingRequest true "Rating value (1-5)"
// @Security Bearer
// @Success 200 {object} dto.JobApplicationResponse "Updated application"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found: Application not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /applications/{id}/rating [patch]
func (h *JobApplicationHandler) UpdateApplicationRating(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, parseErr := validators.ParseUUID(idStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	var req dto.UpdateApplicationRatingRequest
	if jsonErr := validators.ParseJSON(r.Body, &req); jsonErr != nil {
		responseHandlers.RespondWithError(w, jsonErr)
		return
	}
	if valErr := validators.ValidateDto(&req); valErr != nil {
		responseHandlers.RespondWithError(w, valErr)
		return
	}

	app, err := h.Queries.UpdateJobApplicationRating(r.Context(), db.UpdateJobApplicationRatingParams{
		ID:     id,
		Rating: sql.NullInt32{Int32: req.Rating, Valid: true},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Application not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to update application rating", http.StatusInternalServerError))
		return
	}

	responseHandlers.RespondWithSuccess(w, mapJobApplicationToResponse(app), http.StatusOK)
}
