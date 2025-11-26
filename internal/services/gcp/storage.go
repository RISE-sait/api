package gcp

import (
	"api/config"
	"api/internal/libs/errors"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const bucketName = "rise-sports"

// getGCPClient creates and returns a Google Cloud Storage client using credentials from either
// the environment variable or a service account file. If neither is found, an error is returned.
//
// Returns:
//   - *storage.Client: A pointer to the Google Cloud Storage client.
//   - *errLib.CommonError: An error if GCP credentials are not found or if the client creation fails.
//
// Example usage:
//
//	client, err := getGCPClient()  // Creates a GCP storage client using available credentials.
func getGCPClient() (*storage.Client, *errLib.CommonError) {
	var opt option.ClientOption
	if gcpServiceAccountCredentials := config.Env.GcpServiceAccountCredentialsJSON; gcpServiceAccountCredentials != "" {
		opt = option.WithCredentialsJSON([]byte(gcpServiceAccountCredentials))
	} else if _, err := os.Stat("/app/config/gcp-service-account.json"); err == nil {
		opt = option.WithCredentialsFile("/app/config/gcp-service-account.json")
	} else {
		log.Printf("GCP credentials not found")
		return nil, errLib.New("Internal server error: GCP credentials not found", http.StatusInternalServerError)
	}

	client, err := storage.NewClient(context.Background(), opt)
	if err != nil {
		log.Printf("Failed to create GCP storage client: %v", err)
		return nil, errLib.New("Internal server error: Failed to create GCP storage client", http.StatusInternalServerError)
	}

	return client, nil
}

// GetFilesInBucket retrieves a list of file URLs from a specified folder in a Google Cloud Storage bucket.
// It connects to Google Cloud Storage, queries the specified folder, and returns the list of file URLs.
//
// Parameters:
//   - folderName: The name of the folder in the Google Cloud Storage bucket to query.
//
// Returns:
//   - []string: A list of file URLs in the specified folder.
//   - *errLib.CommonError: An error if the client cannot be created or if any issues occur during the file retrieval.
//
// Example usage:
//
//	files, err := GetFilesInBucket("folderName")  // Retrieves file URLs from the specified folder in the bucket.
func GetFilesInBucket(folderName string) ([]string, *errLib.CommonError) {

	client, err := getGCPClient()
	if err != nil {
		return nil, err
	}

	// Get a handle for the bucket
	bucket := client.Bucket("rise-sports")

	// List the objects in the bucket
	var fileURLs []string

	it := bucket.Objects(context.Background(), &storage.Query{
		Prefix: folderName + "/",
	})

	for {
		objAttrs, retrievalErr := it.Next()
		if retrievalErr != nil {
			break
		}

		fileURLs = append(fileURLs, GeneratePublicFileURL(objAttrs.Name))
	}

	return fileURLs, nil
}

func UploadImageToGCP(image io.Reader, fileName string) (string, *errLib.CommonError) {

	client, gcpClientErr := getGCPClient()
	if gcpClientErr != nil {
		return "", gcpClientErr
	}

	// Get a handle for the bucket
	bucket := client.Bucket(bucketName) // Replace with your bucket name

	// Read the image data into a byte slice
	imageData, ioErr := io.ReadAll(image)
	if ioErr != nil {
		return "", errLib.New(fmt.Sprintf("Failed to read image data: %v", ioErr), http.StatusBadRequest)
	}

	// Upload the image to GCP Storage
	object := bucket.Object(fileName)
	writer := object.NewWriter(context.Background())
	writer.ContentType = http.DetectContentType(imageData) // Automatically detect MIME type

	// Write the image data to the object
	if _, writerErr := writer.Write(imageData); writerErr != nil {
		log.Println(fmt.Sprintf("failed to upload image to GCP Storage: %v", writerErr))
		return "", errLib.New("failed to upload image to GCP Storage", http.StatusInternalServerError)
	}

	// Close the writer to finalize the upload
	if writerErr := writer.Close(); writerErr != nil {
		log.Println(fmt.Sprintf("failed to close the writer: %v", writerErr))
		return "", errLib.New("failed to close the writer", http.StatusInternalServerError)
	}

	// Return the public URL for the uploaded file
	return GeneratePublicFileURL(fileName), nil
}

func GeneratePublicFileURL(fileName string) string {
	parts := strings.Split(fileName, "/")
	for i, part := range parts {
		parts[i] = url.QueryEscape(part)
	}
	encodedFileName := strings.Join(parts, "/")                       // Preserve `/` separators
	encodedFileName = strings.ReplaceAll(encodedFileName, "+", "%20") // Fix space encoding
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, encodedFileName)
}

// DeleteFileFromGCP deletes a single file from GCP storage by its public URL
func DeleteFileFromGCP(publicURL string) *errLib.CommonError {
	client, gcpClientErr := getGCPClient()
	if gcpClientErr != nil {
		return gcpClientErr
	}

	// Extract the file path from the public URL
	// URL format: https://storage.googleapis.com/rise-sports/path/to/file.ext
	prefix := fmt.Sprintf("https://storage.googleapis.com/%s/", bucketName)
	if !strings.HasPrefix(publicURL, prefix) {
		return errLib.New("Invalid GCP URL format", http.StatusBadRequest)
	}

	// Get the object path and decode URL encoding
	objectPath := strings.TrimPrefix(publicURL, prefix)
	decodedPath, decodeErr := url.QueryUnescape(objectPath)
	if decodeErr != nil {
		return errLib.New("Failed to decode file path", http.StatusBadRequest)
	}

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(decodedPath)

	if deleteErr := obj.Delete(context.Background()); deleteErr != nil {
		log.Printf("Failed to delete file from GCP: %s, error: %v", decodedPath, deleteErr)
		return errLib.New("Failed to delete file from storage", http.StatusInternalServerError)
	}

	return nil
}

// DeleteOldProfileImages deletes all existing profile images in a user's folder
// to prevent accumulation of old profile photos
func DeleteOldProfileImages(userFolder string, folder string) *errLib.CommonError {
	client, gcpClientErr := getGCPClient()
	if gcpClientErr != nil {
		return gcpClientErr
	}
	
	bucket := client.Bucket(bucketName)
	
	// List all files in the user's profile folder
	it := bucket.Objects(context.Background(), &storage.Query{
		Prefix: fmt.Sprintf("%s/%s/profile", folder, userFolder),
	})
	
	// Delete each existing profile image
	for {
		objAttrs, err := it.Next()
		if err != nil {
			break // No more objects or error occurred
		}
		
		// Delete the object
		obj := bucket.Object(objAttrs.Name)
		if deleteErr := obj.Delete(context.Background()); deleteErr != nil {
			log.Printf("Failed to delete old profile image %s: %v", objAttrs.Name, deleteErr)
			// Continue deleting other files even if one fails
		}
	}
	
	return nil
}
