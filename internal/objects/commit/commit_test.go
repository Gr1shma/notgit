package commit_test

import (
	"testing"
	"time"

	"github.com/Gr1shma/notgit/internal/objects/commit"
	"github.com/stretchr/testify/require"
)

func TestNewCommit(t *testing.T) {
	c := &commit.Commit{
		TreeHash:     "a1b2c3",
		ParentHashes: []string{"deadbeef"},
		Author: commit.Signature{
			Name:  "Grishma",
			Email: "hey@grishmadhakal.com.np",
			Time:  time.Unix(1625151600, 0),
		},
		Committer: commit.Signature{
			Name:  "Grishma",
			Email: "hey@grishmadhakal.com.np",
			Time:  time.Unix(1625155200, 0),
		},
		Message: "Initial commit",
	}

	err := c.ComputeHash()
	require.NoError(t, err)
	require.NotEmpty(t, c.Hash)
}

func TestCommitSerialize(t *testing.T) {
	c := &commit.Commit{
		TreeHash:     "abc123",
		ParentHashes: []string{"def456"},
		Author: commit.Signature{
			Name:  "Grishma",
			Email: "hey@grishmadhakal.com.np",
			Time:  time.Unix(1600000000, 0),
		},
		Committer: commit.Signature{
			Name:  "Grishma",
			Email: "hey@grishmadhakal.com.np",
			Time:  time.Unix(1600003600, 0),
		},
		Message: "hello from serialize test",
	}

	err := c.ComputeHash()
	require.NoError(t, err)

	serialized, err := c.Serialize()
	require.NoError(t, err)
	require.Contains(t, string(serialized), "tree abc123")
	require.Contains(t, string(serialized), "parent def456")
	require.Contains(t, string(serialized), "author Grishma <hey@grishmadhakal.com.np>")
	require.Contains(t, string(serialized), "committer Grishma <hey@grishmadhakal.com.np>")
	require.Contains(t, string(serialized), "hello from serialize test")
}

func TestCommitDeserialize(t *testing.T) {
	original := &commit.Commit{
		TreeHash:     "xyz999",
		ParentHashes: []string{"123abc"},
		Author: commit.Signature{
			Name:  "Grishma",
			Email: "hey@grishmadhakal.com.np",
			Time:  time.Unix(1610000000, 0),
		},
		Committer: commit.Signature{
			Name:  "Grishma",
			Email: "hey@grishmadhakal.com.np",
			Time:  time.Unix(1610003600, 0),
		},
		Message: "Deserialize test message",
	}

	err := original.ComputeHash()
	require.NoError(t, err)

	serialized, err := original.Serialize()
	require.NoError(t, err)

	deserialized, err := commit.DeserializeCommit(serialized)
	require.NoError(t, err)

	require.Equal(t, original.TreeHash, deserialized.TreeHash)
	require.Equal(t, original.ParentHashes, deserialized.ParentHashes)
	require.Equal(t, original.Author.Name, deserialized.Author.Name)
	require.Equal(t, original.Author.Email, deserialized.Author.Email)
	require.Equal(t, original.Author.Time.Unix(), deserialized.Author.Time.Unix())
	require.Equal(t, original.Committer.Name, deserialized.Committer.Name)
	require.Equal(t, original.Committer.Email, deserialized.Committer.Email)
	require.Equal(t, original.Committer.Time.Unix(), deserialized.Committer.Time.Unix())
	require.Equal(t, original.Message, deserialized.Message)
	require.Equal(t, original.Hash, deserialized.Hash)
}
