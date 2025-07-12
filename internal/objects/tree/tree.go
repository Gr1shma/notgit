package tree

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/Gr1shma/notgit/internal/objects"
)

type EntryType int

const (
	EntryTypeBlob EntryType = iota
	EntryTypeTree
)

type Entry struct {
	Name string
	Hash string
	Type EntryType
}

type Tree struct {
	Entries  []Entry
	entryMap map[string]*Entry
	Hash     string
}

var _ objects.Object = (*Tree)(nil)

func NewTree() *Tree {
	return &Tree{
		Entries:  make([]Entry, 0),
		entryMap: make(map[string]*Entry),
	}
}

func (t *Tree) Type() string {
	return "tree"
}

func (t *Tree) AddEntry(name, hash string, entryType EntryType) {
	if existingEntry, exists := t.entryMap[name]; exists {
		existingEntry.Hash = hash
		existingEntry.Type = entryType
		return
	}

	entry := Entry{
		Name: name,
		Hash: hash,
		Type: entryType,
	}
	t.Entries = append(t.Entries, entry)
	t.entryMap[name] = &t.Entries[len(t.Entries)-1] // Point to the actual entry in slice
}

func (t *Tree) GetEntry(name string) *Entry {
	if entry, exists := t.entryMap[name]; exists {
		return entry
	}
	return nil
}

func (t *Tree) Size() int {
	return len(t.Entries)
}

func (t *Tree) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	for _, entry := range t.Entries {
		mode := t.getModeForType(entry.Type)
		typeStr := t.getTypeString(entry.Type)

		line := fmt.Sprintf("%s %s %s\t%s\n",
			mode,
			typeStr,
			entry.Hash,
			entry.Name,
		)
		buffer.WriteString(line)
	}
	return buffer.Bytes(), nil
}

func DeserializeTree(data []byte) (*Tree, error) {
	return nil, nil
}

func (t *Tree) ComputeHash() error {
	serialized, err := t.Serialize()
	if err != nil {
		return err
	}
	hash := sha1.Sum(serialized)
	t.Hash = hex.EncodeToString(hash[:])
	return nil
}

func (t *Tree) GetHash() string {
	return t.Hash
}

func (t *Tree) getModeForType(entryType EntryType) string {
	switch entryType {
	case EntryTypeBlob:
		return "100644"
	case EntryTypeTree:
		return "040000"
	default:
		return "100644"
	}
}

func (t *Tree) getTypeString(entryType EntryType) string {
	switch entryType {
	case EntryTypeBlob:
		return "blob"
	case EntryTypeTree:
		return "tree"
	default:
		return "blob"
	}
}
