package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/*
Swagger Documentation Generator

This tool automates the generation of Swagger/OpenAPI documentation for the Rise API.
It scans specified directories for Go files containing Swagger annotations and
generates corresponding documentation files.

Features:
  - Recursive directory scanning
  - Configurable directory exclusions
  - Multiple format output (JSON, Go)
  - Test file exclusion
  - Smart directory filtering based on project structure

Usage:
  Run this tool from the project root directory:
  > go run cmd/swag_init/swag-init.go

Configuration:
  Base Directories:
    - ./cmd/server/server
    - ./internal/domains

  Excluded Directories:
    - persistence
    - values
    - tests
    - test_utils
    - service
    - router (only in cmd/server)

Output:
  Generates Swagger documentation in both JSON and Go formats
*/

// ngl idek this code well, I just LLMed it and it works, so u know
func main() {
	// Define the base directories to scan
	baseDirs := []string{"./cmd/server/server", "./internal/domains"}

	skipDirs := map[string]bool{
		"persistence": true,
		"values":      true,
		"tests":       true,
		"test_utils":  true,
		"service":     true,
	}

	// Collect all subdirectories containing .go files, excluding "persistence" directories
	var dirs []string
	for _, baseDir := range baseDirs {
		filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {

				if baseDir == "./cmd/server" && info.Name() == "router" {
					return filepath.SkipDir // Skip the "router" directory
				}

				if baseDir == "./internal/domains" {

					if skipDirs[info.Name()] {
						return filepath.SkipDir
					}

					// "entity" is no longer used in the project structure
					// but just keep it in case shit breaks
					if info.Name() == "entity" {
						parentDir := filepath.Base(filepath.Dir(path))
						if parentDir != "identity" {
							return filepath.SkipDir
						}
					}
				}

				// Check if the directory contains .go files
				goFiles, err := filepath.Glob(filepath.Join(path, "*.go"))
				if err != nil {
					return err
				}

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
	cmd := exec.Command("swag", "init", "--dir", dirArg, "--ot", "json,go", "--exclude", "mocks,*_test.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("Error running swag init: " + err.Error())
	}
}
