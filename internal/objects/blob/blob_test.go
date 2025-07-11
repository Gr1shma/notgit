package blob_test

import (
	"testing"

	"github.com/Gr1shma/notgit/internal/objects/blob"
	"github.com/stretchr/testify/require"
)

func TestNewBlob(t *testing.T) {
	emptyStringBlob, err := blob.NewBlob([]byte(""))
	require.NoError(t, err)
	require.Equal(t, int64(0), emptyStringBlob.Size)
	require.Equal(t, "da39a3ee5e6b4b0d3255bfef95601890afd80709", emptyStringBlob.Hash)

	oneLineBlob, err := blob.NewBlob([]byte("Hello World"))
	require.NoError(t, err)
	require.Equal(t, int64(11), oneLineBlob.Size)
	require.Equal(t, "0a4d55a8d778e5022fab701977c5d840bbc486d0", oneLineBlob.Hash)

	multiLineBlob, err := blob.NewBlob([]byte("line1\nline2\nline3\n"))
	require.NoError(t, err)
	require.Equal(t, int64(18), multiLineBlob.Size)
	require.Equal(t, "8c84f6f36dd2230d3e9c954fa436e5fda90b1957", multiLineBlob.Hash)
}

func TestBlobSerialize(t *testing.T) {
	content := []byte("Hello World")
	createdBlob, err := blob.NewBlob(content)
	require.NoError(t, err)

	serializedContent, err := createdBlob.Serialize()
	require.NoError(t, err)

	// The prefix must match "blob <size>\x00"
	expectedPrefix := []byte("blob 11\x00")
	require.Equal(t, expectedPrefix, serializedContent[:len(expectedPrefix)])

	// The remaining bytes after the prefix should be the original content
	require.Equal(t, content, serializedContent[len(expectedPrefix):])

	// The total serialized length must be prefix length + content length
	require.Equal(t, len(expectedPrefix)+len(content), len(serializedContent))
}

func TestBlobDeserialize(t *testing.T) {
	content := []byte("Hello World")
	originalBlob, err := blob.NewBlob(content)
	require.NoError(t, err)

	serializedContent, err := originalBlob.Serialize()
	require.NoError(t, err)

	// Deserialize the blob from the serialized content
	deserializedBlob, err := blob.DeserializeBlob(serializedContent)
	require.NoError(t, err)

	// Verify that the deserialized blob matches the original
	require.Equal(t, content, deserializedBlob.Content)
	require.Equal(t, originalBlob.Content, deserializedBlob.Content)
	require.Equal(t, originalBlob.Size, deserializedBlob.Size)
	require.Equal(t, originalBlob.Hash, deserializedBlob.Hash)
}
