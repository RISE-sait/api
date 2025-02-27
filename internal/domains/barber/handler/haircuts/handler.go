package haircut

import (
	"api/internal/services/s3"
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
// @Router /barbers/{id}/haircuts/upload [post]
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
