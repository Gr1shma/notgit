package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindRepoRoot(startPath string) (string, error) {
	currentPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	for {
		repoDir := filepath.Join(currentPath, ".notgit")
		if info, err := os.Stat(repoDir); err == nil && info.IsDir() {
			return currentPath, nil
		}

		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached filesystem root without finding repo directory
			return "", os.ErrNotExist
		}

		currentPath = parentPath
	}
}
