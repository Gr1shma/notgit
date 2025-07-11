package repository

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateRepo(basePath string) error {
	repoPath := filepath.Join(basePath, ".notgit")
	dirs := []string{
		filepath.Join(repoPath, "reps"),
		filepath.Join(repoPath, "objects"),
	}

	for _, dir := range dirs {
		// 0o755 -> (drwxr-xr-x)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("Error while creating the directory named %s: %v", dir, err)
		}
	}

	headPath := filepath.Join(repoPath, "HEAD")
	headFileContent := []byte("ref: refs/heads/master\n")

	// 0o644 -> (-rw-r--r--)
	if err := os.WriteFile(headPath, headFileContent, 0o644); err != nil {
		return fmt.Errorf("Error while writing content in HEAD: %v", err)
	}

	configPath := filepath.Join(repoPath, "config")
	if err := os.WriteFile(configPath, []byte{}, 0o644); err != nil {
		return fmt.Errorf("Error while creating config file: %v", err)
	}

	return nil
}
