package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

	var dirs []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Function to walk a directory and collect directories containing .go files
	walkDir := func(baseDir string) {
		defer wg.Done()
		filepath.WalkDir(baseDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				// Skip specific directories
				if baseDir == "./cmd/server" && d.Name() == "router" {
					return filepath.SkipDir
				}

				if baseDir == "./internal/domains" {
					if skipDirs[d.Name()] {
						return filepath.SkipDir
					}

					if d.Name() == "entity" {
						parentDir := filepath.Base(filepath.Dir(path))
						if parentDir != "identity" {
							return filepath.SkipDir
						}
					}
				}

				// Check if the directory contains .go files
				matches, _ := filepath.Glob(filepath.Join(path, "*.go"))
				if len(matches) > 0 {
					mu.Lock()
					dirs = append(dirs, path)
					mu.Unlock()
				}
			}
			return nil
		})
	}

	// Start a goroutine for each base directory
	for _, baseDir := range baseDirs {
		wg.Add(1)
		go walkDir(baseDir)
	}

	// Wait for all goroutines to finish
	wg.Wait()

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
