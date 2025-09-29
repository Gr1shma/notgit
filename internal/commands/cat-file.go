package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gr1shma/notgit/internal/objects/blob"
	"github.com/Gr1shma/notgit/internal/objects/commit"
	"github.com/Gr1shma/notgit/internal/objects/tree"
	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/spf13/cobra"
)

var catFilePrettyBool bool
var catFileTypeBool bool
var catFileSizeBool bool

var catFileCmd = &cobra.Command{
	Use:   "cat-file <type-and-size|-t|-s|-p> <object>",
	Short: "Display the contents of a Git object",
	Long: `cat-file provides content or type and size information for repository objects.
With the -p flag, the content of the object is pretty-printed.
With the -t flag, show the object type.
With the -s flag, show the object size.`,
	Args: cobra.ExactArgs(1),
	RunE: catFileCallback,
}

func init() {
	catFileCmd.PersistentFlags().BoolVarP(&catFilePrettyBool, "pretty", "p", false, "Pretty print objects")
	catFileCmd.PersistentFlags().BoolVarP(&catFileTypeBool, "type", "t", false, "Show object type")
	catFileCmd.PersistentFlags().BoolVarP(&catFileSizeBool, "size", "s", false, "Show object size")
	rootCmd.AddCommand(catFileCmd)
}

func catFileCallback(cmd *cobra.Command, args []string) error {
	objectHash := strings.TrimSpace(args[0])
	
	if len(objectHash) < 4 {
		return fmt.Errorf("invalid object hash: %s (too short)", objectHash)
	}
	if len(objectHash) != 40 {
		return fmt.Errorf("invalid object hash: %s (must be 40 characters)", objectHash)
	}

	flagCount := 0
	if catFilePrettyBool {
		flagCount++
	}
	if catFileTypeBool {
		flagCount++
	}
	if catFileSizeBool {
		flagCount++
	}
	
	if flagCount > 1 {
		return fmt.Errorf("only one of -p, -t, or -s can be specified")
	}
	
	repo, err := repository.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	objDir := filepath.Join(repo.NotgitDir, "objects", objectHash[:2])
	objFile := filepath.Join(objDir, objectHash[2:])

	data, err := os.ReadFile(objFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("object %s not found", objectHash)
		}
		return fmt.Errorf("failed to read object file: %w", err)
	}

	objectType, content, err := parseObjectHeader(data)
	if err != nil {
		return fmt.Errorf("failed to parse object: %w", err)
	}

	if catFileTypeBool {
		fmt.Fprintln(cmd.OutOrStdout(), objectType)
		return nil
	}

	if catFileSizeBool {
		fmt.Fprintln(cmd.OutOrStdout(), len(content))
		return nil
	}

	if catFilePrettyBool {
		if err := prettyPrintObject(cmd, objectType, data); err != nil {
			return fmt.Errorf("failed to pretty print object: %w", err)
		}
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Object type: %s\n", objectType)
	fmt.Fprintf(cmd.OutOrStdout(), "Object size: %d bytes\n", len(content))
	return nil
}

func parseObjectHeader(data []byte) (string, []byte, error) {
	nullIndex := bytes.IndexByte(data, 0)
	if nullIndex == -1 {
		return "", nil, fmt.Errorf("invalid object format: missing null byte separator")
	}

	header := string(data[:nullIndex])
	content := data[nullIndex+1:]

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid object header: %s", header)
	}

	objectType := parts[0]
	
	switch objectType {
	case "blob", "tree", "commit":
	default:
		return "", nil, fmt.Errorf("unknown object type: %s", objectType)
	}
	
	return objectType, content, nil
}

func prettyPrintObject(cmd *cobra.Command, objectType string, data []byte) error {
	switch objectType {
	case "blob":
		return prettyPrintBlob(cmd, data)
	case "tree":
		return prettyPrintTree(cmd, data)
	case "commit":
		return prettyPrintCommit(cmd, data)
	default:
		return fmt.Errorf("unknown object type: %s", objectType)
	}
}

func prettyPrintBlob(cmd *cobra.Command, data []byte) error {
	blobObj, err := blob.DeserializeBlob(data)
	if err != nil {
		return fmt.Errorf("failed to deserialize blob: %w", err)
	}
	
	_, err = cmd.OutOrStdout().Write(blobObj.Content)
	if err != nil {
		return fmt.Errorf("failed to write blob content: %w", err)
	}
	return nil
}

func prettyPrintTree(cmd *cobra.Command, data []byte) error {
	treeObj, err := tree.DeserializeTree(data)
	if err != nil {
		return fmt.Errorf("failed to deserialize tree: %w", err)
	}

	if len(treeObj.Entries) == 0 {
		return nil // Empty tree, nothing to print
	}

	for _, entry := range treeObj.Entries {
		mode := getModeForType(entry.Type)
		typeStr := getTypeString(entry.Type)
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s\t%s\n", mode, typeStr, entry.Hash, entry.Name)
	}
	return nil
}

func prettyPrintCommit(cmd *cobra.Command, data []byte) error {
	commitObj, err := commit.DeserializeCommit(data)
	if err != nil {
		return fmt.Errorf("failed to deserialize commit: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "tree %s\n", commitObj.TreeHash)

	for _, parent := range commitObj.ParentHashes {
		fmt.Fprintf(cmd.OutOrStdout(), "parent %s\n", parent)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "author %s <%s> %d %s\n",
		commitObj.Author.Name,
		commitObj.Author.Email,
		commitObj.Author.Time.Unix(),
		commitObj.Author.Time.Format("-0700"),
	)

	fmt.Fprintf(cmd.OutOrStdout(), "committer %s <%s> %d %s\n",
		commitObj.Committer.Name,
		commitObj.Committer.Email,
		commitObj.Committer.Time.Unix(),
		commitObj.Committer.Time.Format("-0700"),
	)

	fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", commitObj.Message)
	return nil
}

func getModeForType(entryType tree.EntryType) string {
	switch entryType {
	case tree.EntryTypeBlob:
		return "100644"
	case tree.EntryTypeTree:
		return "040000"
	default:
		return "100644"
	}
}

func getTypeString(entryType tree.EntryType) string {
	switch entryType {
	case tree.EntryTypeBlob:
		return "blob"
	case tree.EntryTypeTree:
		return "tree"
	default:
		return "blob"
	}
}
