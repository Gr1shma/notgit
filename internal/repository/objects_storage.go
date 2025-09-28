package repository

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Gr1shma/notgit/internal/objects"
	"github.com/Gr1shma/notgit/internal/objects/blob"
	"github.com/Gr1shma/notgit/internal/objects/commit"
	"github.com/Gr1shma/notgit/internal/objects/tree"
)

func (r *Repository) StoreObject(obj objects.Object) (string, error) {
	data, err := obj.Serialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize object: %w", err)
	}

	hash := sha1.Sum(data)
	hashStr := hex.EncodeToString(hash[:])

	dir := filepath.Join(r.NotgitDir, "objects", hashStr[:2])
	file := filepath.Join(dir, hashStr[2:])

	if _, err := os.Stat(file); err == nil {
		return hashStr, nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create object dir: %w", err)
	}

	if err := os.WriteFile(file, data, 0o644); err != nil {
		return "", fmt.Errorf("failed to write object file: %w", err)
	}

	return hashStr, nil
}

func (r *Repository) RetrieveBlob(hash string) (*blob.Blob, error) {
	data, err := retrieveObject(r.NotgitDir, hash)
	if err != nil {
		return nil, err
	}

	return blob.DeserializeBlob(data)
}

func (r *Repository) RetrieveTree(hash string) (*tree.Tree, error) {
	data, err := retrieveObject(r.NotgitDir, hash)
	if err != nil {
		return nil, err
	}

	return tree.DeserializeTree(data)
}

func (r *Repository) RetrieveCommit(hash string) (*commit.Commit, error) {
	data, err := retrieveObject(r.NotgitDir, hash)
	if err != nil {
		return nil, err
	}

	return commit.DeserializeCommit(data)
}

func retrieveObject(repoPath, hash string) ([]byte, error) {
	objectPath := filepath.Join(repoPath, "objects", hash[:2], hash[2:])

	data, err := os.ReadFile(objectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read object file: %v", err)
	}

	return data, nil
}
