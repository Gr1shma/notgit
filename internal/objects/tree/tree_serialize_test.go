package tree_test

import (
	"fmt"
	"testing"

	"github.com/Gr1shma/notgit/internal/objects/blob"
	"github.com/Gr1shma/notgit/internal/objects/tree"
	"github.com/stretchr/testify/require"
)

func TestTreeSerialize(t *testing.T) {
	// Create a new empty tree
	newTree := tree.NewTree()
	require.Equal(t, 0, newTree.Size(), "New tree should be empty")

	// Create a blob to add to the tree
	b, err := blob.NewBlob([]byte("Hello World"))
	require.NoError(t, err, "Creating blob should not fail")
	require.NotNil(t, b, "Blob should not be nil")

	// Add the blob entry to the tree
	newTree.AddEntry("hello.txt", b.Hash, tree.EntryTypeBlob)

	// Serialize the tree
	serialized, err := newTree.Serialize()
	require.NoError(t, err, "Serialization should not fail")
	require.NotEmpty(t, serialized, "Serialized output should not be empty")

	// Check if it contains the expected format
	// Format: "<mode> <type> <hash>\t<name>\n"
	expectedFragment := fmt.Sprintf("100644 blob %s\thello.txt\n", b.Hash)
	require.Contains(t, string(serialized), expectedFragment, "Serialized tree should contain expected entry line")
}

func TestTreeSerialize_MultipleEntries(t *testing.T) {
	// Create a blob to add to trees
	b, err := blob.NewBlob([]byte("Hello World"))
	require.NoError(t, err)
	require.NotNil(t, b)

	// Create a subtree and add a blob entry
	subTree := tree.NewTree()
	subTree.AddEntry("subfile.txt", b.Hash, tree.EntryTypeBlob)
	require.NoError(t, subTree.ComputeHash())

	// Create the main tree with a blob and a subtree
	mainTree := tree.NewTree()
	mainTree.AddEntry("file.txt", b.Hash, tree.EntryTypeBlob)
	mainTree.AddEntry("subdir", subTree.GetHash(), tree.EntryTypeTree)

	// Serialize the main tree
	serialized, err := mainTree.Serialize()
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	// Check that serialized output contains both entries with expected format
	expectedBlob := fmt.Sprintf("100644 blob %s\tfile.txt\n", b.Hash)
	expectedTree := fmt.Sprintf("040000 tree %s\tsubdir\n", subTree.GetHash())

	serializedStr := string(serialized)
	require.Contains(t, serializedStr, expectedBlob, "Serialized tree should contain blob entry line")
	require.Contains(t, serializedStr, expectedTree, "Serialized tree should contain subtree entry line")
}
