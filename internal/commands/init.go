package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize a new notgit repository",
	Long:  `Initialize a new notgit repository in the current directory by creating the necessary metadata and configuration files.`,

	Run: initCallback,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func initCallback(cmd *cobra.Command, args []string) {
	path := "."
	if len(args) == 1 {
		path = args[0]
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("Error while getting the absolute path: %v\n", err)
		return
	}

	pathStat, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(absPath, 0755); err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			return
		}
		fmt.Printf("Created directory: %s\n", absPath)
	} else if err != nil {
		fmt.Printf("Error accessing path: %v\n", err)
		return
	} else if !pathStat.IsDir() {
		fmt.Printf("Path exists but is not a directory: %s\n", absPath)
		return
	}

	if err := repository.CreateRepo(absPath); err != nil {
		fmt.Printf("Error while creating the repo: %v\n", err)
		return
	}

	fmt.Printf("Initialized empty notgit repository in %s/.notgit\n", absPath)
}
