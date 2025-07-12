package tree_test

import (
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
