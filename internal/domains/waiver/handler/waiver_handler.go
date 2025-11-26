package waiver

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"api/internal/di"
	dbIdentity "api/internal/domains/identity/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/services/gcp"
	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type WaiverHandler struct {
	queries *dbIdentity.Queries
}

func NewWaiverHandler(container *di.Container) *WaiverHandler {
	return &WaiverHandler{
		queries: dbIdentity.New(container.DB),
	}
}

// UploadWaiver handles waiver document uploads to cloud storage
// @Summary Upload signed waiver document
// @Description Accepts a signed waiver document (PDF or image) and uploads it to Google Cloud Storage
// @Tags waivers
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Waiver file to upload (pdf, jpg, jpeg, png)"
// @Param user_id query string false "Target user ID (required for admin uploads, optional for self-upload)"
// @Param notes query string false "Optional notes about the waiver"
// @Security Bearer
// @Success 201 {object} map[string]interface{} "Waiver uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid file or file type"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 413 {object} map[string]interface{} "Payload Too Large: File size exceeds limit"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /waivers/upload [post]
func (h *WaiverHandler) UploadWaiver(w http.ResponseWriter, r *http.Request) {
	// Limit file size to 20MB for PDFs
	r.ParseMultipartForm(20 << 20)

	// Get the file from the request
	file, header, err := r.FormFile("file")
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("No file provided", http.StatusBadRequest))
		return
	}
	defer file.Close()

	// Validate file size (20MB limit)
	if header.Size > 20<<20 {
		responseHandlers.RespondWithError(w, errLib.New("File size exceeds 20MB limit", http.StatusRequestEntityTooLarge))
		return
	}

	// Validate file type
	if !isValidWaiverType(header.Filename) {
		responseHandlers.RespondWithError(w, errLib.New("Invalid file type. Only pdf, jpg, jpeg, and png are allowed", http.StatusBadRequest))
		return
	}

	// Get uploader ID
	uploaderID, uploaderErr := contextUtils.GetUserID(r.Context())
	if uploaderErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Unauthorized", http.StatusUnauthorized))
		return
	}

	// Determine target user
	targetUserIDParam := r.URL.Query().Get("user_id")
	var targetUserID uuid.UUID
	var uploadedBy uuid.NullUUID

	if targetUserIDParam != "" {
		// Admin uploading for another user
		isStaff, staffErr := contextUtils.IsStaff(r.Context())
		if staffErr != nil || !isStaff {
			responseHandlers.RespondWithError(w, errLib.New("Only staff can upload waivers for other users", http.StatusForbidden))
			return
		}

		parsedID, parseErr := uuid.Parse(targetUserIDParam)
		if parseErr != nil {
			responseHandlers.RespondWithError(w, errLib.New("Invalid user_id format", http.StatusBadRequest))
			return
		}
		targetUserID = parsedID
		uploadedBy = uuid.NullUUID{UUID: uploaderID, Valid: true}
	} else {
		// User uploading for themselves
		targetUserID = uploaderID
		uploadedBy = uuid.NullUUID{Valid: false}
	}

	// Get notes from query params
	notes := r.URL.Query().Get("notes")

	// Get user's name for the folder path
	userInfo, userErr := h.queries.GetUserByIdOrEmail(r.Context(), dbIdentity.GetUserByIdOrEmailParams{
		ID:    uuid.NullUUID{UUID: targetUserID, Valid: true},
		Email: sql.NullString{Valid: false},
	})

	// Generate folder name using first_last or fallback to user ID
	var folderName string
	if userErr == nil && userInfo.FirstName != "" && userInfo.LastName != "" {
		// Sanitize names: lowercase, replace spaces with underscores, remove special chars
		firstName := strings.ToLower(strings.TrimSpace(userInfo.FirstName))
		lastName := strings.ToLower(strings.TrimSpace(userInfo.LastName))
		// Replace spaces and remove non-alphanumeric characters
		firstName = strings.ReplaceAll(firstName, " ", "_")
		lastName = strings.ReplaceAll(lastName, " ", "_")
		folderName = fmt.Sprintf("%s_%s_%s", firstName, lastName, targetUserID.String()[:8])
	} else {
		folderName = targetUserID.String()
	}

	// Generate filename
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	timestamp := time.Now().Unix()
	fileName := fmt.Sprintf("waivers/%s/waiver_%d%s", folderName, timestamp, fileExt)

	// Upload to GCP Storage
	publicURL, uploadErr := gcp.UploadImageToGCP(file, fileName)
	if uploadErr != nil {
		responseHandlers.RespondWithError(w, uploadErr)
		return
	}

	// Save to database
	waiverUpload, dbErr := h.queries.CreateWaiverUpload(r.Context(), dbIdentity.CreateWaiverUploadParams{
		UserID:        targetUserID,
		FileUrl:       publicURL,
		FileName:      header.Filename,
		FileType:      strings.TrimPrefix(fileExt, "."),
		FileSizeBytes: sql.NullInt64{Int64: header.Size, Valid: true},
		UploadedBy:    uploadedBy,
		Notes:         sql.NullString{String: notes, Valid: notes != ""},
	})
	if dbErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to save waiver record", http.StatusInternalServerError))
		return
	}

	response := map[string]interface{}{
		"id":          waiverUpload.ID,
		"message":     "Waiver uploaded successfully",
		"url":         publicURL,
		"filename":    header.Filename,
		"size_bytes":  header.Size,
		"uploaded_at": waiverUpload.CreatedAt,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusCreated)
}

// GetUserWaivers retrieves all uploaded waivers for a user
// @Summary Get user's uploaded waivers
// @Description Retrieves all waiver documents uploaded for a specific user
// @Tags waivers
// @Produce json
// @Param user_id path string true "User ID"
// @Security Bearer
// @Success 200 {array} map[string]interface{} "List of waiver uploads"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid user ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /waivers/user/{user_id} [get]
func (h *WaiverHandler) GetUserWaivers(w http.ResponseWriter, r *http.Request) {
	userIDParam := chi.URLParam(r, "user_id")
	userID, parseErr := uuid.Parse(userIDParam)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid user_id format", http.StatusBadRequest))
		return
	}

	// Check authorization - user can view their own, staff can view anyone's
	currentUserID, _ := contextUtils.GetUserID(r.Context())
	isStaff, _ := contextUtils.IsStaff(r.Context())

	if currentUserID != userID && !isStaff {
		responseHandlers.RespondWithError(w, errLib.New("Not authorized to view this user's waivers", http.StatusForbidden))
		return
	}

	waivers, err := h.queries.GetWaiverUploadsByUserId(r.Context(), userID)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to retrieve waivers", http.StatusInternalServerError))
		return
	}

	// Transform to response format
	response := make([]map[string]interface{}, len(waivers))
	for i, w := range waivers {
		uploaderName := ""
		if w.UploaderFirstName.Valid && w.UploaderLastName.Valid {
			uploaderName = fmt.Sprintf("%s %s", w.UploaderFirstName.String, w.UploaderLastName.String)
		}

		response[i] = map[string]interface{}{
			"id":            w.ID,
			"file_url":      w.FileUrl,
			"file_name":     w.FileName,
			"file_type":     w.FileType,
			"file_size":     w.FileSizeBytes.Int64,
			"notes":         w.Notes.String,
			"uploaded_by":   uploaderName,
			"uploaded_at":   w.CreatedAt,
		}
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// DeleteWaiver deletes a waiver upload
// @Summary Delete a waiver upload
// @Description Deletes a waiver document (admin only)
// @Tags waivers
// @Produce json
// @Param id path string true "Waiver upload ID"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Waiver deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Admin only"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /waivers/{id} [delete]
func (h *WaiverHandler) DeleteWaiver(w http.ResponseWriter, r *http.Request) {
	// Admin only
	isStaff, staffErr := contextUtils.IsStaff(r.Context())
	if staffErr != nil || !isStaff {
		responseHandlers.RespondWithError(w, errLib.New("Admin access required", http.StatusForbidden))
		return
	}

	idParam := chi.URLParam(r, "id")
	id, parseErr := uuid.Parse(idParam)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid waiver ID format", http.StatusBadRequest))
		return
	}

	// Get the waiver record first to get the file URL
	waiver, getErr := h.queries.GetWaiverUploadById(r.Context(), id)
	if getErr != nil {
		if getErr == sql.ErrNoRows {
			responseHandlers.RespondWithError(w, errLib.New("Waiver not found", http.StatusNotFound))
			return
		}
		responseHandlers.RespondWithError(w, errLib.New("Failed to retrieve waiver", http.StatusInternalServerError))
		return
	}

	// Delete from GCP storage
	if gcpErr := gcp.DeleteFileFromGCP(waiver.FileUrl); gcpErr != nil {
		// Log the error but continue with database deletion
		// The file might have been manually deleted or doesn't exist
		fmt.Printf("Warning: Failed to delete waiver file from GCP: %s, error: %v\n", waiver.FileUrl, gcpErr)
	}

	// Delete from database
	rowsAffected, err := h.queries.DeleteWaiverUpload(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to delete waiver", http.StatusInternalServerError))
		return
	}

	if rowsAffected == 0 {
		responseHandlers.RespondWithError(w, errLib.New("Waiver not found", http.StatusNotFound))
		return
	}

	responseHandlers.RespondWithSuccess(w, map[string]string{"message": "Waiver deleted successfully"}, http.StatusOK)
}

// isValidWaiverType checks if the file has a valid waiver extension
func isValidWaiverType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := []string{".pdf", ".jpg", ".jpeg", ".png"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}
