package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Gr1shma/notgit/internal/objects/blob"
	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Add file contents to the index",
	Long:  `Add file contents to the index by storing them as blob objects in .notgit/objects.`,

	Run: addCallback,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func addCallback(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Please specify a file or directory to add.")
		return
	}

	repo, err := repository.OpenRepository(".")
	if err != nil {
		fmt.Println("Error: not inside a notgit repository.")
		return
	}

	index, err := repo.LoadIndex()
	if err != nil {
		fmt.Printf("Error loading index: %v\n", err)
		return
	}

	rootPath := args[0]

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		b, err := blob.NewBlob(data)
		if err != nil {
			return fmt.Errorf("failed to create blob for %s: %w", path, err)
		}

		hash, err := repo.StoreObject(b)
		if err != nil {
			return fmt.Errorf("failed to store object for %s: %w", path, err)
		}

		rel, _ := filepath.Rel(".", path)
		index.AddEntry(rel, hash)

		fmt.Printf("added %s (%s)\n", rel, hash)

		return nil
	})

	if err != nil {
		fmt.Printf("error while adding: %v\n", err)
		return
	}
	if  err := repo.SaveIndex(index); err != nil {
		fmt.Printf("error saving the index: %s", err)
	}
}
