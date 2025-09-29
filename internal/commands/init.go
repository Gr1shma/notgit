package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/spf13/cobra"
)

type InitArgs struct {
	quietMode bool
	directory string
}

var initArgs = &InitArgs{}

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize a new notgit repository",
	Long:  `Initialize a new notgit repository in the current directory by creating the necessary metadata and configuration files.`,

	Args: cobra.MaximumNArgs(1),
	RunE: initCallback,
}

func init() {
	initCmd.Flags().BoolVarP(&initArgs.quietMode, "quiet", "q", false, "supress output messages")
	rootCmd.AddCommand(initCmd)
}

func initCallback(cmd *cobra.Command, args []string) error {
	initArgs.directory = "."
	if len(args) == 1 {
		initArgs.directory = args[0]
	}

	absPath, err := filepath.Abs(initArgs.directory)
	if err != nil {
		return fmt.Errorf("error while getting the absolute path: %w", err)

	}

	pathStat, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
		if !initArgs.quietMode {
			fmt.Fprintf(cmd.OutOrStdout(), "Created directory: %s\n", absPath)
		}
	} else if err != nil {
		return fmt.Errorf("error accessing path: %w", err)
	} else if !pathStat.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s\n", absPath)
	}

	if err := repository.CreateRepo(absPath); err != nil {
		return fmt.Errorf("error creating the repo: %w", err)
	}

	if !initArgs.quietMode {
		fmt.Fprintf(cmd.OutOrStdout(), "Initialized empty notgit repository in %s/.notgit\n", absPath)
	}
	return nil
}
