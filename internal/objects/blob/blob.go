package blob

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/Gr1shma/notgit/internal/objects"
)

type Blob struct {
	Content []byte
	Size    int64
	Hash    string
}

var _ objects.Object = (*Blob)(nil)

func NewBlob(content []byte) (*Blob, error) {
	hash := sha1.Sum(content)
	return &Blob{
		Content: content,
		Size:    int64(len(content)),
		Hash:    hex.EncodeToString(hash[:]),
	}, nil
}

func (b *Blob) Type() string {
	return "blob"
}

func (b *Blob) GetHash() string {
	return b.Hash
}

func (b *Blob) Serialize() ([]byte, error) {
	// "\x00" is the null byte separator between header and content
	header := fmt.Sprintf("blob %d\x00", b.Size)
	return append([]byte(header), b.Content...), nil
}

func DeserializeBlob(data []byte) (*Blob, error) {
	nullIndex := bytes.IndexByte(data, 0)

	if nullIndex == -1 {
		return nil, fmt.Errorf("invalid blob format: missing null byte separator")
	}

	header := string(data[:nullIndex])
	var size int64
	_, err := fmt.Sscanf(header, "blob %d", &size)
	if err != nil {
		return nil, fmt.Errorf("invalid blob header: %v", err)
	}

	content := data[nullIndex+1:]
	if int64(len(content)) != size {
		return nil, fmt.Errorf("content size mismatch: expected %d, got %d", size, len(content))
	}

	return NewBlob(content)
}
