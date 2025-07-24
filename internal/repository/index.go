package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type IndexEntry struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

type Index struct {
	Entries map[string]IndexEntry `json:"entries"`
}

func NewIndex() *Index {
	return &Index{Entries: make(map[string]IndexEntry)}
}

func (r *Repository) IndexPath() string {
	return filepath.Join(r.NotgitDir, "index")
}

func (r *Repository) LoadIndex() (*Index, error) {
	indexPath := r.IndexPath()
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.ErrNotExist == err {
			return nil, fmt.Errorf("error index file doesn't exist: %w", err)
		}
		return nil, fmt.Errorf("failed to read index: %w", err)
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("failed to parse index file: %w", err)
	}
	return &idx, nil
}

func (r *Repository) SaveIndex(idx *Index) error {
	path := r.IndexPath()
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marsal the index: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func (idx *Index) AddEntry(path, hash string) {
	idx.Entries[path] = IndexEntry{Path: path, Hash: hash}
}
