package commands

import (
	"fmt"
	"strings"

	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/spf13/cobra"
)

type LogArgs struct {
	graph bool
}

var logArgs = &LogArgs{}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show commit logs",
	Long:  `Display the commit history starting from the current HEAD reference. (Only for particular branch for now)`,
	Args:  cobra.NoArgs,
	RunE:  logCallback,
}

func init() {
	logCmd.Flags().BoolVarP(&logArgs.graph, "graph", "g", false, "draw a text-based graphical representation of the commit history")
	rootCmd.AddCommand(logCmd)
}

func logCallback(cmd *cobra.Command, args []string) error {
	repo, err := repository.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("error opening repository: %w", err)
	}

	currentRef, err := repo.GetHEADCommitHash()
	if err != nil {
		return fmt.Errorf("error getting HEAD commit: %w", err)
	}

	if currentRef == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "No commits yet\n")
		return nil
	}

	if logArgs.graph {
		if err := displayGraphLog(cmd, repo, currentRef); err != nil {
			return err
		}
	} else {
		if err := displayStandardLog(cmd, repo, currentRef); err != nil {
			return err
		}
	}

	return nil
}

func displayStandardLog(cmd *cobra.Command, repo *repository.Repository, startRef string) error {
	currentRef := startRef

	for currentRef != "" {
		commit, err := repo.RetrieveCommit(currentRef)
		if err != nil {
			return fmt.Errorf("failed to retrieve commit %s: %w", currentRef, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "commit %s\n", currentRef)
		fmt.Fprintf(cmd.OutOrStdout(), "Author: %s <%s>\n", commit.Author.Name, commit.Author.Email)
		fmt.Fprintf(cmd.OutOrStdout(), "Date:   %s\n", commit.Author.Time.Format("Mon Jan 2 15:04:05 2006 -0700"))
		fmt.Fprintf(cmd.OutOrStdout(), "\n    %s\n\n", commit.Message)

		if len(commit.ParentHashes) > 0 {
			currentRef = commit.ParentHashes[0]
		} else {
			currentRef = ""
		}
	}

	return nil
}

func displayGraphLog(cmd *cobra.Command, repo *repository.Repository, startRef string) error {
	currentRef := startRef

	for currentRef != "" {
		commit, err := repo.RetrieveCommit(currentRef)
		if err != nil {
			return fmt.Errorf("failed to retrieve commit %s: %w", currentRef, err)
		}

		shortHash := currentRef
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}

		message := strings.Split(commit.Message, "\n")[0]
		fmt.Fprintf(cmd.OutOrStdout(), "* %s %s\n", shortHash, message)

		if len(commit.ParentHashes) > 0 {
			currentRef = commit.ParentHashes[0]
		} else {
			currentRef = ""
		}
	}

	return nil
}
