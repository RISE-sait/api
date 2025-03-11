package gcp

import (
	"testing"
)

func TestGeneratePublicFileURL(t *testing.T) {

	tests := []struct {
		fileName string
		expected string
	}{
		{"image.png", "https://storage.googleapis.com/rise-sports/image.png"},
		{"folder/image.png", "https://storage.googleapis.com/rise-sports/folder/image.png"},
		{"folder with space/image.png", "https://storage.googleapis.com/rise-sports/folder%20with%20space/image.png"},
		{"folder/file with @ symbol.png", "https://storage.googleapis.com/rise-sports/folder/file%20with%20%40%20symbol.png"},
		{"deeply/nested/structure/image.png", "https://storage.googleapis.com/rise-sports/deeply/nested/structure/image.png"},
	}

	for _, tt := range tests {
		t.Run(tt.fileName, func(t *testing.T) {
			result := GeneratePublicFileURL("rise-sports", tt.fileName)

			if result != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}
