package utils

import (
	"fmt"
	"os"
	"unicode/utf8"
)

// 0o644 -> (-rw-r--r--)
const defaultFilePerm os.FileMode = 0o644

// ReadFile reads the contents of the file at the given path.
//
// it checks for:
//   - file existence
//   - that the path is not a directory
//   - that the contents are valid UTF-8 (optional for most cases)
//
// returns an error with detailed context if any of those checks fail.
func ReadFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s does not exist", path)
		}
		return nil, fmt.Errorf("cannot access file %s: %w", path, err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("expected file but found directory at %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	if !utf8.Valid(data) {
		return nil, fmt.Errorf("file %s is not valid UTF-8", path)
	}

	return data, nil
}

// WriteFile writes data to the file at path using the given permission.
// if permPtr is nil, default permission (0o644) is used.
// it errors if the path is a directory or writing fails.
func WriteFile(path string, data []byte, permPtr *os.FileMode) error {
	perm := defaultFilePerm
	if permPtr != nil {
		perm = *permPtr
	}

	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("cannot write to path %s: is a directory", path)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("cannot access path %s: %w", path, err)
	}

	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", path, err)
	}

	return nil
}
