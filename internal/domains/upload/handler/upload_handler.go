package upload

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"api/internal/di"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/services/gcp"
	contextUtils "api/utils/context"

	"github.com/google/uuid"
)

type UploadHandler struct{}

func NewUploadHandler(container *di.Container) *UploadHandler {
	return &UploadHandler{}
}

// UploadImage handles image uploads to cloud storage
// @Summary Upload image to cloud storage
// @Description Accepts an image file and uploads it to Google Cloud Storage, returning the public URL
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file to upload (jpg, jpeg, png, gif, webp)"
// @Param folder query string false "Folder name in storage bucket" default("images")
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Image uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid file or file type"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 413 {object} map[string]interface{} "Payload Too Large: File size exceeds limit"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /upload/image [post]
func (h *UploadHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Limit file size to 10MB
	r.ParseMultipartForm(10 << 20)

	// Get the file from the request
	file, header, err := r.FormFile("image")
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("No image file provided", http.StatusBadRequest))
		return
	}
	defer file.Close()

	// Validate file size (10MB limit)
	if header.Size > 10<<20 {
		responseHandlers.RespondWithError(w, errLib.New("File size exceeds 10MB limit", http.StatusRequestEntityTooLarge))
		return
	}

	// Validate file type
	if !isValidImageType(header.Filename) {
		responseHandlers.RespondWithError(w, errLib.New("Invalid file type. Only jpg, jpeg, png, gif, and webp are allowed", http.StatusBadRequest))
		return
	}

	// Get folder from query params (default to "images")
	folder := r.URL.Query().Get("folder")
	if folder == "" {
		folder = "images"
	}

	// Get user ID for organized storage
	userID, userErr := contextUtils.GetUserID(r.Context())
	var userPrefix string
	if userErr == nil {
		userPrefix = userID.String()
	} else {
		userPrefix = "anonymous"
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	uniqueID := uuid.New().String()
	fileName := fmt.Sprintf("%s/%s/%d_%s%s", folder, userPrefix, timestamp, uniqueID, fileExt)

	// Upload to GCP Storage
	publicURL, uploadErr := gcp.UploadImageToGCP(file, fileName)
	if uploadErr != nil {
		responseHandlers.RespondWithError(w, uploadErr)
		return
	}

	response := map[string]interface{}{
		"message":    "Image uploaded successfully",
		"url":        publicURL,
		"filename":   fileName,
		"size_bytes": header.Size,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// isValidImageType checks if the file has a valid image extension
func isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	
	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}