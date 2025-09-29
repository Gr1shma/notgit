package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gr1shma/notgit/internal/objects/blob"
	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/spf13/cobra"
)

var addVerboseBool bool

var addCmd = &cobra.Command{
	Use:   "add <pathspec>...",
	Short: "Add file contents to the index",
	Long:  `Add file contents to the index by storing them as blob objects in .notgit/objects.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  addCallback,
}

func init() {
	addCmd.PersistentFlags().BoolVarP(&addVerboseBool, "verbose", "v", false, "Be verbose and show files as they are added")
	rootCmd.AddCommand(addCmd)
}

func addCallback(cmd *cobra.Command, args []string) error {
	repo, err := repository.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	index, err := repo.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	repoRoot := filepath.Dir(repo.NotgitDir)

	for _, pathSpec := range args {
		if _, err := os.Stat(pathSpec); os.IsNotExist(err) {
			return fmt.Errorf("pathspec '%s' did not match any files", pathSpec)
		}

		err := filepath.Walk(pathSpec, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("could not get absolute path for %s: %w", path, err)
			}
			absNotgitDir, err := filepath.Abs(repo.NotgitDir)
			if err != nil {
				return fmt.Errorf("could not get absolute path for repository dir: %w", err)
			}

			if strings.HasPrefix(absPath, absNotgitDir) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if info.IsDir() {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			b, err := blob.NewBlob(data)
			if err != nil {
				return fmt.Errorf("failed to create blob for %s: %w", path, err)
			}

			hash, err := repo.StoreObject(b)
			if err != nil {
				return fmt.Errorf("failed to store object for %s: %w", path, err)
			}

			relativePath, err := filepath.Rel(repoRoot, absPath)
			if err != nil {
				return fmt.Errorf("failed to determine repository-relative path for %s: %w", path, err)
			}
			relativePath = filepath.ToSlash(relativePath)

			index.AddEntry(relativePath, hash)

			if addVerboseBool {
				fmt.Fprintf(cmd.OutOrStdout(), "add '%s'\n", relativePath)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("error processing path '%s': %w", pathSpec, err)
		}
	}

	if err := repo.SaveIndex(index); err != nil {
		return fmt.Errorf("failed to save the index: %w", err)
	}

	return nil
}
