package tree_test

import (
	"fmt"
	"testing"

	"github.com/Gr1shma/notgit/internal/objects/blob"
	"github.com/Gr1shma/notgit/internal/objects/tree"
	"github.com/stretchr/testify/require"
)

func TestNewTree(t *testing.T) {
	newTree := tree.NewTree()

	// Test that tree is not nil
	require.NotNil(t, newTree)

	// Test that tree starts empty
	require.Equal(t, 0, newTree.Size())

	// Test that type is correct
	require.Equal(t, "tree", newTree.Type())

	// Test that hash is initially empty (will be calculated when needed)
	require.Empty(t, newTree.GetHash())
}

func TestAddEntry(t *testing.T) {
	// Create a new empty tree
	newTree := tree.NewTree()
	require.Equal(t, 0, newTree.Size(), "New tree should be empty")

	// Create a blob to add to the tree
	b, err := blob.NewBlob([]byte("Hello World"))
	require.NoError(t, err, "Creating blob should not fail")
	require.NotNil(t, b, "Blob should not be nil")

	// Add the blob entry to the tree
	newTree.AddEntry("test.txt", b.Hash, tree.EntryTypeBlob)

	// Verify tree size increased
	require.Equal(t, 1, newTree.Size(), "Tree should have 1 entry after adding")

	// Verify entry was added correctly using GetEntry method (better than direct access)
	entry := newTree.GetEntry("test.txt")
	require.NotNil(t, entry, "Entry should be found")
	require.Equal(t, "test.txt", entry.Name, "Entry name should match")
	require.Equal(t, b.Hash, entry.Hash, "Entry hash should match blob hash")
	require.Equal(t, tree.EntryTypeBlob, entry.Type, "Entry type should be blob")

	// Compute hash explicitly
	require.NoError(t, newTree.ComputeHash())

	treeHash := newTree.GetHash()
	require.NotEmpty(t, treeHash, "Tree hash should be calculated")
	require.Len(t, treeHash, 40, "Tree hash should be 40 characters (SHA-1 hex)")
}

func TestAddEntry_ReplaceExisting(t *testing.T) {
	// Test replacing an existing entry
	newTree := tree.NewTree()

	// Add first entry
	b1, err := blob.NewBlob([]byte("Hello World"))
	require.NoError(t, err)
	newTree.AddEntry("test.txt", b1.Hash, tree.EntryTypeBlob)
	require.Equal(t, 1, newTree.Size())

	// Replace with different content
	b2, err := blob.NewBlob([]byte("Goodbye World"))
	require.NoError(t, err)
	newTree.AddEntry("test.txt", b2.Hash, tree.EntryTypeBlob) // Same name, different hash

	// Verify size didn't change (replaced, not added)
	require.Equal(t, 1, newTree.Size(), "Tree should still have 1 entry after replacement")

	// Verify entry was replaced
	entry := newTree.GetEntry("test.txt")
	require.NotNil(t, entry)
	require.Equal(t, b2.Hash, entry.Hash, "Entry should have new hash")

	// Recompute hash after change
	require.NoError(t, newTree.ComputeHash())
	require.Len(t, newTree.GetHash(), 40)
}

func TestAddEntry_MultipleEntries(t *testing.T) {
	// Test adding multiple entries
	newTree := tree.NewTree()

	// Add blob entry
	b, err := blob.NewBlob([]byte("Hello World"))
	require.NoError(t, err)
	newTree.AddEntry("file.txt", b.Hash, tree.EntryTypeBlob)

	// Add tree entry (subdirectory)
	subTree := tree.NewTree()
	subTree.AddEntry("subfile.txt", b.Hash, tree.EntryTypeBlob)

	// Compute hash for the sub-tree before adding
	require.NoError(t, subTree.ComputeHash())
	newTree.AddEntry("subdir", subTree.GetHash(), tree.EntryTypeTree)

	// Verify both entries exist
	require.Equal(t, 2, newTree.Size())

	// Verify blob entry
	blobEntry := newTree.GetEntry("file.txt")
	require.NotNil(t, blobEntry)
	require.Equal(t, tree.EntryTypeBlob, blobEntry.Type)

	// Verify tree entry
	treeEntry := newTree.GetEntry("subdir")
	require.NotNil(t, treeEntry)
	require.Equal(t, tree.EntryTypeTree, treeEntry.Type)

	// Compute hash for parent tree
	require.NoError(t, newTree.ComputeHash())
	require.Len(t, newTree.GetHash(), 40)
}

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
