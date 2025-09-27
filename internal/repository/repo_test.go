package repository_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/Gr1shma/notgit/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestCreateRepo(t *testing.T) {
	// Setup: Create a temporary directory
	tempDir, err := os.MkdirTemp("", "notgit-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Act: Call CreateRepo
	err = repository.CreateRepo(tempDir)
	require.NoError(t, err)

	// Assert: .notgit directory exists
	repoPath := filepath.Join(tempDir, ".notgit")
	require.DirExists(t, repoPath)

	// Assert: refs/heads directory exists
	require.DirExists(t, filepath.Join(repoPath, "refs", "heads"))

	// Assert: objects directory exists
	require.DirExists(t, filepath.Join(repoPath, "objects"))

	// Assert: HEAD file exists and has correct content
	headPath := filepath.Join(repoPath, "HEAD")
	require.FileExists(t, headPath)
	headContent, err := os.ReadFile(headPath)
	require.NoError(t, err)

	expectedBranch := "master" // fallback default
	if cfg, _, err := utils.LoadConfig(true); err == nil {
		if branchName, err := utils.GetConfigKeyValue(cfg, "init.defaultBranch"); err == nil {
			expectedBranch = branchName
		}
	}

	expectedHeadContent := "ref: refs/heads/" + expectedBranch + "\n"
	require.Equal(t, expectedHeadContent, string(headContent))

	// Assert: config file exists and is empty
	configPath := filepath.Join(repoPath, "config")
	require.FileExists(t, configPath)
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.Equal(t, 0, len(configContent))
}
