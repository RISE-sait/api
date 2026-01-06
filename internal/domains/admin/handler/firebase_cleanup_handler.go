package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api/internal/di"
	"api/internal/jobs"
)

// FirebaseCleanupResponse contains the results of a Firebase cleanup operation
type FirebaseCleanupResponse struct {
	TotalFirebaseUsers int      `json:"total_firebase_users"`
	TotalDBUsers       int      `json:"total_db_users"`
	OrphanedCount      int      `json:"orphaned_count"`
	DeletedCount       int      `json:"deleted_count"`
	FailedCount        int      `json:"failed_count"`
	DryRun             bool     `json:"dry_run"`
	OrphanedEmails     []string `json:"orphaned_emails,omitempty"`
	Errors             []string `json:"errors,omitempty"`
}

type FirebaseCleanupHandler struct {
	container *di.Container
}

func NewFirebaseCleanupHandler(container *di.Container) *FirebaseCleanupHandler {
	return &FirebaseCleanupHandler{
		container: container,
	}
}

// CleanupOrphanedFirebaseUsers handles the Firebase cleanup request
// @Summary Clean up orphaned Firebase users
// @Description Finds and optionally deletes Firebase users that don't exist in the database
// @Tags admin
// @Accept json
// @Produce json
// @Param dry_run query bool false "If true, only report what would be deleted without actually deleting (default: true)"
// @Success 200 {object} FirebaseCleanupResponse "Cleanup result"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/firebase/cleanup [post]
// @Security Bearer
func (h *FirebaseCleanupHandler) CleanupOrphanedFirebaseUsers(w http.ResponseWriter, r *http.Request) {
	// Default to dry run for safety
	dryRun := true
	if r.URL.Query().Get("dry_run") == "false" {
		dryRun = false
	}

	// Create and configure the cleanup job
	job := jobs.NewFirebaseCleanupJob(h.container)
	job.SetDryRun(dryRun)

	// Run the cleanup
	result, err := job.RunWithResult(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to run cleanup: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// RecoverMissingFirebaseUsers handles the Firebase recovery request
// @Summary Recover missing Firebase users
// @Description Finds DB users missing from Firebase and recreates their Firebase accounts
// @Tags admin
// @Accept json
// @Produce json
// @Param dry_run query bool false "If true, only report what would be recovered without actually creating (default: true)"
// @Param limit query int false "Max number of users to recover (default: 0 = unlimited). Use limit=1 to test with one user first."
// @Param email query string false "Specific email to recover (for testing). If set, only this email will be processed."
// @Success 200 {object} map[string]interface{} "Recovery result"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/firebase/recover [post]
// @Security Bearer
func (h *FirebaseCleanupHandler) RecoverMissingFirebaseUsers(w http.ResponseWriter, r *http.Request) {
	// Default to dry run for safety
	dryRun := true
	if r.URL.Query().Get("dry_run") == "false" {
		dryRun = false
	}

	// Parse limit parameter
	limit := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Parse target email parameter
	targetEmail := r.URL.Query().Get("email")

	// Create and configure the recovery job
	job := jobs.NewFirebaseRecoveryJob(h.container)
	job.SetDryRun(dryRun)
	job.SetLimit(limit)
	job.SetTargetEmail(targetEmail)

	// Run the recovery
	result, err := job.Run(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to run recovery: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
