package haircut

import (
	responseHandlers "api/internal/libs/responses"
	"api/internal/services/gcp"
	"encoding/json"
	"fmt"
	"net/http"
)

// UploadHaircutImage handles the upload of a haircut image to S3.
// @Summary Upload a haircut image
// @Description Uploads a haircut image to S3 and returns the object URL.
// @Tags haircut
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Haircut image to upload"
// @Success 200 {object} map[string]string "File uploaded successfully"
// @Failure 400 {object} map[string]string "Bad Request: Invalid input"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /haircuts [post]
func UploadHaircutImage(w http.ResponseWriter, r *http.Request) {

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

	fileName := fmt.Sprintf("haircut/%v", header.Filename)

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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode success message", http.StatusInternalServerError)
	}
}

// GetHaircutImages retrieves haircut images from the specified folder in S3.
// @Summary Retrieve haircut images
// @Description Retrieves all haircut images from a folder in S3. Optionally, specify a barber name to get images from that barber's folder.
// @Tags haircut
// @Accept json
// @Produce json
// @Param barber query string false "Barber name to filter images"
// @Success 200 {object} []string "List of image URLs"
// @Failure 400 {object} map[string]string "Bad Request: Invalid input"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /haircuts [get]
func GetHaircutImages(w http.ResponseWriter, r *http.Request) {

	barberName := r.URL.Query().Get("barber")

	folderPath := "haircut"

	// If a barberName is provided, append it to the folder path
	if barberName != "" {
		folderPath = fmt.Sprintf("haircut/%s", barberName)
	}

	images, err := gcp.GetFilesInBucket("rise-sports", folderPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching images: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the list of image URLs
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Respond with the images as a JSON array (You could format it however you want)
	if err := json.NewEncoder(w).Encode(images); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}
