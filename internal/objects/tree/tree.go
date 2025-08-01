package tree

import (
	"bufio"
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

func (t *Tree) SerializeContent() ([]byte, error) {
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

func (t *Tree) Serialize() ([]byte, error) {
	content, err := t.SerializeContent()
	if err != nil {
		return nil, err
	}

	header := fmt.Sprintf("tree %d\x00", len(content))
	full := append([]byte(header), content...)
	return full, nil
}

func DeserializeTree(data []byte) (*Tree, error) {
	if bytes.Contains(data, []byte{0}) {
		nullIndex := bytes.IndexByte(data, 0)
		if nullIndex == -1 {
			return nil, fmt.Errorf("invalid tree format: missing null byte separator")
		}

		header := string(data[:nullIndex])
		content := data[nullIndex+1:]

		if !bytes.HasPrefix([]byte(header), []byte("tree ")) {
			return nil, fmt.Errorf("not a tree object: %s", header)
		}

		return parseTreeContent(content)
	} else {
		return parseTreeContent(data)
	}
}

func parseTreeContent(data []byte) (*Tree, error) {
	tree := NewTree()
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		parts := bytes.SplitN(line, []byte("\t"), 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tree entry format: %q", line)
		}

		meta := string(parts[0])
		name := string(parts[1])

		var mode, typeStr, hash string
		_, err := fmt.Sscanf(meta, "%s %s %s", &mode, &typeStr, &hash)
		if err != nil {
			return nil, fmt.Errorf("invalid tree metadata: %q", meta)
		}

		if !isValidMode(mode) {
			return nil, fmt.Errorf("invalid entry mode: %s", mode)
		}

		entryType, err := parseEntryType(typeStr)
		if err != nil {
			return nil, err
		}

		tree.AddEntry(name, hash, entryType)
	}

	return tree, nil
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

func isValidMode(mode string) bool {
	return mode == "100644" || mode == "040000"
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

func parseEntryType(typeStr string) (EntryType, error) {
	switch typeStr {
	case "blob":
		return EntryTypeBlob, nil
	case "tree":
		return EntryTypeTree, nil
	default:
		return 0, fmt.Errorf("invalid entry type: %s", typeStr)
	}
}
