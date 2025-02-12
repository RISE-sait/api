package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Define the base directories to scan
	baseDirs := []string{"./cmd/server", "./internal/domains"}

	skipDirs := map[string]bool{
		"persistence": true,
		"values":      true,
		"services":    true,
		"tests":       true,
		"test_utils":  true,
	}

	// Collect all subdirectories containing .go files, excluding "persistence" directories
	var dirs []string
	for _, baseDir := range baseDirs {
		filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// Skip "persistence" directories
				if skipDirs[info.Name()] {
					return filepath.SkipDir
				}

				if info.Name() == "entities" {
					parentDir := filepath.Base(filepath.Dir(path))
					if parentDir != "identity" {
						return filepath.SkipDir
					}
				}

				// Check if the directory contains .go files
				goFiles, _ := filepath.Glob(filepath.Join(path, "*.go"))
				if len(goFiles) > 0 {
					dirs = append(dirs, path)
				}
			}
			return nil
		})
	}

	// Join the directories with commas
	dirArg := strings.Join(dirs, ",")

	// Run swag init
	cmd := exec.Command("swag", "init", "--dir", dirArg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Error running swag init:", err)
	}
}
