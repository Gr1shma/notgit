package tree_test

import (
	"testing"

	"github.com/Gr1shma/notgit/internal/objects/blob"
	"github.com/Gr1shma/notgit/internal/objects/tree"
	"github.com/stretchr/testify/require"
)

func TestTreeDeserialize(t *testing.T) {
	// Create a new empty tree and add an entry
	origTree := tree.NewTree()
	b, err := blob.NewBlob([]byte("Hello World"))
	require.NoError(t, err)
	origTree.AddEntry("hello.txt", b.Hash, tree.EntryTypeBlob)

	// Serialize the tree
	serialized, err := origTree.Serialize()
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	// Deserialize the tree from the serialized bytes
	deserializedTree, err := tree.DeserializeTree(serialized)
	require.NoError(t, err)
	require.NotNil(t, deserializedTree)

	// Check that the deserialized tree has exactly one entry
	require.Equal(t, 1, deserializedTree.Size())

	// Verify the entry matches the original
	entry := deserializedTree.GetEntry("hello.txt")
	require.NotNil(t, entry)
	require.Equal(t, "hello.txt", entry.Name)
	require.Equal(t, b.Hash, entry.Hash)
	require.Equal(t, tree.EntryTypeBlob, entry.Type)
}

func TestTreeDeserialize_MultipleEntries(t *testing.T) {
	// Create a blob to add to trees
	b, err := blob.NewBlob([]byte("Hello World"))
	require.NoError(t, err)

	// Create a subtree and add an entry
	subTree := tree.NewTree()
	subTree.AddEntry("subfile.txt", b.Hash, tree.EntryTypeBlob)
	require.NoError(t, subTree.ComputeHash())

	// Create the main tree with two entries:
	// - a blob file
	// - a subtree directory
	mainTree := tree.NewTree()
	mainTree.AddEntry("file.txt", b.Hash, tree.EntryTypeBlob)
	mainTree.AddEntry("subdir", subTree.GetHash(), tree.EntryTypeTree)

	// Serialize the main tree
	serialized, err := mainTree.Serialize()
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	// Deserialize the tree
	deserializedTree, err := tree.DeserializeTree(serialized)
	require.NoError(t, err)
	require.NotNil(t, deserializedTree)

	// Check the number of entries in the deserialized tree
	require.Equal(t, 2, deserializedTree.Size())

	// Validate blob entry
	blobEntry := deserializedTree.GetEntry("file.txt")
	require.NotNil(t, blobEntry)
	require.Equal(t, tree.EntryTypeBlob, blobEntry.Type)
	require.Equal(t, b.Hash, blobEntry.Hash)

	// Validate subtree entry
	treeEntry := deserializedTree.GetEntry("subdir")
	require.NotNil(t, treeEntry)
	require.Equal(t, tree.EntryTypeTree, treeEntry.Type)
	require.Equal(t, subTree.GetHash(), treeEntry.Hash)
}
