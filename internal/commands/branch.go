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
	case cmd.Flags().Changed("move") && len(args) == 2:
		if err := renameBranch(repo, args[0], args[1]); err != nil {
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
		return fmt.Errorf("error while reading the branches directory %w", err)
	}

	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("error while getting the current branch %w", err)
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
	headCommitHash, err := getHEADCommitHash(repo)
	if err != nil {
		return fmt.Errorf("failed to get HEAD commit hash: %w", err)
	}

	branchPath := filepath.Join(repo.NotgitDir, "refs", "heads", branchName)
	if err := os.WriteFile(branchPath, []byte(headCommitHash+"\n"), 0o644); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	fmt.Printf("Created branch '%s'\n", branchName)
	return nil
}

func getHEADCommitHash(repo *repository.Repository) (string, error) {
	headPath := filepath.Join(repo.NotgitDir, "HEAD")
	headContent, err := os.ReadFile(headPath)
	if err != nil {
		return "", fmt.Errorf("failed to read HEAD file: %w", err)
	}

	ref := strings.TrimSpace(string(headContent))
	if refPath, found := strings.CutPrefix(ref, "ref: "); found {
		fullRefPath := filepath.Join(repo.NotgitDir, refPath)
		commitHash, err := os.ReadFile(fullRefPath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", nil
			}
			return "", fmt.Errorf("failed to read ref file: %w", err)
		}
		return strings.TrimSpace(string(commitHash)), nil

	}

	return ref, nil
}

func deleteBranchByName(repo *repository.Repository, name string) error {
	return nil
}

func renameBranch(repo *repository.Repository, oldName, newName string) error {
	return nil
}
