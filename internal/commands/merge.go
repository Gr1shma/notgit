package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge <branch-name>",
	Short: "Merge branches",
	Long: `Merge the specified branch into the current branch.

Currently only supports fast-forward merges.
A fast-forward merge occurs when the current branch is an ancestor of the target branch.

WARNING: Uncommitted changes may be lost during merge.
Make sure to commit your changes before merging.`,
	Args: cobra.ExactArgs(1),
	RunE: mergeCallback,
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}

func mergeCallback(cmd *cobra.Command, args []string) error {
	repo, err := repository.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	targetBranch := args[0]

	if strings.Contains(targetBranch, "/") {
		return fmt.Errorf("nested branch names are not supported")
	}

	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch == targetBranch {
		return fmt.Errorf("cannot merge branch '%s' into itself", targetBranch)
	}

	targetBranchPath := filepath.Join(repo.NotgitDir, "refs", "heads", targetBranch)
	if _, err := os.Stat(targetBranchPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("branch '%s' does not exist", targetBranch)
		}
		return fmt.Errorf("failed to check branch existence: %w", err)
	}

	currentCommitHash, err := repo.GetHEADCommitHash()
	if err != nil {
		return fmt.Errorf("failed to get current commit: %w", err)
	}

	targetCommitHashBytes, err := os.ReadFile(targetBranchPath)
	if err != nil {
		return fmt.Errorf("failed to read target branch: %w", err)
	}
	targetCommitHash := strings.TrimSpace(string(targetCommitHashBytes))

	if targetCommitHash == "" {
		return fmt.Errorf("branch '%s' has no commits", targetBranch)
	}

	if currentCommitHash == targetCommitHash {
		fmt.Printf("Already up to date.\n")
		return nil
	}

	canFastForward, err := isAncestor(repo, currentCommitHash, targetCommitHash)
	if err != nil {
		return fmt.Errorf("failed to check ancestry: %w", err)
	}

	if !canFastForward {
		return fmt.Errorf("cannot fast-forward merge: branches have diverged. Non-fast-forward merges are not yet supported")
	}

	if err := performFastForwardMerge(repo, currentBranch, targetBranch, targetCommitHash); err != nil {
		return fmt.Errorf("failed to perform merge: %w", err)
	}

	fmt.Printf("Fast-forward merge from '%s' to '%s'\n", currentBranch, targetBranch)
	return nil
}

func isAncestor(repo *repository.Repository, ancestorHash, descendantHash string) (bool, error) {
	currentHash := descendantHash

	for currentHash != "" {
		if currentHash == ancestorHash {
			return true, nil
		}

		commit, err := repo.RetrieveCommit(currentHash)
		if err != nil {
			return false, fmt.Errorf("failed to load commit %s: %w", currentHash, err)
		}

		currentHash = commit.ParentHashes[0]
	}

	return false, nil
}

func performFastForwardMerge(repo *repository.Repository, currentBranch, targetBranch, targetCommitHash string) error {
	targetCommit, err := repo.RetrieveCommit(targetCommitHash)
	if err != nil {
		return fmt.Errorf("failed to load target commit: %w", err)
	}

	targetTree, err := repo.RetrieveTree(targetCommit.TreeHash)
	if err != nil {
		return fmt.Errorf("failed to load target tree: %w", err)
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

	currentBranchPath := filepath.Join(repo.NotgitDir, "refs", "heads", currentBranch)
	if err := os.WriteFile(currentBranchPath, []byte(targetCommitHash+"\n"), 0o644); err != nil {
		return fmt.Errorf("failed to update branch ref: %w", err)
	}

	return nil
}
