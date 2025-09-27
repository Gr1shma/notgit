package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Gr1shma/notgit/internal/utils"
)

type Repository struct {
	BaseDir   string
	NotgitDir string
}

func CreateRepo(basePath string) error {
	repoPath := filepath.Join(basePath, ".notgit")
	dirs := []string{
		filepath.Join(repoPath, "refs", "heads"),
		filepath.Join(repoPath, "objects"),
	}

	for _, dir := range dirs {
		// 0o755 -> (drwxr-xr-x)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("error while creating the directory named %s: %w", dir, err)
		}
	}

	cfg, _, _ := utils.LoadConfig(true)
	defaultBranch := "master"
	if cfg != nil {
		if branchName, err := utils.GetConfigKeyValue(cfg, "init.defaultBranch"); err == nil {
			defaultBranch = branchName
		}
	}

	headPath := filepath.Join(repoPath, "HEAD")
	headFileContent := []byte("ref: refs/heads/" + defaultBranch + "\n")

	if err := os.WriteFile(headPath, headFileContent, 0o644); err != nil {
		return fmt.Errorf("error while writing content in HEAD: %w", err)
	}

	configPath := filepath.Join(repoPath, "config")
	if err := os.WriteFile(configPath, []byte{}, 0o644); err != nil {
		return fmt.Errorf("error while creating config file: %w", err)
	}

	return nil
}

func OpenRepository(startPath string) (*Repository, error) {
	repoRoot, err := utils.FindRepoRoot(startPath)
	if err != nil {
		return nil, fmt.Errorf("not a notgit repository: %w", err)
	}

	return &Repository{
		BaseDir:   repoRoot,
		NotgitDir: filepath.Join(repoRoot, ".notgit"),
	}, nil
}
