package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gr1shma/notgit/internal/objects/tree"
	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/spf13/cobra"
)

var createAndSwitch bool

var switchCmd = &cobra.Command{
	Use:   "switch <branch-name>",
	Short: "Switch branches",
	Long: `Switch to a specified branch.

With -c flag, creates a new branch if it doesn't exist and switches to it.

WARNING: Uncommitted changes will be lost when switching branches.
Make sure to commit your changes before switching.
Untracked files will be preserved.`,
	Args: cobra.ExactArgs(1),
	RunE: switchCallback,
}

func init() {
	switchCmd.Flags().BoolVarP(&createAndSwitch, "create", "c", false, "Create the branch if it doesn't exist")
	rootCmd.AddCommand(switchCmd)
}

func switchCallback(cmd *cobra.Command, args []string) error {
	repo, err := repository.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	branchName := args[0]

	if strings.Contains(branchName, "/") {
		return fmt.Errorf("nested branch names are not supported")
	}

	branchPath := filepath.Join(repo.NotgitDir, "refs", "heads", branchName)

	_, err = os.Stat(branchPath)
	branchExists := err == nil

	if !branchExists {
		if createAndSwitch {
			if err := createBranchFromHEAD(repo, branchName); err != nil {
				return fmt.Errorf("failed to create branch: %w", err)
			}
		} else {
			return fmt.Errorf("branch '%s' does not exist. Use -c to create it", branchName)
		}
	}

	if err := switchToBranch(repo, branchName); err != nil {
		return fmt.Errorf("failed to switch branch: %w", err)
	}

	return nil
}

func switchToBranch(repo *repository.Repository, branchName string) error {
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch == branchName {
		fmt.Printf("Already on '%s'\n", branchName)
		return nil
	}

	branchPath := filepath.Join(repo.NotgitDir, "refs", "heads", branchName)
	if _, err := os.Stat(branchPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("branch '%s' does not exist", branchName)
		}
		return fmt.Errorf("failed to check branch existence: %w", err)
	}

	if err := updateWorkingDirectory(repo, branchName); err != nil {
		return fmt.Errorf("failed to update working directory: %w", err)
	}

	headPath := filepath.Join(repo.NotgitDir, "HEAD")
	newHeadContent := fmt.Sprintf("ref: refs/heads/%s\n", branchName)
	if err := os.WriteFile(headPath, []byte(newHeadContent), 0o644); err != nil {
		return fmt.Errorf("failed to update HEAD: %w", err)
	}

	fmt.Printf("Switched to branch '%s'\n", branchName)
	return nil
}

func updateWorkingDirectory(repo *repository.Repository, branchName string) error {
	branchPath := filepath.Join(repo.NotgitDir, "refs", "heads", branchName)
	commitHashBytes, err := os.ReadFile(branchPath)
	if err != nil {
		return fmt.Errorf("failed to read branch ref: %w", err)
	}

	commitHash := strings.TrimSpace(string(commitHashBytes))
	if commitHash == "" {
		return fmt.Errorf("branch '%s' has no commits", branchName)
	}

	commit, err := repo.RetrieveCommit(commitHash)
	if err != nil {
		return fmt.Errorf("failed to load commit: %w", err)
	}

	targetTree, err := repo.RetrieveTree(commit.TreeHash)
	if err != nil {
		return fmt.Errorf("failed to load tree: %w", err)
	}

	repoRoot := filepath.Dir(repo.NotgitDir)

	index, err := repo.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	trackedFiles := make(map[string]bool)
	for _, entry := range index.Entries {
		trackedFiles[entry.Path] = true
	}

	if err := removeTrackedFiles(repoRoot, repo.NotgitDir, trackedFiles); err != nil {
		return fmt.Errorf("failed to remove tracked files: %w", err)
	}

	if err := restoreTreeToWorkingDirectory(repo, targetTree, repoRoot); err != nil {
		return fmt.Errorf("failed to restore tree: %w", err)
	}

	return nil
}

func removeTrackedFiles(repoRoot, notgitDir string, trackedFiles map[string]bool) error {
	absNotgitDir, err := filepath.Abs(notgitDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for .notgit: %w", err)
	}

	for trackedPath := range trackedFiles {
		fullPath := filepath.Join(repoRoot, trackedPath)
		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			continue
		}

		if strings.HasPrefix(absPath, absNotgitDir+string(filepath.Separator)) {
			continue
		}

		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", fullPath, err)
		}
	}

	err = filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil
		}

		if absPath == absNotgitDir || strings.HasPrefix(absPath, absNotgitDir+string(filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if absPath == repoRoot {
			return nil
		}

		if info.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil
			}
			if len(entries) == 0 {
				os.Remove(path)
			}
		}

		return nil
	})

	return err
}

func restoreTreeToWorkingDirectory(repo *repository.Repository, tr *tree.Tree, basePath string) error {
	for _, entry := range tr.Entries {
		entryPath := filepath.Join(basePath, entry.Name)

		if entry.Type == tree.EntryTypeTree {
			subtree, err := repo.RetrieveTree(entry.Hash)
			if err != nil {
				return fmt.Errorf("failed to load subtree %s: %w", entry.Hash, err)
			}

			if err := os.MkdirAll(entryPath, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", entryPath, err)
			}

			if err := restoreTreeToWorkingDirectory(repo, subtree, entryPath); err != nil {
				return err
			}
		} else {
			blob, err := repo.RetrieveBlob(entry.Hash)
			if err != nil {
				return fmt.Errorf("failed to load blob %s: %w", entry.Hash, err)
			}

			parentDir := filepath.Dir(entryPath)
			if err := os.MkdirAll(parentDir, 0o755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", entryPath, err)
			}

			if err := os.WriteFile(entryPath, blob.Content, 0o644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", entryPath, err)
			}
		}
	}

	return nil
}
