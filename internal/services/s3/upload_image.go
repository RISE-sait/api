package s3

import (
	"api/config"
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadImageToS3(image io.Reader, fileName string) (string, error) {
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(config.Envs.AwsConfig.AccessKeyId, config.Envs.AwsConfig.SecretKey, ""),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %v", err)
	}

	// Create an S3 service client
	svc := s3.New(sess)

	// Read the image data into a byte slice
	imageData, err := io.ReadAll(image)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %v", err)
	}

	// Upload the image to S3
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String("risesports"),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String(http.DetectContentType(imageData)), // Automatically detect MIME type
		ACL:         aws.String("public-read"),                     // Make the object publicly accessible
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image to S3: %v", err)
	}

	return "", nil
}
