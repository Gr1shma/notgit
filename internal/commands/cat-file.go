package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Gr1shma/notgit/internal/objects/blob"
	"github.com/spf13/cobra"
)

var catFilePrettyBool bool

var catFileCmd = &cobra.Command{
	Use:   "cat-file <object>",
	Short: "Display the contents of a Git object",
	Long: `cat-file provides content or type and size information for repository objects.

With the -p flag, the content of the object is pretty-printed.`,
	Args: cobra.ExactArgs(1),
	Run:  catFileCallback,
}

func init() {
	catFileCmd.PersistentFlags().BoolVarP(&catFilePrettyBool, "preety", "p", false, "Preety print objects")
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

	blobObj, err := blob.DeserializeBlob(data)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Failed to deserialize blob: %v\n", err)
		return
	}

	if catFilePrettyBool {
		fmt.Println(string(blobObj.Content))
	} else {
		fmt.Printf("Object size: %d bytes\n", blobObj.Size)
	}
}

