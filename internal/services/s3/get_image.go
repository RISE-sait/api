package s3

import (
	"api/config"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"strings"
)

func GetImagesFromFolder(folderName string) ([]string, error) {
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(config.Envs.AwsConfig.AccessKeyId, config.Envs.AwsConfig.SecretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	// Create an S3 service client
	svc := s3.New(sess)

	// Create the input for ListObjectsV2
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("risesports"),
		Prefix: aws.String(folderName), // folder name to list objects inside that folder
	}

	// List the objects in the specified folder
	result, err := svc.ListObjectsV2(input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in folder: %v", err)
	}

	var imageURLs []string
	for _, object := range result.Contents {
		if !strings.HasSuffix(*object.Key, "/") {

			fullURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", "risesports", "us-east-2", *object.Key)

			imageURLs = append(imageURLs, fullURL) // Add object key to result list
		}
	}

	return imageURLs, nil
}
