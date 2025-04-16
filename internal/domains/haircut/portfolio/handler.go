package haircut

import (
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"api/internal/services/gcp"
	contextUtils "api/utils/context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

// UploadHaircutImage handles the upload of a haircut image to Google Cloud Storage.
// @Description Uploads a haircut image to Google Cloud Storage and returns the object URL.
// @Tags haircuts
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param file formData file true "Haircut image to upload"
// @Success 200 {object} map[string]string "File uploaded successfully"
// @Failure 400 {object} map[string]string "Bad Request: Invalid input"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /haircuts [post]
func UploadHaircutImage(w http.ResponseWriter, r *http.Request) {

	userId, ctxErr := contextUtils.GetUserID(r.Context())
	if ctxErr != nil {
		responseHandlers.RespondWithError(w, ctxErr)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the form data
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileName := fmt.Sprintf("haircut/%v/%v", userId.String(), header.Filename)

	url, uploadErr := gcp.UploadImageToGCP(
		file,     // File (io.Reader)
		fileName, // Use the original file name
	)
	if uploadErr != nil {
		responseHandlers.RespondWithError(w, uploadErr)
		return
	}

	successMessage := "Success. URL generated: " + url

	// Encode the success message as JSON and write it to the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Using json.NewEncoder directly to encode the response to JSON
	response := map[string]string{"message": successMessage}
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode success message", http.StatusInternalServerError)
	}
}

// GetHaircutImages retrieves haircut images from Google Cloud Storage.
// @Description Retrieves all haircut images from a folder in Google Cloud Storage. Optionally, specify a barber name to get images from that barber's folder.
// @Tags haircuts
// @Accept json
// @Produce json
// @Param barber_id query string false "Barber ID to filter images"
// @Success 200 {object} []string "List of image URLs"
// @Failure 400 {object} map[string]string "Bad Request: Invalid input"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /haircuts [get]
func GetHaircutImages(w http.ResponseWriter, r *http.Request) {

	var barberID uuid.UUID

	folderPath := "haircut"

	// If a barber ID is provided, append it to the folder path
	if barberIdStr := r.URL.Query().Get("barber_id"); barberIdStr != "" {
		id, err := validators.ParseUUID(barberIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		barberID = id
	}

	// If a barberName is provided, append it to the folder path
	if barberID != uuid.Nil {
		folderPath = fmt.Sprintf("haircut/%s", barberID.String())
	}

	images, err := gcp.GetFilesInBucket(folderPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching images: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the list of image URLs
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if encodingErr := json.NewEncoder(w).Encode(images); encodingErr != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", encodingErr), http.StatusInternalServerError)
	}
}
