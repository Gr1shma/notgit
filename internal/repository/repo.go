package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Gr1shma/notgit/internal/utils"
)

func CreateRepo(basePath string) error {
	repoPath := filepath.Join(basePath, ".notgit")
	dirs := []string{
		filepath.Join(repoPath, "refs", "heads"),
		filepath.Join(repoPath, "objects"),
	}

	for _, dir := range dirs {
		// 0o755 -> (drwxr-xr-x)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("Error while creating the directory named %s: %w", dir, err)
		}
	}

	headPath := filepath.Join(repoPath, "HEAD")
	headFileContent := []byte("ref: refs/heads/master\n")

	if err := utils.WriteFile(headPath, headFileContent, nil); err != nil {
		return fmt.Errorf("Error while writing content in HEAD: %w", err)
	}

	configPath := filepath.Join(repoPath, "config")
	if err := utils.WriteFile(configPath, []byte{}, nil); err != nil {
		return fmt.Errorf("Error while creating config file: %w", err)
	}

	return nil
}

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
