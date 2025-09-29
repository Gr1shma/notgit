package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/Gr1shma/notgit/internal/objects/commit"
	"github.com/Gr1shma/notgit/internal/objects/tree"
	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/Gr1shma/notgit/internal/utils"
	"github.com/spf13/cobra"
)

var commitMessageString string

var commitCmd = &cobra.Command{
	Use:   "commit -m <message>",
	Short: "Record changes to the repository",
	Long: `Stores the current contents of the index in a new commit object.
A commit message is required to describe the changes being recorded.`,
	RunE: commitCallback,
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessageString, "message", "m", "", "Commit message (required)")
	commitCmd.MarkFlagRequired("message")
	rootCmd.AddCommand(commitCmd)
}

func commitCallback(cmd *cobra.Command, args []string) error {
	repo, err := repository.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("error while getting notgit repository: %w", err)
	}

	authorName, authorEmail, err := getUserIdentity()
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		fmt.Fprintln(cmd.ErrOrStderr(), "\nPlease configure your user identity using:")
		fmt.Fprintln(cmd.ErrOrStderr(), "  notgit config set --global user.name \"Your Name\"")
		fmt.Fprintln(cmd.ErrOrStderr(), "  notgit config set --global user.email \"you@example.com\"")
		return fmt.Errorf("user identity not configured")
	}

	idx, err := repo.LoadIndex()
	if err != nil {
		return fmt.Errorf("error while getting index from the repository: %w", err)
	}

	if len(idx.Entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Nothing to commit, working tree clean")
		return nil
	}

	commitTree := tree.NewTree()
	for path, entry := range idx.Entries {
		commitTree.AddEntry(path, entry.Hash, tree.EntryTypeBlob)
	}
	if err := commitTree.ComputeHash(); err != nil {
		return fmt.Errorf("fatal: failed to compute tree hash: %w", err)
	}
	treeSHA := commitTree.GetHash()

	_, err = repo.StoreObject(commitTree)
	if err != nil {
		return fmt.Errorf("fatal: failed to write tree object: %w", err)
	}

	parentSHA, err := repo.GetHEADCommitHash()
	if err != nil {
		return fmt.Errorf("failed to read HEAD commit hash: %w", err)
	}

	now := time.Now()
	authorSig := commit.Signature{
		Name:  authorName,
		Email: authorEmail,
		Time:  now,
	}

	parentHashes := []string{}
	if parentSHA != "" {
		parentHashes = append(parentHashes, parentSHA)
	}

	commitObj := commit.NewCommit(treeSHA, commitMessageString, parentHashes, authorSig, authorSig)

	if err := commitObj.ComputeHash(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: failed to compute commit hash: %v\n", err)
		os.Exit(1)
	}
	commitSHA := commitObj.GetHash()

	_, err = repo.StoreObject(commitObj)
	if err != nil {
		return fmt.Errorf("fatal: failed to write commit object: %w\n", err)
	}

	if err := repo.UpdateHEAD(commitSHA); err != nil {
		return fmt.Errorf("failed to update HEAD: %w", err)
	}

	if parentSHA == "" {
		fmt.Printf("[root-commit %s] %s\n", commitSHA[:7], commitMessageString)
	} else {
		branchName, err := repo.GetCurrentBranch()
		if err != nil {
			return fmt.Errorf("failed to get current branch")
		}
		if branchName == "" {
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", commitSHA[:7], commitMessageString) // detached HEAD
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "[%s %s] %s\n", branchName, commitSHA[:7], commitMessageString)
		}
	}
	return nil
}

func getUserIdentity() (name, email string, err error) {
	localCfg, _, localErr := utils.LoadConfig(false)
	if localErr == nil {
		name, _ = utils.GetConfigKeyValue(localCfg, "user.name")
		email, _ = utils.GetConfigKeyValue(localCfg, "user.email")
	}

	if name == "" || email == "" {
		globalCfg, _, globalErr := utils.LoadConfig(true)
		if globalErr == nil {
			if name == "" {
				name, _ = utils.GetConfigKeyValue(globalCfg, "user.name")
			}
			if email == "" {
				email, _ = utils.GetConfigKeyValue(globalCfg, "user.email")
			}
		}
	}

	if name == "" || email == "" {
		return "", "", fmt.Errorf("fatal: user identity not set")
	}
	return name, email, nil
}
