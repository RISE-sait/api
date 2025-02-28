package haircut

import (
	"api/internal/services/s3"
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

	fileName := fmt.Sprintf("haircuts/%v", header.Filename)

	// Upload the file to S3
	url, err := s3.UploadImageToS3(
		file,     // File (io.Reader)
		fileName, // Use the original file name
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload file: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the S3 object URL in the response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully. URL: %s", url)
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

	folderPath := "haircuts"

	// If a barberName is provided, append it to the folder path
	if barberName != "" {
		folderPath = fmt.Sprintf("haircuts/%s/", barberName)
	}

	images, err := s3.GetImagesFromFolder(folderPath)
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
