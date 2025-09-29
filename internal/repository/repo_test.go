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

func TestRepositoryBranchAndHead(t *testing.T) {
	tmp := t.TempDir()
	repo := &repository.Repository{NotgitDir: tmp}

	// HEAD pointing to a branch
	headPath := filepath.Join(tmp, "HEAD")
	err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0o644)
	require.NoError(t, err)

	branchName, err := repo.GetCurrentBranch()
	require.NoError(t, err)
	require.Equal(t, "main", branchName)

	// Create branch ref with fake commit hash
	fakeHash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	refPath := filepath.Join(tmp, "refs", "heads", branchName)
	require.NoError(t, os.MkdirAll(filepath.Dir(refPath), 0o755))
	err = os.WriteFile(refPath, []byte(fakeHash+"\n"), 0o644)
	require.NoError(t, err)

	hash, err := repo.GetHEADCommitHash()
	require.NoError(t, err)
	require.Equal(t, fakeHash, hash)

	// HEAD detached with direct commit hash
	detachedHash := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	err = os.WriteFile(headPath, []byte(detachedHash+"\n"), 0o644)
	require.NoError(t, err)

	branchName, err = repo.GetCurrentBranch()
	require.NoError(t, err)
	require.Equal(t, "", branchName) // detached => no branch

	hash, err = repo.GetHEADCommitHash()
	require.NoError(t, err)
	require.Equal(t, detachedHash, hash)
}

func TestUpdateHEAD(t *testing.T) {
	tmp := t.TempDir()

	// Initialize a fake repo
	repo := &repository.Repository{NotgitDir: tmp}

	// Create HEAD pointing to branch "main"
	headPath := filepath.Join(tmp, "HEAD")
	branchName := "main"
	refPath := filepath.Join("refs", "heads", branchName)
	err := os.MkdirAll(filepath.Join(tmp, "refs", "heads"), 0o755)
	require.NoError(t, err)

	err = os.WriteFile(headPath, []byte("ref: "+refPath+"\n"), 0o644)
	require.NoError(t, err)

	// Write commit SHA to branch
	commitSHA := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	err = repo.UpdateHEAD(commitSHA)
	require.NoError(t, err)

	// Check that the branch ref file contains the commit SHA
	branchRefFullPath := filepath.Join(tmp, refPath)
	data, err := os.ReadFile(branchRefFullPath)
	require.NoError(t, err)
	require.Equal(t, commitSHA+"\n", string(data))

	// Now test detached HEAD
	detachedSHA := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	err = os.WriteFile(headPath, []byte("detached-content\n"), 0o644)
	require.NoError(t, err)

	err = repo.UpdateHEAD(detachedSHA)
	require.NoError(t, err)

	// HEAD file itself should now contain the detached SHA
	data, err = os.ReadFile(headPath)
	require.NoError(t, err)
	require.Equal(t, detachedSHA+"\n", string(data))
}
