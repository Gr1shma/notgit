package repository_test

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Gr1shma/notgit/internal/objects/blob"
	"github.com/Gr1shma/notgit/internal/objects/commit"
	"github.com/Gr1shma/notgit/internal/objects/tree"
	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestStoreBlobObject(t *testing.T) {
	tmpDir := t.TempDir()

	err := repository.CreateRepo(tmpDir)
	require.NoError(t, err)

	repo, err := repository.OpenRepository(tmpDir)
	require.NoError(t, err)

	content := []byte("hello world")
	bl, err := blob.NewBlob(content)
	require.NoError(t, err)

	serialized, err := bl.Serialize()
	require.NoError(t, err)

	expectedHash := sha1.Sum(serialized)
	expectedHashStr := hex.EncodeToString(expectedHash[:])

	hashStr, err := repo.StoreObject(bl)
	require.NoError(t, err)
	require.Equal(t, expectedHashStr, hashStr)

	objPath := filepath.Join(repo.NotgitDir, "objects", hashStr[:2], hashStr[2:])
	_, err = os.Stat(objPath)
	require.NoError(t, err, "object file should exist")

	storedData, err := os.ReadFile(objPath)
	require.NoError(t, err)
	require.Equal(t, serialized, storedData, "stored object data must match serialized")

	retrievedBlob, err := repo.RetrieveBlob(hashStr)
	require.NoError(t, err)
	require.Equal(t, bl.Content, retrievedBlob.Content, "retrieved blob content should match original")
}

func TestStoreTreeObject(t *testing.T) {
	tempDir := t.TempDir()

	repoPath := filepath.Join(tempDir, "repo")
	err := repository.CreateRepo(repoPath)
	require.NoError(t, err)

	repo, err := repository.OpenRepository(repoPath)
	require.NoError(t, err)

	entries := []tree.Entry{
		{
			Name: "file.txt",
			Hash: "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
			Type: tree.EntryTypeBlob,
		},
		{
			Name: "subdir",
			Hash: "4b825dc642cb6eb9a060e54bf8d69288fbee4904",
			Type: tree.EntryTypeTree,
		},
	}

	tr := &tree.Tree{
		Entries: entries,
	}

	err = tr.ComputeHash()
	require.NoError(t, err)

	hash, err := repo.StoreObject(tr)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	objPath := filepath.Join(repo.NotgitDir, "objects", hash[:2], hash[2:])
	_, err = os.Stat(objPath)
	require.NoError(t, err)

	retrievedTree, err := repo.RetrieveTree(hash)
	require.NoError(t, err)
	require.Len(t, retrievedTree.Entries, len(entries), "retrieved tree should have same number of entries")
	
	for i, entry := range entries {
		require.Equal(t, entry.Name, retrievedTree.Entries[i].Name, "entry name should match")
		require.Equal(t, entry.Hash, retrievedTree.Entries[i].Hash, "entry hash should match")
		require.Equal(t, entry.Type, retrievedTree.Entries[i].Type, "entry type should match")
	}
}

func TestStoreCommitObject(t *testing.T) {
	tempDir := t.TempDir()

	repoPath := filepath.Join(tempDir, "repo")
	err := repository.CreateRepo(repoPath)
	require.NoError(t, err)

	repo, err := repository.OpenRepository(repoPath)
	require.NoError(t, err)

	author := commit.Signature{
		Name:  "John Doe",
		Email: "john@example.com",
		Time:  time.Now(),
	}

	committer := author

	c := &commit.Commit{
		TreeHash:  "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
		ParentHashes: nil,
		Author:    author,
		Committer: committer,
		Message:   "Initial commit",
	}

	err = c.ComputeHash()
	require.NoError(t, err)

	hash, err := repo.StoreObject(c)
	require.NoError(t, err)
	require.Equal(t, c.Hash, hash)

	objPath := filepath.Join(repo.NotgitDir, "objects", hash[:2], hash[2:])
	_, err = os.Stat(objPath)
	require.NoError(t, err)

	retrievedCommit, err := repo.RetrieveCommit(hash)
	require.NoError(t, err)
	require.Equal(t, c.TreeHash, retrievedCommit.TreeHash, "tree hash should match")
	require.Equal(t, c.Message, retrievedCommit.Message, "commit message should match")
	require.Equal(t, c.Author.Name, retrievedCommit.Author.Name, "author name should match")
	require.Equal(t, c.Author.Email, retrievedCommit.Author.Email, "author email should match")
	require.Equal(t, c.Committer.Name, retrievedCommit.Committer.Name, "committer name should match")
	require.Equal(t, c.Committer.Email, retrievedCommit.Committer.Email, "committer email should match")
	require.Equal(t, len(c.ParentHashes), len(retrievedCommit.ParentHashes), "parent hashes length should match")
}
