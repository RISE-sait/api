package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
