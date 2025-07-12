package tree_test

import (
	"testing"

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
