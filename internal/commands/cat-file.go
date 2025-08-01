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
	Run:  catFileCallback,
}

func init() {
	catFileCmd.PersistentFlags().BoolVarP(&catFilePrettyBool, "pretty", "p", false, "Pretty print objects")
	catFileCmd.PersistentFlags().BoolVarP(&catFileTypeBool, "type", "t", false, "Show object type")
	catFileCmd.PersistentFlags().BoolVarP(&catFileSizeBool, "size", "s", false, "Show object size")
	rootCmd.AddCommand(catFileCmd)
}

func catFileCallback(cmd *cobra.Command, args []string) {
	objectHash := args[0]
	if len(objectHash) < 3 {
		fmt.Fprintf(cmd.ErrOrStderr(), "Invalid object hash: %s\n", objectHash)
		return
	}

	objDir := filepath.Join(".notgit", "objects", objectHash[:2])
	objFile := filepath.Join(objDir, objectHash[2:])

	data, err := os.ReadFile(objFile)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Failed to read object file: %v\n", err)
		return
	}

	objectType, content, err := parseObjectHeader(data)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Failed to parse object: %v\n", err)
		return
	}

	if catFileTypeBool {
		fmt.Println(objectType)
		return
	}

	if catFileSizeBool {
		fmt.Println(len(content))
		return
	}

	if catFilePrettyBool {
		err := prettyPrintObject(objectType, data)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Failed to pretty print object: %v\n", err)
		}
		return
	}

	fmt.Printf("Object type: %s\n", objectType)
	fmt.Printf("Object size: %d bytes\n", len(content))
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
	return objectType, content, nil
}

func prettyPrintObject(objectType string, data []byte) error {
	switch objectType {
	case "blob":
		return prettyPrintBlob(data)
	case "tree":
		return prettyPrintTree(data)
	case "commit":
		return prettyPrintCommit(data)
	default:
		return fmt.Errorf("unknown object type: %s", objectType)
	}
}

func prettyPrintBlob(data []byte) error {
	blobObj, err := blob.DeserializeBlob(data)
	if err != nil {
		return err
	}
	fmt.Print(string(blobObj.Content))
	return nil
}

func prettyPrintTree(data []byte) error {
	treeObj, err := tree.DeserializeTree(data)
	if err != nil {
		return err
	}

	for _, entry := range treeObj.Entries {
		mode := getModeForType(entry.Type)
		typeStr := getTypeString(entry.Type)
		fmt.Printf("%s %s %s\t%s\n", mode, typeStr, entry.Hash, entry.Name)
	}
	return nil
}

func prettyPrintCommit(data []byte) error {
	commitObj, err := commit.DeserializeCommit(data)
	if err != nil {
		return err
	}

	fmt.Printf("tree %s\n", commitObj.TreeHash)

	for _, parent := range commitObj.ParentHashes {
		fmt.Printf("parent %s\n", parent)
	}

	fmt.Printf("author %s <%s> %d %s\n",
		commitObj.Author.Name,
		commitObj.Author.Email,
		commitObj.Author.Time.Unix(),
		commitObj.Author.Time.Format("-0700"),
	)

	fmt.Printf("committer %s <%s> %d %s\n",
		commitObj.Committer.Name,
		commitObj.Committer.Email,
		commitObj.Committer.Time.Unix(),
		commitObj.Committer.Time.Format("-0700"),
	)

	fmt.Printf("\n%s\n", commitObj.Message)
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
