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
	userRepo "api/internal/domains/identity/persistence/repository/user"

	"github.com/google/uuid"
)

type UploadHandler struct{
	userRepo *userRepo.UsersRepository
}

func NewUploadHandler(container *di.Container) *UploadHandler {
	return &UploadHandler{
		userRepo: userRepo.NewUserRepository(container),
	}
}

// UploadImage handles image uploads to cloud storage
// @Summary Upload image to cloud storage
// @Description Accepts an image file and uploads it to Google Cloud Storage, returning the public URL
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file to upload (jpg, jpeg, png, gif, webp)"
// @Param folder query string false "Folder name in storage bucket" default("images")
// @Param target_user_id query string false "Target user ID for admin uploads (admins only)"
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

	// Determine target user for upload (support admin uploads to other users)
	targetUserID, targetUserErr := contextUtils.GetUserID(r.Context())
	
	// Check if admin is uploading for another user
	targetUserIDParam := r.URL.Query().Get("target_user_id")
	if targetUserIDParam != "" {
		// Verify the requesting user is an admin
		if isStaff, staffErr := contextUtils.IsStaff(r.Context()); staffErr != nil || !isStaff {
			responseHandlers.RespondWithError(w, errLib.New("Only admins can upload for other users", http.StatusForbidden))
			return
		}
		
		// Parse the target user ID
		if parsedTargetID, parseErr := uuid.Parse(targetUserIDParam); parseErr == nil {
			targetUserID = parsedTargetID
			targetUserErr = nil
		} else {
			responseHandlers.RespondWithError(w, errLib.New("Invalid target_user_id format", http.StatusBadRequest))
			return
		}
	}

	// Get target user information for organized storage
	var userFolder string
	var roleFolder string
	if targetUserErr == nil {
		// Get user information to create readable folder name
		userInfo, infoErr := h.userRepo.GetUserInfo(r.Context(), "", targetUserID)
		if infoErr == nil {
			// Create readable folder name: "FirstName_LastName_UserID"
			userFolder = fmt.Sprintf("%s_%s_%s", 
				strings.ReplaceAll(userInfo.FirstName, " ", "_"),
				strings.ReplaceAll(userInfo.LastName, " ", "_"),
				targetUserID.String()[:8]) // First 8 chars of UUID for uniqueness
			
			// Determine role-based main folder
			switch strings.ToLower(userInfo.Role) {
			case "superadmin", "admin", "instructor", "coach", "barber":
				roleFolder = "staff"
			case "athlete":
				roleFolder = "athletes"
			case "parent":
				roleFolder = "parents"
			case "child":
				roleFolder = "children"
			default:
				roleFolder = "users" // fallback for unknown roles
			}
		} else {
			userFolder = targetUserID.String()
			roleFolder = "users"
		}
	} else {
		userFolder = "anonymous"
		roleFolder = "users"
	}

	// Create full folder path: folder/roleFolder/userFolder (e.g., profiles/staff/Anthony_Barber_a1b2c3d4)
	fullFolderPath := fmt.Sprintf("%s/%s", folder, roleFolder)
	
	// Delete any existing profile images for this user to prevent accumulation
	if deleteErr := gcp.DeleteOldProfileImages(userFolder, fullFolderPath); deleteErr != nil {
		// Log the error but don't fail the upload - this is cleanup, not critical
		fmt.Printf("Warning: Failed to delete old profile images: %v\n", deleteErr)
	}

	// Generate filename with timestamp for cache busting
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	timestamp := time.Now().Unix()
	fileName := fmt.Sprintf("%s/%s/%s/profile_%d%s", folder, roleFolder, userFolder, timestamp, fileExt)

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