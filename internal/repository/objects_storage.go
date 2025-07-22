package repository

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Gr1shma/notgit/internal/objects"
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
