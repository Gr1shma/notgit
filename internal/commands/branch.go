package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/spf13/cobra"
)

var (
	deleteBranch string
	renameOld    string
	renameNew    string
)

var branchCmd = &cobra.Command{
	Use:   "branch [branchName]",
	Short: "List, create, delete, or rename branches",
	Long: `Manage branches in notgit.

Without arguments, this shows the existing branches.
With a branch name, it creates a new branch pointing to the current commit.
With -d, deletes the specified branch.
With -m, renames the given branch.`,
	Args: cobra.MaximumNArgs(2),
	RunE: branchCallback,
}

func init() {
	branchCmd.Flags().StringVarP(&deleteBranch, "delete", "d", "", "Delete the specified branch")
	branchCmd.Flags().StringVarP(&renameOld, "move", "m", "", "Old branch name (use with --new-name)")
	branchCmd.Flags().StringVar(&renameNew, "new-name", "", "New branch name for rename")
	rootCmd.AddCommand(branchCmd)
}

func branchCallback(cmd *cobra.Command, args []string) error {
	repo, err := repository.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	switch {
	case cmd.Flags().Changed("move"):
		if renameNew == "" {
			return fmt.Errorf("--new-name is required when using -m/--move")
		}
		if err := renameBranch(repo, renameOld, renameNew); err != nil {
			return fmt.Errorf("failed to rename branch: %w", err)
		}
	case deleteBranch != "":
		if err := deleteBranchByName(repo, deleteBranch); err != nil {
			return fmt.Errorf("failed to delete branch: %w", err)
		}
	case len(args) == 1:
		if err := createBranch(repo, args[0]); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
	default:
		if err := listBranches(repo); err != nil {
			return fmt.Errorf("failed to list branches: %w", err)
		}
	}

	return nil
}

func listBranches(repo *repository.Repository) error {
	branchesDir := filepath.Join(repo.NotgitDir, "refs", "heads")
	entries, err := os.ReadDir(branchesDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No branches found")
			return nil
		}
		return fmt.Errorf("error while reading the branches directory %w", err)
	}

	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("error while getting the current branch %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No branches found")
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		branchName := entry.Name()
		if branchName == currentBranch {
			fmt.Printf("* %s\n", branchName)
		} else {
			fmt.Printf("  %s\n", branchName)
		}
	}
	return nil
}

func createBranch(repo *repository.Repository, branchName string) error {
	if strings.Contains(branchName, "/") {
		return fmt.Errorf("nested branch names are not supported")
	}

	branchPath := filepath.Join(repo.NotgitDir, "refs", "heads", branchName)
	if _, err := os.Stat(branchPath); err == nil {
		return fmt.Errorf("branch '%s' already exists", branchName)
	}

	headCommitHash, err := repo.GetHEADCommitHash()
	if err != nil {
		return fmt.Errorf("failed to get HEAD commit hash: %w", err)
	}

	if headCommitHash == "" {
		return fmt.Errorf("cannot create branch: no commits exist yet")
	}

	refsHeadsDir := filepath.Join(repo.NotgitDir, "refs", "heads")
	if err := os.MkdirAll(refsHeadsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create refs/heads directory: %w", err)
	}

	if err := os.WriteFile(branchPath, []byte(headCommitHash+"\n"), 0o644); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	fmt.Printf("Created branch '%s'\n", branchName)
	return nil
}

func deleteBranchByName(repo *repository.Repository, name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch == name {
		return fmt.Errorf("cannot delete the current branch '%s'", name)
	}

	branchRefPath := filepath.Join(repo.NotgitDir, "refs", "heads", name)

	if _, err := os.Stat(branchRefPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("branch '%s' not found", name)
		}
		return fmt.Errorf("failed to check branch existence: %w", err)
	}

	if err := os.Remove(branchRefPath); err != nil {
		return fmt.Errorf("failed to delete branch '%s': %w", name, err)
	}

	fmt.Printf("Deleted branch '%s'\n", name)
	return nil
}

func renameBranch(repo *repository.Repository, oldName, newName string) error {
	if oldName == "" {
		return fmt.Errorf("old branch name cannot be empty")
	}
	if newName == "" {
		return fmt.Errorf("new branch name cannot be empty")
	}
	if oldName == newName {
		return fmt.Errorf("old and new branch name are same")
	}

	if strings.Contains(oldName, "/") || strings.Contains(newName, "/") {
		return fmt.Errorf("nested branch names are not supported")
	}

	oldBranchPath := filepath.Join(repo.NotgitDir, "refs", "heads", oldName)
	newBranchPath := filepath.Join(repo.NotgitDir, "refs", "heads", newName)

	if _, err := os.Stat(oldBranchPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("branch '%s' not found", oldName)
		}
		return fmt.Errorf("failed to check old branch existence: %w", err)
	}

	if _, err := os.Stat(newBranchPath); err == nil {
		return fmt.Errorf("branch '%s' already exists", newName)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check new branch existence: %w", err)
	}

	if err := os.Rename(oldBranchPath, newBranchPath); err != nil {
		return fmt.Errorf("failed to rename branch from '%s' to '%s': %w", oldName, newName, err)
	}

	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		// If we can't get current branch, assume it's not the one we renamed
		// The rename was successful, so don't fail here
		fmt.Printf("Renamed branch '%s' to '%s'\n", oldName, newName)
		return nil
	}

	if currentBranch == oldName {
		headPath := filepath.Join(repo.NotgitDir, "HEAD")
		newHeadContent := fmt.Sprintf("ref: refs/heads/%s\n", newName)
		if err := os.WriteFile(headPath, []byte(newHeadContent), 0o644); err != nil {
			// Branch was renamed successfully but HEAD update failed
			// Try to rollback the rename
			os.Rename(newBranchPath, oldBranchPath)
			return fmt.Errorf("failed to update HEAD after renaming current branch: %w", err)
		}
	}

	fmt.Printf("Renamed branch '%s' to '%s'\n", oldName, newName)
	return nil
}
