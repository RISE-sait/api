package handler

import (
	"encoding/json"
	"net/http"

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
